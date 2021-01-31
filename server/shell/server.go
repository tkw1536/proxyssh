// Package shell implements an OpenSSH-like ssh server
package shell

import (
	"errors"
	"strings"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh"
	"github.com/tkw1536/proxyssh/server"
	"github.com/tkw1536/proxyssh/utils"
)

// Options are options for the NewProxySSHServer function.
type Options struct {
	// ListenAddress is the address to listen on.
	// It should be of the form 'address:port'.
	ListenAddress string

	// Shell is the shell to use for the server
	// This is called like `shell -c "command"`` when an ssh command is provided or like `shell` when not.
	// The shell is passed to exec.LookPath().
	Shell string

	// ForwardAddresses are addresses that port forwarding is allowed for.
	ForwardAddresses []utils.NetworkAddress
	// ReverseAddresses are addresses that reverse port forwarding is allowed for.
	ReverseAddresses []utils.NetworkAddress

	// IdleTimeout is the timeout after which a connection is considered idle.
	IdleTimeout time.Duration
}

// NewProxySSHServer makes a new OpenSSH-like ssh server.
//
// When a user connects to the server, it executes a shell and proxies the session input and ouput streams accordingly.
// It furthermore allows sending traffic to and from specific ip addresses.
// This can be configured using Options.
//
// It uses the provided options, and returns the new server that was created.
// It returns the new server that was created.
func NewProxySSHServer(logger utils.Logger, opts Options) (sshserver *ssh.Server) {
	sshserver = &ssh.Server{
		Handler: HandleCommand(logger, func(s ssh.Session) ([]string, error) {
			command := s.Command()
			if len(command) == 0 {
				// no arguments were provided => run /bin/bash
				command = []string{opts.Shell}
			} else {
				// some arguments were provided => run /bin/bash -c
				command = []string{opts.Shell, "-c", strings.Join(command, " ")}
			}
			return command, nil
		}),

		Addr:        opts.ListenAddress,
		IdleTimeout: opts.IdleTimeout,
	}
	server.AllowPortForwarding(logger, sshserver, opts.ForwardAddresses, opts.ReverseAddresses)

	return
}

// HandleCommand creates an ssh.Handler that runs a shell command for every ssh.Session that connects.
//
// shellCommand is a function.
// It is called for every ssh session and should return the shell command along with any arguments to execute for the provided session.
// The returned array must be at least of length 1.
// The first argument will be passed to exec.LookPath.
// When shell command returns a non-nil error, no command will be executed and the session will be aborted.
//
// logger is called for every significant event that occurs.
//
// See also the CommandSession struct and NewCommandSession func.
func HandleCommand(logger utils.Logger, shellCommand func(session ssh.Session) (command []string, err error)) ssh.Handler {
	return proxyssh.HandleProcess(logger, func(session ssh.Session) (proxyssh.Process, error) {
		command, err := shellCommand(session)
		if err != nil {
			return nil, err
		}

		if len(command) == 0 {
			return nil, errors.New("didn't create a process")
		}

		return NewSystemProcess(command[0], command[1:])
	})
}
