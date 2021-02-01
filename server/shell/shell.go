package shell

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

	Cmd *exec.Cmd
}

// Init initializes this process
func (sp *SystemProcess) Init(ctx context.Context, isPty bool) error {
	// exec.Command internally does use LookPath(), but doesn't return an error
	// Instead we explicitly call LookPath() to intercept the error

	exe, err := exec.LookPath(sp.command)
	if err != nil {
		err = errors.Wrapf(err, "Can't find %s in path", sp.command)
		return err
	}

	sp.Cmd = exec.Command(exe, sp.args...)
	return nil
}

// String turns ShellProcess into a string
func (sp *SystemProcess) String() string {
	if sp == nil || sp.Cmd == nil {
		return ""
	}

	return strings.Join(append([]string{sp.Cmd.Path}, sp.Cmd.Args...), " ")
}

// Stdout returns a pipe to Stdout
func (sp *SystemProcess) Stdout() (io.ReadCloser, error) {
	return sp.Cmd.StdoutPipe()
}

// Stderr returns a pipe to Stderr
func (sp *SystemProcess) Stderr() (io.ReadCloser, error) {
	return sp.Cmd.StderrPipe()
}

// Stdin returns a pipe to Stdin
func (sp *SystemProcess) Stdin() (io.WriteCloser, error) {
	return sp.Cmd.StdinPipe()
}

// Start starts this process
func (sp *SystemProcess) Start(Term string, resizeChan <-chan proxyssh.WindowSize, isPty bool) (*os.File, error) {
	// not a tty => start the process and be done!
	if !isPty {
		return nil, sp.Cmd.Start()
	}

	// add the terminal environment variable
	sp.Cmd.Env = append(sp.Cmd.Env, fmt.Sprintf("TERM=%s", Term))

	// start the pty
	f, err := pty.Start(sp.Cmd)
	if err != nil {
		return nil, err
	}

	// start tracking window size
	go func() {
		for size := range resizeChan {
			pty.Setsize(f, &pty.Winsize{
				Rows: size.Height,
				Cols: size.Width,
			})
		}
	}()

	// and return a function for this
	return f, nil
}

// Wait waits for the process and returns the exit code
func (sp *SystemProcess) Wait() (code int, err error) {
	// wait for the command
	err = sp.Cmd.Wait()
	code = 255

	// if we have a failure and it's not an exit code
	// we need to return an error
	_, isExitError := err.(*exec.ExitError)
	if err != nil && !isExitError {
		err = errors.Wrap(err, "cmd.Wait() returned non-exit-error")
		return
	}

	// return the exit code
	code = sp.Cmd.ProcessState.ExitCode()
	return code, nil
}

// Cleanup cleans up this process, typically killing it
func (sp *SystemProcess) Cleanup() (killed bool) {
	// no process => return
	if sp.Cmd.Process == nil {
		return true
	}

	// silence any panic()ing errors, but return false!
	defer func() {
		recover()
	}()

	// kill the process, and prevent further attempts
	if sp.Cmd.Process.Kill() == nil {
		sp.Cmd.Process = nil
	}
	return true
}
