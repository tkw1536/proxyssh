// Package logging provides Logger.
package logging

import (
	"fmt"
	"log"
	"net"

	"github.com/gliderlabs/ssh"
)

// LogSessionOrContext represents either an ssh.Session or an ssh.Context.
// It is used by logging functions that accept both.
//
// See also the github.com/gliderlabs/ssh.
type LogSessionOrContext interface {
	User() string
	RemoteAddr() net.Addr
}

func init() {
	// both ssh.Context and ssh.Session implement LogSessionOrContext
	var _ LogSessionOrContext = (ssh.Context)(nil)
	var _ LogSessionOrContext = (ssh.Session)(nil)
}

// Logger represents an object that can be used for log messages.
// It is assumed to be goroutine-safe.
//
// The log.Logger type of the builtin log package implements this type.
type Logger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
}

// log.Logger fullfills Logger
var _ Logger = (*log.Logger)(nil)

// FmtSSHLog formats a log message, prefixes it with information about s, and then prints it to Logger.
//
// This message behaves like log.Printf except that it takes a Logger and LogSessionOrContext as arguments.
func FmtSSHLog(logger Logger, s LogSessionOrContext, message string, args ...interface{}) {
	prefix := fmt.Sprintf("[%s@%s] ", s.User(), s.RemoteAddr().String())
	actual := fmt.Sprintf(message, args...)
	logger.Print(prefix + actual)
}
