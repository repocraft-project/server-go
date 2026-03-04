package git

import "io"

type ReadonlyFile interface {
	io.Reader
	io.Seeker
	io.Closer
}

type WriteonlyFile interface {
	io.Writer
	io.Seeker
	io.Closer
}

type FS interface {
	Create(entry string) error
	Reader(entry string) (ReadonlyFile, error)
	Writer(entry string) (WriteonlyFile, error)
	Delete(entry string) error
	Rename(src, dst string) error
	Listdir(entry string) ([]string, error)
	Exist(entry string) bool
}
