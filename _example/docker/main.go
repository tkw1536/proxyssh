package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/tkw1536/proxyssh/dockerproxy"
	"github.com/tkw1536/proxyssh/simpleproxy"
	"github.com/tkw1536/proxyssh/utils"

	"github.com/docker/docker/client"
)

var logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	// make the server
	server := dockerproxy.NewDockerProxyServer(logger, dockerproxy.ServerOptions{
		Client:              cli,
		ListenAddress:       ListenAddress,
		DockerLabelUser:     DockerLabelUser,
		DockerLabelAuthFile: DockerLabelAuthFile,
		DockerLabelKey:      "",
		ContainerShell:      ContainerShell,
		NoAuth:              NoAuth,
		IdleTimeout:         time.Duration(IdleTimeout) * time.Second,
		ForwardPorts:        ForwardPorts,
		ReversePorts:        ReversePorts,
	})

	// load rsa host key
	_, err := simpleproxy.UseOrMakeHostKey(logger, server, HostKeyPath+"_rsa", simpleproxy.RSAAlgorithm)
	if err != nil {
		logger.Fatal(err)
	}

	// load ed25519 host key
	_, err = simpleproxy.UseOrMakeHostKey(logger, server, HostKeyPath+"_ed25519", simpleproxy.ED25519Algorithm)
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
	// NoAuth disables authentication
	NoAuth = false
	// IdleTimeout is the timeout after which an idle connection is killed
	IdleTimeout = 30
	// ForwardPorts are ports that are allowed to be forwarded
	ForwardPorts = utils.PortListVar(nil)
	// ReversePorts are ports that are allowed to be forwarded (in reverse)
	ReversePorts = utils.PortListVar(nil)

	// HostKeyPath is the path to the host key
	HostKeyPath = "hostkey.pem"
)

func init() {
	flag.StringVar(&ListenAddress, "port", ListenAddress, "Port to listen on")
	flag.IntVar(&IdleTimeout, "timeout", IdleTimeout, "Idle Timeout in seconds")

	flag.StringVar(&DockerLabelUser, "userlabel", DockerLabelUser, "Label to find docker containers by")
	flag.StringVar(&DockerLabelAuthFile, "keylabel", DockerLabelAuthFile, "Label to find docker containers by")

	flag.StringVar(&ContainerShell, "shell", ContainerShell, "Shell to execute within the container")

	flag.BoolVar(&NoAuth, "unsafe", NoAuth, "Disable ssh server authentication and alllow anyone to connect")

	flag.Var(&ForwardPorts, "L", "Ports to allow local forwarding for")
	flag.Var(&ReversePorts, "R", "Ports to allow reverse forwarding for")

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
