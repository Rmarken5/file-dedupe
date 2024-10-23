package hasher

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
)

type (
	PathGetter interface {
		GetFilePath() string
	}
	LineReader interface {
		ReadLine(ctx context.Context, lineNumber int) ([]byte, error)
	}
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

func (f *File) ReadOffset(ctx context.Context, offset int) ([]byte, error) {
	panic("implement me")
}

func (f *File) GetFilePath() string {
	return f.FilePath
}
