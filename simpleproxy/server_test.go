package simpleproxy

import (
	"net"
	"os"
	"testing"

	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh/testutils"
	"github.com/tkw1536/proxyssh/utils"
)

var testServer *ssh.Server

// make addresses for forward and reverse forwarding
var (
	forwardPortsAllow = utils.MustParsePortWithHost(testutils.NewTestListenAddress())
	forwardPortsDeny  = utils.MustParsePortWithHost(testutils.NewTestListenAddress())

	reversePortsAllow = utils.MustParsePortWithHost(testutils.NewTestListenAddress())
	reversePortsDeny  = utils.MustParsePortWithHost(testutils.NewTestListenAddress())
)

func TestMain(m *testing.M) {

	// create a new server and start listening
	testServer = NewSimpleProxyServer(
		testutils.TestLogger(),
		ServerOptions{
			Shell:         "/bin/bash",
			ForwardPorts:  []utils.PortWithHost{forwardPortsAllow},
			ReversePorts:  []utils.PortWithHost{reversePortsAllow},
			ListenAddress: testutils.NewTestListenAddress(),
		},
	)
	// start listening and then serving
	testListener, err := net.Listen("tcp", testServer.Addr)
	if err != nil {
		panic(err)
	}
	go testServer.Serve(testListener)

	// run the code
	code := m.Run()

	// shutdown the testserver
	testServer.Close()

	// and exit
	os.Exit(code)
}
