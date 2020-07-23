package testutils

import (
	"bytes"
	"net"

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
	_, session, err := NewTestServerSession(addr, opts)
	if err != nil {
		return
	}
	defer session.Close()

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
