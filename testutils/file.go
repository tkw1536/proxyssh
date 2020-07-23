package testutils

import (
	"io/ioutil"
	"os"
)

// WriteTempFile writes content into a temporary file.
// The caller is expected to call cleanup when done with the file.
// If something goes wrong, calls panic().
func WriteTempFile(pattern, content string) (filepath string, cleanup func()) {
	// Create a temporary file
	tmpFile, err := ioutil.TempFile("", pattern)
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
