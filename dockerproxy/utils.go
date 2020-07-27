package dockerproxy

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/gliderlabs/ssh"
	"github.com/pkg/errors"
)

// DockerExec wraps a command into a 'docker exec'
// this depends on the 'docker' executable being available.
func DockerExec(s ssh.Session, containerID string, command []string, workdir string, user string) (exec []string) {
	exec = []string{"docker", "exec", "--interactive"}

	// ensure it's a tty when we asked for one
	if _, _, isPty := s.Pty(); isPty {
		exec = append(exec, "--tty")
	}

	// append the workdir
	if workdir != "" {
		exec = append(exec, "--workdir", workdir)
	}

	// append the user
	if user != "" {
		exec = append(exec, "--user", user)
	}

	// append the container id and command
	exec = append(exec, containerID)
	exec = append(exec, command...)
	return
}

// FindUniqueContainer finds a unique running container with the given label key and value
// If there is no unique container fullfilling this condition, returns an error.
func FindUniqueContainer(cli *client.Client, key string, value string) (types.Container, error) {
	// Setup a filter for a running container with the given key/value label
	Filters := filters.NewArgs()
	Filters.Add("label", fmt.Sprintf("%s=%s", key, value))
	Filters.Add("status", "running")

	// do the list
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: Filters,
	})
	if err != nil {
		return types.Container{}, errors.Wrap(err, "Unable to list containers")
	}

	// make sure there is *exactly* one
	if len(containers) != 1 {
		return types.Container{}, errors.New("Found more than a single container")
	}

	return containers[0], nil
}

// DockerSSHAuthOptions represents options for docker ssh auth
type DockerSSHAuthOptions struct {
	// label that may contain a key
	LabelKey string

	// label that may contain the path to a file within the container
	LabelKeypath string
}

// FindContainerKeys finds the keys in a certain docker container
// silences all errors, and will return an empty slice instead.
func FindContainerKeys(cli *client.Client, container types.Container, opts DockerSSHAuthOptions) (keys []ssh.PublicKey) {
	// check if we have a label
	if opts.LabelKey != "" {
		keyString, hasKey := container.Labels[opts.LabelKey]
		if hasKey {
			keys = parseAllKeys([]byte(keyString))
		}
	}

	// if we have a keypath label we need to do more work
	if opts.LabelKeypath == "" {
		return
	}

	// read from filepath
	filePath, hasFilePath := container.Labels[opts.LabelKeypath]
	if !hasFilePath {
		return
	}

	// if there isn't a filepath, done
	if len(filePath) == 0 {
		return
	}

	filePathAry := strings.Split(filePath, ";")
	for _, path := range filePathAry {
		// grab the bytes, and ignore errors
		// so that non-existent paths do not fail
		content, _, err := cli.CopyFromContainer(context.Background(), container.ID, path)
		if err != nil {
			continue
		}
		defer content.Close()

		// read all the keys from the content
		bytes, err := ioutil.ReadAll(content)
		if err != nil {
			continue
		}
		keys = append(keys, parseAllKeys(bytes)...)
	}

	return
}

// parseAllKeys parses all keys from the given list of bytes
func parseAllKeys(bytes []byte) (keys []ssh.PublicKey) {
	var key ssh.PublicKey
	var err error
	for true {
		key, _, _, bytes, err = ssh.ParseAuthorizedKey(bytes)
		if err != nil {
			break
		}
		keys = append(keys, key)
	}
	return
}
