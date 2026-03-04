# Protocol Design

The protocol layer implements git-upload-pack and git-receive-pack, the core Git transfer protocols. This is encapsulated in the Transferer type.

## Transferer

The Transferer is the main entry point for the protocol layer. It wraps the FS interface and provides high-level methods:

```go
type Transferer struct {
    fs FS
}

func NewTransferer(fs FS) *Transferer

func (t *Transferer) UploadPack(ctx context.Context, pat string, client io.Reader, server io.Writer) error
func (t *Transferer) ReceivePack(ctx context.Context, pat string, client io.Reader, server io.Writer) error
```

### User API

```go
fs := git.NewLocalFS("/data/git")
transferer := git.NewTransferer(fs)

transferer.UploadPack(ctx, "user/repo", r, w)   // fetch/clone
transferer.ReceivePack(ctx, "user/repo", r, w)  // push
```

## upload-pack (fetch/clone)

Steps:
1. Read client wants (requested object hashes)
2. Send server advertised refs
3. Negotiation phase (have/want exchange)
4. Generate and send packfile
5. Send done marker

## receive-pack (push)

Steps:
1. Read client refs and capabilities
2. Send server refs (advertised refs)
3. Receive packfile from client
4. Extract and store objects
5. Update references atomically
6. Send status report

## Internal Components

The Transferer uses internal (unexported) components:

- **objectStorage** (object.go): Manages Git objects (loose + packfiles)
- **refStorage** (ref.go): Manages Git references (loose + packed-refs)

These are created using subFS to isolate each repository:

```go
func (t *Transferer) ReceivePack(ctx context.Context, pat string, r io.Reader, w io.Writer) error {
    objects := newObjectStorage(newSubFS(t.fs, path.Join(pat, "objects")))
    refs := newRefStorage(newSubFS(t.fs, path.Join(pat, "refs")))
    // ...
}
```

## Packfile Handling

We implement our own packfile encoding and decoding (in `packfile.go`), independent of go-git. This keeps our FS interface minimal and avoids heavy dependencies.

### Packfile Format

Packfile is stored in `objects/pack/` with `.pack` extension. Format:
- Header: `PACK` + version (4 bytes) + num objects (4 bytes)
- Entries: object data (zlib compressed) with type/size info
- Footer: SHA1 checksum (20 bytes)

### Implementation

Two symmetric functions handle packfile encoding and decoding:

```go
// encodePackfile encodes objects as a packfile.
// want: object hashes to include. If nil, encodes all reachable objects (used for repack).
// Returns the packfile SHA-1 checksum.
func encodePackfile(w io.Writer, storage *objectStorage, want []Hash) (Hash, error)

// decodePackfile decodes a packfile and stores objects via storage.
// Returns the packfile SHA-1 checksum.
func decodePackfile(r io.Reader, storage *objectStorage) (Hash, error)
```

The Transferer uses these for:
- **UploadPack**: Call encodePackfile to generate packfile for client
- **ReceivePack**: Call decodePackfile to parse packfile from client

These functions also support local repack (GC):
- **Local repack**: Call encodePackfile with want=nil to pack all reachable objects

This is NOT exposed to users—the Transferer handles all packfile operations transparently.

## pkt-line Format

All Git protocol communication uses pkt-line encoding: 4-byte length (hex) + payload.

## Reference Updates

Reference updates are atomic:
1. Receive all objects
2. Validate all references
3. Update all references together
4. Report success/failure

This prevents partial updates on failure.
