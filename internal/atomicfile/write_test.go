package atomicfile

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAtomicallyCreatesAndReplacesPrivateFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "resume.pdf")
	if err := Write(path, []byte("first")); err != nil {
		t.Fatalf("Write(first) error = %v", err)
	}
	if err := Write(path, []byte("replacement")); err != nil {
		t.Fatalf("Write(replacement) error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !bytes.Equal(data, []byte("replacement")) {
		t.Fatalf("file contents = %q", data)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("file permissions = %o, want owner-only", info.Mode().Perm())
	}
}
