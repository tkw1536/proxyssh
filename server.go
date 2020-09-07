package proxyssh

import (
	"strings"
	"time"

	"github.com/gliderlabs/ssh"
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
func NewProxySSHServer(logger utils.Logger, opts Options) (server *ssh.Server) {
	server = &ssh.Server{
		Handler: HandleShellCommand(logger, func(s ssh.Session) ([]string, error) {
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
	AllowPortForwarding(logger, server, opts.ForwardAddresses, opts.ReverseAddresses)

	return
}
