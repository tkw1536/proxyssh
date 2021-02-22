package asyncio

import (
	"context"
	"io"
	"time"
)

// Read is a replacement for reader.Read() that returns within timeout after context is canncelled.
//
// Read can not cancel the underlying I/O; instead it is left running in a seperate goroutine.
func Read(ctx context.Context, timeout time.Duration, reader io.Reader, p []byte) (n int, err error) {

	// channel for the read
	readDone := make(chan struct{})

	// prepare return values for the inner read value
	innerP := bufferFromPool(len(p))
	defer bufferPool.Put(innerP)

	var innerN int
	var innerErr error

	go func() {
		defer close(readDone)
		innerN, innerErr = reader.Read(innerP)
	}()

	// wait for the read to be done!
	select {
	case <-readDone:
		goto readOK
	case <-ctx.Done():
	}

	// wait again for the timeout to expire
	select {
	case <-readDone:
		goto readOK
	case <-time.After(timeout):
	}

	return 0, ErrCanceled

readOK:
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
