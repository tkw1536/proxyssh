package testutils

import (
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

func TestGenerateRSATestKeyPair(t *testing.T) {
	N := 1

	for i := 0; i < N; i++ {
		gotPrivate, gotPublic := GenerateRSATestKeyPair()

		if !ssh.KeysEqual(gotPrivate.PublicKey(), gotPublic) {
			t.Error("GenerateRSATestKeyPair(): private and public key do not match")
		}

		if gotPublic.Type() != "ssh-rsa" {
			t.Error("GenerateRSTestKeyPair(): did not get type 'ssh-rsa'")
		}
	}
}

func TestAuthorizedKeysString(t *testing.T) {
	got := AuthorizedKeysString(rsaPrivateKey.PublicKey())
	want := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC8SMGdkclCjuZek9b3gmcf6lcGLTB1qpDjCAVowAKOaQKdPZhSnrMBH0ZdqIpxkYgi0GtnndU8RsYszKwujV9uDeLyf2rIaSMeRts6TQUpvWyBv1iycRGsadqKOGW0TclX1Xuvx2NwVGwTuyo6/D/6IEAWnFMtSdrexoaRumUZAIt5kfxR1L5PMDayon0bxOadTnhTcQhoyfDjVAYakqo+JkE89YUC8jDIcCdJ5EjAlCrG8o9Ttl5O6muCr+bdSiWWlc0Xa+Ma1T/CE7WsCOEHefRX/a5L22rVbO5FrjpVSt4XrnGUVRouqdE9t80oxYOSVDa+j1AnxY0859+3rOB7"

	if got != want {
		t.Errorf("AuthorizedKeysString() = %s, want = %s", got, want)
	}
}

var rsaPrivateKey gossh.Signer

func init() {
	block, _ := pem.Decode([]byte(`-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`))

	private, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	rsaPrivateKey, err = gossh.NewSignerFromKey(private)
	if err != nil {
		panic(err)
	}

}
