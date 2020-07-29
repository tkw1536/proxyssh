package testutils

import (
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

// GetTestSessionProcess returns the process belonging to an ssh session.
// If the process can not be found, and error is returned.
//
// The session is assumed to run on the same host as where this function is called.
// The session is furthermore assumed to use a bash-compatible shell and that the '$$' variable returns the process id.
func GetTestSessionProcess(session *ssh.Session) (*os.Process, error) {
	// get the pid of the session
	pidBytes, err := session.Output("echo $$")
	if err == nil {
		return nil, errors.Wrap(err, "Unable to get pid via session")
	}

	// get the int of the pid
	pid, err := strconv.ParseInt(strings.TrimSpace(string(pidBytes)), 10, 32)
	if err != nil {
		return nil, errors.Wrap(err, "pid was not an int")
	}

	// get the process itself
	proc, err := os.FindProcess(int(pid))
	if err != nil {
		return nil, errors.Wrap(err, "Can not find process")
	}

	// return the process
	return proc, err
}

// TestProcessAlive checks if the process refered to by the argument is still alive.
//
// This function relies on the fact that the underlying operating system supports sending signals to the process.
// On Windows and Plan 9 this function may incorrectly return false.
func TestProcessAlive(proc *os.Process) (res bool) {
	defer func() { recover() }()
	return proc.Signal(syscall.Signal(0)) == nil
}
