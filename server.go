// Package proxyssh provides a simple interface to make an ssh server.
package proxyssh

import (
	"strings"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh/utils"
)

// ServerOptions are options for the simple proxy server
type ServerOptions struct {
	// ListenAddress is the address to listen on
	ListenAddress string
	// Shell is the shell to use for the server
	// This is called like `shell -c "command"`` when an ssh command is provided or like `shell` when not.
	// This is passed to exec.LookPath()
	Shell string
	// ForwardAddresses are addresses that port forwarding is allowed for.
	ForwardAddresses []utils.NetworkAddress
	// ReverseAddresses are addresses that reverse port forwarding is allowed for.
	ReverseAddresses []utils.NetworkAddress
	// IdleTimeout is the timeout after which a connection is considered idle
	IdleTimeout time.Duration
}

// NewProxySSHServer makes a new simple proxy server with the given logger and options.
// It returns a new server that was created.
func NewProxySSHServer(logger utils.Logger, opts ServerOptions) (server *ssh.Server) {
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
	server = AllowPortForwarding(logger, server, opts.ForwardAddresses, opts.ReverseAddresses)

	return
}
