package shell

import (
	"net"
	"os"
	"testing"

	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh/server"
	"github.com/tkw1536/proxyssh/testutils"
	"github.com/tkw1536/proxyssh/utils"
)

var testServer *ssh.Server

// make addresses for forward and reverse forwarding
var (
	forwardPortsAllow = utils.MustParseNetworkAddress(testutils.NewTestListenAddress())
	forwardPortsDeny  = utils.MustParseNetworkAddress(testutils.NewTestListenAddress())

	reversePortsAllow = utils.MustParseNetworkAddress(testutils.NewTestListenAddress())
	reversePortsDeny  = utils.MustParseNetworkAddress(testutils.NewTestListenAddress())
)

func TestMain(m *testing.M) {

	// make a new server
	var err error
	testServer, err = server.NewServer(
		testutils.GetTestLogger(),
		server.Options{
			ForwardAddresses: []utils.NetworkAddress{forwardPortsAllow},
			ReverseAddresses: []utils.NetworkAddress{reversePortsAllow},
		},
		&SystemExecConfig{
			Shell: "/bin/bash",
		},
	)
	if err != nil {
		panic(err)
	}

	// start listening and then serving
	addr := testutils.NewTestListenAddress()
	testListener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	go testServer.Serve(testListener)
	testServer.Addr = addr

	// run the code
	code := m.Run()

	// shutdown the testserver
	testServer.Close()

	// and exit
	os.Exit(code)
}
