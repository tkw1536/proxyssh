package proxyssh

import (
	"io"

	"github.com/gliderlabs/ssh"
	"github.com/pkg/errors"
	"github.com/tkw1536/proxyssh/logging"
)

// Handler represents a Configuration that provides a handler.
type Handler interface {
	Handle(logger logging.Logger, session ssh.Session) (Process, error)
}

// ErrHandlerAlreadySet is returned by ApplyHandler when a handler is already applies to the server.
var ErrHandlerAlreadySet = errors.New("ApplyHandler: Handler already set")

// ApplyHandler applies a handler to a server by setting server.Handler.
// When server.Handler is already set, returns an error.
func ApplyHandler(logger logging.Logger, server *ssh.Server, handler Handler) error {

	if server.Handler != nil {
		return ErrHandlerAlreadySet
	}

	server.Handler = makeProcessHandler(logger, handler)
	return nil
}

// makeProcessHandler creates a new ssh.Handler that implements handler.
func makeProcessHandler(logger logging.Logger, handler Handler) ssh.Handler {
	return func(session ssh.Session) {
		// logging
		logging.FmtSSHLog(logger, session, "session_start %s", session.User())
		defer logging.FmtSSHLog(logger, session, "session_end")

		// handle the provided session
		process, err := handler.Handle(logger, session)
		if err != nil {
			abortsession(logger, session, errors.Wrap(err, "Failed to create process"))
			return
		}

		sshcmd := &Session{
			Session: session,
			Logger:  logger,
			Process: process,
		}
		if err != nil {
			abortsession(logger, session, errors.Wrap(err, "Failed to create ssh command"))
			return
		}

		logging.FmtSSHLog(logger, session, "session_valid %s", process)
		sshcmd.Run()
	}
}

// abortsession exits an SSH session with code 255.
// It also prints the error message to the user on STDERR.
func abortsession(logger logging.Logger, s ssh.Session, err error) {
	errmsg := err.Error()
	logging.FmtSSHLog(logger, s, "session_command %s", errmsg)
	io.WriteString(s.Stderr(), errmsg+"\n")
	s.Exit(255)
}
