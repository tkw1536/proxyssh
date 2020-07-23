// Package simpleproxy provides a simple interface to make an ssh server.
package simpleproxy

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
	// Shell is the shell to run
	Shell string
	// ForwardPorts to allow forwarding for
	ForwardPorts []utils.NetworkAddress
	// ReversePorts to allow forwarding for
	ReversePorts []utils.NetworkAddress
	// IdleTimeout is the timeout after which a connection is considered idle
	IdleTimeout time.Duration
}

// NewSimpleProxyServer makes a new simple proxy server
func NewSimpleProxyServer(logger utils.Logger, opts ServerOptions) (server *ssh.Server) {
	server = &ssh.Server{
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
	server = AllowPortForwarding(logger, server, opts.ForwardPorts, opts.ReversePorts)

	return
}
