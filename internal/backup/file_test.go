package backup

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndReadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile.json")
	want := []byte(`{"schemaVersion":1}`)
	if err := Write(path, want); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("Read() = %q, want %q", got, want)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("backup permissions = %o, want owner-only", info.Mode().Perm())
	}
}

func TestReadRejectsOversizedBackup(t *testing.T) {
	path := filepath.Join(t.TempDir(), "large.json")
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := file.Truncate(MaxFileSize + 1); err != nil {
		t.Fatalf("Truncate() error = %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if _, err := Read(path); err == nil {
		t.Fatal("Read() expected size error")
	}
}
