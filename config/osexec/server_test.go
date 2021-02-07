package osexec

import (
	"net"
	"os"
	"testing"

	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh"
	"github.com/tkw1536/proxyssh/feature"
	"github.com/tkw1536/proxyssh/internal/testutils"
)

var testServer *ssh.Server

// make addresses for forward and reverse forwarding
var (
	forwardPortsAllow = feature.MustParseNetworkAddress(testutils.NewTestListenAddress())
	forwardPortsDeny  = feature.MustParseNetworkAddress(testutils.NewTestListenAddress())

	reversePortsAllow = feature.MustParseNetworkAddress(testutils.NewTestListenAddress())
	reversePortsDeny  = feature.MustParseNetworkAddress(testutils.NewTestListenAddress())
)

func TestMain(m *testing.M) {

	// make a new server
	var err error
	testServer, err = proxyssh.NewServer(
		testutils.GetTestLogger(),
		&proxyssh.Options{
			ForwardAddresses: []feature.NetworkAddress{forwardPortsAllow},
			ReverseAddresses: []feature.NetworkAddress{reversePortsAllow},
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
