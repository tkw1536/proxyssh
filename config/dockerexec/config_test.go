package dockerexec

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/tkw1536/proxyssh/internal/integrationtest"
	"github.com/tkw1536/proxyssh/internal/testutils"

	gossh "golang.org/x/crypto/ssh"
)

var execTestConfig = &ContainerExecConfig{
	DockerLabelUser:     "de.tkw01536.test.user",
	DockerLabelAuthFile: "de.tkw01536.test.file",

	ContainerShell: "/bin/sh",
}

func TestConnectContainer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping docker-compose test in short mode")
	}

	integrationtest.RunComposeTest(configComposeTest, map[string]string{
		"authorized_key": testutils.AuthorizedKeysString(testPublicKeyConfig) + "\n",
	}, func(cli client.APIClient, findService func(name string) types.Container, stopService func(name string)) error {
		// create a new server with this client
		execTestConfig.Client = cli
		testServer, _, cleanup := integrationtest.NewServer(nil, execTestConfig)
		defer cleanup()

		tests := []struct {
			name    string
			command string
			stdin   string

			wantOut  string
			wantErr  string
			wantCode int
		}{
			{
				name:     "echo on stdout",
				command:  "echo 'Hello world'",
				stdin:    "",
				wantOut:  "Hello world\n",
				wantErr:  "",
				wantCode: 0,
			},

			{
				name:     "echo on stderr",
				command:  "echo 'Hello world' 1>&2",
				stdin:    "",
				wantOut:  "",
				wantErr:  "Hello world\n",
				wantCode: 0,
			},

			{
				name:     "echo on both",
				command:  "echo 'stderr' 1>&2 && echo 'stdout'",
				stdin:    "",
				wantOut:  "stdout\n",
				wantErr:  "stderr\n",
				wantCode: 0,
			},

			{
				name:     "exit code != 0",
				command:  "false",
				stdin:    "",
				wantOut:  "",
				wantErr:  "",
				wantCode: 1,
			},

			{
				name:     "send stdin to stdout",
				command:  "cat",
				stdin:    "Hello world",
				wantOut:  "Hello world",
				wantErr:  "",
				wantCode: 0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				gotOut, gotErr, gotCode, err := testutils.RunTestServerCommand(testServer.Addr, gossh.ClientConfig{
					Auth: []gossh.AuthMethod{
						gossh.PublicKeys(testPrivateKeyConfig),
					},
				}, tt.command, tt.stdin)
				if err != nil {
					t.Errorf("Unable to create test server session: %s", err)
					t.FailNow()
				}

				if gotOut != tt.wantOut {
					t.Errorf("Command() got out = %s, want = %s", gotOut, tt.wantOut)
				}
				if gotErr != tt.wantErr {
					t.Errorf("Command() got err = %s, want = %s", gotErr, tt.wantErr)
				}
				if gotCode != tt.wantCode {
					t.Errorf("Command() got err = %s, want = %s", gotErr, tt.wantErr)
				}
			})
		}

		integrationtest.AssertLeakDetector(t, len(tests))

		return nil
	})
}

var testPrivateKeyConfig, testPublicKeyConfig = testutils.GenerateRSATestKeyPair()

// this compose file contains one container:
var configComposeTest = `
version: '2'

services:
	runner:
		labels:
			de.tkw01536.test.file: "/authorized_key"
			de.tkw01536.test.user: "user"
		volumes:
			- "./authorized_key:/authorized_key"
		image: alpine
		command: sh -c 'while sleep 3600; do :; done'
`
