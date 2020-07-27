package proxyssh

import (
	"encoding/pem"
	"io/ioutil"
	"os"

	"github.com/gliderlabs/ssh"
	"github.com/pkg/errors"
	"github.com/tkw1536/proxyssh/utils"
	gossh "golang.org/x/crypto/ssh"
)

// UseOrMakeHostKey attempts to load a host key from the given privateKeyPath.
// If the path does not exist, a new host key is generated.
// It then adds this hostkey to the priovided server.
//
// All parameters except the server are passed to ReadOrMakeHostKey.
// Please see the appropriate documentation for that function.
//
// logger is called whenever a new host key algorithm is being generated.
//
// The server returned from this function is returned for convenience only.
// It is the same server that was originally passed.
func UseOrMakeHostKey(logger utils.Logger, server *ssh.Server, privateKeyPath string, algorithm HostKeyAlgorithm) (*ssh.Server, error) {
	key, err := ReadOrMakeHostKey(logger, privateKeyPath, algorithm)
	if err != nil {
		return server, err
	}

	// use the host key
	server.AddHostKey(key)
	return server, nil
}

// ReadOrMakeHostKey attempts to load a host key from the given privateKeyPath.
// If the path does not exist, a new host key is generated.
//
// This function assumes that if there is a host key in privateKeyPath it uses the provided HostKeyAlgorithm.
// It makes no attempt at verifiying this; the key mail fail to load and return an error, or it may load incorrect data.
//
// logger is called whenever a new host key algorithm is being generated.
func ReadOrMakeHostKey(logger utils.Logger, privateKeyPath string, algorithm HostKeyAlgorithm) (key gossh.Signer, err error) {
	hostKey := newHostKey(algorithm)
	// path doesn't exist => generate a new key there!
	if _, e := os.Stat(privateKeyPath); os.IsNotExist(e) {
		err = makeHostKey(logger, hostKey, privateKeyPath)
		if err != nil {
			err = errors.Wrap(err, "Unable to generate new host key")
			return
		}
	}
	err = loadHostKey(logger, hostKey, privateKeyPath)
	if err != nil {
		return nil, err
	}
	return hostKey.Signer()
}

// HostKeyAlgorithm is an enumerated value that represents a specific algorithm used for host keys.
type HostKeyAlgorithm string

const (
	// RSAAlgorithm represents the RSA Algorithm
	RSAAlgorithm HostKeyAlgorithm = "rsa"

	// ED25519Algorithm represents the ED25519 algorithm
	ED25519Algorithm HostKeyAlgorithm = "ed25519"
)

type hostKey interface {
	// Algorithm returns the algorithm of this hostkey
	Algorithm() HostKeyAlgorithm

	// signer turns this hostkey into a signer
	Signer() (ssh.Signer, error)

	// Generate generates a new private key for this key algorithm
	Generate() error
	// MarshalPEM writes this key to a pem
	MarshalPEM() (*pem.Block, error)

	// UnmarshalPEM reads this key from a pem
	UnmarshalPEM(block *pem.Block) error
}

// newHostKey returns a hostKey of the given algorithm
func newHostKey(algorithm HostKeyAlgorithm) hostKey {
	switch algorithm {
	case RSAAlgorithm:
		return &rsaHostKey{bitSize: 4096}
	case ED25519Algorithm:
		return &ed25519HostKey{}
	default:
		panic("Unsupported HostKeyAlgorithm")
	}
}

// loadHostKey loadsa host key
func loadHostKey(logger utils.Logger, key hostKey, path string) (err error) {
	logger.Printf("load_hostkey %s %s", key.Algorithm(), path)

	// read all the bytes from the file
	privateKeyBytes, err := ioutil.ReadFile(path)
	if err != nil {
		err = errors.Wrap(err, "Unable to read private key bytes")
		return
	}

	// if the length is nil, return
	if len(privateKeyBytes) == 0 {
		err = errors.New("No bytes were read from the private key")
		return
	}

	// decode the pem and unmarshal it
	privateKeyPEM, _ := pem.Decode(privateKeyBytes)
	if privateKeyPEM == nil {
		err = errors.New("pem.Decode() returned nil")
		return
	}
	return key.UnmarshalPEM(privateKeyPEM)
}

// makeHostKey makes a new host key
func makeHostKey(logger utils.Logger, key hostKey, path string) error {
	logger.Printf("generate_hostkey %s %s", key.Algorithm(), path)

	if err := key.Generate(); err != nil {
		return errors.Wrap(err, "Failed to generate key")
	}

	privateKeyPEM, err := key.MarshalPEM()
	if err != nil {
		return errors.Wrap(err, "Failed to marshal key")
	}

	// generate and write private key as PEM
	privateKeyFile, err := os.Create(path)
	defer privateKeyFile.Close()
	if err != nil {
		return err
	}
	return pem.Encode(privateKeyFile, privateKeyPEM)
}
