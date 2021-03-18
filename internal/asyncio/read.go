package asyncio

import (
	"context"
	"io"
	"time"
)

// This file is untested because Wait() is tested.

// Read is a replacement for reader.Read() that returns within timeout after context is canncelled.
//
// Read can not cancel the underlying I/O; instead it is left running in a seperate goroutine.
func Read(ctx context.Context, timeout time.Duration, reader io.Reader, p []byte) (n int, err error) {
	// prepare return values for the inner read value
	innerP := bufferFromPool(len(p))
	defer bufferPool.Put(innerP)

	// prepare inner return values
	var innerN int
	var innerErr error

	// wait for the read to finish
	ok := Wait(ctx, timeout, func() {
		innerN, innerErr = reader.Read(innerP)
	}, nil)

	if !ok {
		return 0, ErrCanceled // read timed out
	}

	// copy the return values and exit
	copy(p, innerP[:innerN])
	return innerN, innerErr
}

// Reader returns a new reader with the Read method calling Read().
func Reader(ctx context.Context, timeout time.Duration, reader io.Reader) io.Reader {
	return asyncReader{reader: reader, ctx: ctx, timeout: timeout}
}

type asyncReader struct {
	reader  io.Reader
	ctx     context.Context
	timeout time.Duration
}

func (r asyncReader) Read(p []byte) (int, error) {
	return Read(r.ctx, r.timeout, r.reader, p)
}
