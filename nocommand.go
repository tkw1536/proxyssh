package proxyssh

import (
	"io"

	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh/utils"
	"golang.org/x/crypto/ssh/terminal"
)

// NewForwardingSSHServer makes a new Server that has port forwarding enabled, but no shell access.
//
// It uses the provided options, and returns the new server that was created.
// The 'shell' argument from options is ignored.
// It returns the new server that was created.
func NewForwardingSSHServer(logger utils.Logger, opts Options) (server *ssh.Server) {
	server = &ssh.Server{
		Handler: HandleNoCommand(logger),

		Addr:        opts.ListenAddress,
		IdleTimeout: opts.IdleTimeout,
	}
	AllowPortForwarding(logger, server, opts.ForwardAddresses, opts.ReverseAddresses)

	return
}

// HandleNoCommand creates an ssh.Handler that provides a no-op handler for a shell session.
func HandleNoCommand(logger utils.Logger) ssh.Handler {
	return func(session ssh.Session) {
		// logging
		utils.FmtSSHLog(logger, session, "session_start %s", session.User())
		defer utils.FmtSSHLog(logger, session, "session_end")

		// start a exit terminal
		go makeExitTerminal(logger, session)

		// wait for the session to be done
		<-session.Context().Done()
	}
}

// makeExitTerminal reads from the session
func makeExitTerminal(logger utils.Logger, session ssh.Session) {
	utils.FmtSSHLog(logger, session, "terminal_start")
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
			utils.FmtSSHLog(logger, session, "Error reading from terminal: %s", err)
			session.Exit(255)
		}
	}

	session.Exit(0)

}
