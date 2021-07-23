// +build dockertest

package dockerexec

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/tkw1536/proxyssh/internal/integrationtest"
	"github.com/tkw1536/proxyssh/internal/testutils"
)

func TestFindUniqueContainer(t *testing.T) {
	integrationtest.RunComposeTest(findUniqueContainerCompose, nil, func(cli client.APIClient, findService func(name string) types.Container, stopService func(name string)) error {
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
