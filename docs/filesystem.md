# Git Server Filesystem Design

## Overview

A Git server is essentially a transport server built on top of a filesystem. The core of Git is object storage (objects) and reference management (refs). All git operations (clone, fetch, push) revolve around these files.

## Minimal Filesystem Interface

To support multiple storage backends (local filesystem, KV cache, remote object storage), we define the minimal filesystem interface:

```go
type ReadonlyFile interface {
    io.Reader
    io.Seeker
    io.Closer
}

type WriteonlyFile interface {
    io.Writer
    io.Seeker
    io.Closer
}

type FS interface {
    Create(entry string) error
    Reader(entry string) (ReadonlyFile, error)
    Writer(entry string) (WriteonlyFile, error)
    Delete(entry string) error
    Rename(src, dst string) error
    Listdir(entry string) ([]string, error)
    Exist(entry string) bool
}
```

## Interface Explanation

| Method    | Purpose                              |
| --------- | ------------------------------------ |
| `Create`  | Create a new file                    |
| `Reader`  | Read file contents                   |
| `Writer`  | Write file contents                  |
| `Delete`  | Delete a file                        |
| `Rename`  | Rename a file (for temp file writes) |
| `Listdir` | List directory contents              |
| `Exist`   | Check if file exists                 |

## Storage Backends

### Local Filesystem

Implemented using os package:
- `Create` → os.MkdirAll + os.Create (just create file, no handle returned)
- `Reader` → os.Open
- `Writer` → os.OpenFile(O_WRONLY|O_CREATE)
- `Delete` → os.Remove
- `Rename` → os.Rename
- `Listdir` → os.ReadDir
- `Exist` → os.Stat != nil

### KV Cache

Implemented using Get/Put/Scan:
- `Reader` → Get(key) → return Seekable Reader
- `Writer` → Put(key) → return Seekable Writer
- `Listdir` → Prefix scan
- `Exist` → Exists(key)
- `Rename` → Create + Copy + Delete (fallback)

### Remote Object Storage

Implemented using S3/MinIO APIs:
- Read/Write objects via Range requests
- Listdir via ListObjects
- Rename via CopyObject + DeleteObject

## Git Operations

On top of this minimal filesystem, we can implement:

- **git-upload-pack**: Read objects, generate packfile
- **git-receive-pack**: Receive packfile, update objects and refs

This design makes storage backends fully pluggable without modifying transport logic.
