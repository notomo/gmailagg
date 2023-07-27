package fstestext

import (
	"io"
	"io/fs"
	"testing"
	"testing/fstest"
)

func GetFileContent(t *testing.T, tmpfs fs.FS, fileName string) []byte {
	t.Helper()

	if err := fstest.TestFS(tmpfs, fileName); err != nil {
		t.Fatal(err)
	}

	f, err := tmpfs.Open(fileName)
	if err != nil {
		t.Fatal(err)
	}

	got, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	return got
}
