package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const MaxFileSize = 16 << 20

func Read(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open backup: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("inspect backup: %w", err)
	}
	if info.Size() > MaxFileSize {
		return nil, fmt.Errorf("backup exceeds the %d MiB size limit", MaxFileSize>>20)
	}
	data, err := io.ReadAll(io.LimitReader(file, MaxFileSize+1))
	if err != nil {
		return nil, fmt.Errorf("read backup: %w", err)
	}
	if len(data) > MaxFileSize {
		return nil, fmt.Errorf("backup exceeds the %d MiB size limit", MaxFileSize>>20)
	}
	return data, nil
}

func Write(path string, data []byte) error {
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".tailorcv-backup-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary backup: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("protect temporary backup: %w", err)
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write temporary backup: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync temporary backup: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary backup: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace backup file: %w", err)
	}
	return nil
}
