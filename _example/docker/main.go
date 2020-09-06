package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/tkw1536/proxyssh"
	"github.com/tkw1536/proxyssh/dockerproxy"
	"github.com/tkw1536/proxyssh/utils"

	"github.com/docker/docker/client"
)

var logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	// make the server
	server := dockerproxy.NewProxy(logger, dockerproxy.Options{
		Client:                cli,
		ListenAddress:         ListenAddress,
		DockerLabelUser:       DockerLabelUser,
		DockerLabelAuthFile:   DockerLabelAuthFile,
		DockerLabelKey:        "",
		ContainerShell:        ContainerShell,
		DisableAuthentication: DisableAuthentication,
		IdleTimeout:           IdleTimeout,
		ForwardAddresses:      ForwardAddresses,
		ReverseAddresses:      ReverseAddresses,
	})

	// load rsa host key
	_, err := proxyssh.UseOrMakeHostKey(logger, server, HostKeyPath+"_rsa", proxyssh.RSAAlgorithm)
	if err != nil {
		logger.Fatal(err)
	}

	// load ed25519 host key
	_, err = proxyssh.UseOrMakeHostKey(logger, server, HostKeyPath+"_ed25519", proxyssh.ED25519Algorithm)
	if err != nil {
		logger.Fatal(err)
	}

	// and serve
	logger.Printf("Listening on %s", ListenAddress)
	logger.Fatal(server.ListenAndServe())
}

var (
	// ListenAddress is the address to listen on
	ListenAddress = ":2222"
	// DockerLabelUser is the label to find the container by
	DockerLabelUser = "de.tkw1536.proxyssh.user"
	// DockerLabelAuthFile is the label to find the authorized_keys file by
	DockerLabelAuthFile = "de.tkw1536.proxyssh.authfile"
	// ContainerShell is the executable to run within the container
	ContainerShell = "/bin/sh"
	// DisableAuthentication disables authentication
	DisableAuthentication = false
	// IdleTimeout is the timeout after which an idle connection is killed
	IdleTimeout = 1 * time.Hour
	// ForwardAddresses are ports that are allowed to be forwarded
	ForwardAddresses = utils.NetworkAddressListVar(nil)
	// ReverseAddresses are ports that are allowed to be forwarded (in reverse)
	ReverseAddresses = utils.NetworkAddressListVar(nil)

	// HostKeyPath is the path to the host key
	HostKeyPath = "hostkey.pem"
)

func init() {
	flag.StringVar(&ListenAddress, "port", ListenAddress, "Port to listen on")
	flag.DurationVar(&IdleTimeout, "timeout", IdleTimeout, "Idle Timeout")

	flag.StringVar(&DockerLabelUser, "userlabel", DockerLabelUser, "Label to find docker containers by")
	flag.StringVar(&DockerLabelAuthFile, "keylabel", DockerLabelAuthFile, "Label to find docker containers by")

	flag.StringVar(&ContainerShell, "shell", ContainerShell, "Shell to execute within the container")

	flag.BoolVar(&DisableAuthentication, "unsafe", DisableAuthentication, "Disable ssh server authentication and alllow anyone to connect")

	flag.Var(&ForwardAddresses, "L", "Ports to allow local forwarding for")
	flag.Var(&ReverseAddresses, "R", "Ports to allow reverse forwarding for")

	flag.StringVar(&HostKeyPath, "hostkey", HostKeyPath, "Path to the host key")

	flag.Parse()
}

var cli *client.Client

func init() {
	var err error
	cli, err = client.NewEnvClient()
	if err != nil {
		panic(err)
	}
}
