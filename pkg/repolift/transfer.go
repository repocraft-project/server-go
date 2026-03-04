package repolift

import (
	"context"
	"io"
	"log"

	"github.com/go-git/go-git/v6/plumbing/transport"
)

type Service byte

const (
	UploadPackService Service = iota
	ReceivePackService
)

var (
	uploadPackOpts  = &transport.UploadPackOptions{StatelessRPC: true, AdvertiseRefs: true}
	receivePackOpts = &transport.ReceivePackOptions{StatelessRPC: true, AdvertiseRefs: true}
)

type Transferer struct {
	rootFS FS
}

func NewTransferer(rootFS FS) *Transferer {
	return &Transferer{rootFS: rootFS}
}

func (t *Transferer) storer(path string) *fsStorer {
	repoFS := newSubFS(t.rootFS, path)
	return NewStorer(repoFS)
}

func (t *Transferer) checkRepoExists(path string) error {
	fs := newSubFS(t.rootFS, path)
	entries, err := fs.List("")
	if err != nil {
		log.Printf("[checkRepoExists] path=%s error=%v\n", path, err)
		return err
	}
	if len(entries) == 0 {
		log.Printf("[checkRepoExists] path=%s empty=true\n", path)
		return transport.ErrRepositoryNotFound
	}
	log.Printf("[checkRepoExists] path=%s entries=%v\n", path, entries)
	return nil
}

func (t *Transferer) AdvertiseReferences(ctx context.Context, path string, w io.Writer, service Service) error {
	log.Printf("[AdvertiseReferences] path=%s service=%d\n", path, service)

	if err := t.checkRepoExists(path); err != nil {
		log.Printf("[AdvertiseReferences] checkRepoExists failed: %v\n", err)
		return err
	}

	st := t.storer(path)
	log.Printf("[AdvertiseReferences] storer=%T\n", st)

	var svc transport.Service
	switch service {
	case UploadPackService:
		svc = transport.UploadPackService
	case ReceivePackService:
		svc = transport.ReceivePackService
	}

	log.Printf("[AdvertiseReferences] calling transport.AdvertiseReferences with svc=%s\n", svc.Name())
	err := transport.AdvertiseReferences(ctx, st, w, svc, true)
	log.Printf("[AdvertiseReferences] result: %v\n", err)
	return err
}

func (t *Transferer) UploadPack(ctx context.Context, path string, r io.Reader, w io.Writer) error {
	log.Printf("[UploadPack] path=%s\n", path)

	if err := t.checkRepoExists(path); err != nil {
		return err
	}

	st := t.storer(path)
	err := transport.UploadPack(ctx, st, nopReadCloser{r}, nopWriteCloser{w}, uploadPackOpts)
	log.Printf("[UploadPack] result: %v\n", err)
	return err
}

func (t *Transferer) ReceivePack(ctx context.Context, path string, r io.Reader, w io.Writer) error {
	log.Printf("[ReceivePack] path=%s\n", path)

	if err := t.checkRepoExists(path); err != nil {
		return err
	}

	st := t.storer(path)
	err := transport.ReceivePack(ctx, st, nopReadCloser{r}, nopWriteCloser{w}, receivePackOpts)
	log.Printf("[ReceivePack] result: %v\n", err)
	return err
}
