package testutils

import (
	"os"
)

// WriteTempFile writes content into a temporary file.
//
// Pattern is used to generate the filename of the file, it is passed on to os.CreateTemp.
// Content is written as a string into the file.
//
// This function returns a pair of (filepath, cleanup) where
// filepath is the path to the temporary file and
// cleanup is a function that removes the temporary file.
//
// It is the callers responsibility to call the cleanup function.
// A typical invocation of this function is something like:
//
//	tempfile, cleanup = WriteTempFile(pattern, content)
//	defer cleanup()
//
// If something goes wrong during the creation of the temporary file, panic() is called.
func WriteTempFile(pattern, content string) (filepath string, cleanup func()) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", pattern)
	if err != nil {
		panic(err)
	}

	// cleanup: delete the file when done
	filepath = tmpFile.Name()
	cleanup = func() {
		os.Remove(filepath)
	}

	// if we have some content, write it into the file
	if content != "" {
		if _, err := tmpFile.Write([]byte(content)); err != nil {
			tmpFile.Close()
			cleanup()
			panic(err)
		}
	}

	// close the file, so that other code can read it
	tmpFile.Close()

	return
}
