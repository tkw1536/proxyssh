// Package terminal provides REPLConfig.
package terminal

import (
	"flag"
	"io"

	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh/logging"
	"golang.org/x/crypto/ssh/terminal"
)

// REPLConfig implements a configuration that does not provide any shell access.
type REPLConfig struct {
}

// Apply applies this configuration to a server.
// This is a no-op.
func (REPLConfig) Apply(logger logging.Logger, sshserver *ssh.Server) error {
	sshserver.Handler = handleNoCommand(logger)
	return nil
}

// RegisterFlags registers flags representing the config to the provided flagset.
// When flagset is nil, uses flag.CommandLine.
func (REPLConfig) RegisterFlags(flagset *flag.FlagSet) {
}

// Handle handles a new server process.
// TODO: this should be implemented in the futture
//func (cfg *EmptyConfiguration) Handle(logger logging.Logger, session ssh.Session) (proxyssh.Process, error) {
//	panic("not yet implemented")
//}

// handleNoCommand creates an ssh.Handler that provides a no-op handler for a shell session.
func handleNoCommand(logger logging.Logger) ssh.Handler {
	return func(session ssh.Session) {
		// logging
		logging.FmtSSHLog(logger, session, "session_start %s", session.User())
		defer logging.FmtSSHLog(logger, session, "session_end")

		// start a exit terminal
		go makeExitTerminal(logger, session)

		// wait for the session to be done
		<-session.Context().Done()
	}
}

// makeExitTerminal reads from the session
func makeExitTerminal(logger logging.Logger, session ssh.Session) {
	logging.FmtSSHLog(logger, session, "terminal_start")
	term := terminal.NewTerminal(session, "")

	// tell the user we don't provide a shell
	io.WriteString(term, "No shell access provided. Use CTRL-C / CTRL-D to close. \n")

loop:
	for {
		_, err := term.ReadLine()
		switch err {
		case nil: /* keep going */
		case io.EOF:
			break loop
		default:
			logging.FmtSSHLog(logger, session, "Error reading from terminal: %s", err)
			session.Exit(255)
		}
	}

	session.Exit(0)

}
