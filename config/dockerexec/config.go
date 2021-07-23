// Package dockerexec provides ContainerExecConfig.
package dockerexec

import (
	"flag"
	"strings"

	"github.com/docker/docker/client"
	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh"
	"github.com/tkw1536/proxyssh/feature"
	"github.com/tkw1536/proxyssh/logging"
)

// ContainerExecConfig implements a proxyssh.Configuration and proxyssh.Handler that execute user processes within running docker containers.
// For this purpose it makes use of 'docker exec'.
//
// The association of incoming user to a docker container happens via the username.
// To find a docker container, the server looks for a docker container where a specific label
// has a value equal to the username.
// If there is no running docker container with the provided label (or there is more than one) the connection will fail.
//
// To authenticate a user, the server uses ssh keys.
// A user is considered authenticated if they can prove the ownership of at least one of the ssh keys associated with this user.
// To find the ssh keys associated to a user, the server uses labels on the associated docker container.
// However in this case, two different labels are checked.
//
// One label can contain an ssh key (in authorized_keys) format.
// The second label may contain comma-seperated file paths.
// These file paths are interpreted relative to the filesystem of the docker container.
// Each file (if it exists) may contain several ssh public keys (in authorized_keys format).
//
// Once a user is authenticated, a session within the associated container will be started.
// For this, a process inside the docker container (called the shell) will be started.
// When no arguments are provided, it will run the shell without any arguments.
// When some arguments are provided by the user, it will run the shell with two arguments, '-c' and a concatination of the arguments provided.
//
// For example, assume the shell is '/bin/sh' and the command provided by the user is 'whoami'.
// Then the server will execute the command '/bin/sh -c whoami' inside the container.
//
// When the ssh user requested a tty, a tty will be allocated within the container.
// When no tty was requested, none will be allocated.
//
// Both the shell and labels to be used can be configured via opts.
type ContainerExecConfig struct {

	// Client is the docker client to be used to the docker daemon.
	Client client.APIClient

	// DockerLabelUser is the label to use for associating a user to a container.
	DockerLabelUser string

	// DockerLabelAuthFile is the label of a container that may contain paths to authorized_keys files.
	DockerLabelAuthFile string

	// DockerLabelKey is the label that may contain an authorized_key for a user.
	DockerLabelKey string

	// ContainerShell is the executable to run within the container.
	ContainerShell string
}

// BUG: We find the unique container associated to a provided server twice.
// This could technically be abused for timing attacks.
// But I am not sure how to avoid this.

// Apply applies this configuration to the server.
func (cfg *ContainerExecConfig) Apply(logger logging.Logger, sshserver *ssh.Server) error {
	sshserver.PublicKeyHandler = feature.AuthorizeKeys(logger, func(ctx ssh.Context) ([]ssh.PublicKey, error) {
		// find the (unique) associated container
		container, err := FindUniqueContainer(cfg.Client, cfg.DockerLabelUser, ctx.User())
		if err != nil {
			return nil, err
		}

		// find the keys associated to this container
		keys := FindContainerKeys(cfg.Client, container, SSHAuthOptions{
			LabelFile: cfg.DockerLabelAuthFile,
		})

		return keys, nil
	})
	return nil
}

// Handle implements the handler
func (cfg *ContainerExecConfig) Handle(logger logging.Logger, session ssh.Session) (proxyssh.Process, error) {
	userCommand := session.Command()

	// determine the command to run inside the docker container
	// when no arguments are given, use the shell.
	// else use shell -c 'arguments'
	command := make([]string, 1, 3)
	command[0] = cfg.ContainerShell
	if len(userCommand) > 0 {
		command = append(command, "-c", strings.Join(userCommand, " "))
	}

	// find the (unique) associated container
	container, err := FindUniqueContainer(cfg.Client, cfg.DockerLabelUser, session.User())
	if err != nil {
		return nil, err
	}

	return NewContainerExecProcess(cfg.Client, container.ID, command), nil
}

// RegisterFlags registers flags representing the config to the provided flagset.
// When flagset is nil, uses flag.CommandLine.
func (cfg *ContainerExecConfig) RegisterFlags(flagset *flag.FlagSet) {
	if flagset == nil {
		flagset = flag.CommandLine
	}

	flagset.StringVar(&cfg.DockerLabelUser, "userlabel", cfg.DockerLabelUser, "Label to find docker files by")
	flagset.StringVar(&cfg.DockerLabelAuthFile, "keylabel", cfg.DockerLabelAuthFile, "Label to find the authorized_keys file by")

	flagset.StringVar(&cfg.ContainerShell, "shell", cfg.ContainerShell, "Shell to execute within the container")
}
