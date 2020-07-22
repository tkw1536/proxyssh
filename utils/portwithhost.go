package utils

import (
	"flag"
	"net"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// PortWithHost represents a port in combination with a hostname
type PortWithHost struct {
	Host string
	Port uint32
}

// ParsePortWithHost parses a port with a host from a string
func ParsePortWithHost(s string) (p PortWithHost, err error) {
	var pp string

	// host
	p.Host, pp, err = net.SplitHostPort(s)
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

// MustParsePortWithHost is like ParsePort execpt that it panics on error
func MustParsePortWithHost(s string) PortWithHost {
	p, err := ParsePortWithHost(s)
	if err != nil {
		panic(err)
	}
	return p
}

func (p PortWithHost) String() string {
	return net.JoinHostPort(p.Host, strconv.FormatInt(int64(p.Port), 10))
}

// PortListVar represents a list of ports with hosts
type PortListVar []PortWithHost

func (p *PortListVar) String() string {
	ports := make([]string, len(*p))
	for i, ph := range *p {
		ports[i] = ph.String()
	}

	return strings.Join(ports, ",")
}

// Set sets the value of this PortVar
func (p *PortListVar) Set(value string) (err error) {
	newPort, err := ParsePortWithHost(value)
	if err != nil {
		return err
	}
	*p = append(*p, newPort)
	return
}

func init() {
	// ensure that PortVar fullfills the flag.Value interface
	var _ flag.Value = (*PortListVar)(nil)
}
