package testutils

import (
	"net"
	"strconv"
)

// NewTestListenAddress returns a new unused address that can be used to start a listener on.
// The address returned will be off the form 'host:port' and compatible with net.Dial() and friends.
//
// The address is guaranteed to be on a port higher than 1024.
// Furthermore, it is guaranteed not be used by any other server, and listens only on the loopback interface.
// It will typically be something like "127.0.0.1:12345", but is not guaranteed to be of this form.
//
// If no address is available, or something unexpected happens, panic() is called.
func NewTestListenAddress() string {
	// fetch a new unused address from the kernel
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}

	// open a listener, and get the actual address
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	address := listener.Addr().(*net.TCPAddr)
	defer listener.Close()

	// join it back into a string
	return net.JoinHostPort(address.IP.String(), strconv.Itoa(address.Port))
}

// TCPConstantTestResponse starts accepting connections on the listener.
// It then sends a constant response and closes the accepted connection.
//
// This function performs blocking work on the goroutine it was called on.
// As such, it should typically be called like:
//
//  listener, err := net.Listen("tcp", address)
//  go TCPConstantTestResponse(listener, response)
//  defer listener.Close()
func TCPConstantTestResponse(listener net.Listener, response string) {
	// respond to everything with a constant response
	responseBytes := []byte(response)
	for {
		client, _ := listener.Accept()
		if client == nil {
			continue
		}
		client.Write(responseBytes)
		client.Close()
	}
}
