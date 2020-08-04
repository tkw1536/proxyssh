package testutils

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"

	"golang.org/x/crypto/ssh"
)

// rsaTestKeyPairSize is the bitsize forthe test key
const rsaTestKeyPairSize = 2048

// GenerateRSATestKeyPair generates a new RSA Keypair for use within testing.
// A keypair consists of a private key (in the form of a Signer) and a public key.
//
// If something goes wrong during the generation of the keypair, panic() is called.
//
// The bitsize of the keypair is determined internally, but will be the same for subsequent calls of this function.
func GenerateRSATestKeyPair() (ssh.Signer, ssh.PublicKey) {
	privkey, err := rsa.GenerateKey(rand.Reader, rsaTestKeyPairSize)
	if err != nil {
		panic(err)
	}
	signer, err := ssh.NewSignerFromKey(privkey)
	if err != nil {
		panic(err)
	}

	return signer, signer.PublicKey()
}

// AuthorizedKeysString turns an (assumed rsa-key) into a string that can be written to an authorized_keys file.
// If something goes wrong, calls panic()
func AuthorizedKeysString(key ssh.PublicKey) string {
	buffer := &bytes.Buffer{}
	encoder := base64.NewEncoder(base64.StdEncoding, buffer)
	if _, err := encoder.Write(key.Marshal()); err != nil {
		panic(err)
	}
	if err := encoder.Close(); err != nil {
		panic(err)
	}
	return "ssh-rsa " + buffer.String()
}
