package utils

import (
	"flag"
	"net"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// NetworkAddress is a network address consisting of a hostname and a port
type NetworkAddress struct {
	Hostname string
	Port     Port
}

// Port represents the Port of a NetworkAddress
type Port uint16

// ParseNetworkAddress parses a network address of the form 'Hostname:Port'.
// See function net.Dial, in particular the hostport for allowed combinations.
//
// When parsing fails a non-nil error is returned.
// Otherwise, error is nil.
func ParseNetworkAddress(s string) (p NetworkAddress, err error) {
	var pp string

	// split into hostname and port
	p.Hostname, pp, err = net.SplitHostPort(s)
	if err != nil {
		return
	}

	// parse the port into an int
	port, err := strconv.ParseUint(pp, 10, 32)
	if err != nil {
		err = errors.Wrapf(err, "Unable to parse port: %s", err)
		return
	}
	p.Port = Port(port)

	return p, nil
}

// MustParseNetworkAddress is like ParseNetworkAddress except that it calls panic() instead of returning an error.
//
// This function is intended to be used for test cases.
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
type NetworkAddressListVar struct {
	Addresses *[]NetworkAddress
}

// String turns this NetworkAddressListVar into a comma-seperated list of network addresses.
func (p *NetworkAddressListVar) String() string {
	ports := make([]string, len(*p.Addresses))
	for i, ph := range *p.Addresses {
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
	*p.Addresses = append(*p.Addresses, newPort)
	return
}

func init() {
	// ensure that NetworkAddressListVar fullfills the flag.Value interface
	var _ flag.Value = (*NetworkAddressListVar)(nil)
}
