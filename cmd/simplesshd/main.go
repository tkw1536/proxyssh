// Command simplesshd provides a simple ssh daemon that works similar to OpenSSH.
// It accepts connections on port 2222 from any interface by default.
//
// This command is not intended to be used in production.
// The various defaults are unsafe.
// It exists to demonstrate the functionality of the proxyssh package.
//
// Overview
//
// When a connection is received no authentication is performed and it is accepted by default.
//
// When an SSH client requests a session this daemon executes a shell command.
// By default, the shell used is /bin/bash, but this can be configured.
// This shell command is executed as the user running the simplesshd command.
// When arguments are provided via the ssh command, these are passed to the command.
// When the client requests to allocate a pty, a pty is created.
// The daemon then proxies the input and output streams between the connection and the executed process.
//
// This daemon furthermore allows Port Forwarding and Reverse Port Forwarding.
// This is only allowed to a limited set of Network Addresses, these have to be provided via arguments.
//
// Configuration
//
// All configuration is performed using command line flags.
//
//  -port hostname:port
// By default connections on any interface on port 2222 will be accepted.
// This can be changed using this argument.
//
//  -shell executable
//
// When executing a user program the '/bin/bash' shell is used by default.
// This argument allows to use a different shell instead.
// The shell is looked up in $PATH.
//
// When an ssh session is started without a provided command, proxyssh starts the shell without any arguments.
// When the user provides a command to run, it is passed to the shell using a '-c' argument.
// For example, suppose the shell is /bin/bash and the user requests the command 'ls -alh'.
// Then this program will execute the command:
//   /bin/bash -c "ls -alh"
// No escaping is performed on the user-provided shell command.
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
// By default, SSH connections are terminated after one hour of inactivity.
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
	"github.com/tkw1536/proxyssh/config/osexec"
	"github.com/tkw1536/proxyssh/legal"
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

	logger.Printf("Listening on %s", options.ListenAddress)
	logger.Fatal(sshserver.ListenAndServe())
}

var options = &proxyssh.Options{
	ListenAddress: ":2222",
	IdleTimeout:   time.Hour,

	DisableAuthentication: false,

	ForwardAddresses: nil,
	ReverseAddresses: nil,

	HostKeyPath: "hostkey.pem",
}

var config = &osexec.SystemExecConfig{
	Shell: "/bin/bash",
}

func init() {
	defer flag.Parse()

	legal.RegisterFlag(nil)
	options.RegisterFlags(nil, false)
	config.RegisterFlags(nil)
}
