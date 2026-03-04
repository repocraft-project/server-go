# Git Server Implementation Design

## Architecture Overview

A Git server is essentially a transport server that provides two core services: upload-pack (for fetch/clone) and receive-pack (for push). The server does not need to understand the content of Git objects—it only needs to store, retrieve, and transmit them efficiently.

## Layered Architecture

```
┌─────────────────────────────────────────┐
│           Application Layer              │
│  (HTTP/SSH handlers, auth, repo mgmt)  │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│          Transferer                      │
│  UploadPack / ReceivePack                │
│  (unexported: objectStorage, refStorage)│
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│          FS Interface                   │
│  (pluggable: LocalFS / KV / S3)        │
│  + subfs.go (path prefix wrapper)      │
└─────────────────────────────────────────┘
```

## Directory Structure

```
pkg/git/
  fs.go           # FS interface
  subfs.go        # subFS helper (path prefix wrapper)
  localfs.go      # LocalFS implementation
  
  object.go       # objectStorage (unexported)
  ref.go          # refStorage (unexported)
  packfile.go     # packfile encode/decode (unexported)
  transfer.go     # Transferer
```

## User API

```go
fs := git.NewLocalFS("/data/git")
transferer := git.NewTransferer(fs)

transferer.UploadPack(ctx, "user/repo", r, w)   // fetch/clone
transferer.ReceivePack(ctx, "user/repo", r, w)  // push
```

## Design Principles

1. **Minimal FS First**: Define the smallest possible filesystem interface that can support Git operations
2. **Storage Agnostic**: All storage implementations (local, KV, S3) share the same interface
3. **Protocol Independent**: upload-pack and receive-pack are implemented once, usable over any transport (HTTP, SSH, git protocol)
4. **No Submodule Complexity**: Submodules are a client-side concern; the server treats all repositories uniformly
5. **No go-git Exposure**: All go-git dependencies are internal; user-facing API is completely separate
6. **Unexported Storage**: objectStorage and refStorage are helper structs (unexported)

## Implementation Sequence

See detailed designs in:
- `docs/filesystem.md` - FS interface and implementations
- `docs/storage.md` - Internal storage design (objectStorage, refStorage, subFS)
- `docs/protocol.md` - Protocol layer design (Transferer, UploadPack, ReceivePack)
