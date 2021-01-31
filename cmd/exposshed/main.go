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
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tkw1536/proxyssh/legal"
	"github.com/tkw1536/proxyssh/server"
	"github.com/tkw1536/proxyssh/server/forwarder"
	"github.com/tkw1536/proxyssh/server/shell"
	"github.com/tkw1536/proxyssh/utils"
)

var logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	// init
	sshserver := forwarder.NewForwardingSSHServer(logger, shell.Options{
		ListenAddress: listenAddress,

		IdleTimeout: idleTimeout,

		ForwardAddresses: forwardPorts,
		ReverseAddresses: reversePorts,
	})

	// load host keys
	err := server.UseOrMakeHostKey(logger, sshserver, hostKeyPath+"_rsa", server.RSAAlgorithm)
	if err != nil {
		logger.Fatal(err)
	}
	err = server.UseOrMakeHostKey(logger, sshserver, hostKeyPath+"_ed25519", server.ED25519Algorithm)
	if err != nil {
		logger.Fatal(err)
	}

	// and run
	logger.Printf("Listening on %s", listenAddress)
	logger.Fatal(sshserver.ListenAndServe())
}

var (
	listenAddress = ":2222"

	idleTimeout = 12 * time.Hour

	forwardPorts = utils.NetworkAddressListVar(nil)
	reversePorts = utils.NetworkAddressListVar(nil)

	hostKeyPath = "hostkey.pem"
)

func init() {
	var legalFlag bool
	flag.BoolVar(&legalFlag, "legal", legalFlag, "Print legal notices and exit")
	defer func() {
		if legalFlag {
			fmt.Println("This executable contains code from several different go packages. ")
			fmt.Println("Some of these packages require licensing information to be made available to the end user. ")
			fmt.Println(legal.Notices)
			os.Exit(0)
		}
	}()

	flag.StringVar(&listenAddress, "port", listenAddress, "Port to listen on")
	flag.DurationVar(&idleTimeout, "timeout", idleTimeout, "Timeout to kill inactive connections after")

	flag.Var(&forwardPorts, "L", "Ports to allow local forwarding for")
	flag.Var(&reversePorts, "R", "Ports to allow reverse forwarding for")

	flag.StringVar(&hostKeyPath, "hostkey", hostKeyPath, "Path to the host key")

	flag.Parse()
}
