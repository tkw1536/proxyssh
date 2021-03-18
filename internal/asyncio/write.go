package asyncio

import (
	"context"
	"io"
	"time"
)

// This file is untested because Wait() is tested.

// Write is a replacement for writer.Write() that returns within timeout after context is cancelled.
//
// Write can not cancel the underlying I/O; instead it is left running in a seperate goroutine.
func Write(ctx context.Context, timeout time.Duration, writer io.Writer, p []byte) (n int, err error) {
	// copy over a buffer for the inner write call
	innerP := bufferFromPool(len(p))
	copy(innerP, p)

	// inner return values
	var innerN int
	var innerErr error

	// make a write call
	ok := Wait(ctx, timeout, func() {
		defer bufferPool.Put(innerP)
		innerN, innerErr = writer.Write(innerP)
	}, nil)

	// write was canceled
	if !ok {
		return 0, ErrCanceled
	}

	// done!
	return innerN, innerErr
}

// Writer returns a new writer with the Write method calling Write().
func Writer(ctx context.Context, timeout time.Duration, writer io.Writer) io.Writer {
	return asyncWriter{writer: writer, ctx: ctx, timeout: timeout}
}

type asyncWriter struct {
	writer  io.Writer
	ctx     context.Context
	timeout time.Duration
}

func (w asyncWriter) Write(p []byte) (int, error) {
	return Write(w.ctx, w.timeout, w.writer, p)
}
