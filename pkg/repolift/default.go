package repolift

import (
	"context"
	"io"

	"github.com/go-git/go-billy/v6"
	"github.com/go-git/go-git/v6/plumbing/cache"
	"github.com/go-git/go-git/v6/plumbing/transport"
	"github.com/go-git/go-git/v6/storage"
	"github.com/go-git/go-git/v6/storage/filesystem"
)

type DefaultTransferer struct {
	fs billy.Filesystem
}

func NewDefaultTransferer(fs billy.Filesystem) *DefaultTransferer {
	return &DefaultTransferer{fs: fs}
}

func (t *DefaultTransferer) UploadPack(
	ctx context.Context,
	repo string,
	input io.ReadCloser,
	output io.WriteCloser,
	options *UploadPackOptions,
) error {
	sto, err := t.openRepo(repo)
	if err != nil {
		return err
	}

	return transport.UploadPack(ctx, sto, input, output, nil)
}

func (t *DefaultTransferer) ReceivePack(
	ctx context.Context,
	repo string,
	input io.ReadCloser,
	output io.WriteCloser,
	options *ReceivePackOptions,
) error {
	sto, err := t.openRepo(repo)
	if err != nil {
		return err
	}

	return transport.ReceivePack(ctx, sto, input, output, nil)
}

func (t *DefaultTransferer) openRepo(repo string) (storage.Storer, error) {
	if _, err := t.fs.Stat(repo); err != nil {
		return nil, err
	}

	fs, err := t.fs.Chroot(repo)
	if err != nil {
		return nil, err
	}

	return filesystem.NewStorage(fs, cache.NewObjectLRUDefault()), nil
}
