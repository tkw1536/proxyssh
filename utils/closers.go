package utils

import "io"

// AsWriterCloser is an io.Writer which falls back to a noop close method
type AsWriterCloser struct {
	io.Writer
}

// Close calls Close() on Writer if supported, or returns nil.
func (nc *AsWriterCloser) Close() error {
	if closer, isCloser := nc.Writer.(io.Closer); isCloser {
		return closer.Close()
	}
	return nil
}

// AsReaderCloser is an io.Reader which falls back to a noop close method
type AsReaderCloser struct {
	io.Reader
}

// Close calls Close() on Reader if supported, or returns nil.
func (nc *AsReaderCloser) Close() error {
	if closer, isCloser := nc.Reader.(io.Closer); isCloser {
		return closer.Close()
	}
	return nil
}

// AsReadWriterCloser is an io.ReadWriter which falls back to a noop close method
type AsReadWriterCloser struct {
	io.ReadWriter
}

// Close calls Close() on ReadWriter if supported, or returns nil.
func (nc *AsReadWriterCloser) Close() error {
	if closer, isCloser := nc.ReadWriter.(io.Closer); isCloser {
		return closer.Close()
	}
	return nil
}
