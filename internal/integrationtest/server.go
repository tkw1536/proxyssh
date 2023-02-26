package integrationtest

import (
	"net"
	"testing"

	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh"
	"github.com/tkw1536/proxyssh/internal/testutils"
	"github.com/tkw1536/proxyssh/logging"
)

// NewServer creates a new proxyssh.Server for an integration test.
// Returns the test server and a cleanup function to be defered.
//
// When options is nil, sets up empty options.
//
// When testServer.Handler remains unset, sets it to respond with "success" on stdout and exit immediatly.
// When testServer.HostSigners remains unset, generates a single reusable test server key.
//
// This function should be used like:
//
//	testServer, cleanup := NewServer(testServer)
//	defer cleanup()
//
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

	// reset the leak detector stats
	logging.ResetGlobalLeakDetectorStats()

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

// AssertLeakDetector creates a new subtest that ensures that the leak detector succeeded with exactly 'want' subprocesses.
// When the leak detector is not enabled, the subtest is skipped.
//
// This function is untested.
func AssertLeakDetector(t *testing.T, want int) {
	t.Run("Memory Leak Detector", func(t *testing.T) {
		if !logging.MemoryLeakEnabled {
			t.Skip("MemoryLeakEnabled = false, use the 'leak' tag to enable")
		}

		success, failure := logging.GetGlobalLeakDetectorStats()
		got := int(success)

		if got != want {
			t.Errorf("Leak Detector: wanted %d finished calls, received %d (and %d failures)", want, got, failure)
		}
	})
}
