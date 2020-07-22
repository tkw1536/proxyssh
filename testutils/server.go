package testutils

import (
	"net"
	"strconv"
)

// NewTestListenAddress returns a new address for a test server to listen under
// if something goes wrong, calls panic().
func NewTestListenAddress() string {
	// fetch a new unused address from the kernel
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	// check that it is actually open
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	// return it
	return net.JoinHostPort(listener.Addr().(*net.TCPAddr).IP.String(), strconv.Itoa(listener.Addr().(*net.TCPAddr).Port))
}

// AcceptAllWith accepts anything sent to the listener with a constant respone
// if err is not nil, panic()s.
func AcceptAllWith(listener net.Listener, err error, response string) {
	// make a listener
	if err != nil {
		panic(err)
	}

	// respond to everything with 'success'
	go func() {
		for {
			client, _ := listener.Accept()
			if client == nil {
				continue
			}
			client.Write([]byte(response))
			client.Close()
		}
	}()
}
