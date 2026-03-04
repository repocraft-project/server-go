# Storage Design

The storage layer provides internal helper structs for the Transferer. These are NOT exported to users—they are implementation details.

## subFS

The subFS wraps an FS interface with a path prefix, allowing each repository to have its own isolated view:

```go
type subFS struct {
    root string  // e.g., "user/repo/objects" or "user/repo/refs"
    fs   FS
}

func newSubFS(fs FS, root string) *subFS

func (s *subFS) Create(entry string) error {
    return s.fs.Create(path.Join(s.root, entry))
}

func (s *subFS) Reader(entry string) (ReadonlyFile, error) {
    return s.fs.Reader(path.Join(s.root, entry))
}

// ... all other FS methods forwarded with path.Join
```

## objectStorage (in object.go)

Git objects are stored in `objects/{first-two-hash}/{rest-of-hash}` format. The server doesn't need to understand object types—it only stores and retrieves raw compressed data.

```go
type objectStorage struct {
    fs FS
}

func newObjectStorage(fs FS) *objectStorage

func (s *objectStorage) Get(hash hash) (io.ReadCloser, error)
func (s *objectStorage) Set(hash hash, data io.Reader) error
func (s *objectStorage) Has(hash hash) bool
func (s *objectStorage) Iter() (*hashIter, error)
```

- **Get**: Returns zlib-compressed raw object data
- **Set**: Stores zlib-compressed object data
- **Has**: Checks if object exists
- **Iter**: Iterates over all object hashes

## refStorage (in ref.go)

Git references (branches, tags) are stored in:
- Loose refs: `refs/{category}/{name}` - one file per reference
- Packed refs: `packed-refs` - all refs packed into one file

```go
type refStorage struct {
    fs FS
}

func newRefStorage(fs FS) *refStorage

func (s *refStorage) Get(refName string) (hash, error)
func (s *refStorage) Set(refName string, hash hash) error
func (s *refStorage) Delete(refName string) error
func (s *refStorage) Iter() (*refIter, error)
func (s *refStorage) Pack() error
```

- **Get**: Returns the hash for a reference
- **Set**: Sets a reference to a hash
- **Delete**: Removes a reference
- **Iter**: Iterates over all references
- **Pack**: Writes all loose refs into packed-refs

## hash Type

```go
type hashType byte

const (
    algoSHA1 hashType = iota
    algoSHA256
)

type hash struct {
    alg      hashType
    checksum [32]byte
}
```

## Usage in Transferer

```go
func (t *Transferer) ReceivePack(ctx context.Context, pat string, r io.Reader, w io.Writer) error {
    objects := newObjectStorage(newSubFS(t.fs, path.Join(pat, "objects")))
    refs := newRefStorage(newSubFS(t.fs, path.Join(pat, "refs")))
    // ...
}
```

The Transferer creates subFS instances for each repository path (pat), passing them to the internal objectStorage (object.go) and refStorage (ref.go).
