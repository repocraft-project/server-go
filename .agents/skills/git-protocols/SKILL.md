---
name: git-protocols
description: "Implementation guide for Git server protocols (Smart HTTP). Use when building a Git hosting platform that needs to implement git clone/fetch/push over HTTP, handle service discovery endpoints, process packfile upload/download, or manage bare Git repositories. Covers Smart HTTP protocol which is the modern standard for Git hosting."
---

# Git Protocols

## Overview

This skill enables implementing a Git server that supports Smart HTTP protocol for Git clone, fetch, and push operations. Smart HTTP is the most popular protocol for Git hosting because it uses standard HTTPS ports, supports authentication, and works through corporate firewalls.

A Git server's job is to:
1. **Serve Git objects** - blobs, trees, commits
2. **Serve refs** - branches, tags
3. **Receive pushes** - accept and store objects + update refs

## Git Data Model (Prerequisites)

Understanding how Git stores data is essential for implementing a Git server.

### Objects

Git is a content-addressable filesystem. Every piece of data is stored as an **object** identified by its SHA-1 hash (40 hex characters).

**Object types:**

| Type       | Description                                                       |
| ---------- | ----------------------------------------------------------------- |
| **blob**   | File content (any file data)                                      |
| **tree**   | Directory structure (list of filenames + blob/tree pointers)      |
| **commit** | Snapshot metadata (tree pointer, parent commits, author, message) |
| **tag**    | Named pointer to another object (usually a commit)                |

**Object storage:**
- Objects are stored in `.git/objects/` directory
- First 2 chars of SHA-1 = subdirectory name
- Remaining 38 chars = filename
- Example: object `d670460b4b4aece5915caf5c68d12f560a9fe3e4` stored at `.git/objects/d6/70460b4b4aece5915caf5c68d12f560a9fe3e4`

### Packfiles

For efficiency, Git packs multiple objects into a single **packfile**:
- Compressed storage (zlib)
- Delta compression (stores differences between versions)
- Two files: `.pack` (data) and `.idx` (index)

When transferring data over the network, Git uses packfiles to reduce transfer size.

### References (Refs)

Refs are named pointers to commits:
- **Branches**: `refs/heads/<name>` (e.g., `refs/heads/master`)
- **Tags**: `refs/tags/<name>`
- **HEAD**: Special ref pointing to current branch

Refs are stored as plain text files in `.git/refs/`:
```
.refs/heads/master   # Contains: ca82a6dff817ec66f44342007202690a93763949
```

## Protocol Types

| Protocol     | Port   | Authentication     | Use Case                             |
| ------------ | ------ | ------------------ | ------------------------------------ |
| Smart HTTP   | 443/80 | Yes (Basic/Bearer) | **Recommended** - modern Git hosting |
| SSH          | 22     | SSH Keys           | Developer access                     |
| Git (daemon) | 9418   | None               | Anonymous read-only                  |

This skill focuses on **Smart HTTP**.

## HTTP Endpoints

For a repository at `{repo}`:

```
GET  /{repo}/info/refs?service=git-upload-pack   # Service discovery (fetch/clone)
GET  /{repo}/info/refs?service=git-receive-pack  # Service discovery (push)
POST /{repo}/git-upload-pack                     # Fetch/clone data
POST /{repo}/git-receive-pack                    # Push data
```

## Smart HTTP Flow

### Service Discovery

When a client wants to clone or fetch, first it asks the server "what do you have?":

```
GET /repo/info/refs?service=git-upload-pack
```

Server responds with refs and capabilities:

```
001f# service=git-upload-pack
00ab6c5f0e45abd7832bf23074a333f739977c9e8188 refs/heads/master\0multi_ack thin-pack side-band side-band-64k ofs-delta shallow no-progress include-tag multi_ack_detailed symref=HEAD:refs/heads/master agent=git/2.x.x
0000
```

Format: 4-digit hex length + content + `0000` (end marker)

### Fetch (Downloading Data)

1. Client POSTs to `/git-upload-pack`
2. Client sends "want" list (commits they want)
3. Client sends "have" list (commits they already have)
4. Client sends `done`
5. Server calculates needed objects and sends packfile

Client request format:
```
0032want 0a53e9ddeaddad63ad106860237bbf53411d11a7
0032have 441b40d833fdfa93eb2908e52742248faf0ee993
0009done
0000
```

### Push (Uploading Data)

1. Client connects to `git-receive-pack`
2. Server sends current refs
3. Client sends updated refs + packfile with new objects
4. Server unpacks and updates refs
5. Server reports success/failure

Ref update format:
```
0076ca82a6dff817ec66f44342007202690a93763949 15027957951b64cf874c3557a0f3547bd83b3ff6 refs/heads/master
0000
<packfile data>
```

Server response:
```
000eunpack ok
```

## Implementation Approaches

There are several ways to implement Git Smart HTTP:

1. **Use a Git library** - Libraries like go-git (Go), JGit (Java), or Dulwich (Python) handle the protocol
2. **Invoke git executables** - Call `git-upload-pack` and `git-receive-pack` as external processes
3. **Implement protocol directly** - Parse and generate protocol messages yourself

For wire format details, see [references/protocol.md](references/protocol.md).

## Repository Storage

Git hosting requires **bare repositories** (no working directory):

```
repo.git/
├── HEAD                    # Points to current branch (e.g., "ref: refs/heads/master")
├── config                  # Repository configuration
├── hooks/                  # Server-side hooks
├── info/                   # Additional info
├── objects/
│   ├── info/              # List of loose objects
│   ├── pack/              # Packfiles (*.pack, *.idx)
│   ├── [xx]/[40-char]    # Loose objects (xx = first 2 chars of SHA)
├── refs/
│   ├── heads/             # Branch refs
│   └── tags/              # Tag refs
└── packed-refs            # Packed refs (combined from refs/)
```

Create bare repo: `git init --bare repo.git`

## References

For detailed protocol specifications:

- [protocol.md](references/protocol.md) - Complete wire format and state machine
- [endpoints.md](references/endpoints.md) - HTTP endpoint design
