package docker

// TODO: Set raw mode at the end and at the beginning
// https://github.com/docker/cli/blob/88c6089300a82d3373892adf6845a4fed1a4ba8d/cli/command/container/exec.go
// https://github.com/docker/cli/blob/88c6089300a82d3373892adf6845a4fed1a4ba8d/cli/command/container/hijack.go

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/creack/pty"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gliderlabs/ssh"
)

// NewEngineProcess creates a process that executes inside a container
//
// The command returned will depend on the 'docker' executable being availabel on the underlying system.
// The commnd will not prefix the entrypoing.
func NewEngineProcess(ctx context.Context, s ssh.Session, containerID string, client *client.Client, command []string) (*EngineProcess, error) {

	// channel for the pty
	_, _, isTerm := s.Pty()

	config := types.ExecConfig{
		Tty:          isTerm,
		AttachStdin:  true,
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          command,
	}

	ep := &EngineProcess{
		client: client,
		ctx:    ctx,

		containerID: containerID,
		config:      config,
	}

	if err := ep.init(s, isTerm); err != nil {
		return nil, err
	}
	return ep, nil
}

// EngineProcess represents a process running inside a docker engine
type EngineProcess struct {
	client *client.Client
	ctx    context.Context

	containerID string
	config      types.ExecConfig

	// input / output
	stdout, stderr io.ReadCloser
	stdin          io.WriteCloser

	// conn is the connection to close
	conn *types.HijackedResponse

	// input / output
	stdoutPipe, stderrPipe *os.File
	stdinPipe              *os.File
	pty                    *os.File

	// TTY: Use the pty
	term ssh.Pty

	outputErrChan chan error
	inputDoneChan chan struct{}

	// runtime id
	execID string
	exited bool
}

// String turns EngineProcess into a string
func (ep *EngineProcess) String() string {
	if ep == nil {
		return ""
	}

	return ep.containerID + " " + strings.Join(ep.config.Cmd, " ")
}

func (ep *EngineProcess) init(s ssh.Session, isTerm bool) error {
	if isTerm {
		return ep.initTerm(s)
	}

	return ep.initPlain(s)
}

func (ep *EngineProcess) initPlain(s ssh.Session) error {
	var err error

	ep.stdout, ep.stdoutPipe, err = os.Pipe()
	if err != nil {
		return err
	}

	ep.stderr, ep.stderrPipe, err = os.Pipe()
	if err != nil {
		return err
	}

	ep.stdinPipe, ep.stdin, err = os.Pipe()
	if err != nil {
		return err
	}

	return nil
}

func (ep *EngineProcess) initTerm(s ssh.Session) error {
	fpty, ftty, err := pty.Open()
	if err != nil {
		return err
	}

	// setup pty for setting up stuff
	ep.pty = fpty

	ep.stdout = ftty
	ep.stdoutPipe = ftty

	ep.stderr = ftty
	ep.stderrPipe = ftty

	ep.stdin = ftty
	ep.stdinPipe = ftty

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

func (ep *EngineProcess) startCommon(isTerm bool) error {
	// TODO: Set raw mode and revert at the end!

	// create the exec
	res, err := ep.client.ContainerExecCreate(ep.ctx, ep.containerID, ep.config)
	if err != nil {
		return err
	}
	ep.execID = res.ID

	// attach to it
	conn, err := ep.client.ContainerExecAttach(ep.ctx, ep.execID, ep.config)
	ep.conn = &conn

	// setup channels
	ep.outputErrChan = make(chan error)
	ep.inputDoneChan = make(chan struct{})

	// read output
	go func() {
		fmt.Println("copy over output")
		if isTerm { // tty => regular copy
			_, err = io.Copy(ep.stdoutPipe, conn.Reader)
		} else {
			_, err = stdcopy.StdCopy(ep.stdoutPipe, ep.stderrPipe, conn.Reader)
		}

		// send error (if any)
		ep.outputErrChan <- err
	}()

	// write input
	go func() {
		fmt.Println("copy over input")
		io.Copy(conn.Conn, ep.stdinPipe)
		conn.CloseWrite()
		close(ep.inputDoneChan)
	}()

	return nil
}

// Start starts the process
func (ep *EngineProcess) Start() error {
	return ep.startCommon(false)
}

// StartPty starts the process inside a pseudo tty
func (ep *EngineProcess) StartPty(Term string) (*os.File, error) {
	ep.config.Env = append(ep.config.Env, "TERM="+Term)
	return ep.pty, ep.startCommon(true)
}

// waitStreams waits for the streams to finish
func (ep *EngineProcess) waitStreams() error {
	select {
	case err := <-ep.outputErrChan:
		return err
	case <-ep.inputDoneChan:
		// wait for output also
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

	if err := ep.waitStreams(); err != nil {
		return 0, err
	}

	resp, err := ep.client.ContainerExecInspect(ep.ctx, ep.execID)
	if err == nil {
		ep.exited = true
	}
	return resp.ExitCode, err
}

// TryKill tries to kill the process
func (ep *EngineProcess) TryKill() (success bool) {
	if ep.conn != nil {
		ep.conn.Close()
		ep.conn = nil
	}

	return ep.exited // return if we exited
}
