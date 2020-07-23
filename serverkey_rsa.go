package proxyssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/gliderlabs/ssh"
	"github.com/pkg/errors"
	gossh "golang.org/x/crypto/ssh"
)

type rsaHostKey struct {
	*rsa.PrivateKey
	bitSize int
}

func (rk *rsaHostKey) Algorithm() HostKeyAlgorithm {
	return RSAAlgorithm
}

func (rk *rsaHostKey) Generate() (err error) {
	rk.PrivateKey, err = rsa.GenerateKey(rand.Reader, rk.bitSize)
	return
}

func (rk *rsaHostKey) MarshalPEM() (block *pem.Block, err error) {
	block = &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rk.PrivateKey)}
	return
}

func (rk *rsaHostKey) UnmarshalPEM(block *pem.Block) (err error) {
	if block.Type != "RSA PRIVATE KEY" {
		err = errors.New("Expected 'RSA PRIVATE KEY' in PEM format")
		return
	}

	// parse either a PKCS1 or PKCS8
	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
		if parsedKey, err = x509.ParsePKCS8PrivateKey(block.Bytes); err != nil { // note this returns type `interface{}`
			err = errors.Wrap(err, "Expected PKCS1 or PKCS8 private key")
			return
		}
	}

	var isRSA bool
	rk.PrivateKey, isRSA = parsedKey.(*rsa.PrivateKey)
	if !isRSA {
		err = errors.New("Expected an rsa.PrivateKey")
		return
	}

	return
}

func (rk *rsaHostKey) Signer() (ssh.Signer, error) {
	return gossh.NewSignerFromKey(rk.PrivateKey)
}

func init() {
	var _ hostKey = (*rsaHostKey)(nil)
}
