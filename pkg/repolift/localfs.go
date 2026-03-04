package repolift

import (
	"os"
	"path"
)

var (
	_ FS = (*LocalFS)(nil)
)

type LocalFS struct {
	root string
}

func NewLocalFS(root string) *LocalFS {
	return &LocalFS{root: root}
}

func (fs *LocalFS) Create(entry string) error {
	fullPath := fs.fullPath(entry)
	dir := path.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	return f.Close()
}

func (fs *LocalFS) Delete(entry string) error {
	fullPath := fs.fullPath(entry)
	return os.Remove(fullPath)
}

func (fs *LocalFS) List(entry string) ([]string, error) {
	fullPath := fs.fullPath(entry)

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, nil
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}

	return names, nil
}

func (fs *LocalFS) Reader(entry string) (ReadonlyFile, error) {
	fullPath := fs.fullPath(entry)
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs *LocalFS) Writer(entry string) (WriteonlyFile, error) {
	fullPath := fs.fullPath(entry)
	dir := path.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs *LocalFS) fullPath(entry string) string {
	if entry == "" {
		return fs.root
	}
	return path.Join(fs.root, entry)
}
