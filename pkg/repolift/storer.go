package repolift

import "github.com/go-git/go-git/v6/storage"

var (
	_ storage.Storer = (*storer)(nil)
)

type storer struct {
	rootFS FS
}

func newStorer(rootFS FS) *storer {
	return &storer{rootFS: rootFS}
}
