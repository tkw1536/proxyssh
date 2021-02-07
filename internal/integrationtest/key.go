package integrationtest

import (
	"sync"

	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh/internal/testutils"
)

var testKeyKey ssh.Signer
var testKeyMutex sync.Mutex

// getTestKey returns an ssh.Signer generated for testing.
// Multiple calls to getTestKey will return the same key.
func getTestKey() ssh.Signer {
	testKeyMutex.Lock()
	defer testKeyMutex.Unlock()

	if testKeyKey != nil {
		return testKeyKey
	}

	signer, _ := testutils.GenerateRSATestKeyPair()
	testKeyKey = signer
	return testKeyKey
}
