// Package asyncio provides context-aware alternatives to io methods.
package asyncio

import (
	"context"
	"testing"
	"time"
)

func TestWait(t *testing.T) {
	t.Run("finish before context cancels without cancel func", func(t *testing.T) {
		ok := Wait(context.Background(), 0, func() {}, nil)
		if ok != true {
			t.Error("Wait(): got ok=false, wanted ok=true")
		}
	})

	t.Run("finish before context cancels with cancel func", func(t *testing.T) {
		ok := Wait(context.Background(), 0, func() {}, func() { panic("shouldn't be called") })
		if ok != true {
			t.Error("Wait(): got ok=false, wanted ok=true")
		}
	})

	t.Run("finish after context, but before timeout without cancel func", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(time.Second / 10)
			cancel()
		}()

		ok := Wait(ctx, 1*time.Second, func() {
			<-ctx.Done()
		}, nil)

		if ok != true {
			t.Error("Wait(): got ok=false, wanted ok=true")
		}
	})

	t.Run("finish after context, but before timeout with cancel func", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(time.Second / 10)
			cancel()
		}()

		var cancelCalled bool

		ok := Wait(ctx, time.Second, func() {
			time.Sleep(2 * time.Second) // finish after two seconds
		}, func() {
			cancelCalled = true
		})

		if ok != false {
			t.Error("Wait(): got ok=true, wanted ok=false")
		}

		if cancelCalled != true {
			t.Error("Wait() did not call cancel!")
		}
	})
}
