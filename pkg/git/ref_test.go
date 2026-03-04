package git

import (
	"path/filepath"
	"testing"

	"github.com/repocraft-project/server-go/internal/gitcmd"
)

func TestRefStorage_SetAndGet(t *testing.T) {
	tmp := t.TempDir()
	if err := gitcmd.Init(tmp); err != nil {
		t.Fatal(err)
	}

	fs := NewLocalFS(filepath.Join(tmp, ".git"))
	storage := newRefStorage(fs)

	content := []byte("hello world\n")
	gitHash, err := gitcmd.HashObject(content)
	if err != nil {
		t.Fatal(err)
	}

	h, err := hashFromHex(algoSHA1, gitHash)
	if err != nil {
		t.Fatal(err)
	}

	err = storage.Set("refs/heads/test", h)
	if err != nil {
		t.Fatal(err)
	}

	gitHash, err = gitcmd.RevParseWithDir(filepath.Join(tmp, ".git"), "refs/heads/test")
	if err != nil {
		t.Fatal(err)
	}

	gitHashExpected, _ := hashFromHex(algoSHA1, gitHash)
	if !h.Equal(gitHashExpected) {
		t.Errorf("hash mismatch: ours=%s, git=%s", h, gitHashExpected)
	}
}

func TestRefStorage_Get_NotFound(t *testing.T) {
	tmp := t.TempDir()
	if err := gitcmd.Init(tmp); err != nil {
		t.Fatal(err)
	}

	fs := NewLocalFS(filepath.Join(tmp, ".git"))
	storage := newRefStorage(fs)

	_, err := storage.Get("refs/heads/nonexistent")
	if err != ErrRefNotFound {
		t.Errorf("expected ErrRefNotFound, got %v", err)
	}
}

func TestRefStorage_Delete(t *testing.T) {
	tmp := t.TempDir()
	if err := gitcmd.Init(tmp); err != nil {
		t.Fatal(err)
	}

	fs := NewLocalFS(filepath.Join(tmp, ".git"))
	storage := newRefStorage(fs)

	content := []byte("hello world\n")
	gitHash, err := gitcmd.HashObject(content)
	if err != nil {
		t.Fatal(err)
	}

	h, err := hashFromHex(algoSHA1, gitHash)
	if err != nil {
		t.Fatal(err)
	}

	err = storage.Set("refs/heads/to-delete", h)
	if err != nil {
		t.Fatal(err)
	}

	err = storage.Delete("refs/heads/to-delete")
	if err != nil {
		t.Fatal(err)
	}

	_, err = storage.Get("refs/heads/to-delete")
	if err != ErrRefNotFound {
		t.Errorf("expected ErrRefNotFound after delete, got %v", err)
	}
}

func TestRefStorage_Iter(t *testing.T) {
	tmp := t.TempDir()
	if err := gitcmd.Init(tmp); err != nil {
		t.Fatal(err)
	}

	fs := NewLocalFS(filepath.Join(tmp, ".git"))
	storage := newRefStorage(fs)

	contents := [][]byte{
		[]byte("commit a\n"),
		[]byte("commit b\n"),
		[]byte("commit c\n"),
	}

	refs := []string{
		"refs/heads/branch-a",
		"refs/heads/branch-b",
		"refs/heads/branch-c",
	}

	for i, content := range contents {
		gitHash, err := gitcmd.HashObject(content)
		if err != nil {
			t.Fatal(err)
		}
		h, err := hashFromHex(algoSHA1, gitHash)
		if err != nil {
			t.Fatal(err)
		}
		err = storage.Set(refs[i], h)
		if err != nil {
			t.Fatal(err)
		}
	}

	iter, err := storage.Iter()
	if err != nil {
		t.Fatal(err)
	}

	found := make(map[string]bool)
	iter.ForEach(func(r ref) error {
		found[r.Name] = true
		return nil
	})

	for _, ref := range refs {
		if !found[ref] {
			t.Errorf("ref %s not found in iteration", ref)
		}
	}
}
