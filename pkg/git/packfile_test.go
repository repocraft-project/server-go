package git

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"testing"

	"github.com/repocraft-project/server-go/internal/gitcmd"
)

func TestPackfileHeader_WriteAndRead(t *testing.T) {
	tests := []struct {
		name    string
		version uint32
		count   uint32
	}{
		{"version 2 count 0", 2, 0},
		{"version 2 count 1", 2, 1},
		{"version 2 count 100", 2, 100},
		{"version 3 count 1", 3, 1},
		{"version 3 count 1000", 3, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := writePackfileHeader(&buf, tt.version, tt.count)
			if err != nil {
				t.Fatal(err)
			}

			version, count, err := readPackfileHeader(bytes.NewReader(buf.Bytes()))
			if err != nil {
				t.Fatal(err)
			}

			if version != tt.version {
				t.Errorf("version = %d, want %d", version, tt.version)
			}
			if count != tt.count {
				t.Errorf("count = %d, want %d", count, tt.count)
			}
		})
	}
}

func TestObjectHeader_WriteAndRead(t *testing.T) {
	tests := []struct {
		name    string
		objType byte
		size    uint64
	}{
		{"small blob", typeBlob, 10},
		{"size 14", typeBlob, 14},
		{"size 15 threshold", typeBlob, 15},
		{"size 16", typeBlob, 16},
		{"size 127", typeBlob, 127},
		{"size 128 threshold", typeBlob, 128},
		{"size 129", typeBlob, 129},
		{"size 1000", typeBlob, 1000},
		{"size 10000", typeBlob, 10000},
		{"large size", typeBlob, 1 << 20},
		{"tree type", typeTree, 100},
		{"commit type", typeCommit, 200},
		{"tag type", typeTag, 150},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := writeObjectHeader(&buf, tt.objType, tt.size)
			if err != nil {
				t.Fatal(err)
			}

			t.Logf("size %d: wrote %x", tt.size, buf.Bytes())

			objType, size, err := readObjectHeader(bytes.NewReader(buf.Bytes()))
			if err != nil {
				t.Fatal(err)
			}

			if objType != tt.objType {
				t.Errorf("objType = %d, want %d", objType, tt.objType)
			}
			if size != tt.size {
				t.Errorf("size = %d, want %d", size, tt.size)
			}
		})
	}
}

func TestDetectObjectType(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wantType byte
	}{
		{"blob", []byte("blob 12\000hello world"), typeBlob},
		{"commit", []byte("commit 100\ntree abc\nparent def"), typeCommit},
		{"tree", []byte("tree 50\x00"), typeTree},
		{"tag", []byte("tag 50\x00object commit"), typeTag},
		{"invalid short", []byte("blob"), 0},
		{"invalid", []byte("invalid data"), 0},
		{"empty", []byte{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectObjectType(tt.data)
			if got != tt.wantType {
				t.Errorf("detectObjectType() = %d, want %d", got, tt.wantType)
			}
		})
	}
}

func TestEncodePackfile_SingleBlob(t *testing.T) {
	tmp := t.TempDir()
	if err := gitcmd.Init(tmp); err != nil {
		t.Fatal(err)
	}

	gitDir := filepath.Join(tmp, ".git")
	fs := NewLocalFS(gitDir)
	storage := newObjectStorage(fs)

	content := []byte("hello world\n")
	_, err := gitcmd.HashObject(content)
	if err != nil {
		t.Fatal(err)
	}

	// Create the object in Git format (with header)
	objData := []byte(fmt.Sprintf("blob %d\x00%s", len(content), content))
	objHash, err := hashFromReader(algoSHA1, bytes.NewReader(objData))
	if err != nil {
		t.Fatal(err)
	}

	if err := storage.Set(objHash, bytes.NewReader(objData)); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	packHash, err := encodePackfile(&buf, storage, []hash{objHash})
	if err != nil {
		t.Fatal(err)
	}

	if packHash.IsZero() {
		t.Error("packfile hash should not be zero")
	}

	data := buf.Bytes()
	if len(data) < 12 {
		t.Errorf("packfile too small: %d bytes", len(data))
	}

	if string(data[:4]) != "PACK" {
		t.Error("packfile should start with PACK")
	}
}

func TestEncodePackfile_MultipleBlobs(t *testing.T) {
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
		objData := []byte(fmt.Sprintf("blob %d\x00%s", len(content), content))
		objHash, err := hashFromReader(algoSHA1, bytes.NewReader(objData))
		if err != nil {
			t.Fatal(err)
		}
		if err := storage.Set(objHash, bytes.NewReader(objData)); err != nil {
			t.Fatal(err)
		}
		hashes[i] = objHash
	}

	var buf bytes.Buffer
	_, err := encodePackfile(&buf, storage, hashes)
	if err != nil {
		t.Fatal(err)
	}

	data := buf.Bytes()
	if len(data) < 12 {
		t.Errorf("packfile too small: %d bytes", len(data))
	}
}

func TestEncodePackfile_Empty(t *testing.T) {
	tmp := t.TempDir()
	if err := gitcmd.Init(tmp); err != nil {
		t.Fatal(err)
	}

	fs := NewLocalFS(filepath.Join(tmp, ".git"))
	storage := newObjectStorage(fs)

	var buf bytes.Buffer
	h, err := encodePackfile(&buf, storage, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !h.IsZero() {
		t.Error("empty packfile hash should be zero")
	}

	if buf.Len() != 0 {
		t.Errorf("empty packfile should have 0 bytes, got %d", buf.Len())
	}
}

func TestDecodePackfile_Basic(t *testing.T) {
	tmp := t.TempDir()
	if err := gitcmd.Init(tmp); err != nil {
		t.Fatal(err)
	}

	fs := NewLocalFS(filepath.Join(tmp, ".git"))
	storage := newObjectStorage(fs)

	content := []byte("hello world\n")
	objData := []byte(fmt.Sprintf("blob %d\x00%s", len(content), content))
	objHash, err := hashFromReader(algoSHA1, bytes.NewReader(objData))
	if err != nil {
		t.Fatal(err)
	}

	if err := storage.Set(objHash, bytes.NewReader(objData)); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	_, err = encodePackfile(&buf, storage, []hash{objHash})
	if err != nil {
		t.Fatal(err)
	}

	newStorage := newObjectStorage(fs)
	_, err = decodePackfile(bytes.NewReader(buf.Bytes()), newStorage)
	if err != nil {
		t.Fatal(err)
	}

	if !newStorage.Has(objHash) {
		t.Error("decoded object should exist in storage")
	}

	r, err := newStorage.Get(objHash)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	decoded, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(decoded, objData) {
		t.Errorf("decoded object content = %q, want %q", decoded, objData)
	}
}

func TestEncodeDecode_Loop(t *testing.T) {
	tmp := t.TempDir()
	if err := gitcmd.Init(tmp); err != nil {
		t.Fatal(err)
	}

	fs := NewLocalFS(filepath.Join(tmp, ".git"))
	storage := newObjectStorage(fs)

	contents := [][]byte{
		[]byte("hello world\n"),
		[]byte("content a\n"),
		[]byte("content b\n"),
		[]byte("some other content\n"),
	}

	hashes := make([]hash, len(contents))
	for i, content := range contents {
		objData := []byte(fmt.Sprintf("blob %d\x00%s", len(content), content))
		objHash, err := hashFromReader(algoSHA1, bytes.NewReader(objData))
		if err != nil {
			t.Fatal(err)
		}
		if err := storage.Set(objHash, bytes.NewReader(objData)); err != nil {
			t.Fatal(err)
		}
		hashes[i] = objHash
	}

	var buf bytes.Buffer
	_, err := encodePackfile(&buf, storage, hashes)
	if err != nil {
		t.Fatal(err)
	}

	newStorage := newObjectStorage(fs)
	_, err = decodePackfile(bytes.NewReader(buf.Bytes()), newStorage)
	if err != nil {
		t.Fatal(err)
	}

	for _, h := range hashes {
		if !newStorage.Has(h) {
			t.Errorf("object %s should exist after decode", h)
		}

		r, err := newStorage.Get(h)
		if err != nil {
			t.Fatal(err)
		}
		r.Close()
	}
}
