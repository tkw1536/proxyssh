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

	logger = log.New(os.Stderr, "test", log.LstdFlags)
	return logger
}
