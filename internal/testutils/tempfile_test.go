package testutils

import (
	"os"
	"sync"
	"testing"
)

func TestWriteTempFile(t *testing.T) {
	testFileText := "hello world"

	// write the temp file and make sure it's cleaned up
	var once sync.Once
	path, cleanup := WriteTempFile("", testFileText)
	defer once.Do(cleanup)

	if path == "" {
		t.Errorf("WriteTempFile() got path = '', want path != ''")
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	gotString := string(bytes)
	if gotString != testFileText {
		t.Errorf("WriteTempFile() wrote text=%s, want text=%s", gotString, testFileText)
	}

	// cleanup should delete the file
	once.Do(cleanup)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("WriteTempFile() cleanup did not delete temporary file")
	}

}
