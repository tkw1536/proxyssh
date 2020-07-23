package testutils

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

// NewTestServerSession makes a new test server session
func NewTestServerSession(addr string, opts ssh.ClientConfig) (*ssh.Client, *ssh.Session, error) {
	if opts.HostKeyCallback == nil {
		opts.HostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		}
	}
	// create a new client
	conn, err := ssh.Dial("tcp", addr, &opts)
	if err != nil {
		return nil, nil, err
	}

	// create a new session
	session, err := conn.NewSession()
	if err != nil {
		conn.Close()
		return nil, nil, err
	}

	return conn, session, err
}

// RunTestServerCommand runs a command on the test server and returns its stdout, stderr and code
func RunTestServerCommand(addr string, opts ssh.ClientConfig, command, stdin string) (stdout string, stderr string, code int, err error) {
	// create a new session
	client, session, err := NewTestServerSession(addr, opts)
	if err != nil {
		return
	}
	defer session.Close()
	defer client.Close()

	// setup input
	stdinBuf := bytes.NewBufferString(stdin)
	session.Stdin = stdinBuf

	// setup output
	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	// run the command and get the exit code of the error
	err = session.Run(command)
	if err == nil {
		code = 0
	} else if eerr, iseerr := err.(*ssh.ExitError); iseerr {
		code = eerr.ExitStatus()
		err = nil
	} else {
		return
	}

	// save the buffers
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	return
}

// GenerateRSATestKeyPair generates a new rsa keypair to be used for testing
// If generation fails, calls panic()
func GenerateRSATestKeyPair() (ssh.Signer, ssh.PublicKey) {
	privkey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	signer, err := ssh.NewSignerFromKey(privkey)
	if err != nil {
		panic(err)
	}

	return signer, signer.PublicKey()
}

// GetTestSessionProcess returns the process belonging to a test session
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

// TestProcessAlive checks if a process is alive
func TestProcessAlive(proc *os.Process) (res bool) {
	defer func() { recover() }()
	return proc.Signal(syscall.Signal(0)) == nil
}
