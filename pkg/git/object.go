package git

import (
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
)

var ErrObjectNotFound = errors.New("object not found")

type objectStorage struct {
	fs FS
}

func newObjectStorage(fs FS) *objectStorage {
	return &objectStorage{fs: fs}
}

func (s *objectStorage) Get(hash hash) (io.ReadCloser, error) {
	r, err := s.getRaw(hash)
	if err != nil {
		return nil, err
	}
	return &lazyCloseReader{
		Reader: r,
		closer: r.(io.Closer),
	}, nil
}

func (s *objectStorage) Set(hash hash, data io.Reader) error {
	entry := s.objectPath(hash)
	if s.fs.Exist(entry) {
		return nil
	}
	if err := s.fs.Create(entry); err != nil {
		return err
	}
	f, err := s.fs.Writer(entry)
	if err != nil {
		return err
	}
	zlibWriter := zlib.NewWriter(f)
	_, err = io.Copy(zlibWriter, data)
	if err != nil {
		zlibWriter.Close()
		f.Close()
		return err
	}
	if err := zlibWriter.Close(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func (s *objectStorage) Has(hash hash) bool {
	return s.fs.Exist(s.objectPath(hash))
}

func (s *objectStorage) Iter() (*hashIter, error) {
	objectsDir := "objects"
	if !s.fs.Exist(objectsDir) {
		return &hashIter{}, nil
	}
	dirs, err := s.fs.Listdir(objectsDir)
	if err != nil {
		return nil, err
	}
	ch := make(chan hash, 64)
	iter := &hashIter{ch: ch}
	go func() {
		defer close(ch)
		for _, dir := range dirs {
			if len(dir) != 2 {
				continue
			}
			files, err := s.fs.Listdir(path.Join(objectsDir, dir))
			if err != nil {
				continue
			}
			for _, file := range files {
				h, err := hashFromHex(algoSHA1, dir+file)
				if err != nil {
					continue
				}
				ch <- h
			}
		}
	}()
	return iter, nil
}

func (s *objectStorage) objectPath(hash hash) string {
	hex := hash.String()
	return path.Join("objects", hex[:2], hex[2:])
}

type hashIter struct {
	ch chan hash
}

func (i *hashIter) Next() (hash, error) {
	h, ok := <-i.ch
	if !ok {
		return hash{}, io.EOF
	}
	return h, nil
}

func (i *hashIter) ForEach(fn func(hash) error) error {
	for {
		h, err := i.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := fn(h); err != nil {
			return err
		}
	}
}

func (i *hashIter) Close() error {
	return nil
}

func (s *objectStorage) getRaw(hash hash) (io.Reader, error) {
	entry := s.objectPath(hash)
	f, err := s.fs.Reader(entry)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrObjectNotFound
		}
		return nil, err
	}
	zlibReader, err := zlib.NewReader(f)
	if err != nil {
		f.Close()
		return nil, err
	}
	return &rawObjectReader{
		Reader:    zlibReader,
		closeFunc: f.Close,
	}, nil
}

type rawObjectReader struct {
	io.Reader
	closeFunc func() error
}

func (r *rawObjectReader) Close() error {
	return r.closeFunc()
}

type lazyCloseReader struct {
	io.Reader
	closer io.Closer
}

func (r *lazyCloseReader) Close() error {
	return r.closer.Close()
}

type sizeCheckReader struct {
	data   io.Reader
	size   int64
	readed bool
}

func (r *sizeCheckReader) Read(p []byte) (n int, err error) {
	n, err = r.data.Read(p)
	r.size += int64(n)
	if err == nil && !r.readed && r.size > 10*1024*1024*1024 {
		return 0, fmt.Errorf("object too large")
	}
	r.readed = true
	return n, err
}
