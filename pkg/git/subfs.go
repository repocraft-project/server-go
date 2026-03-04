package git

import "path"

type subFS struct {
	root string
	fs   FS
}

func newSubFS(fs FS, root string) *subFS {
	return &subFS{
		root: root,
		fs:   fs,
	}
}

func (s *subFS) Create(entry string) error {
	return s.fs.Create(path.Join(s.root, entry))
}

func (s *subFS) Reader(entry string) (ReadonlyFile, error) {
	return s.fs.Reader(path.Join(s.root, entry))
}

func (s *subFS) Writer(entry string) (WriteonlyFile, error) {
	return s.fs.Writer(path.Join(s.root, entry))
}

func (s *subFS) Delete(entry string) error {
	return s.fs.Delete(path.Join(s.root, entry))
}

func (s *subFS) Rename(src, dst string) error {
	return s.fs.Rename(path.Join(s.root, src), path.Join(s.root, dst))
}

func (s *subFS) Listdir(entry string) ([]string, error) {
	return s.fs.Listdir(path.Join(s.root, entry))
}

func (s *subFS) Exist(entry string) bool {
	return s.fs.Exist(path.Join(s.root, entry))
}
