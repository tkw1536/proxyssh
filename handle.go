package proxyssh

import (
	"io"

	"github.com/gliderlabs/ssh"
	"github.com/pkg/errors"
	"github.com/tkw1536/proxyssh/utils"
)

// HandleProcess does whatever
func HandleProcess(logger utils.Logger, makeProcess func(session ssh.Session) (Process, error)) ssh.Handler {
	return func(session ssh.Session) {
		// logging
		utils.FmtSSHLog(logger, session, "session_start %s", session.User())
		defer utils.FmtSSHLog(logger, session, "session_end")

		process, err := makeProcess(session)
		if err != nil {
			abortsession(logger, session, errors.Wrap(err, "Failed to create process"))
			return
		}

		// TODO: Have (command, args) returned
		sshcmd := &Session{
			Session: session,
			Logger:  logger,
			Process: process,
		}
		if err != nil {
			abortsession(logger, session, errors.Wrap(err, "Failed to create ssh command"))
			return
		}

		utils.FmtSSHLog(logger, session, "session_valid %s", process)
		sshcmd.Run()
	}
}

// abortsession exits an SSH session with code 255.
// It also prints the error message to the user on STDERR.
func abortsession(logger utils.Logger, s ssh.Session, err error) {
	errmsg := err.Error()
	utils.FmtSSHLog(logger, s, "session_command %s", errmsg)
	io.WriteString(s.Stderr(), errmsg+"\n")
	s.Exit(255)
}
