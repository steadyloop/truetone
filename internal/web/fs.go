package web

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileSystemStore uses file systems to implement ContentStore.
// Files live at {base}/{id}/{filename}.
type FileSystemStore struct {
	base string
}

// Ensures FileSystemStore implements ContentStore
var _ ContentStore = (*FileSystemStore)(nil)

func NewFileSystemStore(path string) *FileSystemStore {
	os.MkdirAll(path, 0755)
	return &FileSystemStore{base: path}
}

// Read returns the contents of {base}/{id}/{filename}.
func (s *FileSystemStore) Read(id string, filename string) ([]byte, error) {
	path := filepath.Join(s.base, id, filename)
	video, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return video, nil
}

// Write writes data to {base}/{id}/{filename}, creating the video's directory if needed.
func (s *FileSystemStore) Write(id string, filename string, data []byte) error {
	// 1. Creates directory with corresponding id
	path := filepath.Join(s.base, id)
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	// 2. Actually writes it out
	if err := os.WriteFile(filepath.Join(path, filename), data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}
	return nil
}
