package integrationtest

import (
	"sync"
	"testing"

	"github.com/tkw1536/proxyssh/logging"
)

func TestGetLogger(t *testing.T) {
	N := 10000

	// make a list of loggers
	var wg sync.WaitGroup
	loggers := make([]logging.Logger, N)

	// wait to fill all of them
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func(i int) {
			loggers[i] = GetLogger()
			wg.Done()
		}(i)
	}

	wg.Wait()

	// check that the first logger is not nil
	firstLogger := loggers[0]
	if firstLogger == nil {
		t.Error("GetLogger() first call returned nil")
	}

	// check that the rest are all equal
	for _, logger := range loggers[1:] {
		if logger != firstLogger {
			t.Error("GetLogger() returned different instances")
		}
	}

}
