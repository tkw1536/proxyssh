// Package osexec provides SystemExecConfig.
package osexec

import (
	"flag"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh"
	"github.com/tkw1536/proxyssh/internal/logging"
)

// SystemExecConfig implements a proxyssh.Configuration and proxyssh.Handler that execute user processes using the real system.
type SystemExecConfig struct {
	// Shell is the shell to use for the server
	// This is called like `shell -c "command"`` when an ssh command is provided or like `shell` when not.
	// The shell is passed to exec.LookPath().
	Shell string
}

// Apply applies this configuration to the server.
// This is a no-op.
func (cfg *SystemExecConfig) Apply(logger logging.Logger, sshserver *ssh.Server) error {
	return nil
}

// Handle handles a new configuration thingy
func (cfg *SystemExecConfig) Handle(logger logging.Logger, session ssh.Session) (proxyssh.Process, error) {
	userCommand := session.Command()

	// determine the arguments to pass to the shell.
	var args []string
	if len(userCommand) > 0 {
		args = []string{"-c", strings.Join(userCommand, " ")}
	}

	// create a new system process
	return NewSystemProcess(cfg.Shell, args), nil
}

// RegisterFlags registers flags representing the config to the provided flagset.
// When flagset is nil, uses flag.CommandLine.
func (cfg *SystemExecConfig) RegisterFlags(flagset *flag.FlagSet) {
	if flagset == nil {
		flagset = flag.CommandLine
	}

	flag.StringVar(&cfg.Shell, "shell", cfg.Shell, "Shell to use")
}
