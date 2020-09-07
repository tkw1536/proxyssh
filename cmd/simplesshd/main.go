// Command simplesshd provides a simple ssh server that executes commands on the local machine.
package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/tkw1536/proxyssh"
	"github.com/tkw1536/proxyssh/utils"
)

var logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	// start the server
	server := proxyssh.NewProxySSHServer(logger, proxyssh.Options{
		ListenAddress:    ListenAddress,
		IdleTimeout:      IdleTimeout,
		Shell:            Shell,
		ForwardAddresses: ForwardPorts,
		ReverseAddresses: ReversePorts,
	})

	// load rsa host key
	err := proxyssh.UseOrMakeHostKey(logger, server, HostKeyPath+"_rsa", proxyssh.RSAAlgorithm)
	if err != nil {
		logger.Fatal(err)
	}

	// load ed25519 host key
	err = proxyssh.UseOrMakeHostKey(logger, server, HostKeyPath+"_ed25519", proxyssh.ED25519Algorithm)
	if err != nil {
		logger.Fatal(err)
	}

	// serve the server
	logger.Printf("Listening on %s", ListenAddress)
	logger.Fatal(server.ListenAndServe())
}

var (
	// ListenAddress is the address to listen on
	ListenAddress = ":2222"
	// IdleTimeout is the timeout after which an idle connection is killed
	IdleTimeout = time.Hour
	// Shell is the shell to use
	Shell = "/bin/bash"
	// ForwardPorts are ports that are allowed to be forwarded
	ForwardPorts = utils.NetworkAddressListVar(nil)
	// ReversePorts are ports that are allowed to be forwarded (in reverse)
	ReversePorts = utils.NetworkAddressListVar(nil)
	// HostKeyPath is the base path to the host key
	HostKeyPath = "hostkey.pem"
)

func init() {
	flag.StringVar(&ListenAddress, "port", ListenAddress, "Port to listen on")
	flag.DurationVar(&IdleTimeout, "timeout", IdleTimeout, "Timeout to kill inactive connections after")
	flag.StringVar(&Shell, "shell", Shell, "Shell to use")

	flag.Var(&ForwardPorts, "L", "Ports to allow local forwarding for")
	flag.Var(&ReversePorts, "R", "Ports to allow reverse forwarding for")

	flag.StringVar(&HostKeyPath, "hostkey", HostKeyPath, "Path to the host key")

	flag.Parse()
}
