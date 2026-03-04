package git

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	hashlib "hash"
	"io"
)

type hashType byte

const (
	algoSHA1 hashType = iota
	algoSHA256
)

type hash struct {
	alg      hashType
	checksum [32]byte
}

func (h hash) Algorithm() hashType {
	return h.alg
}

func (h hash) Bytes() []byte {
	switch h.alg {
	case algoSHA256:
		return h.checksum[:32]
	default:
		return h.checksum[:20]
	}
}

func (h hash) String() string {
	switch h.alg {
	case algoSHA256:
		return hex.EncodeToString(h.checksum[:32])
	default:
		return hex.EncodeToString(h.checksum[:20])
	}
}

func (h hash) IsZero() bool {
	switch h.alg {
	case algoSHA256:
		for _, b := range h.checksum[:32] {
			if b != 0 {
				return false
			}
		}
	default:
		for _, b := range h.checksum[:20] {
			if b != 0 {
				return false
			}
		}
	}
	return true
}

func (h hash) Equal(other hash) bool {
	if h.alg != other.alg {
		return false
	}
	switch h.alg {
	case algoSHA256:
		return h.checksum == other.checksum
	default:
		return bytes.Equal(h.checksum[:20], other.checksum[:20])
	}
}

func newHash(alg hashType) hash {
	return hash{alg: alg}
}

func hashFromBytes(alg hashType, data []byte) (hash, error) {
	h := hash{alg: alg}
	switch alg {
	case algoSHA256:
		if len(data) != 32 {
			return hash{}, errors.New("sha256 hash must be 32 bytes")
		}
		copy(h.checksum[:32], data)
	default:
		if len(data) != 20 {
			return hash{}, errors.New("sha1 hash must be 20 bytes")
		}
		copy(h.checksum[:20], data)
	}
	return h, nil
}

func hashFromHex(alg hashType, hexStr string) (hash, error) {
	var data []byte
	var err error
	switch alg {
	case algoSHA256:
		data, err = hex.DecodeString(hexStr)
		if err != nil {
			return hash{}, err
		}
		if len(data) != 32 {
			return hash{}, errors.New("sha256 hex must be 64 characters")
		}
	default:
		data, err = hex.DecodeString(hexStr)
		if err != nil {
			return hash{}, err
		}
		if len(data) != 20 {
			return hash{}, errors.New("sha1 hex must be 40 characters")
		}
	}
	return hashFromBytes(alg, data)
}

type hashReader struct {
	alg hashType
	h   hashlib.Hash
}

func newHashReader(alg hashType) *hashReader {
	var hh hashlib.Hash
	switch alg {
	case algoSHA256:
		hh = sha256.New()
	default:
		hh = sha1.New()
	}
	return &hashReader{
		alg: alg,
		h:   hh,
	}
}

func (r *hashReader) Write(p []byte) (n int, err error) {
	return r.h.Write(p)
}

func (r *hashReader) sum() (hash, error) {
	return hashFromBytes(r.alg, r.h.Sum(nil))
}

func hashFromReader(alg hashType, r io.Reader) (hash, error) {
	hr := newHashReader(alg)
	_, err := io.Copy(hr, r)
	if err != nil {
		return hash{}, err
	}
	return hr.sum()
}
