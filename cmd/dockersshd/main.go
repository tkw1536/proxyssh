// Command dockersshd provides an ssh server that executes commands inside docker.
// It accepts connections on port 2222 from any interface by default.
//
//
// Overview
//
// When a connection is received the daemon will first attempt to find a matching docker container.
// This happens by finding docker containers where the label 'de.tkw1536.proxyssh.user' is equal to the username of the ssh connection.
// If no unique matching container is found, the connection is rejected immediatly.
// The label used can be customized, see Configuration below.
//
// Next, the server investigates the container further and attempts to find an OpenSSH-like authorized_keys file.
// The location of this file is dervived from the 'de.tkw1536.proxyssh.authfile' label.
// If the ssh key of the client connection matches any of the public keys in the authorized_key files, the connection is accepted.
// Otherwise it is rejected.
//
// Afterwards, a new shell process is executed inside the Docker Container.
// By default, the shell used is /bin/sh, but this can be configured.
// When arguments are provided via the ssh command, these are passed to the command.
// When the client requests to allocate a pty, a pty is created.
// The daemon then proxies the input and output streams between the connection and the executed process.
//
// This daemon furthermore allows Port Forwarding and Reverse Port Forwarding.
// This is only allowed to a limited set of Network Addresses, these have to be provided via arguments.
// These are evaluated relative to the 'dockersshd' host, not the docker container in question.
//
//
// Configuration
//
// All configuration is performed using command line flags.
//
//  -port hostname:port
// By default connections on any interface on port 2222 will be accepted.
// This can be changed using this argument.
//
//  -userlabel label
//
// To associate a docker container with an incoming connection by default the 'de.tkw1536.proxyssh.user' label is used.
// In order for a connection to succeed, there must be a single running container with the label value equal to the username
// of the incoming connection.
// This argument can be used to use a different label instead.
//
//  -keylabel label
//
// By default to authenticate a user the 'de.tkw1536.proxyssh.authfile' label of docker containers is used.
// The value of this label should contain comma-seperated file paths to authorized_keys files within the docker container.
// If a file does not exist or is invalid, it is silently ignored.
// Connections are accepted if any of the public key signatures match the incoming ssh key.
// This argument can be used to use a different label instead.
//
//  -unsafe
//
// This flag can be used to turn off authentication completly.
// It should not be used in production, and is for debugging purposes only.
//
//  -shell executable
//
// When executing a user program inside a docker container the '/bin/sh' shell is used by default.
// This argument allows to use a different shell instead.
// The shell is not looked up in $PATH.
//
// When an ssh session is started without a provided command, dockersshd starts the shell without any arguments.
// When the user provides a command to run, it is passed to the shell using a '-c' argument.
// For example, suppose the shell is /bin/sh and the user requests the command 'ls -alh'.
// Then this program will execute the command:
//   /bin/sh -c "ls -alh"
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
// The daemon supports two kinds of ssh host keys, an RSA and an ED25519 key.
// By default these are stored in two files called 'hostkey.pem_rsa' and 'hostkey.pem_ed25519' in the working directory of the proxysshd process.
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
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tkw1536/proxyssh/legal"
	"github.com/tkw1536/proxyssh/server"
	"github.com/tkw1536/proxyssh/server/docker"
	"github.com/tkw1536/proxyssh/utils"

	"github.com/docker/docker/client"
)

var logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	// init
	sshserver := docker.NewProxy(logger, docker.Options{
		Client: cli,

		ListenAddress: listenAddress,

		DockerLabelUser:     dockerLabelUser,
		DockerLabelAuthFile: dockerLabelAuthFile,

		ContainerShell: containerShell,

		DisableAuthentication: disableAuthentication,

		IdleTimeout: idleTimeout,

		ForwardAddresses: forwardAddresses,
		ReverseAddresses: reverseAddresses,
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

	dockerLabelUser     = "de.tkw1536.proxyssh.user"
	dockerLabelAuthFile = "de.tkw1536.proxyssh.authfile"

	containerShell = "/bin/sh"

	disableAuthentication = false

	idleTimeout = 1 * time.Hour

	forwardAddresses = utils.NetworkAddressListVar(nil)
	reverseAddresses = utils.NetworkAddressListVar(nil)

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
	flag.DurationVar(&idleTimeout, "timeout", idleTimeout, "Idle Timeout")

	flag.StringVar(&dockerLabelUser, "userlabel", dockerLabelUser, "Label to find docker files by")
	flag.StringVar(&dockerLabelAuthFile, "keylabel", dockerLabelAuthFile, "Label to find the authorized_keys file by")

	flag.StringVar(&containerShell, "shell", containerShell, "Shell to execute within the container")

	flag.BoolVar(&disableAuthentication, "unsafe", disableAuthentication, "Disable ssh server authentication and alllow anyone to connect")

	flag.Var(&forwardAddresses, "L", "Ports to allow local forwarding for")
	flag.Var(&reverseAddresses, "R", "Ports to allow reverse forwarding for")

	flag.StringVar(&hostKeyPath, "hostkey", hostKeyPath, "Path to the host key")

	flag.Parse()

}

var cli *client.Client

func init() {
	var err error
	cli, err = client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	cli.NegotiateAPIVersion(context.Background())
}
