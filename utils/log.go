// Package utils contains utility functions that are used by dockerproxy and proxyssh.
// These methods are intended to be used by github.com/tkw1536/proxyssh/dockerproxy and github.com/tkw1536/proxyssh only
// and may change without notice.
package utils

import (
	"fmt"
	"log"
	"net"

	"github.com/gliderlabs/ssh"
)

// SSHSessionOrContext represents unifies "github.com/gliderlabs/ssh".Session and "github.com/gliderlabs/ssh".Context.
type SSHSessionOrContext interface {
	User() string
	RemoteAddr() net.Addr
}

// Logger is an interface for "log".Logger
// Logger is assumed to be goroutine-safe.
type Logger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
}

// FmtSSHLog works like "log".Printf except that it takes a Logger and SSHSessionOrContext as parameters.
// The log output will be prefixed with an identifier of the current session or context.
// message and arguments will be passed to "fmt".Sprintf.
func FmtSSHLog(logger Logger, s SSHSessionOrContext, message string, args ...interface{}) {
	prefix := fmt.Sprintf("[%s@%s] ", s.User(), s.RemoteAddr().String())
	actual := fmt.Sprintf(message, args...)
	logger.Print(prefix + actual)
}

func init() {
	// check that ssh.Context and ssh.Session fullfill the SSHLike interface
	var _ SSHSessionOrContext = (ssh.Context)(nil)
	var _ SSHSessionOrContext = (ssh.Session)(nil)

	// check that log.Logger represents an actual logger
	var _ Logger = (*log.Logger)(nil)
}
