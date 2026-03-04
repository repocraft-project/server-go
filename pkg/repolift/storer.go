package repolift

import "github.com/go-git/go-git/v6/storage"

var (
	_ storage.Storer = (*fstorer)(nil)
)

type fstorer struct {
	rootFS FS
}

func newStorer(rootFS FS) *fstorer {
	return &fstorer{rootFS: rootFS}
}
