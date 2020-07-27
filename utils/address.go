package utils

import (
	"flag"
	"net"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// NetworkAddress represents a network address consisting of a hostname and a port.
type NetworkAddress struct {
	Hostname string
	Port     uint32
}

// ParseNetworkAddress parses a network address of the form 'host:port'.
// See func "net".Dial, in particular the hostport for allowed combinations.
func ParseNetworkAddress(s string) (p NetworkAddress, err error) {
	var pp string

	// host
	p.Hostname, pp, err = net.SplitHostPort(s)
	if err != nil {
		return
	}

	// port
	port, err := strconv.ParseUint(pp, 10, 32)
	if err != nil {
		err = errors.Errorf("Unable to parse port")
		return p, err
	}
	p.Port = uint32(port)

	return p, nil
}

// MustParseNetworkAddress is like ParseNetworkAddress except that it calls panic() instead of returning an error.
// This func is intended for test cases.
func MustParseNetworkAddress(s string) NetworkAddress {
	p, err := ParseNetworkAddress(s)
	if err != nil {
		panic(err)
	}
	return p
}

// String turns this NetworkAddress into a string of the form "host:port".
// This string is guaranteed to be parsable by ParseNetworkAddress() as well as "net".Dial and friends.
func (p NetworkAddress) String() string {
	return net.JoinHostPort(p.Hostname, strconv.FormatInt(int64(p.Port), 10))
}

// NetworkAddressListVar represents a "flag".Value that contains a list of network addresses.
// It can be passed multiple times, and collects all NetworkAddress in an ordered list.
type NetworkAddressListVar []NetworkAddress

// String turns this NetworkAddressListVar into a comma-seperated list of network addresses.
func (p *NetworkAddressListVar) String() string {
	ports := make([]string, len(*p))
	for i, ph := range *p {
		ports[i] = ph.String()
	}

	return strings.Join(ports, ",")
}

// Set sets the value of this NetworkAddressListVar
// This function is intedned to be called by flag.Var()
func (p *NetworkAddressListVar) Set(value string) (err error) {
	newPort, err := ParseNetworkAddress(value)
	if err != nil {
		return err
	}
	*p = append(*p, newPort)
	return
}

func init() {
	// ensure that NetworkAddressListVar fullfills the flag.Value interface
	var _ flag.Value = (*NetworkAddressListVar)(nil)
}
