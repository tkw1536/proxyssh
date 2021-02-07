// Package testutils contains various functions to be used by the tests of the proxyssh package.
// It is not intended to be used publicly, and no versioning guarantees apply.
package testutils

import (
	"log"
	"sync"
	"testing"
)

func TestGetTestLogger(t *testing.T) {
	N := 10000

	// make a list of loggers
	var wg sync.WaitGroup
	loggers := make([]*log.Logger, N)

	// wait to fill all of them
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func(i int) {
			loggers[i] = GetTestLogger()
			wg.Done()
		}(i)
	}

	wg.Wait()

	// check that the first logger is not nil
	firstLogger := loggers[0]
	if firstLogger == nil {
		t.Error("TestLogger() first call returned nil")
	}

	// check that the rest are all equal
	for _, logger := range loggers[1:] {
		if logger != firstLogger {
			t.Error("TestLogger() returned different instances")
		}
	}

}
