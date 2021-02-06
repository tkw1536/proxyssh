package docker

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

// NewEngineProcess creates a process that executes within a docker container.
//
// The commnd will not prefix the entrypoint.
func NewEngineProcess(client *client.Client, containerID string, command []string) *EngineProcess {
	config := types.ExecConfig{
		AttachStdin:  true,
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          command,
	}

	return &EngineProcess{
		client: client,

		containerID: containerID,
		config:      config,
	}
}

// EngineProcess represents a process running inside a docker engine
type EngineProcess struct {
	// environment
	client *client.Client
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
func (ep *EngineProcess) String() string {
	if ep == nil {
		return ""
	}

	return ep.containerID + " " + strings.Join(ep.config.Cmd, " ")
}

// Init initializes this EngineProcess
func (ep *EngineProcess) Init(ctx context.Context, isTerm bool) error {
	ep.ctx = ctx
	if isTerm {
		ep.config.Tty = true
		return ep.initTerm()
	}

	return ep.initPlain()
}

func (ep *EngineProcess) initPlain() error {
	var err error

	ep.stdout, ep.stdoutTerm, err = utils.NewWritePipe()
	if err != nil {
		return err
	}

	ep.stderr, ep.stderrTerm, err = utils.NewWritePipe()
	if err != nil {
		return err
	}

	ep.stdinTerm, ep.stdin, err = utils.NewReadPipe()
	if err != nil {
		return err
	}

	return nil
}

func (ep *EngineProcess) initTerm() error {
	// create a new pty
	pty, tty, err := pty.Open()
	if err != nil {
		return err
	}

	// store the pty and tty use
	ep.ptyTerm = utils.GetTerminal(pty)
	ep.ttyTerm = utils.GetTerminal(tty)

	// standard output is the tty
	ep.stdout = tty
	ep.stdoutTerm = utils.GetTerminal(tty)

	// standard input is the tty
	ep.stdin = tty
	ep.stdinTerm = utils.GetTerminal(tty)

	return nil
}

// Stdout returns a pipe to Stdout
func (ep *EngineProcess) Stdout() (io.ReadCloser, error) {
	return ep.stdout, nil
}

// Stderr returns a pipe to Stderr
func (ep *EngineProcess) Stderr() (io.ReadCloser, error) {
	return ep.stderr, nil
}

// Stdin returns a pipe to Stdin
func (ep *EngineProcess) Stdin() (io.WriteCloser, error) {
	return ep.stdin, nil
}

// setRawTerminals sets all the terminals to raw mode
func (ep *EngineProcess) setRawTerminals() error {
	if err := ep.stdoutTerm.SetRawInput(); err != nil {
		return err
	}

	if err := ep.stderrTerm.SetRawInput(); err != nil {
		return err
	}

	if err := ep.stdoutTerm.SetRawOutput(); err != nil {
		return err
	}

	return nil
}

// restoreTerminals restores all the terminal modes
func (ep *EngineProcess) restoreTerminals() {
	ep.restoreTerms.Do(func() {
		ep.stdoutTerm.RestoreTerminal()
		ep.stderrTerm.RestoreTerminal()
		ep.stdinTerm.RestoreTerminal()

		// this check has been adapted from upstream; for some reason they hang on specific platforms
		if in := ep.stdinTerm.File(); in != nil && runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
			in.Close()
		}
	})
}

// Start starts this process
func (ep *EngineProcess) Start(Term string, resizeChan <-chan proxyssh.WindowSize, isPty bool) (*os.File, error) {
	if isPty {
		ep.config.Env = append(ep.config.Env, "TERM="+Term)

		// start resizing the terminal
		go func() {
			for size := range resizeChan {
				ep.ptyTerm.ResizeTo(size.Height, size.Width)

				ep.client.ContainerExecResize(ep.ctx, ep.execID, types.ResizeOptions{
					Height: uint(size.Height),
					Width:  uint(size.Width),
				})
			}
		}()
	}

	// start streaming
	if err := ep.execAndStream(true); err != nil {
		return nil, err
	}

	// and return
	return ep.ptyTerm.File(), nil
}

func (ep *EngineProcess) execAndStream(isPty bool) error {

	// set all the streams into raw mode
	if err := ep.setRawTerminals(); err != nil {
		return err
	}

	// create the exec
	res, err := ep.client.ContainerExecCreate(ep.ctx, ep.containerID, ep.config)
	if err != nil {
		return err
	}
	ep.execID = res.ID

	// attach to it
	conn, err := ep.client.ContainerExecAttach(ep.ctx, ep.execID, types.ExecStartCheck{
		Detach: false,
		Tty:    isPty,
	})
	ep.conn = &conn

	// setup channels
	ep.outputErrChan = make(chan error)
	ep.inputDoneChan = make(chan struct{})

	// read output
	go func() {
		if isPty {
			_, err = io.Copy(ep.stdoutTerm.File(), conn.Reader)
			ep.restoreTerminals()
		} else {
			_, err = stdcopy.StdCopy(ep.stdoutTerm.File(), ep.stderrTerm.File(), conn.Reader)
		}

		// close output and send error (if any)
		ep.outputErrChan <- err
	}()

	// write input
	go func() {
		io.Copy(conn.Conn, ep.stdinTerm.File())
		conn.CloseWrite()
		close(ep.inputDoneChan)
	}()

	return nil
}

// waitStreams waits for the streams to finish
func (ep *EngineProcess) waitStreams() error {
	defer ep.restoreTerminals()

	select {
	case err := <-ep.outputErrChan:
		return err
	case <-ep.inputDoneChan: // wait for output also
		select {
		case err := <-ep.outputErrChan:
			return err
		case <-ep.ctx.Done():
			return ep.ctx.Err()
		}
	case <-ep.ctx.Done():
		return ep.ctx.Err()
	}
}

// Wait waits for the process and returns the exit code
func (ep *EngineProcess) Wait() (code int, err error) {

	// wait for the streams to close
	if err := ep.waitStreams(); err != nil {
		return 0, err
	}

	// inspect and get the actual exit code
	resp, err := ep.client.ContainerExecInspect(ep.ctx, ep.execID)
	if err == nil {
		ep.exited = true
	}
	return resp.ExitCode, err
}

// Cleanup cleans up this process, typically to kill it.
func (ep *EngineProcess) Cleanup() (killed bool) {

	if ep.ptyTerm != nil { // cleanup the pty
		ep.ptyTerm.Close()
		ep.ptyTerm = nil
	}

	if ep.ttyTerm != nil {
		ep.ttyTerm.Close()
		ep.ttyTerm = nil
	}

	if ep.conn != nil { // cleanup the connection
		ep.conn.Close()
		ep.conn = nil
	}

	return ep.exited // return if we exited
}
