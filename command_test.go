package proxyssh

import (
	"testing"
	"time"

	"github.com/tkw1536/proxyssh/testutils"
	gossh "golang.org/x/crypto/ssh"
)

func TestCommand(t *testing.T) {

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
			gotOut, gotErr, gotCode, err := testutils.RunTestServerCommand(testServer.Addr, gossh.ClientConfig{}, tt.command, tt.stdin)
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
}

func TestCommandIsKilled(t *testing.T) {
	waitKillTimeout := 1 * time.Second
	t.Run("closing the network connection kills the process", func(t *testing.T) {
		// run a new session
		client, session, err := testutils.NewTestServerSession(testServer.Addr, gossh.ClientConfig{})
		defer client.Close()
		if err != nil {
			t.Errorf("Unable to create test server session: %s", err)
			t.FailNow()
		}

		// process
		process, err := testutils.GetTestSessionProcess(session)
		if err != nil {
			t.Errorf("Unable to get test session pid: %s", err)
			t.FailNow()
		}

		// close the client
		client.Close()
		time.Sleep(waitKillTimeout)

		// check if the process ist still alive
		if testutils.TestProcessAlive(process) {
			t.Errorf("TestCommandKilled(): Process still alive")
		}
	})

	t.Run("closing the session kills the process", func(t *testing.T) {
		// run a new session
		client, session, err := testutils.NewTestServerSession(testServer.Addr, gossh.ClientConfig{})
		defer client.Close()
		if err != nil {
			t.Errorf("Unable to create test server session: %s", err)
			t.FailNow()
		}

		// process
		process, err := testutils.GetTestSessionProcess(session)
		if err != nil {
			t.Errorf("Unable to get test session pid: %s", err)
			t.FailNow()
		}

		// close the session
		session.Close()
		time.Sleep(waitKillTimeout)

		// check if the process ist still alive
		if testutils.TestProcessAlive(process) {
			t.Errorf("TestCommandKilled(): Process still alive")
		}
	})
}
