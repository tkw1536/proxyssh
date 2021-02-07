package osexec

import (
	"testing"

	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh/feature"
	"github.com/tkw1536/proxyssh/internal/testutils"
	gossh "golang.org/x/crypto/ssh"
)

var (
	allowedPrivateKeyA, allowedPublicKeyA = testutils.GenerateRSATestKeyPair()
	allowedPrivateKeyB, allowedPublicKeyB = testutils.GenerateRSATestKeyPair()
	deniedPrivateKey, deniedPublicKeyB    = testutils.GenerateRSATestKeyPair()
)

func TestAuthorizeKeys(t *testing.T) {
	testServer.PublicKeyHandler = feature.AuthorizeKeys(testutils.GetTestLogger(), func(ctx ssh.Context) (keys []ssh.PublicKey, err error) {
		switch ctx.User() {
		// user1 has keys allowedPublicKeyA dn allowedPublicKeyB
		case "user1":
			keys = []ssh.PublicKey{allowedPublicKeyA, allowedPublicKeyB}

		// user2 has keys allowedPublicKeyB
		case "user2":
			keys = []ssh.PublicKey{allowedPublicKeyA}
		}
		return
	})
	defer func() {
		testServer.PublicKeyHandler = nil
	}()

	var tests = []struct {
		name      string
		username  string
		clientKey gossh.Signer
		wantOK    bool
	}{
		{
			"user1 with first authorized key gets access",
			"user1",
			allowedPrivateKeyA,
			true,
		},
		{
			"user1 with second authorized key gets access",
			"user1",
			allowedPrivateKeyB,
			true,
		},
		{
			"user1 with non-authorized key gets denied access",
			"user1",
			deniedPrivateKey,
			false,
		},

		{
			"user2 with authorized key gets access",
			"user2",
			allowedPrivateKeyA,
			true,
		},
		{
			"user2 with first non-authorized key gets denied access",
			"user2",
			allowedPrivateKeyB,
			false,
		},
		{
			"user2 with first non-authorized key gets denied access",
			"user2",
			deniedPrivateKey,
			false,
		},

		{
			"user3 with first non-authorized key gets denied access",
			"user3",
			allowedPrivateKeyA,
			false,
		},
		{
			"user3 with second non-authorized key gets denied access",
			"user3",
			allowedPrivateKeyB,
			false,
		},
		{
			"user3 with third non-authorized key gets denied access",
			"user3",
			deniedPrivateKey,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := testutils.RunTestServerCommand(testServer.Addr, gossh.ClientConfig{
				User: tt.username,
				Auth: []gossh.AuthMethod{
					gossh.PublicKeys(tt.clientKey),
				},
			}, "exit 0", "")
			if tt.wantOK && err != nil {
				t.Error("AuthorizeKeys() unexpectedly denied access even though it shouldn't have")
			} else if !tt.wantOK && err == nil {
				t.Error("AuthorizeKeys() unexpectedly granted access even though it shouldn't have")
			}
		})
	}

}
