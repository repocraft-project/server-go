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

Packfile encoding/decoding is handled internally using go-git's packfile package. This is NOT exposed to users—the Transferer handles all packfile operations transparently.

## pkt-line Format

All Git protocol communication uses pkt-line encoding: 4-byte length (hex) + payload.

## Reference Updates

Reference updates are atomic:
1. Receive all objects
2. Validate all references
3. Update all references together
4. Report success/failure

This prevents partial updates on failure.
