package osexec

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/creack/pty"
	"github.com/pkg/errors"
	"github.com/tkw1536/proxyssh"
	"github.com/tkw1536/proxyssh/internal/term"
	"github.com/tkw1536/proxyssh/logging"
)

// NewSystemProcess creates a new system process
func NewSystemProcess(command string, args []string) *SystemProcess {
	return &SystemProcess{
		command: command,
		args:    args,
	}
}

// SystemProcess represents a process that is run using a the shell on the current machine
type SystemProcess struct {
	command string
	args    []string

	cmd      *exec.Cmd
	terminal *term.Pair
}

// Init initializes this process
func (sp *SystemProcess) Init(ctx context.Context, detector logging.MemoryLeakDetector, isPty bool) error {
	// exec.Command internally does use LookPath(), but doesn't return an error
	// Instead we explicitly call LookPath() to intercept the error

	exe, err := exec.LookPath(sp.command)
	if err != nil {
		err = errors.Wrapf(err, "Can't find %s in path", sp.command)
		return err
	}

	sp.cmd = exec.Command(exe, sp.args...)
	return nil
}

// String turns ShellProcess into a string
func (sp *SystemProcess) String() string {
	if sp == nil || sp.cmd == nil {
		return ""
	}

	return strings.Join(append([]string{sp.cmd.Path}, sp.cmd.Args...), " ")
}

// Stdout returns a pipe to Stdout
func (sp *SystemProcess) Stdout() (io.ReadCloser, error) {
	return sp.cmd.StdoutPipe()
}

// Stderr returns a pipe to Stderr
func (sp *SystemProcess) Stderr() (io.ReadCloser, error) {
	return sp.cmd.StderrPipe()
}

// Stdin returns a pipe to Stdin
func (sp *SystemProcess) Stdin() (io.WriteCloser, error) {
	return sp.cmd.StdinPipe()
}

// Start starts this process
func (sp *SystemProcess) Start(detector logging.MemoryLeakDetector, Term string, resizeChan <-chan proxyssh.WindowSize, isPty bool) (*os.File, error) {
	// not a tty => start the process and be done!
	if !isPty {
		return nil, sp.cmd.Start()
	}

	// add the terminal environment variable
	sp.cmd.Env = append(sp.cmd.Env, fmt.Sprintf("TERM=%s", Term))

	// start the pty
	f, err := pty.Start(sp.cmd)
	if err != nil {
		return nil, err
	}

	// use a new terminal
	sp.terminal = &term.Pair{}
	sp.terminal.Use(f)

	// handle resize events on the terminal
	sp.terminal.Handle(resizeChan)

	// and return a function for this
	return sp.terminal.External(), nil
}

// Wait waits for the process and returns the exit code
func (sp *SystemProcess) Wait(detector logging.MemoryLeakDetector) (code int, err error) {

	// wait for the command
	detector.Add("Wait")
	err = sp.cmd.Wait()
	code = 255
	detector.Done("Wait")

	// if we have a failure and it's not an exit code
	// we need to return an error
	_, isExitError := err.(*exec.ExitError)
	if err != nil && !isExitError {
		err = errors.Wrap(err, "cmd.Wait() returned non-exit-error")
		return
	}

	// return the exit code
	code = sp.cmd.ProcessState.ExitCode()
	return code, nil
}

// Cleanup cleans up this process, typically killing it
func (sp *SystemProcess) Cleanup() (killed bool) {
	sp.terminal.Close()

	// no process => return
	if sp.cmd.Process == nil {
		return true
	}

	// silence any panic()ing errors, but return false!
	defer func() {
		recover()
	}()

	// kill the process, and prevent further attempts
	if sp.cmd.Process.Kill() == nil {
		sp.cmd.Process = nil
	}
	return true
}
