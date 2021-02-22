// Package asyncio provides context-aware alternatives to io methods.
package asyncio

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/tkw1536/proxyssh/logging"
)

// TODO: Test this package!

// ErrCanceled is returned by Read() or Write() when the context was canceled and the timeout expired.
var ErrCanceled = errors.New("asyncio: Context canceled and timeout expired")

// Copy is like io.Copy, but wrapping dst and src using Writer and Reader methods.
func Copy(ctx context.Context, timeout time.Duration, dst io.Writer, src io.Reader) (written int64, err error) {
	return io.Copy(
		Writer(ctx, timeout, dst),
		Reader(ctx, timeout, src),
	)
}

// CopyLeak is like Copy with timeout being half the memory leak timeout.
// It also prints a debug message when the memory leak detector is enabled.
func CopyLeak(ctx context.Context, dst io.Writer, src io.Reader) (written int64, err error) {
	written, err = Copy(ctx, logging.MemoryLeakTimeout/2, dst, src)
	if logging.MemoryLeakEnabled && err == ErrCanceled {
		fmt.Println("CopyLeak: Timed out on copy!")
	}
	return
}

// StdCopy is like stdcopy.StdCopy, but wrapping readers using Reader() and writers using Writer().
func StdCopy(ctx context.Context, timeout time.Duration, dstout, dsterr io.Writer, src io.Reader) (written int64, err error) {
	return stdcopy.StdCopy(
		Writer(ctx, timeout, dstout),
		Writer(ctx, timeout, dsterr),
		Reader(ctx, timeout, src),
	)
}

// StdCopyLeak is like StdCopy with timeout being half the memory leak timeout.
func StdCopyLeak(ctx context.Context, dstout, dsterr io.Writer, src io.Reader) (written int64, err error) {
	written, err = StdCopy(ctx, logging.MemoryLeakTimeout/2, dstout, dsterr, src)
	if logging.MemoryLeakEnabled && err == ErrCanceled {
		fmt.Println("StdCopyLeak: Timed out on copy!")
	}
	return
}
