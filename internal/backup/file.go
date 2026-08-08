package backup

import (
	"fmt"
	"io"
	"os"

	"github.com/jscyril/tailorcv/internal/atomicfile"
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
	if err := atomicfile.Write(path, data); err != nil {
		return fmt.Errorf("write backup: %w", err)
	}
	return nil
}
