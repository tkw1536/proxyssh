// Command exposshed provides an daemon that allows clients to use local and remote port forwarding.
// It accepts connections on port 2222 from any interface by default, but does not provide shell interface.
//
// This command is not intended to be used on a public interface
//
// Overview
//
// When a connection is received no authentication is performed and it is accepted by default.
// It then permits port forwarding and reverse port forwarding as configured using the '-L' and '-R' flags.
//
// Configuration
//
// All configuration is performed using command line flags.
//
//  -port hostname:port
// By default connections on any interface on port 2222 will be accepted.
// This can be changed using this argument.
//
//
//  -L host:port, -R host:port
//
// To configure the ports to allow traffic to and from certain hosts in the local network via the ssh server, the '-L' and '-R' flags can be used.
// '-L' enables the ssh client to send connections to the provided host:port combination.
// '-R' enables the reverse, enabling the ssh client to accept connections at the provided host and port.
// Both flags can be passed multiple times.
//
//  -hostkey prefix
//
// Te daemon supports two kinds of ssh host keys, an RSA and an ED25519 key.
// By default these are stored in two files called 'hostkey.pem_rsa' and 'hostkey.pem_ed25519' in the working directory of the simplesshd process.
// If either of these files do not exist, they are generated when the program runs for the first time.
//
// It is possible to customize where these files are stored.
// Using this argument their prefix (by default 'hostkey.pem') can be set.
//
//  -timeout time
//
// By default, SSH connections are terminated after twelve hours of inactivity.
// This timeout can be customized using this flag.
// The time argument should be a sequence of  numbers follows by their units.
// Valid units are "ns", "us", "ms", "s", "m" and "h".
// See also golang's time.ParseDuration function.
package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/tkw1536/proxyssh"
	"github.com/tkw1536/proxyssh/legal"
	"github.com/tkw1536/proxyssh/server"
	"github.com/tkw1536/proxyssh/server/forwarder"
)

var logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	sshserver, err := proxyssh.NewServer(
		logger,
		options,
		config,
	)

	if err != nil {
		logger.Fatalf("Failed to initialize server: %s", err)
	}

	// and run
	logger.Printf("Listening on %s", options.ListenAddress)
	logger.Fatal(sshserver.ListenAndServe())
}

var options = &server.Options{
	ListenAddress: ":2222",
	IdleTimeout:   12 * time.Hour,

	DisableAuthentication: true,

	ForwardAddresses: nil,
	ReverseAddresses: nil,

	HostKeyPath: "hostkey.pem",
}

var config = &forwarder.EmptyConfiguration{}

func init() {
	defer flag.Parse()

	legal.RegisterFlag(nil)
	options.RegisterFlags(nil, false)
}
