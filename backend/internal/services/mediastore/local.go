package mediastore

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// BlobStorage exposes the S3-compatible interface for all media handles.
type BlobStorage interface {
	Driver() string
	Save(id string, reader io.Reader) (string, error)
	Delete(id string) error
	GetURL(id string) string
	Open(id string) (io.ReadCloser, error)
}

type Config struct {
	Driver string

	LocalPath string
	BaseURL   string

	S3 S3Config
}

func New(ctx context.Context, cfg Config) (BlobStorage, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Driver)) {
	case "", "local":
		if err := os.MkdirAll(filepath.Clean(cfg.LocalPath), 0755); err != nil {
			return nil, err
		}
		return NewLocalStorage(cfg.LocalPath, cfg.BaseURL), nil
	case "s3":
		return NewS3Storage(ctx, cfg.S3)
	default:
		return nil, fmt.Errorf("unsupported storage driver %q", cfg.Driver)
	}
}

type LocalStorage struct {
	baseDir string
	baseURL string
}

func NewLocalStorage(baseDir string, baseURL string) *LocalStorage {
	return &LocalStorage{
		baseDir: baseDir,
		baseURL: baseURL,
	}
}

func (s *LocalStorage) Driver() string {
	return "local"
}

func (s *LocalStorage) Save(id string, reader io.Reader) (string, error) {
	path := filepath.Join(s.baseDir, id)

	outFile, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, reader); err != nil {
		return "", err
	}

	return path, nil
}

func (s *LocalStorage) Delete(id string) error {
	path := filepath.Join(s.baseDir, id)
	return os.Remove(path)
}

// GetURL returns the accessible URL for the media asset.
// Example: baseURL could be "/media" mapping to a static Echo route.
func (s *LocalStorage) GetURL(id string) string {
	return s.baseURL + "/" + id
}

func (s *LocalStorage) Open(id string) (io.ReadCloser, error) {
	path := filepath.Join(s.baseDir, id)
	return os.Open(path)
}
