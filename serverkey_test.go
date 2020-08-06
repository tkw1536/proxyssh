package proxyssh

import (
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh/testutils"
	gossh "golang.org/x/crypto/ssh"
)

// testKeys represents
var testKeys = []struct {
	algorithm  HostKeyAlgorithm
	publicKey  string
	privateKey string
}{
	{
		algorithm: RSAAlgorithm,
		privateKey: `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAvEjBnZHJQo7mXpPW94JnH+pXBi0wdaqQ4wgFaMACjmkCnT2Y
Up6zAR9GXaiKcZGIItBrZ53VPEbGLMysLo1fbg3i8n9qyGkjHkbbOk0FKb1sgb9Y
snERrGnaijhltE3JV9V7r8djcFRsE7sqOvw/+iBAFpxTLUna3saGkbplGQCLeZH8
UdS+TzA2sqJ9G8TmnU54U3EIaMnw41QGGpKqPiZBPPWFAvIwyHAnSeRIwJQqxvKP
U7ZeTuprgq/m3UollpXNF2vjGtU/whO1rAjhB3n0V/2uS9tq1WzuRa46VUreF65x
lFUaLqnRPbfNKMWDklQ2vo9QJ8WNPOfft6zgewIDAQABAoIBACkbKUohXfMuB5V2
aWQ4EBOjscQjcYT+7Ark4WlxIh29R1jU7cB77VC9ZztjZHZO843GOuywRLGYMgPt
21l+e+snFPkkYEfIzGX7yjj8P7hRJrNc9xxeGyGtKo0qqummYeLPNOW3fjoz9DSK
lDm0gLM2/0bwcihdC2+/n/mI3DGM0KxiHHEWniZ0QBBt5msBcnStNqXl1HqI7zRM
WYfcw99tR6oDuXxUeckrx089Zm04ngoz8fN0TMzB/E06B9Ghp+VOFrxWvZtkdmlr
zQSG8FHRBj6Pb3q+xwC5+LWwCc2Rcy6117AFTUyoQbOUxr2xRMCKN/mXYuTMI6kJ
1Ek+bsECgYEA74C1nmtvuLIw6b06E+C5TQnYgjCRt0HOYgEdg40V9DphjDcHUmSP
t/DNvjeb52vuXkQlRvA9Fy0jnYoA4I+u3mRbDNuA69UDL0G/CI4BvG7E2pdliiGc
91gYcCidlI4T/bJ2KQljuBDYvHO/4dDrRtQfKstDiMQGxIWKMDPAJfMCgYEAyUDh
XKBGQttKVK0EYQ4+9bfSzDaLih05XLEyU4dKi9uGvsTiwXfYhZ+j1Uzib+I6Ggpj
bi/5z2EQkpVyvs/6Syar5lYoMZASTPFeglBFCNCBtqM4ofFzKm1aVeXxqucBCzC9
7h5bIz6RMzVZI5+QVcJyurlD+B/YHG4f+FEiVVkCgYACyOJTtzgTU68R6KtWM9Sz
upuT1/C5ysAVj6HCN8+7iTo9IR6qrJSnNNuPjKH5bN3WpsAwNPbg4Bt753DfK4yC
9XPBkIPNOirRT9hixxPqFvee+wepNX7XuWR/WVmLsqM03fBVxdAtAbUja80dWQqD
RlAedUKRwoW5nzveF5vyVQKBgQDAN0mHcETrIgsSaNWf5T1Y7qNVDFumJFdfIpbQ
lI0smxVNKzrwUYjpxxlxZid3ePjJWlaxLP1JhomPl1Gq0XVbRao1KuDkxZfVCUuc
5xGoY40gZTB36+Z1BVCcbiytcEjc6UbeIPwm42FHKZLjpUjzv+5YTQ6P88ozDTNX
thY2sQKBgQDascCJ/C3QUHgeqxquz5WJorOPKCc6yuAEcrxzRLqcpeuc+hedn2/n
cYkgpFn+ZrD7OFj3d/DbO38redzOl6Bi9u0VPCsyw6ejjGuu53ZJuws1CV8vSLP3
+2PxiwNUnMA/V3eGKPpKWnEP+psmf3/F0quZ8C2wh87jxQllszsKgQ==
-----END RSA PRIVATE KEY-----`,
		publicKey: `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC8SMGdkclCjuZek9b3gmcf6lcGLTB1qpDjCAVowAKOaQKdPZhSnrMBH0ZdqIpxkYgi0GtnndU8RsYszKwujV9uDeLyf2rIaSMeRts6TQUpvWyBv1iycRGsadqKOGW0TclX1Xuvx2NwVGwTuyo6/D/6IEAWnFMtSdrexoaRumUZAIt5kfxR1L5PMDayon0bxOadTnhTcQhoyfDjVAYakqo+JkE89YUC8jDIcCdJ5EjAlCrG8o9Ttl5O6muCr+bdSiWWlc0Xa+Ma1T/CE7WsCOEHefRX/a5L22rVbO5FrjpVSt4XrnGUVRouqdE9t80oxYOSVDa+j1AnxY0859+3rOB7
		`,
	},
	{
		algorithm: ED25519Algorithm,
		privateKey: `-----BEGIN PRIVATE KEY-----
5XWK7d7RseuAEPacEEKBCBrTVVX83VAQrLhdfwphSr4=
-----END PRIVATE KEY-----
		`,
		publicKey: `ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBOk1XkUeNB+ICjiJVqYqm4xvdXa8nq/oN4VeQldzB2V`,
	},
}

func TestReadOrMakeHostKey(t *testing.T) {
	for _, tt := range testKeys {

		// get the public key
		ttPublic, _, _, _, err := ssh.ParseAuthorizedKey([]byte(tt.publicKey))
		if err != nil {
			t.Fatal("Unable to parse public key")
		}

		t.Run(fmt.Sprintf("read an existing %s key", tt.algorithm), func(t *testing.T) {
			// create a temporary file
			tmpFile, cleanup := testutils.WriteTempFile("privkey.pem", tt.privateKey)
			defer cleanup()

			// test actual: try to load the key
			signer, err := ReadOrMakeHostKey(testutils.GetTestLogger(), tmpFile, tt.algorithm)
			if err != nil {
				t.Errorf("ReadOrMakeHostKey() error = %v, wantError = nil", err)
			}
			if signer == nil {
				t.Errorf("ReadOrMakeHostKey() signer = nil, wantSigner != nil")
			}
			if !ssh.KeysEqual(signer.PublicKey(), ttPublic) {
				t.Errorf("ReadOrMakeHostKey() returned wrong key")
			}

		})

		t.Run(fmt.Sprintf("make a new %s key", tt.algorithm), func(t *testing.T) {
			// create a new temp file path
			// by making a file, and then immediatly removing it
			tmpFile, cleanup := testutils.WriteTempFile("privkey.pem", "")
			cleanup()

			// test actual: ReadOrMakeHostKey should make a new file
			signer, err := ReadOrMakeHostKey(testutils.GetTestLogger(), tmpFile, tt.algorithm)
			if err != nil {
				t.Errorf("ReadOrMakeHostKey() error = %v, wantError = nil", err)
			}
			if signer == nil {
				t.Errorf("ReadOrMakeHostKey() signer = nil, wantSigner != nil")
			}

			if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
				t.Errorf("ReadOrMakeHostKey(): %s was not created", tmpFile)
			}
		})

		t.Run(fmt.Sprintf("error reading %s key", tt.algorithm), func(t *testing.T) {
			// create a temporary file, and write an invalid key into i
			tmpFile, cleanup := testutils.WriteTempFile("privkey.pem", "not-a-secret-key")
			defer cleanup()

			// test actual: ReadOrMakeHostKey should error
			signer, err := ReadOrMakeHostKey(testutils.GetTestLogger(), tmpFile, tt.algorithm)
			if err == nil {
				t.Errorf("ReadOrMakeHostKey() error = %v, wantError != nil", err)
			}
			if signer != nil {
				t.Errorf("ReadOrMakeHostKey() signer != nil, wantSigner = nil")
			}

			if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
				t.Errorf("ReadOrMakeHostKey(): %s was deleted", tmpFile)
			}
		})

	}
}

func TestUseOrMakeHostKey(t *testing.T) {
	for _, tt := range testKeys {

		// get the public key
		ttPublic, _, _, _, err := ssh.ParseAuthorizedKey([]byte(tt.publicKey))
		if err != nil {
			t.Fatal("Unable to parse public key")
		}

		t.Run(fmt.Sprintf("use %s key", tt.algorithm), func(t *testing.T) {
			// create a temporary file
			tmpFile, cleanup := testutils.WriteTempFile("privkey.pem", tt.privateKey)
			defer cleanup()

			_, err := UseOrMakeHostKey(testutils.GetTestLogger(), testServer, tmpFile, tt.algorithm)
			if err != nil {
				t.Errorf("UseOrMakeHostKey() error = %v, wantError = nil", err)
				t.FailNow()
			}

			// make a new session and store the public key
			var gotPublicKey ssh.PublicKey
			client, _, err := testutils.NewTestServerSession(testServer.Addr, gossh.ClientConfig{
				HostKeyAlgorithms: []string{ttPublic.Type()},
				HostKeyCallback: func(hostname string, remote net.Addr, key gossh.PublicKey) error {
					gotPublicKey = key
					return nil
				},
			})
			if err != nil {
				t.Errorf("Unable to create test server session: %s", err)
				t.FailNow()
			}
			defer client.Close()

			if !ssh.KeysEqual(gotPublicKey, ttPublic) {
				t.Errorf("UseOrMakeHostKey(): wrong public key presented")
			}
		})
	}
}
