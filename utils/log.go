package utils

import (
	"fmt"
	"log"
	"net"

	"github.com/gliderlabs/ssh"
)

// SSHLike represents anything that looks like an ssh connection with a user and remote address
type SSHLike interface {
	User() string
	RemoteAddr() net.Addr
}

// LogLike represents anything that looks like a logger
type LogLike interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
}

// FmtSSHLog is like logger.Printf except that it prefixes session information
func FmtSSHLog(logger LogLike, s SSHLike, message string, args ...interface{}) {
	prefix := fmt.Sprintf("[%s@%s] ", s.User(), s.RemoteAddr().String())
	actual := fmt.Sprintf(message, args...)
	logger.Print(prefix + actual)
}

func init() {
	// check that ssh.Context and ssh.Session fullfill the SSHLike interface
	var _ SSHLike = (ssh.Context)(nil)
	var _ SSHLike = (ssh.Session)(nil)

	// check that log.Logger represents an actual logger
	var _ LogLike = (*log.Logger)(nil)
}
