package hasher

import (
	"context"
	"crypto/md5"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"os"
)

type (
	File struct {
		tracer   trace.Tracer
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
	return &File{tracer: otel.Tracer("hasher.file"), FilePath: filePath, MD5Hash: fmt.Sprintf("%x", hash)}, nil
}

func (f *File) Delete(ctx context.Context) error {
	_, span := f.tracer.Start(ctx, "delete")
	defer span.End()
	if err := os.Remove(f.FilePath); err != nil {
		return fmt.Errorf("failed to delete file %s: %w", f.FilePath, err)
	}
	return nil
}
