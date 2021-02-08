// Package terminal provides REPLConfig.
package terminal

import (
	"context"
	"flag"
	"io"

	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh"
	"github.com/tkw1536/proxyssh/logging"
)

// REPLConfig implements a configuration that does not provide any shell access.
type REPLConfig struct {
	WelcomeMessage string
	Prompt         string
	Loop           func(ctx context.Context, term io.Writer, input string) (exit bool, code int)
}

// Apply applies this configuration to a server.
// This is a no-op.
func (r *REPLConfig) Apply(logger logging.Logger, sshserver *ssh.Server) error {
	return nil
}

// Handle handles
func (r *REPLConfig) Handle(logger logging.Logger, session ssh.Session) (proxyssh.Process, error) {
	return &REPLProcess{
		WelcomeMessage: r.WelcomeMessage,
		Prompt:         r.Prompt,
		Loop:           r.Loop,
	}, nil
}

// RegisterFlags registers flags representing the config to the provided flagset.
// When flagset is nil, uses flag.CommandLine.
func (r *REPLConfig) RegisterFlags(flagset *flag.FlagSet) {
}
