package docker

import (
	"strings"
	"time"

	"github.com/docker/docker/client"
	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh/server"
	"github.com/tkw1536/proxyssh/server/shell"
	"github.com/tkw1536/proxyssh/utils"
)

// Options are options for the a DockerProxy to be created.
// For a more detailed description of some of these see the NewProxy method.
type Options struct {
	// Client is the docker client to be used to the docker daemon.
	// Note that in addition to this docker client, a 'docker' binary must be available and configured appropriatly.
	Client *client.Client

	// ListenAddress is the Address this Proxy will listen on.
	// It should be of the form "hostname:port".
	ListenAddress string

	// DockerLabelUser is the label to use for associating a user to a container.
	DockerLabelUser string

	// DockerLabelAuthFile is the label of a container that may contain paths to authorized_keys files.
	DockerLabelAuthFile string

	// DockerLabelKey is the label that may contain an authorized_key for a user.
	DockerLabelKey string

	// ContainerShell is the executable to run within the container.
	ContainerShell string

	// DisableAuthentication allows to completly skip the authentication.
	DisableAuthentication bool

	// IdleTimeout is the timeout after which an idle connection is killed.
	IdleTimeout time.Duration

	// ForwardAddresses are addresses that port forwarding is allowed for.
	ForwardAddresses []utils.NetworkAddress
	// ReverseAddresses are addresses that reverse port forwarding is allowed for.
	ReverseAddresses []utils.NetworkAddress
}

// NewProxy makes a new docker ssh server with the given logger and options.
//
// This package expands upon the simple ssh server.
// Instead of running a command on the real system, it associates each user that connects with an existing docker container..
// It then runs a command within this docker container.
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
// When no pty was requested, none will be allocated.
//
// Both the shell and labels to be used can be configured via opts.
//
// This server optionally allows forwarding (and reverse forwarding) from a given list of hostnames and ports.
//
// This function assumed that the 'docker' client binary is available on the host that it is running on.
// This is because of the unreliability of the docker API.
//
// This function calls the logger for every important event.
// Furthermore, it returns a new pre-configured ssh.Server instance.
// The instance may be modified, however it is in the responsibility of the caller to ensure that this does not interfere with the provided functionality.
func NewProxy(logger utils.Logger, opts Options) (sshserver *ssh.Server) {

	// BUG: We find the unique container associated to a provided server twice.
	// This could technically be abused for timing attacks.
	// But I am not sure how to avoid this.

	sshserver = &ssh.Server{
		Handler: shell.HandleCommand(logger, func(s ssh.Session) (command []string, err error) {
			userCommand := s.Command()

			// determine the command to run inside the docker container
			// when no arguments are given, use the shell.
			// else use shell -c 'arguments'
			command = make([]string, 1, 3)
			command[0] = opts.ContainerShell
			if len(userCommand) > 0 {
				command = append(command, "-c", strings.Join(userCommand, " "))
			}

			// find the (unique) associated container
			container, err := FindUniqueContainer(opts.Client, opts.DockerLabelUser, s.User())
			if err != nil {
				return nil, err
			}

			// wrap this inside a 'docker exec' command
			command = Exec(s, container.ID, command, "", "")
			return
		}),
		PublicKeyHandler: server.AuthorizeKeys(logger, func(ctx ssh.Context) ([]ssh.PublicKey, error) {
			// find the (unique) associated container
			container, err := FindUniqueContainer(opts.Client, opts.DockerLabelUser, ctx.User())
			if err != nil {
				return nil, err
			}

			// find the keys associated to this container
			keys := FindContainerKeys(opts.Client, container, DockerSSHAuthOptions{
				LabelFile: opts.DockerLabelAuthFile,
			})

			return keys, nil
		}),

		Addr:        opts.ListenAddress,
		IdleTimeout: opts.IdleTimeout,
	}

	// if the explicitly requested to turn off authentication, do it
	if opts.DisableAuthentication {
		logger.Print("WARNING: Disabling authentication. Anyone will be able to connect. ")
		sshserver.PublicKeyHandler = nil
	}

	// setup forwarding
	server.AllowPortForwarding(logger, sshserver, opts.ForwardAddresses, opts.ReverseAddresses)

	return
}
