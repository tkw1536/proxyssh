package testutils

import (
	"io"
	"log"
)

// InterceptAndLog returns a new io.Reader that logs every Read() call result.
// It is intended for debugging processes and should not be used in production code.
func InterceptAndLog(reader io.Reader, s string) *InterceptReader {
	return &InterceptReader{
		Reader: reader,
		Intercept: func(p []byte, err error) {
			log.Printf("InterceptPrint: %s %v %v\n", s, err, p)
		},
	}
}

// InterceptReader wraps an io.Reader.
// It calls the Intercept function on every received set of bytes.
//
// It is intended for debugging processes and should not be used in production code.
type InterceptReader struct {
	Reader    io.Reader
	Intercept func(p []byte, err error)
}

func (r *InterceptReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	if r.Intercept != nil {
		// call the function on a copy of the byte slice
		cp := make([]byte, n)
		copy(cp, p[:n])
		r.Intercept(cp, err)
	}
	return
}
