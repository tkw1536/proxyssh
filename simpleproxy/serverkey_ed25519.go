package simpleproxy

import (
	"crypto/ed25519"
	"encoding/pem"

	"github.com/gliderlabs/ssh"
	"github.com/pkg/errors"
	gossh "golang.org/x/crypto/ssh"
)

type ed25519HostKey struct {
	*ed25519.PrivateKey
}

func (ek *ed25519HostKey) Algorithm() HostKeyAlgorithm {
	return ED25519Algorithm
}

func (ek *ed25519HostKey) Generate() (err error) {
	_, pr, err := ed25519.GenerateKey(nil)
	if err != nil {
		return
	}
	ek.PrivateKey = &pr
	return
}

func (ek *ed25519HostKey) MarshalPEM() (block *pem.Block, err error) {
	block = &pem.Block{Type: "PRIVATE KEY", Bytes: ek.PrivateKey.Seed()}
	return
}

func (ek *ed25519HostKey) UnmarshalPEM(block *pem.Block) (err error) {
	if block.Type != "PRIVATE KEY" {
		err = errors.New("Expected 'PRIVATE KEY' in PEM format")
		return
	}

	// parse either a PKCS1 or PKCS8
	pk := ed25519.NewKeyFromSeed(block.Bytes)
	ek.PrivateKey = &pk
	return nil
}

func (ek *ed25519HostKey) Signer() (ssh.Signer, error) {
	return gossh.NewSignerFromKey(ek.PrivateKey)
}

func init() {
	var _ hostKey = (*ed25519HostKey)(nil)
}
