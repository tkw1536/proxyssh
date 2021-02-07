package osexec

import (
	"io/ioutil"
	"net"
	"testing"

	"github.com/tkw1536/proxyssh/testutils"
	gossh "golang.org/x/crypto/ssh"
)

func TestPortForwardingForward(t *testing.T) {
	t.Run("forward port forwarding works on an allowed port", func(t *testing.T) {
		// make a new session with port forwarding
		conn, _, err := testutils.NewTestServerSession(
			testServer.Addr,
			gossh.ClientConfig{},
		)
		if err != nil {
			t.Errorf("Unable to create test server session: %s", err)
			t.FailNow()
		}
		defer conn.Close()

		// start a new local listener
		ll, le := net.Listen("tcp", forwardPortsAllow.String())
		if le != nil {
			t.Errorf("Failed to create test server: %s", err)
			t.FailNow()
		}
		go testutils.TCPConstantTestResponse(ll, "success\n")
		defer ll.Close()

		// dial
		cc, err := conn.Dial("tcp", forwardPortsAllow.String())
		if err != nil {
			t.Errorf("Unable to dial forward: %s", err)
			t.FailNow()
		}
		defer cc.Close()

		// read everything
		out, err := ioutil.ReadAll(cc)
		if err != nil {
			t.Errorf("Unable to read from connection: %s", err)
		}

		gotOut := string(out)
		wantOut := "success\n"
		if gotOut != wantOut {
			t.Errorf("Forward() got out = %s, want = %s", gotOut, wantOut)
		}
	})
	t.Run("forward port forwarding does not work on a denied port", func(t *testing.T) {
		// make a new session
		conn, _, err := testutils.NewTestServerSession(
			testServer.Addr,
			gossh.ClientConfig{},
		)
		if err != nil {
			t.Errorf("Unable to create test server session: %s", err)
			t.FailNow()
		}
		defer conn.Close()

		// start a new local server
		ll, le := net.Listen("tcp", forwardPortsDeny.String())
		if le != nil {
			t.Errorf("Failed to create test server: %s", err)
			t.FailNow()
		}
		go testutils.TCPConstantTestResponse(ll, "success\n")
		defer ll.Close()

		// dial
		cc, err := conn.Dial("tcp", forwardPortsDeny.String())
		if err == nil {
			t.Errorf("Unexpectedly able to dial: %s", err)
			cc.Close()
		}
	})
}

func TestPortForwardingReverse(t *testing.T) {
	t.Run("reverse port forwarding works on an allowed port", func(t *testing.T) {
		// make a new session with port forwarding
		conn, _, err := testutils.NewTestServerSession(
			testServer.Addr,
			gossh.ClientConfig{},
		)
		if err != nil {
			t.Errorf("Unable to create test server session: %s", err)
			t.FailNow()
		}
		defer conn.Close()

		// start a server to listen to
		ll, le := conn.Listen("tcp", reversePortsAllow.String())
		if le != nil {
			t.Errorf("Failed to create test server: %s", err)
			t.FailNow()
		}
		go testutils.TCPConstantTestResponse(ll, "success\n")
		defer ll.Close()

		// dial
		cc, err := net.Dial("tcp", reversePortsAllow.String())
		if err != nil {
			t.Errorf("Unable to dial: %s", err)
		}
		defer cc.Close()

		// read everything
		out, err := ioutil.ReadAll(cc)
		if err != nil {
			t.Errorf("Unable to read from connection: %s", err)
		}

		// compare
		gotOut := string(out)
		wantOut := "success\n"
		if gotOut != wantOut {
			t.Errorf("Reverse() got out = %s, want = %s", gotOut, wantOut)
		}
	})
	t.Run("reverse port forwarding does not work on denied port", func(t *testing.T) {
		// make a new session with port forwarding
		conn, _, err := testutils.NewTestServerSession(
			testServer.Addr,
			gossh.ClientConfig{},
		)
		if err != nil {
			t.Errorf("Unable to create test server session: %s", err)
			t.FailNow()
		}
		defer conn.Close()

		// start a server to listen to
		listener, err := conn.Listen("tcp", reversePortsDeny.String())
		if err == nil {
			t.Errorf("Unexpectedly able to listen: %s", err)
			listener.Close()
		}

	})
}
