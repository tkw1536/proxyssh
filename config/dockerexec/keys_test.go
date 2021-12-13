//go:build dockertest
// +build dockertest

package dockerexec

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh/internal/integrationtest"
	"github.com/tkw1536/proxyssh/internal/testutils"
)

func TestFindContainerKeys(t *testing.T) {
	integrationtest.RunComposeTest(findAuthContainerCompose, map[string]string{
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
