package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/creack/pty"
	"github.com/pkg/errors"
)

// NewSystemProcess creates a new system process
func NewSystemProcess(command string, args []string) (*SystemProcess, error) {
	// exec.Command internally does use LookPath(), but doesn't return an error
	// Instead we explicitly call LookPath() to intercept the error

	exe, err := exec.LookPath(command)
	if err != nil {
		err = errors.Wrapf(err, "Can't find %s in path", command)
		return nil, err
	}

	return &SystemProcess{
		Cmd: exec.Command(exe, args...),
	}, nil
}

// SystemProcess represents a process that is run using a the shell on the current machine
type SystemProcess struct {
	Cmd *exec.Cmd
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

// Start starts the process
func (sp *SystemProcess) Start() error {
	return sp.Cmd.Start()
}

// StartPty starts the process inside a pseudo tty
func (sp *SystemProcess) StartPty(Term string) (*os.File, error) {
	// add the terminal environment variable
	sp.Cmd.Env = append(sp.Cmd.Env, fmt.Sprintf("TERM=%s", Term))

	// and go!
	return pty.Start(sp.Cmd)
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

// TryKill tries to kill the process
func (sp *SystemProcess) TryKill() (success bool) {
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
