// Package copy provides WithReadDeadline.
package copy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/tkw1536/proxyssh/logging"
)

// WithContextTimeout is an additional timeout used by WithContext.
var WithContextTimeout = time.Second

// Buffer size used by withContext
const withContextBufferSize = 32 * 1024

// copyBufferPool is a pool for buffers of constant size.
var copyBufferPool = &sync.Pool{
	New: func() interface{} {
		return make([]byte, withContextBufferSize)
	},
}

// WithContext copies from src to dst, respecting ctx to cancel the operation.
// After ctx is cancelled, an additional internal timeout is used.
func WithContext(ctx context.Context, dst io.Writer, src io.Reader) (written int64, err error) {
	buf := copyBufferPool.Get().([]byte)
	defer copyBufferPool.Put(buf)

	for {
		nr, er := ReadWithContext(ctx, WithContextTimeout, src, buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}

var errTimeout = errors.New("Read Timeout")

// ReadWithContext is like src.Read(b), except when ctx is canceled and timeout is passed returns 0, ctx.Err().
//
// This somewhat works around the problem src.Read() being a blocking operation.
// The approach used in this function instead spins of the work of reading into a seperate goroutine.
// When the Read() operation is aborted, this might still cause src to leak in memory, however the current goroutine will return.
func ReadWithContext(ctx context.Context, timeout time.Duration, src io.Reader, b []byte) (n int, err error) {

	// when in memory leak checking mode, don't give an extra deadline!
	// this is so that we don't get "fire"s of the memory leaker
	if logging.MemoryLeakEnabled {
		timeout = 0
	}

	// grab a temporary buffer!
	secondBuffer := copyBufferPool.Get().([]byte)

	// if the buffer is too small, we need to extend the buffer!
	wantLen := len(b)
	if cap(secondBuffer) < wantLen {
		secondBuffer = make([]byte, len(b))
	}
	secondBuffer = secondBuffer[:len(b)]

	read := make(chan struct{})

	// start the call to src.Read() in a goroutine
	// and write it into the secondBuffer.
	var nInternal int
	var errInternal error
	go func() {
		nInternal, errInternal = src.Read(secondBuffer)
		close(read)
	}()

	// wait for the goroutine to finish and hope it does before the context.
	// if it finishes before the context, copy over the bytes and return.

	select {
	case <-read:
		copy(b, secondBuffer[:nInternal])

		copyBufferPool.Put(secondBuffer)
		return nInternal, errInternal
	case <-ctx.Done():
	}

	// back off and try and give it a bit more time to finish.
	// in the hope that it just deallocates eventually.

	select {
	case <-read:
		copy(b, secondBuffer[:nInternal])

		copyBufferPool.Put(secondBuffer)
		return nInternal, errInternal
	case <-time.After(timeout):
	}

	// Give up, it's probably blocking.
	// When in MemoryLeakDetector mode, at least tell the user.
	if logging.MemoryLeakEnabled {
		fmt.Println("ReadWithContext: Gave up src.Read()")
	}

	n = 0
	return 0, ctx.Err()
}
