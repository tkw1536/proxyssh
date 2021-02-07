package proxyssh

import (
	"flag"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh/feature"
	"github.com/tkw1536/proxyssh/utils"
)

// Options are options that implement features shared by several server implementations.
type Options struct {
	// ListenAddress is the address to listen on.
	// It should be of the form 'address:port'.
	ListenAddress string

	// HostKeyPath is the path to which the host keys are stored.
	// HostKeyAlgorithms are the algorithms to use for host keys.
	//
	// When HostKeyPath is not the empty string, will pass both to UseOrMakeHostKeys.
	HostKeyPath       string
	HostKeyAlgorithms []feature.HostKeyAlgorithm

	// DisableAuthentication allows to completly skip the authentication.
	// This will result in a warning printed to the server
	DisableAuthentication bool

	// ForwardAddresses are addresses that port forwarding is allowed for.
	// ReverseAddresses are addresses that reverse port forwarding is allowed for.
	//
	// See the AllowPortForwarding method for details.
	ForwardAddresses []feature.NetworkAddress
	ReverseAddresses []feature.NetworkAddress

	// IdleTimeout is the timeout after which a connection is considered idle.
	IdleTimeout time.Duration
}

// Apply applies the common options to server.
func (opts *Options) Apply(logger utils.Logger, sshserver *ssh.Server) error {
	// store address and idle timeout
	sshserver.Addr = opts.ListenAddress
	sshserver.IdleTimeout = opts.IdleTimeout

	// turn off authentication if requested
	if opts.DisableAuthentication {
		logger.Print("WARNING: Disabling authentication. Anyone will be able to connect. ")
		sshserver.PublicKeyHandler = nil
	}

	// setup port-forwarding
	feature.AllowPortForwarding(logger, sshserver, opts.ForwardAddresses, opts.ReverseAddresses)

	// setup host keys
	if opts.HostKeyPath != "" {
		if err := feature.UseOrMakeHostKeys(logger, sshserver, opts.HostKeyPath, opts.HostKeyAlgorithms); err != nil {
			return err
		}
	}

	return nil
}

// RegisterFlags registers flags representing the options to the provided flagset.
// When flagset is nil, uses flag.CommandLine.
//
// addUnsafeFlags indiciates if unsafe flags should be added as well.
func (opts *Options) RegisterFlags(flagset *flag.FlagSet, addUnsafeFlags bool) {
	if flagset == nil {
		flagset = flag.CommandLine
	}

	flagset.StringVar(&opts.ListenAddress, "port", opts.ListenAddress, "Port to listen on")
	flagset.DurationVar(&opts.IdleTimeout, "timeout", opts.IdleTimeout, "Timeout to kill inactive connections after")

	if opts.ForwardAddresses == nil {
		opts.ForwardAddresses = []feature.NetworkAddress{}
	}
	fw := feature.NetworkAddressListVar{Addresses: &opts.ForwardAddresses}
	flagset.Var(&fw, "L", "Ports to allow local forwarding for")

	if opts.ReverseAddresses == nil {
		opts.ReverseAddresses = []feature.NetworkAddress{}
	}
	bw := feature.NetworkAddressListVar{Addresses: &opts.ReverseAddresses}
	flagset.Var(&bw, "R", "Ports to allow reverse forwarding for")

	flagset.StringVar(&opts.HostKeyPath, "hostkey", opts.HostKeyPath, "Path hostkeys should be loaded from or created at")

	if addUnsafeFlags {
		flagset.BoolVar(&opts.DisableAuthentication, "unsafe", opts.DisableAuthentication, "Disable ssh server authentication and allow anyone to connect")
	}
}
