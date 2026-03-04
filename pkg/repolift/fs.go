package repolift

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
	Delete(entry string) error
	List(entry string) ([]string, error)
	Reader(entry string) (ReadonlyFile, error)
	Writer(entry string) (WriteonlyFile, error)
}
