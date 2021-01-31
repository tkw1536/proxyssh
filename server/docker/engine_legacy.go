package docker

import (
	"io"
	"os"

	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh/server/shell"
)

// TODO: Make this use the docker SDK somehow!
// So far this has been an issue, but let's continue.

// NewLegacyEngineProcess creates a process that executes inside a container
//
// The command returned will depend on the 'docker' executable being availabel on the underlying system.
// The commnd will not prefix the entrypoing.
func NewLegacyEngineProcess(s ssh.Session, containerID string, command []string, workdir string, user string) (*LegacyEngineProcess, error) {
	args := []string{"exec", "--interactive"}

	// ensure it's a tty when we asked for one
	if _, _, isPty := s.Pty(); isPty {
		args = append(args, "--tty")
	}

	// append the workdir
	if workdir != "" {
		args = append(args, "--workdir", workdir)
	}

	// append the user
	if user != "" {
		args = append(args, "--user", user)
	}

	// append the container id and command
	args = append(args, containerID)
	args = append(args, command...)

	// make a system process
	sp, err := shell.NewSystemProcess("docker", args)
	if err != nil {
		return nil, err
	}

	return &LegacyEngineProcess{process: *sp}, nil
}

// LegacyEngineProcess represents a process running inside a docker engine
type LegacyEngineProcess struct {
	process shell.SystemProcess
}

// String turns EngineProcess into a string
func (ep *LegacyEngineProcess) String() string {
	if ep == nil {
		return ""
	}

	return ep.process.String()
}

// Stdout returns a pipe to Stdout
func (ep *LegacyEngineProcess) Stdout() (io.ReadCloser, error) {
	return ep.process.Stdout()
}

// Stderr returns a pipe to Stderr
func (ep *LegacyEngineProcess) Stderr() (io.ReadCloser, error) {
	return ep.process.Stderr()
}

// Stdin returns a pipe to Stdin
func (ep *LegacyEngineProcess) Stdin() (io.WriteCloser, error) {
	return ep.process.Stdin()
}

// Start starts the process
func (ep *LegacyEngineProcess) Start() error {
	return ep.Start()
}

// StartPty starts the process inside a pseudo tty
func (ep *LegacyEngineProcess) StartPty(Term string) (*os.File, error) {
	return ep.process.StartPty(Term)
}

// Wait waits for the process and returns the exit code
func (ep *LegacyEngineProcess) Wait() (code int, err error) {
	return ep.process.Wait()
}

// TryKill tries to kill the process
func (ep *LegacyEngineProcess) TryKill() (success bool) {
	return ep.process.TryKill()
}
