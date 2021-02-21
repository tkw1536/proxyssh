package testutils

import (
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
)

func TestNewTestListenAddress(t *testing.T) {
	N := 1000
	for i := 0; i < N; i++ {
		testAddress := NewTestListenAddress()
		if !strings.HasPrefix(testAddress, "127.0.0.1:") {
			t.Error("NewTestListenAddress() returned invalid address")
			t.FailNow()
		}

		// parse a port
		port := strings.Split(testAddress, ":")[1]
		thePort, err := strconv.ParseUint(port, 10, 32)
		if err != nil {
			t.Error(err)
		}

		// check that it's > 1024
		if thePort < 1024 {
			t.Error("NewTestListenAddress() returned port < 1024")
		}
	}

	return
}

func TestTCPConstantTestResponse(t *testing.T) {
	testResponse := "hello world"

	// start listening
	address := NewTestListenAddress()
	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}
	go TCPConstantTestResponse(listener, testResponse)
	defer listener.Close()

	// check several times that the response is always the same
	// by reading everything and checking that it's what was set above
	N := 1000
	for i := 0; i < N; i++ {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			panic(err)
		}

		gotBytes, err := io.ReadAll(conn)
		if err != nil {
			t.Error(err)
		}

		conn.Close()

		gotString := string(gotBytes)
		if gotString != testResponse {
			t.Errorf("NewTestListenAddress() responded = %s, want = %s", gotString, testResponse)
		}
	}
}
