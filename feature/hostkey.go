package feature

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"os"

	"github.com/gliderlabs/ssh"
	"github.com/pkg/errors"
	"github.com/tkw1536/proxyssh/logging"
	gossh "golang.org/x/crypto/ssh"
)

// Because of import cyles, tests for this file reside in config/feature_hostkey_test.go.

// UseOrMakeHostKeys is like UseOrMakeHostKey except that it accepts multiple HostKeyAlgorithms.
// For each key algorithm, the privateKeyPath is appended with "_" + the name of the algorithm in question.
//
// When algorithms is nil, picks a reasonable set of default algorithms.
func UseOrMakeHostKeys(logger logging.Logger, server *ssh.Server, privateKeyPath string, algorithms []HostKeyAlgorithm) error {
	if algorithms == nil {
		algorithms = []HostKeyAlgorithm{RSAAlgorithm, ED25519Algorithm}
	}

	for _, algorithm := range algorithms {
		path := privateKeyPath + "_" + string(algorithm)
		if err := UseOrMakeHostKey(logger, server, path, algorithm); err != nil {
			return err
		}
	}
	return nil
}

// UseOrMakeHostKey attempts to load a host key from the given privateKeyPath.
// If the path does not exist, a new host key is generated.
// It then adds this hostkey to the priovided server.
//
// All parameters except the server are passed to ReadOrMakeHostKey.
// Please see the appropriate documentation for that function.
//
// logger is called whenever a new host key algorithm is being generated.
func UseOrMakeHostKey(logger logging.Logger, server *ssh.Server, privateKeyPath string, algorithm HostKeyAlgorithm) error {
	key, err := ReadOrMakeHostKey(logger, privateKeyPath, algorithm)
	if err != nil {
		return err
	}

	// use the host key
	server.AddHostKey(key)
	return nil
}

// ReadOrMakeHostKey attempts to load a host key from the given privateKeyPath.
// If the path does not exist, a new key is generated.
//
// This function assumes that if there is a host key in privateKeyPath it uses the provided HostKeyAlgorithm.
// It makes no attempt at verifiying this; the key mail fail to load and return an error, or it may load incorrect data.
//
// logger is called whenever a new host key algorithm is being generated.
func ReadOrMakeHostKey(logger logging.Logger, privateKeyPath string, algorithm HostKeyAlgorithm) (key gossh.Signer, err error) {
	hostKey := NewHostKey(algorithm)

	if _, e := os.Stat(privateKeyPath); os.IsNotExist(e) { // path doesn't exist => generate a new key there!
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
	return hostKey, nil
}

// loadHostKey loadsa host key
func loadHostKey(logger logging.Logger, key HostKey, path string) (err error) {
	logger.Printf("load_hostkey %s %s", key.Algorithm(), path)

	// read all the bytes from the file
	privateKeyBytes, err := os.ReadFile(path)
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
func makeHostKey(logger logging.Logger, key HostKey, path string) error {
	logger.Printf("generate_hostkey %s %s", key.Algorithm(), path)

	if err := key.Generate(0, nil); err != nil {
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

// HostKey represents an pair of ssh private key and algorithm.
// Once the hostkey is generated or loaded, it is safe for concurrent accesses.
type HostKey interface {
	ssh.Signer

	// Algorithm is the Algorithm used by this HostKey implementation.
	Algorithm() HostKeyAlgorithm

	// Generate generates a new HostKey, discarding whatever was previsouly contained.
	//
	// keySize is the desired public key size in bits. When keySize is 0, a sensible default is used.
	// random is the source of randomness. If random is nil, crypto/rand.Reader will be used.
	Generate(keySize int, random io.Reader) error

	// MarshalPEM marshals the private key into a pem.Block to be used for exporting.
	// The format is not guaranteed to follow any kind of standard, only that it is readable with the corresponding UnmarshalPEM.
	MarshalPEM() (*pem.Block, error)

	// UnmarshalPEM unmarshals the private key from a pem.Block.
	// It is only compatible with whatever MarshalPEM() outputted.
	UnmarshalPEM(block *pem.Block) error
}

// HostKeyAlgorithm is an enumerated value that represents a specific algorithm used for host keys.
type HostKeyAlgorithm string

const (
	// RSAAlgorithm represents the RSA Algorithm
	RSAAlgorithm HostKeyAlgorithm = "rsa"

	// ED25519Algorithm represents the ED25519 algorithm
	ED25519Algorithm HostKeyAlgorithm = "ed25519"
)

// NewHostKey returns a new empty HostKey for the provided HostKey Algorithm.
// An unsupported HostKeyAlgorithm will result in a call to panic().
func NewHostKey(algorithm HostKeyAlgorithm) HostKey {
	switch algorithm {
	case RSAAlgorithm:
		return &rsaHostKey{defaultBitSize: 4096}
	case ED25519Algorithm:
		return &ed25519HostKey{}
	default:
		panic("Unsupported HostKeyAlgorithm")
	}
}

//
// ed25519 key
//

type ed25519HostKey struct {
	ssh.Signer
	pk *ed25519.PrivateKey
}

func init() {
	var _ HostKey = (*ed25519HostKey)(nil)
}

func (ek *ed25519HostKey) Algorithm() HostKeyAlgorithm {
	return ED25519Algorithm
}

var errKeySizeUnsupported = errors.New("ed25519HostKey.Generate(): keySize not supported")

func (ek *ed25519HostKey) Generate(keySize int, random io.Reader) (err error) {
	if keySize != 0 && keySize != ed25519.PublicKeySize {
		return errKeySizeUnsupported
	}
	if random == nil {
		random = rand.Reader
	}

	_, pr, err := ed25519.GenerateKey(random)
	if err != nil {
		return
	}

	// store the private key and setup the signer
	ek.pk = &pr
	ek.Signer, err = gossh.NewSignerFromKey(ek.pk)

	// return
	return
}

func (ek *ed25519HostKey) MarshalPEM() (block *pem.Block, err error) {
	block = &pem.Block{Type: "PRIVATE KEY", Bytes: ek.pk.Seed()}
	return
}

func (ek *ed25519HostKey) UnmarshalPEM(block *pem.Block) (err error) {
	if block.Type != "PRIVATE KEY" {
		err = errors.New("Expected 'PRIVATE KEY' in PEM format")
		return
	}

	pk := ed25519.NewKeyFromSeed(block.Bytes)

	// store the private key and setup the signer
	ek.pk = &pk
	ek.Signer, err = gossh.NewSignerFromKey(ek.pk)

	return nil
}

//
// rsa key
//

type rsaHostKey struct {
	ssh.Signer

	pk *rsa.PrivateKey

	defaultBitSize int
}

func init() {
	var _ HostKey = (*rsaHostKey)(nil)
}

func (rk *rsaHostKey) Algorithm() HostKeyAlgorithm {
	return RSAAlgorithm
}

func (rk *rsaHostKey) Generate(keySize int, random io.Reader) (err error) {
	if keySize == 0 {
		keySize = rk.defaultBitSize
	}
	if random == nil {
		random = rand.Reader
	}

	rk.pk, err = rsa.GenerateKey(random, keySize)
	if err != nil {
		return err
	}

	// store the signer
	rk.Signer, err = gossh.NewSignerFromKey(rk.pk)
	return
}

func (rk *rsaHostKey) MarshalPEM() (block *pem.Block, err error) {
	block = &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rk.pk)}
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

	pk, isRSA := parsedKey.(*rsa.PrivateKey)
	if !isRSA {
		err = errors.New("Expected an rsa.PrivateKey")
		return
	}

	// store the private key and setup the signer
	rk.pk = pk
	rk.Signer, err = gossh.NewSignerFromKey(rk.pk)

	return
}
