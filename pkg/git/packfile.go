package git

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	typeCommit   byte = 1
	typeTree     byte = 2
	typeBlob     byte = 3
	typeTag      byte = 4
	typeOfsDelta byte = 6
	typeRefDelta byte = 7

	packVersion = 2
)

var (
	ErrInvalidPackfile  = errors.New("invalid packfile")
	ErrChecksumMismatch = errors.New("checksum mismatch")
	ErrInvalidObject    = errors.New("invalid object")
)

type packfileHeader struct {
	signature [4]byte
	version   uint32
	count     uint32
}

func writePackfileHeader(w io.Writer, version, count uint32) error {
	h := packfileHeader{
		signature: [4]byte{'P', 'A', 'C', 'K'},
		version:   version,
		count:     count,
	}
	if err := binary.Write(w, binary.BigEndian, &h); err != nil {
		return err
	}
	return nil
}

func readPackfileHeader(r io.Reader) (version, count uint32, err error) {
	sig := make([]byte, 4)
	if _, err := io.ReadFull(r, sig); err != nil {
		return 0, 0, err
	}
	if string(sig) != "PACK" {
		return 0, 0, ErrInvalidPackfile
	}

	versionBytes := make([]byte, 4)
	if _, err := io.ReadFull(r, versionBytes); err != nil {
		return 0, 0, err
	}
	version = binary.BigEndian.Uint32(versionBytes)

	if version != 2 && version != 3 {
		return 0, 0, fmt.Errorf("%w: unsupported version %d", ErrInvalidPackfile, version)
	}

	countBytes := make([]byte, 4)
	if _, err := io.ReadFull(r, countBytes); err != nil {
		return 0, 0, err
	}
	count = binary.BigEndian.Uint32(countBytes)

	return version, count, nil
}

func writeObjectHeader(w io.Writer, objType byte, size uint64) error {
	c := objType << 4

	if size < 15 {
		c |= byte(size)
		_, err := w.Write([]byte{c})
		return err
	}

	c |= 15
	if _, err := w.Write([]byte{c}); err != nil {
		return err
	}

	size -= 15
	for size >= 128 {
		b := byte(size&0x7f) | 0x80
		if _, err := w.Write([]byte{b}); err != nil {
			return err
		}
		size >>= 7
	}
	if _, err := w.Write([]byte{byte(size)}); err != nil {
		return err
	}
	return nil
}

func readObjectHeader(r io.Reader) (objType byte, size uint64, err error) {
	var first byte
	if err := binary.Read(r, binary.BigEndian, &first); err != nil {
		return 0, 0, err
	}

	objType = (first >> 4) & 0x07
	size = uint64(first & 0x0f)

	if size < 15 {
		return objType, size, nil
	}

	shift := uint(0)
	for {
		var b byte
		if err := binary.Read(r, binary.BigEndian, &b); err != nil {
			return 0, 0, err
		}
		size += uint64(b&0x7f) << shift
		if b&0x80 == 0 {
			break
		}
		shift += 7
	}

	return objType, size, nil
}

func encodePackfile(w io.Writer, storage *objectStorage, want []hash) (hash, error) {
	if len(want) == 0 {
		return hash{}, nil
	}

	hasher := sha1.New()
	multiWriter := io.MultiWriter(w, hasher)

	if err := writePackfileHeader(multiWriter, packVersion, uint32(len(want))); err != nil {
		return hash{}, err
	}

	for _, h := range want {
		objReader, err := storage.Get(h)
		if err != nil {
			return hash{}, fmt.Errorf("get object %s: %w", h, err)
		}

		data, err := io.ReadAll(objReader)
		objReader.Close()
		if err != nil {
			return hash{}, fmt.Errorf("read object %s: %w", h, err)
		}

		objType := detectObjectType(data)
		if objType == 0 {
			return hash{}, fmt.Errorf("%w: %s", ErrInvalidObject, h)
		}

		if err := writeObjectHeader(multiWriter, objType, uint64(len(data))); err != nil {
			return hash{}, err
		}

		zw := zlib.NewWriter(multiWriter)
		if _, err := zw.Write(data); err != nil {
			return hash{}, err
		}
		if err := zw.Close(); err != nil {
			return hash{}, err
		}
	}

	hashBytes := hasher.Sum(nil)
	_, err := w.Write(hashBytes)
	if err != nil {
		return hash{}, err
	}

	return hashFromBytes(algoSHA1, hashBytes)
}

func decodePackfile(r io.Reader, storage *objectStorage) (hash, error) {
	// Read all data including header and objects and checksum
	allData, err := io.ReadAll(r)
	if err != nil {
		return hash{}, err
	}

	if len(allData) < 20 {
		return hash{}, ErrInvalidPackfile
	}

	// Separate data and checksum
	dataLen := len(allData) - 20
	packData := allData[:dataLen]
	checksum := allData[dataLen:]

	// Compute hash of pack data
	hasher := sha1.New()
	hasher.Write(packData)
	expectedHash, err := hashFromBytes(algoSHA1, hasher.Sum(nil))
	if err != nil {
		return hash{}, err
	}

	// Verify checksum
	actualHash, err := hashFromBytes(algoSHA1, checksum)
	if err != nil {
		return hash{}, err
	}
	if !expectedHash.Equal(actualHash) {
		return hash{}, ErrChecksumMismatch
	}

	// Parse header and objects from packData
	dataReader := bytes.NewReader(packData)

	_, count, err := readPackfileHeader(dataReader)
	if err != nil {
		return hash{}, err
	}

	for i := uint32(0); i < count; i++ {
		_, _, err := readObjectHeader(dataReader)
		if err != nil {
			return hash{}, err
		}

		zlibReader, err := zlib.NewReader(dataReader)
		if err != nil {
			return hash{}, err
		}

		data, err := io.ReadAll(zlibReader)
		zlibReader.Close()
		if err != nil {
			return hash{}, err
		}

		objectHash, err := hashFromReader(algoSHA1, &bytesReader{data: data})
		if err != nil {
			return hash{}, err
		}

		if err := storage.Set(objectHash, &bytesReader{data: data}); err != nil {
			return hash{}, err
		}
	}

	return expectedHash, nil
}

func detectObjectType(data []byte) byte {
	if len(data) < 5 {
		return 0
	}
	switch {
	case bytesHasPrefix(data, []byte("commit ")):
		return typeCommit
	case bytesHasPrefix(data, []byte("tree ")):
		return typeTree
	case bytesHasPrefix(data, []byte("blob ")):
		return typeBlob
	case bytesHasPrefix(data, []byte("tag ")):
		return typeTag
	}
	return 0
}

func bytesHasPrefix(b, prefix []byte) bool {
	if len(b) < len(prefix) {
		return false
	}
	for i := range prefix {
		if b[i] != prefix[i] {
			return false
		}
	}
	return true
}

type bytesReader struct {
	data []byte
	pos  int
}

func (r *bytesReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (r *bytesReader) Seek(offset int64, whence int) (int64, error) {
	var newPos int64
	switch whence {
	case 0:
		newPos = offset
	case 1:
		newPos = int64(r.pos) + offset
	case 2:
		newPos = int64(len(r.data)) + offset
	default:
		return 0, errors.New("invalid whence")
	}
	if newPos < 0 {
		return 0, errors.New("negative position")
	}
	r.pos = int(newPos)
	return newPos, nil
}
