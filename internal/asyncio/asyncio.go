// Package asyncio provides context-aware alternatives to io methods.
package asyncio

import (
	"context"
	"time"
)

// Wait performs action in a concurrent goroutine and returns true once it has completed.
//
// When ctx is canceled, and f is not yet finished, calls cancel.
// Then waits at most timeout and, if f has not returned, returns false.
// When cancel is nil, it is not called.
func Wait(ctx context.Context, timeout time.Duration, action, cancel func()) (ok bool) {
	// context is already closed; don't do anything!
	if ctx.Err() != nil {
		return false
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		action()
	}()

	// select 1: context
	select {
	case <-done:
		return true
	case <-ctx.Done():
	}

	// cancel the operation
	if cancel != nil {
		cancel()
	}

	// no timeout => return immediatly
	if timeout <= 0 {
		return
	}

	// select 2: timeout
	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}
