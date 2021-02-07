package integrationtest

import (
	"log"
	"os"
	"sync"

	"github.com/tkw1536/proxyssh/logging"
)

var testLogger *log.Logger
var testLogMutex sync.Mutex

// GetLogger returns a pointer to a logger that is to be shared amongst all test cases.
//
// This function can safely be called from different test cases, and in different goroutines at once.
// It will always return the same logger.
func GetLogger() logging.Logger {
	testLogMutex.Lock()
	defer testLogMutex.Unlock()

	// if we don't have a logger, create it
	if testLogger == nil {
		testLogger = log.New(os.Stdout, "test", log.LstdFlags)
	}

	return testLogger
}
