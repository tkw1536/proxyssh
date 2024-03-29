package dockerexec

import (
	"context"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gliderlabs/ssh"
)

// SSHAuthOptions contain options that configure authentication via ssh
type SSHAuthOptions struct {
	// If set, check if a candidate container contains an ssh key in the provided label
	LabelKey string

	// If set, check if a candidate container contains an authorized_keys file at the provided path(s)
	// Paths may be an array seperated by commas.
	LabelFile string
}

// FindContainerKeys finds the public keys desired by a particular container and returns them
//
// Location of stored credentials is determined by options.
//
// This function will ignore all errors and or invalid values.
func FindContainerKeys(cli client.APIClient, container types.Container, options SSHAuthOptions) (keys []ssh.PublicKey) {

	// Check the key label of a provided container for ssh public keys
	// Note that if LabelKey is "", hasKey will return false because a docker label can not be blank.
	keyString, hasKey := container.Labels[options.LabelKey]
	if hasKey && keyString != "" {
		keys = parseAllKeys([]byte(keyString))
	}

	// Check the filepath label and if it exists and is non-empty
	// we can proceed to try and get each file

	filePath, hasFilePath := container.Labels[options.LabelFile]
	if !hasFilePath || filePath == "" {
		return
	}

	// iterate over all files listed in the label and try to read the file pointed to by each one.
	// If something goes wrong, ignore the error and skip ahead to the next one.
	for _, path := range strings.Split(filePath, ",") {

		content, _, err := cli.CopyFromContainer(context.Background(), container.ID, path)
		if err != nil {
			continue
		}
		defer content.Close()

		bytes, err := io.ReadAll(content)
		if err != nil {
			continue
		}

		keys = append(keys, parseAllKeys(bytes)...)
	}

	return
}

// parseAllKeys parses all keys from the given list of bytes
// ignores all errors
func parseAllKeys(bytes []byte) (keys []ssh.PublicKey) {
	var key ssh.PublicKey
	var err error
	for {
		key, _, _, bytes, err = ssh.ParseAuthorizedKey(bytes)
		if err != nil {
			break
		}
		keys = append(keys, key)
	}
	return
}
