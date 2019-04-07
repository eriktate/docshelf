package mock

import (
	"errors"
	"strings"
)

// FileStore is a mock implementation of the FileStore interface.
type FileStore struct {
	files map[string][]byte

	ReadFileCalled   int
	WriteFileCalled  int
	RemoveFileCalled int
	ListDirCalled    int
	ForceError       bool
}

// NewFileStore returns a new mock FileStore struct.
func NewFileStore() *FileStore {
	return &FileStore{files: make(map[string][]byte)}
}

// ReadFile implements the FileStore interface.
func (m *FileStore) ReadFile(path string) ([]byte, error) {
	if m.ForceError {
		return nil, errors.New("forced error")
	}

	return m.files[path], nil
}

// WriteFile implements the FileStore interface.
func (m *FileStore) WriteFile(path string, data []byte) error {
	if m.ForceError {
		return errors.New("forced error")
	}

	m.files[path] = data
	return nil
}

// RemoveFile implements the FileStore interface.
func (m *FileStore) RemoveFile(path string) error {
	if m.ForceError {
		return errors.New("forced error")
	}

	delete(m.files, path)
	return nil
}

// ListDir implements the FileStore interface.
func (m *FileStore) ListDir(path string) ([]string, error) {
	if m.ForceError {
		return nil, errors.New("forced error")
	}

	listing := make([]string, 0)
	for k := range m.files {
		if strings.HasPrefix(k, path) {
			listing = append(listing, k)
		}
	}

	return listing, nil
}
