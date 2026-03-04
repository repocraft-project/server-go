package git

import (
	"os"
	"path/filepath"
)

type localfs struct {
	root string
}

func NewLocalFS(root string) FS {
	return &localfs{
		root: root,
	}
}

func (fs *localfs) Create(entry string) error {
	fullPath := filepath.Join(fs.root, entry)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	return f.Close()
}

func (fs *localfs) Reader(entry string) (ReadonlyFile, error) {
	fullPath := filepath.Join(fs.root, entry)
	return os.Open(fullPath)
}

func (fs *localfs) Writer(entry string) (WriteonlyFile, error) {
	fullPath := filepath.Join(fs.root, entry)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
}

func (fs *localfs) Delete(entry string) error {
	fullPath := filepath.Join(fs.root, entry)
	return os.Remove(fullPath)
}

func (fs *localfs) Rename(src, dst string) error {
	oldPath := filepath.Join(fs.root, src)
	newPath := filepath.Join(fs.root, dst)
	return os.Rename(oldPath, newPath)
}

func (fs *localfs) Listdir(entry string) ([]string, error) {
	fullPath := filepath.Join(fs.root, entry)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names, nil
}

func (fs *localfs) Exist(entry string) bool {
	fullPath := filepath.Join(fs.root, entry)
	_, err := os.Stat(fullPath)
	return err == nil
}
