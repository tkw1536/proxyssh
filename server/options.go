package server

import (
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh/utils"
)

// Options are options shared by multiple server implementations
type Options struct {
	// ListenAddress is the address to listen on.
	// It should be of the form 'address:port'.
	ListenAddress string

	// HostKeyPath is the path to which the host keys are stored.
	// HostKeyAlgorithms are the algorithms to use for host keys.
	//
	// When HostKeyPath is not the empty string, will pass both to UseOrMakeHostKeys.
	HostKeyPath       string
	HostKeyAlgorithms []HostKeyAlgorithm

	// DisableAuthentication allows to completly skip the authentication.
	// This will result in a warning printed to the server
	DisableAuthentication bool

	// ForwardAddresses are addresses that port forwarding is allowed for.
	// ReverseAddresses are addresses that reverse port forwarding is allowed for.
	//
	// See the AllowPortForwarding method for details.
	ForwardAddresses []utils.NetworkAddress
	ReverseAddresses []utils.NetworkAddress

	// IdleTimeout is the timeout after which a connection is considered idle.
	IdleTimeout time.Duration
}

// Apply applies the common options to server.
func (c Options) Apply(logger utils.Logger, sshserver *ssh.Server) error {
	// store address and idle timeout
	sshserver.Addr = c.ListenAddress
	sshserver.IdleTimeout = c.IdleTimeout

	// turn off authentication if requested
	if c.DisableAuthentication {
		logger.Print("WARNING: Disabling authentication. Anyone will be able to connect. ")
		sshserver.PublicKeyHandler = nil
	}

	// setup port-forwarding
	AllowPortForwarding(logger, sshserver, c.ForwardAddresses, c.ReverseAddresses)

	// setup host keys
	if c.HostKeyPath != "" {
		if err := UseOrMakeHostKeys(logger, sshserver, c.HostKeyPath, c.HostKeyAlgorithms); err != nil {
			return err
		}
	}

	return nil
}
