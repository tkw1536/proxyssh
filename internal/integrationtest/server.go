package integrationtest

import (
	"net"

	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh"
	"github.com/tkw1536/proxyssh/internal/testutils"
	"github.com/tkw1536/proxyssh/logging"
)

// NewServer creates a new proxyssh.Server for an integration test.
// Returns the test server and a cleanup function to be defered.
//
//
// When options is nil, sets up empty options.
//
// When testServer.Handler remains unset, sets it to respond with "success" on stdout and exit immediatly.
// When testServer.HostSigners remains unset, generates a single reusable test server key.
//
// This function should be used like:
//
//   testServer, cleanup := NewServer(testServer)
//   defer cleanup()
// This function is untested.
func NewServer(options *proxyssh.Options, configurations ...proxyssh.Configuration) (testServer *ssh.Server, testLogger logging.Logger, cleanup func()) {

	testLogger = GetLogger()

	// ensure that options are valid
	if options == nil {
		options = &proxyssh.Options{}
	}

	// create a new server with the provided options and configuration
	var err error
	testServer, err = proxyssh.NewServer(
		testLogger,
		options,
		configurations...,
	)
	if err != nil {
		panic(err)
	}

	// ensure that the server has at least a single host key pair
	if testServer.HostSigners == nil {
		testServer.HostSigners = append(testServer.HostSigners, getTestKey())
	}

	if testServer.Handler == nil {
		testServer.Handler = func(s ssh.Session) {
			s.Write([]byte("success"))
			s.Exit(0)
		}
	}

	// create a new address to listen on
	addr := testutils.NewTestListenAddress()
	testListener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	// start servering it
	testServer.Addr = addr
	go testServer.Serve(testListener)

	// setup a cleanup function
	cleanup = func() {
		if err := testServer.Close(); err != nil {
			panic(err)
		}
	}

	return
}
