// Package testutils contains various functions to be used by the tests of the proxyssh package.
// It is not intended to be used publicly, and no versioning guarantees apply.
package testutils

import (
	"log"
	"os"
	"sync"
)

var logger *log.Logger
var logMutex sync.Mutex

// TestLogger returns a logger for use within tests
func TestLogger() *log.Logger {
	logMutex.Lock()
	defer logMutex.Unlock()

	if logger != nil {
		return logger
	}

	logger = log.New(os.Stdout, "test", log.LstdFlags)
	return logger
}
