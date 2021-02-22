package dockerexec

import (
	"context"
	"os"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/tkw1536/proxyssh"
	"github.com/tkw1536/proxyssh/internal/asyncio"
	"github.com/tkw1536/proxyssh/internal/term"
	"github.com/tkw1536/proxyssh/logging"
)

// Code is this file is roughly adapted from https://github.com/docker/cli/blob/master/cli/command/container/exec.go
// and also https://github.com/docker/cli/blob/master/cli/command/container/hijack.go.
//
// These are licensed under the Apache 2.0 License.
// This license requires to state changes made to the code and inclusion of the original NOTICE file.
//
// The code was modified to be independent of the docker cli utility classes where applicable.
//
// The original license and NOTICE can be found below:
//
// Docker
// Copyright 2012-2017 Docker, Inc.
//
// This product includes software developed at Docker, Inc. (https://www.docker.com).
//
// This product contains software (https://github.com/creack/pty) developed
// by Keith Rarick, licensed under the MIT License.
//
// The following is courtesy of our legal counsel:
//
// Use and transfer of Docker may be subject to certain restrictions by the
// United States and other governments.
// It is your responsibility to ensure that your use and/or transfer does not
// violate applicable laws.
//
// For more information, please see https://www.bis.doc.gov
//
// See also https://www.apache.org/dev/crypto.html and/or seek legal counsel.

// NewContainerExecProcess creates a process that executes within a docker container.
//
// The command will not prefix the entrypoint.
func NewContainerExecProcess(client client.APIClient, containerID string, command []string) *ContainerExecProcess {
	config := types.ExecConfig{
		AttachStdin:  true,
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          command,
	}

	return &ContainerExecProcess{
		client: client,

		containerID: containerID,
		config:      config,
	}
}

// ContainerExecProcess represents a process running inside a docker engine
type ContainerExecProcess struct {

	// environment
	client client.APIClient
	ctx    context.Context

	// parameters
	containerID string
	config      types.ExecConfig

	// internal streams
	term.Pipes
	terminal *term.Pair // used in tty mode

	// state
	execID string
	conn   *types.HijackedResponse

	// for result handling
	outputErrChan chan error
	inputDoneChan chan struct{}
	restoreTerms  sync.Once

	// for cleanup
	exited bool
}

// String turns EngineProcess into a string
func (cep *ContainerExecProcess) String() string {
	if cep == nil {
		return ""
	}

	return cep.containerID + " " + strings.Join(cep.config.Cmd, " ")
}

// Init initializes this EngineProcess
func (cep *ContainerExecProcess) Init(ctx context.Context, detector logging.MemoryLeakDetector, isTerm bool) error {
	cep.ctx = ctx
	if isTerm {
		cep.config.Tty = true
		return cep.initTerm()
	}

	return cep.initPlain()
}

func (cep *ContainerExecProcess) initPlain() error {
	return nil
}

func (cep *ContainerExecProcess) initTerm() error {
	cep.terminal = &term.Pair{}
	if err := cep.terminal.Open(true); err != nil {
		return err
	}

	return nil
}

// Start starts this process
func (cep *ContainerExecProcess) Start(detector logging.MemoryLeakDetector, Term string, resizeChan <-chan proxyssh.WindowSize, isPty bool) (*os.File, error) {
	if isPty {
		cep.config.Env = append(cep.config.Env, "TERM="+Term)

		cep.terminal.HandleWith(resizeChan, func(size proxyssh.WindowSize) {
			cep.client.ContainerExecResize(cep.ctx, cep.execID, types.ResizeOptions{
				Height: uint(size.Height),
				Width:  uint(size.Width),
			})
		})
	}

	// start streaming
	if err := cep.execAndStream(detector, isPty); err != nil {
		return nil, err
	}

	// and return
	return cep.terminal.External(), nil
}

func (cep *ContainerExecProcess) execAndStream(detector logging.MemoryLeakDetector, isPty bool) error {

	// create the exec
	res, err := cep.client.ContainerExecCreate(cep.ctx, cep.containerID, cep.config)
	if err != nil {
		return err
	}
	cep.execID = res.ID

	// attach to it
	conn, err := cep.client.ContainerExecAttach(cep.ctx, cep.execID, types.ExecStartCheck{
		Detach: false,
		Tty:    isPty,
	})
	cep.conn = &conn

	// setup channels
	cep.outputErrChan = make(chan error, 1)
	cep.inputDoneChan = make(chan struct{})

	// read output
	detector.Add("dockerexec: output")
	go func() {
		defer detector.Done("dockerexec: output")

		if isPty {
			_, err = asyncio.CopyLeak(cep.ctx, cep.terminal.Internal(), conn.Reader)
			cep.terminal.RestoreMode()
		} else {
			_, err = asyncio.StdCopyLeak(cep.ctx, cep.StdoutPipe, cep.StderrPipe, conn.Reader)
		}

		// close output and send error (if any)
		cep.outputErrChan <- err
	}()

	// write input
	detector.Add("dockerexec: input")
	go func() {
		defer detector.Done("dockerexec: input")

		if isPty {
			asyncio.CopyLeak(cep.ctx, conn.Conn, cep.terminal.Internal())
		} else {
			asyncio.CopyLeak(cep.ctx, conn.Conn, cep.StdinPipe)
		}
		conn.CloseWrite()
		close(cep.inputDoneChan)
	}()

	return nil
}

// Wait waits for the process and returns the exit code
func (cep *ContainerExecProcess) Wait(detector logging.MemoryLeakDetector) (code int, err error) {

	// wait for streams to close
	detector.Add("dockerexec: wait")
	err = cep.waitStreams()
	detector.Done("dockerexec: wait")
	if err != nil {
		return 0, err
	}

	// inspect and get the actual exit code
	resp, err := cep.client.ContainerExecInspect(cep.ctx, cep.execID)
	if err == nil {
		cep.exited = true
	}
	return resp.ExitCode, err
}

// waitStreams waits for the streams to finish
func (cep *ContainerExecProcess) waitStreams() error {
	select {
	case err := <-cep.outputErrChan:
		return err
	case <-cep.inputDoneChan: // wait for output also
		select {
		case err := <-cep.outputErrChan:
			return err
		case <-cep.ctx.Done():
			return cep.ctx.Err()
		}
	case <-cep.ctx.Done():
		return cep.ctx.Err()
	}
}

// Cleanup cleans up this process, typically to kill it.
func (cep *ContainerExecProcess) Cleanup() (killed bool) {
	cep.terminal.UnhangHack()
	cep.terminal.Close()
	cep.ClosePipes()

	if cep.conn != nil { // cleanup the connection
		cep.conn.Close()
		cep.conn = nil
	}

	return cep.exited // return if we exited
}
