package repolift

import (
	"context"
	"io"
)

type Transferer interface {
	UploadPack(
		ctx context.Context,
		repo string,
		input io.ReadCloser,
		writer io.WriteCloser,
		options *UploadPackOptions,
	) error

	ReceivePack(
		ctx context.Context,
		repo string,
		input io.ReadCloser,
		writer io.WriteCloser,
		options *ReceivePackOptions,
	) error
}

type UploadPackOptions struct{}
type ReceivePackOptions struct{}
