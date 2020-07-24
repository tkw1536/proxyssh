package testutils

import (
	"bytes"
	"net"

	"golang.org/x/crypto/ssh"
)

// NewTestServerSession connects and starts a new ssh session on the sever listening at address.
//
// This function sets reasonable defaults for the options.
// If options.HostKeyCallback is not set, sets it to a function that accepts every host key.
// If options.User is the empty string, uses the username "user".
// This function can thus be called with an empty options struct.
//
// If no error occurs, the function expects the caller the Close() method on the client.
// If an error occurs during initialization, the client and session will be closed and an error will be returned.
// A typical invocation of this function should look like:
//
//  client, session, err := NewTestServerSession(address, ssh.ClientConfig{})
//  if err != nil {
//  	return err
//  }
//  defer client.Close()
//
func NewTestServerSession(address string, options ssh.ClientConfig) (*ssh.Client, *ssh.Session, error) {
	// set default options
	if options.HostKeyCallback == nil {
		options.HostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		}
	}
	if options.User == "" {
		options.User = "user"
	}

	// create a new client
	conn, err := ssh.Dial("tcp", address, &options)
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

// RunTestServerCommand runs a command on the ssh server listening at address, and returns its standard output and input.
// The address and options parameters are passed to NewTestServerSession.
// The command being run is determined by command.
// The standard input passed to the command is determined from stdin.
//
// The output of the command (consisting of stdout and stderr), along with it's exit code will be returned.
// If something goes wrong running the command, an error will be returned.
func RunTestServerCommand(address string, options ssh.ClientConfig, command, stdin string) (stdout string, stderr string, code int, err error) {
	// create a new session
	client, session, err := NewTestServerSession(address, options)
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
