package web

import (
	"fmt"
	"os"
	"path/filepath"
)

// FSVideoContentService implements VideoContentService using the local filesystem.
// Files live at {baseDir}/{videoId}/{filename}.
type FileSystemStore struct {
	baseDir string
}

// Uncomment the following line to ensure FSVideoContentService implements VideoContentService
var _ ContentStore = (*FileSystemStore)(nil)

func NewFSVideoContentService(path string) *FileSystemStore {
	return &FileSystemStore{baseDir: path}
}

// Read returns the contents of {baseDir}/{id}/{filename}.
func (s *FileSystemStore) Read(id string, filename string) ([]byte, error) {
	path := filepath.Join(s.baseDir, id, filename)
	video, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return video, nil
}

// Write writes data to {baseDir}/{videoId}/{filename}, creating the video's directory if needed.
func (s *FileSystemStore) Write(id string, filename string, data []byte) error {
	// 1. Creates directory with corresponding id
	path := filepath.Join(s.baseDir, id)
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	// 2. Actually writes it out
	if err := os.WriteFile(filepath.Join(path, filename), data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}
	return nil
}
