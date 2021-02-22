package asyncio

import (
	"context"
	"io"
	"time"
)

// Write is a replacement for writer.Write() that returns within timeout after context is cancelled.
//
// Write can not cancel the underlying I/O; instead it is left running in a seperate goroutine.
func Write(ctx context.Context, timeout time.Duration, writer io.Writer, p []byte) (n int, err error) {
	// channel for the read
	writeDone := make(chan struct{})

	// copy over a buffer for the inner write call
	innerP := bufferFromPool(len(p))
	copy(innerP, p)

	var innerN int
	var innerErr error

	go func() {
		defer bufferPool.Put(innerP)
		defer close(writeDone)
		innerN, innerErr = writer.Write(innerP)
	}()

	// wait for the read to be done!
	select {
	case <-writeDone:
		goto writeOK
	case <-ctx.Done():
	}

	// wait again for the timeout to expire
	select {
	case <-writeDone:
		goto writeOK
	case <-time.After(timeout):
	}

	return 0, ErrCanceled

writeOK:
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
