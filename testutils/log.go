package testutils

import (
	"log"
	"os"
	"sync"
)

var testLogger *log.Logger
var testLogMutex sync.Mutex

// GetTestLogger returns a pointer to a logger that is to be shared amongst all test cases.
//
// This function can safely be called from different test cases, and in different goroutines at once.
// It will always return the same logger.
func GetTestLogger() *log.Logger {
	testLogMutex.Lock()
	defer testLogMutex.Unlock()

	// if we don't have a logger, create it
	if testLogger == nil {
		testLogger = log.New(os.Stdout, "test", log.LstdFlags)
	}

	return testLogger
}
