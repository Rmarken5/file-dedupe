package hasher

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
)

type (
	File struct {
		FilePath string
		MD5Hash  string
	}
)

func New(filePath string) (*File, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	hash := md5.Sum(data)
	return &File{FilePath: filePath, MD5Hash: fmt.Sprintf("%x", hash)}, nil
}

func (f *File) Delete(_ context.Context) error {
	if err := os.Remove(f.FilePath); err != nil {
		return fmt.Errorf("failed to delete file %s: %w", f.FilePath, err)
	}
	return nil
}
