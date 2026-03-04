package repolift

import "path"

var (
	_ FS = (*subFS)(nil)
)

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

func (s *subFS) Delete(entry string) error {
	return s.fs.Delete(path.Join(s.root, entry))
}

func (s *subFS) Reader(entry string) (ReadonlyFile, error) {
	return s.fs.Reader(path.Join(s.root, entry))
}

func (s *subFS) Writer(entry string) (WriteonlyFile, error) {
	return s.fs.Writer(path.Join(s.root, entry))
}

func (s *subFS) List(entry string) ([]string, error) {
	return s.fs.List(path.Join(s.root, entry))
}
