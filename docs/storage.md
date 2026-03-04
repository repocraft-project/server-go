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

func (s *objectStorage) get(hash Hash) (io.Reader, error)
func (s *objectStorage) set(hash Hash, data io.Reader) error
func (s *objectStorage) has(hash Hash) bool
func (s *objectStorage) iter() (hashIter, error)
func (s *objectStorage) packfileWriter() (io.WriteCloser, error)
```

- **get**: Returns zlib-compressed raw object data
- **set**: Stores zlib-compressed object data
- **has**: Checks if object exists
- **iter**: Iterates over all object hashes
- **packfileWriter**: Returns writer for creating packfiles

## refStorage (in ref.go)

Git references (branches, tags) are stored in:
- Loose refs: `refs/{category}/{name}` - one file per reference
- Packed refs: `packed-refs` - all refs packed into one file

```go
type refStorage struct {
    fs FS
}

func newRefStorage(fs FS) *refStorage

func (s *refStorage) get(refName string) (Hash, error)
func (s *refStorage) set(refName string, hash Hash) error
func (s *refStorage) delete(refName string) error
func (s *refStorage) iter() (refIter, error)
func (s *refStorage) pack() error
```

- **get**: Returns the hash for a reference
- **set**: Sets a reference to a hash
- **delete**: Removes a reference
- **iter**: Iterates over all references
- **pack**: Writes all loose refs into packed-refs

## Hash Type

```go
type Hash [20]byte
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
