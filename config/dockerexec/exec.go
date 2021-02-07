package dockerexec

import (
	"context"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/creack/pty"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/tkw1536/proxyssh"
	"github.com/tkw1536/proxyssh/utils"
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

	// external streams
	stdout, stderr io.ReadCloser
	stdin          io.WriteCloser

	// internal streams
	stdoutTerm, stderrTerm, stdinTerm, ptyTerm, ttyTerm *utils.Terminal

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
func (cep *ContainerExecProcess) Init(ctx context.Context, isTerm bool) error {
	cep.ctx = ctx
	if isTerm {
		cep.config.Tty = true
		return cep.initTerm()
	}

	return cep.initPlain()
}

func (cep *ContainerExecProcess) initPlain() error {
	var err error

	cep.stdout, cep.stdoutTerm, err = utils.NewWritePipe()
	if err != nil {
		return err
	}

	cep.stderr, cep.stderrTerm, err = utils.NewWritePipe()
	if err != nil {
		return err
	}

	cep.stdinTerm, cep.stdin, err = utils.NewReadPipe()
	if err != nil {
		return err
	}

	return nil
}

func (cep *ContainerExecProcess) initTerm() error {
	// create a new pty
	pty, tty, err := pty.Open()
	if err != nil {
		return err
	}

	// store the pty and tty use
	cep.ptyTerm = utils.GetTerminal(pty)
	cep.ttyTerm = utils.GetTerminal(tty)

	// standard output is the tty
	cep.stdout = tty
	cep.stdoutTerm = utils.GetTerminal(tty)

	// standard input is the tty
	cep.stdin = tty
	cep.stdinTerm = utils.GetTerminal(tty)

	return nil
}

// Stdout returns a pipe to Stdout
func (cep *ContainerExecProcess) Stdout() (io.ReadCloser, error) {
	return cep.stdout, nil
}

// Stderr returns a pipe to Stderr
func (cep *ContainerExecProcess) Stderr() (io.ReadCloser, error) {
	return cep.stderr, nil
}

// Stdin returns a pipe to Stdin
func (cep *ContainerExecProcess) Stdin() (io.WriteCloser, error) {
	return cep.stdin, nil
}

// setRawTerminals sets all the terminals to raw mode
func (cep *ContainerExecProcess) setRawTerminals() error {
	if err := cep.stdoutTerm.SetRawInput(); err != nil {
		return err
	}

	if err := cep.stderrTerm.SetRawInput(); err != nil {
		return err
	}

	if err := cep.stdoutTerm.SetRawOutput(); err != nil {
		return err
	}

	return nil
}

// restoreTerminals restores all the terminal modes
func (cep *ContainerExecProcess) restoreTerminals() {
	cep.restoreTerms.Do(func() {
		cep.stdoutTerm.RestoreTerminal()
		cep.stderrTerm.RestoreTerminal()
		cep.stdinTerm.RestoreTerminal()

		// this check has been adapted from upstream; for some reason they hang on specific platforms
		if in := cep.stdinTerm.File(); in != nil && runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
			in.Close()
		}
	})
}

// Start starts this process
func (cep *ContainerExecProcess) Start(Term string, resizeChan <-chan proxyssh.WindowSize, isPty bool) (*os.File, error) {
	if isPty {
		cep.config.Env = append(cep.config.Env, "TERM="+Term)

		// start resizing the terminal
		go func() {
			for size := range resizeChan {
				cep.ptyTerm.ResizeTo(size.Height, size.Width)

				cep.client.ContainerExecResize(cep.ctx, cep.execID, types.ResizeOptions{
					Height: uint(size.Height),
					Width:  uint(size.Width),
				})
			}
		}()
	}

	// start streaming
	if err := cep.execAndStream(true); err != nil {
		return nil, err
	}

	// and return
	return cep.ptyTerm.File(), nil
}

func (cep *ContainerExecProcess) execAndStream(isPty bool) error {

	// set all the streams into raw mode
	if err := cep.setRawTerminals(); err != nil {
		return err
	}

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
	cep.outputErrChan = make(chan error)
	cep.inputDoneChan = make(chan struct{})

	// read output
	go func() {
		if isPty {
			_, err = io.Copy(cep.stdoutTerm.File(), conn.Reader)
			cep.restoreTerminals()
		} else {
			_, err = stdcopy.StdCopy(cep.stdoutTerm.File(), cep.stderrTerm.File(), conn.Reader)
		}

		// close output and send error (if any)
		cep.outputErrChan <- err
	}()

	// write input
	go func() {
		io.Copy(conn.Conn, cep.stdinTerm.File())
		conn.CloseWrite()
		close(cep.inputDoneChan)
	}()

	return nil
}

// waitStreams waits for the streams to finish
func (cep *ContainerExecProcess) waitStreams() error {
	defer cep.restoreTerminals()

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

// Wait waits for the process and returns the exit code
func (cep *ContainerExecProcess) Wait() (code int, err error) {

	// wait for the streams to close
	if err := cep.waitStreams(); err != nil {
		return 0, err
	}

	// inspect and get the actual exit code
	resp, err := cep.client.ContainerExecInspect(cep.ctx, cep.execID)
	if err == nil {
		cep.exited = true
	}
	return resp.ExitCode, err
}

// Cleanup cleans up this process, typically to kill it.
func (cep *ContainerExecProcess) Cleanup() (killed bool) {

	if cep.ptyTerm != nil { // cleanup the pty
		cep.ptyTerm.Close()
		cep.ptyTerm = nil
	}

	if cep.ttyTerm != nil {
		cep.ttyTerm.Close()
		cep.ttyTerm = nil
	}

	if cep.conn != nil { // cleanup the connection
		cep.conn.Close()
		cep.conn = nil
	}

	return cep.exited // return if we exited
}
