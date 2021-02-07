package dockerexec

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh/internal/testutils"
)

func TestFindUniqueContainer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping docker-compose test in short mode")
	}

	testutils.RunComposeTest(findUniqueContainerCompose, nil, func(cli client.APIClient, findService func(name string) types.Container, stopService func(name string)) error {
		t.Run("find a single container", func(t *testing.T) {
			container, err := FindUniqueContainer(cli, "de.tkw01536.test", "a")
			if err != nil {
				t.Errorf("FindUniqueContainer(): got err = %s, want err = nil", err.Error())
				return
			}

			if !testutils.SliceContainsString(container.Names, "/samplea") {
				t.Error("FindUniqueContainer(): did not find container 'samplea'")
			}
		})

		t.Run("do not find a stopped container", func(t *testing.T) {
			stopService("samplea")
			_, err := FindUniqueContainer(cli, "de.tkw01536.test", "a")
			if err == nil {
				t.Error("FindUniqueContainer(): got err = nil, want err != nil")
				return
			}
			return
		})

		t.Run("do not find non-existent container", func(t *testing.T) {
			_, err := FindUniqueContainer(cli, "de.tkw01536.test", "c")
			if err == nil {
				t.Error("FindUniqueContainer(): got err = nil, want err != nil")
				return
			}
			return
		})

		t.Run("do not find a container with multiple matches", func(t *testing.T) {
			_, err := FindUniqueContainer(cli, "de.tkw01536.test", "b")
			if err == nil {
				t.Error("FindUniqueContainer(): got err = nil, want err != nil")
				return
			}
			return
		})

		t.Run("find a single running container", func(t *testing.T) {
			stopService("sampleb2")
			container, err := FindUniqueContainer(cli, "de.tkw01536.test", "b")
			if err != nil {
				t.Errorf("FindUniqueContainer(): got err = %s, want err = nil", err.Error())
				return
			}

			if !testutils.SliceContainsString(container.Names, "/sampleb1") {
				t.Error("FindUniqueContainer(): did not find container 'sampleb1'")
			}
		})

		return nil
	})
}

// this compose file contains three containers:
//
// - one with label de.tkw01536.test=a
// - two with labels de.tkw01536.test=b
var findUniqueContainerCompose = `
version: '2'

services:
	samplea:
		container_name: samplea
		labels:
			de.tkw01536.test: "a"
		image: alpine
		command: sh -c 'while sleep 3600; do :; done'
	sampleb1:
		container_name: sampleb1
		labels:
			de.tkw01536.test: "b"
		image: alpine
		command: sh -c 'while sleep 3600; do :; done'
	sampleb2:
		container_name: sampleb2
		labels:
			de.tkw01536.test: "b"
		image: alpine
		command: sh -c 'while sleep 3600; do :; done'
`

func TestFindContainerKeys(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping docker-compose test in short mode")
	}

	testutils.RunComposeTest(findAuthContainerCompose, map[string]string{
		"authorized_key_a": testutils.AuthorizedKeysString(testPublicKeyA) + "\n",
		"authorized_key_b": testutils.AuthorizedKeysString(testPublicKeyB) + "\n",
	}, func(cli client.APIClient, findService func(name string) types.Container, stopService func(name string)) error {
		// tests for finding keys in environment variable

		t.Run("container with unset LabelKey returns no keys", func(t *testing.T) {
			keys := FindContainerKeys(cli, findService("nolabels"), SSHAuthOptions{LabelKey: "de.tkw01536.test.key"})
			gotLen := len(keys)
			if gotLen != 0 {
				t.Errorf("FindContainerKeys(): got len(keys) = %d, want len(keys) = 0", gotLen)
			}
		})

		t.Run("container with empty LabelKey returns no keys", func(t *testing.T) {
			keys := FindContainerKeys(cli, findService("keylabel_empty"), SSHAuthOptions{LabelKey: "de.tkw01536.test.key"})
			gotLen := len(keys)
			if gotLen != 0 {
				t.Errorf("FindContainerKeys(): got len(keys) = %d, want len(keys) = 0", gotLen)
			}
		})

		t.Run("container with invalid LabelKey returns no keys", func(t *testing.T) {
			keys := FindContainerKeys(cli, findService("keylabel_invalid"), SSHAuthOptions{LabelKey: "de.tkw01536.test.key"})
			gotLen := len(keys)
			if gotLen != 0 {
				t.Errorf("FindContainerKeys(): got len(keys) = %d, want len(keys) = 0", gotLen)
			}
		})

		t.Run("container with valid label returns single key", func(t *testing.T) {
			keys := FindContainerKeys(cli, findService("keylabel_valid"), SSHAuthOptions{LabelKey: "de.tkw01536.test.key"})
			gotLen := len(keys)
			if gotLen != 1 {
				t.Errorf("FindContainerKeys(): got len(keys) = %d, want len(keys) = 1", gotLen)
				return
			}
			if !ssh.KeysEqual(keys[0], testPublicKeyA) {
				t.Error("FindContainerKeys(): missing test public key")
			}
		})

		// tests for finding key in files

		t.Run("container without labels returns no keys", func(t *testing.T) {
			keys := FindContainerKeys(cli, findService("nolabels"), SSHAuthOptions{LabelFile: "de.tkw01536.test.file"})
			gotLen := len(keys)
			if gotLen != 0 {
				t.Errorf("FindContainerKeys(): got len(keys) = %d, want len(keys) = 0", gotLen)
			}
		})

		t.Run("container with non-existent file returns no keys", func(t *testing.T) {
			keys := FindContainerKeys(cli, findService("filelabel_invalid"), SSHAuthOptions{LabelFile: "de.tkw01536.test.file"})
			gotLen := len(keys)
			if gotLen != 0 {
				t.Errorf("FindContainerKeys(): got len(keys) = %d, want len(keys) = 0", gotLen)
			}
		})

		t.Run("container with one valid file label returns single key", func(t *testing.T) {
			keys := FindContainerKeys(cli, findService("filelabel_valid_one"), SSHAuthOptions{LabelFile: "de.tkw01536.test.file"})
			gotLen := len(keys)
			if gotLen != 1 {
				t.Errorf("FindContainerKeys(): got len(keys) = %d, want len(keys) = 1", gotLen)
				return
			}
			if !ssh.KeysEqual(keys[0], testPublicKeyA) {
				t.Error("FindContainerKeys(): missing test public key")
			}
		})

		t.Run("container with two files but only one valid one returns single key", func(t *testing.T) {
			keys := FindContainerKeys(cli, findService("filelabel_two_one_valid"), SSHAuthOptions{LabelFile: "de.tkw01536.test.file"})
			gotLen := len(keys)
			if gotLen != 1 {
				t.Errorf("FindContainerKeys(): got len(keys) = %d, want len(keys) = 1", gotLen)
				return
			}
			if !ssh.KeysEqual(keys[0], testPublicKeyA) {
				t.Error("FindContainerKeys(): missing test public key")
			}
		})

		t.Run("container with two valid files returns both keys", func(t *testing.T) {
			keys := FindContainerKeys(cli, findService("filelabel_two_two_valid"), SSHAuthOptions{LabelFile: "de.tkw01536.test.file"})
			gotLen := len(keys)
			if gotLen != 2 {
				t.Errorf("FindContainerKeys(): got len(keys) = %d, want len(keys) = 2", gotLen)
				return
			}
			if !ssh.KeysEqual(keys[0], testPublicKeyA) {
				t.Error("FindContainerKeys(): missing test public key A")
			}
			if !ssh.KeysEqual(keys[1], testPublicKeyB) {
				t.Error("FindContainerKeys(): missing test public key B")
			}
		})

		return nil
	})

}

var _, testPublicKeyA = testutils.GenerateRSATestKeyPair()
var _, testPublicKeyB = testutils.GenerateRSATestKeyPair()

var findAuthContainerCompose = `
version: '2'

services:
	nolabels:
		image: alpine
		command: sh -c 'while sleep 3600; do :; done'

	keylabel_empty:
		labels:
			de.tkw01536.test.key: ""
		image: alpine
		command: sh -c 'while sleep 3600; do :; done'
	keylabel_invalid:
		labels:
			de.tkw01536.test.key: "junk"
		image: alpine
		command: sh -c 'while sleep 3600; do :; done'
	keylabel_valid:
		labels:
			de.tkw01536.test.key: "` + testutils.AuthorizedKeysString(testPublicKeyA) + `"
		image: alpine
		command: sh -c 'while sleep 3600; do :; done'

	filelabel_invalid:
		labels:
			de.tkw01536.test.file: "/authorized_key_a"
		image: alpine
		command: sh -c 'while sleep 3600; do :; done'

	filelabel_valid_one:
		labels:
			de.tkw01536.test.file: "/authorized_key_a"
		volumes:
			- "./authorized_key_a:/authorized_key_a"
		image: alpine
		command: sh -c 'while sleep 3600; do :; done'
	
	filelabel_two_one_valid:
		labels:
			de.tkw01536.test.file: "/authorized_key_a,/authorized_key_b"
		volumes:
			- "./authorized_key_a:/authorized_key_a"
		image: alpine
		command: sh -c 'while sleep 3600; do :; done'
	filelabel_two_two_valid:
		labels:
			de.tkw01536.test.file: "/authorized_key_a,/authorized_key_b"
		volumes:
			- "./authorized_key_a:/authorized_key_a"
			- "./authorized_key_b:/authorized_key_b"
		image: alpine
		command: sh -c 'while sleep 3600; do :; done'
`
