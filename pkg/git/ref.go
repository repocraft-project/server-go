package git

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path"
	"sort"
	"strings"
)

var ErrRefNotFound = errors.New("ref not found")

type ref struct {
	Name   string
	Target hash
	Peeled hash
}

type refStorage struct {
	fs FS
}

func newRefStorage(fs FS) *refStorage {
	return &refStorage{fs: fs}
}

func (s *refStorage) Get(refName string) (hash, error) {
	if refs, err := s.readPackedRefs(); err == nil {
		if r, ok := refs[refName]; ok {
			return r.Target, nil
		}
	}

	entry := s.refPath(refName)
	f, err := s.fs.Reader(entry)
	if err != nil {
		if os.IsNotExist(err) {
			return hash{}, ErrRefNotFound
		}
		return hash{}, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		hex := strings.TrimSpace(scanner.Text())
		if len(hex) == 40 {
			return hashFromHex(algoSHA1, hex)
		}
		if len(hex) == 64 {
			return hashFromHex(algoSHA256, hex)
		}
	}
	return hash{}, errors.New("invalid ref format")
}

func (s *refStorage) Set(refName string, hash hash) error {
	entry := s.refPath(refName)
	parts := strings.Split(refName, "/")
	dir := path.Join(parts[:len(parts)-1]...)
	if dir != "" {
		if err := s.fs.Create(path.Join("refs", dir)); err != nil {
			if !os.IsExist(err) {
				return err
			}
		}
	}
	f, err := s.fs.Writer(entry)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(hash.String() + "\n"))
	if err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func (s *refStorage) Delete(refName string) error {
	entry := s.refPath(refName)
	if !s.fs.Exist(entry) {
		return nil
	}
	return s.fs.Delete(entry)
}

func (s *refStorage) Iter() (*refIter, error) {
	refs := make(map[string]hash)

	if packedRefsExist := s.fs.Exist("packed-refs"); packedRefsExist {
		if packed, err := s.readPackedRefs(); err == nil {
			for k, v := range packed {
				refs[k] = v.Target
			}
		}
	}

	refsDir := "refs"
	if s.fs.Exist(refsDir) {
		if err := s.walkRefs(refsDir, "", refs); err != nil {
			return nil, err
		}
	}

	ch := make(chan ref, 64)
	iter := &refIter{ch: ch}

	go func() {
		defer close(ch)
		names := make([]string, 0, len(refs))
		for name := range refs {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			ch <- ref{Name: name, Target: refs[name]}
		}
	}()

	return iter, nil
}

func (s *refStorage) walkRefs(dir, prefix string, refs map[string]hash) error {
	entries, err := s.fs.Listdir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		fullPath := path.Join(dir, entry)
		_, err := s.fs.Listdir(fullPath)
		if errors.Is(err, ErrNotDirectory) {
			refName := path.Join("refs", prefix, entry)
			if h, err := s.Get(refName); err == nil {
				refs[refName] = h
			}
		} else if err != nil {
			return err
		} else {
			newPrefix := path.Join(prefix, entry)
			if err := s.walkRefs(fullPath, newPrefix, refs); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *refStorage) Pack() error {
	refs := make(map[string]hash)
	if err := s.walkRefs("refs", "", refs); err != nil {
		return err
	}

	var buf bytes.Buffer
	names := make([]string, 0, len(refs))
	for name := range refs {
		names = append(names, name)
	}
	sort.Strings(names)
	buf.WriteString("# packed-refs with: peeled ")
	buf.WriteString(string(algoSHA1))
	buf.WriteString("\n")
	for _, name := range names {
		buf.WriteString(refs[name].String())
		buf.WriteString(" ")
		buf.WriteString(name)
		buf.WriteString("\n")
	}

	f, err := s.fs.Writer("packed-refs")
	if err != nil {
		return err
	}
	_, err = buf.WriteTo(f)
	if err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func (s *refStorage) readPackedRefs() (map[string]ref, error) {
	refs := make(map[string]ref)
	f, err := s.fs.Reader("packed-refs")
	if err != nil {
		if os.IsNotExist(err) {
			return refs, nil
		}
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}
		hashStr := parts[0]
		refName := parts[1]
		var h hash
		var err error
		if len(hashStr) == 40 {
			h, err = hashFromHex(algoSHA1, hashStr)
		} else if len(hashStr) == 64 {
			h, err = hashFromHex(algoSHA256, hashStr)
		} else {
			continue
		}
		if err != nil {
			continue
		}
		refs[refName] = ref{Name: refName, Target: h}
	}
	return refs, scanner.Err()
}

func (s *refStorage) refPath(refName string) string {
	return refName
}

type refIter struct {
	ch chan ref
}

func (i *refIter) Next() (ref, error) {
	r, ok := <-i.ch
	if !ok {
		return ref{}, io.EOF
	}
	return r, nil
}

func (i *refIter) ForEach(fn func(ref) error) error {
	for {
		r, err := i.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := fn(r); err != nil {
			return err
		}
	}
}

func (i *refIter) Close() error {
	return nil
}
