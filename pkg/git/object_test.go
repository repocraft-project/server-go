package git

import (
	"bytes"
	"io"
	"path/filepath"
	"testing"

	"github.com/repocraft-project/server-go/internal/gitcmd"
)

func TestObjectStorage_SetAndGet(t *testing.T) {
	tmp := t.TempDir()
	if err := gitcmd.Init(tmp); err != nil {
		t.Fatal(err)
	}

	fs := NewLocalFS(filepath.Join(tmp, ".git"))
	storage := newObjectStorage(fs)

	content := []byte("hello world\n")

	gitHash, err := gitcmd.HashObject(content)
	if err != nil {
		t.Fatal(err)
	}

	h, err := hashFromHex(algoSHA1, gitHash)
	if err != nil {
		t.Fatal(err)
	}

	err = storage.Set(h, bytes.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}

	if !storage.Has(h) {
		t.Error("object should exist")
	}

	r, err := storage.Get(h)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, content) {
		t.Errorf("got %s, want %s", got, content)
	}
}

func TestObjectStorage_NotFound(t *testing.T) {
	tmp := t.TempDir()
	if err := gitcmd.Init(tmp); err != nil {
		t.Fatal(err)
	}

	fs := NewLocalFS(filepath.Join(tmp, ".git"))
	storage := newObjectStorage(fs)

	h, err := hashFromHex(algoSHA1, "0000000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}

	_, err = storage.Get(h)
	if err != ErrObjectNotFound {
		t.Errorf("expected ErrObjectNotFound, got %v", err)
	}

	if storage.Has(h) {
		t.Error("object should not exist")
	}
}

func TestObjectStorage_Iter(t *testing.T) {
	tmp := t.TempDir()
	if err := gitcmd.Init(tmp); err != nil {
		t.Fatal(err)
	}

	fs := NewLocalFS(filepath.Join(tmp, ".git"))
	storage := newObjectStorage(fs)

	contents := [][]byte{
		[]byte("content a\n"),
		[]byte("content b\n"),
		[]byte("content c\n"),
	}

	hashes := make([]hash, len(contents))
	for i, content := range contents {
		gitHash, err := gitcmd.HashObject(content)
		if err != nil {
			t.Fatal(err)
		}
		h, err := hashFromHex(algoSHA1, gitHash)
		if err != nil {
			t.Fatal(err)
		}
		err = storage.Set(h, bytes.NewReader(content))
		if err != nil {
			t.Fatal(err)
		}
		hashes[i] = h
	}

	iter, err := storage.Iter()
	if err != nil {
		t.Fatal(err)
	}

	found := make(map[hash]bool)
	iter.ForEach(func(h hash) error {
		found[h] = true
		return nil
	})

	for _, h := range hashes {
		if !found[h] {
			t.Errorf("hash %s not found in iteration", h)
		}
	}
}
