# HTTP Endpoint Design

This document describes the HTTP endpoint structure for Git Smart HTTP protocol. Language-agnostic.

## Route Structure

```
/                      # Web UI (HTML)
/api/v1/               # REST API
  /repos               # Repository management
  /user                # User management
/{owner}/{repo}.git/   # Git HTTP endpoints
```

## Git HTTP Routes

### With .git suffix (GitHub style)

```
GET    /{owner}/{repo}.git/info/refs?service=git-upload-pack
POST   /{owner}/{repo}.git/git-upload-pack
GET    /{owner}/{repo}.git/info/refs?service=git-receive-pack
POST   /{owner}/{repo}.git/git-receive-pack
```

### Without .git suffix (Gitea style)

```
GET    /{owner}/{repo}/info/refs?service=git-upload-pack
POST   /{owner}/{repo}/git-upload-pack
GET    /{owner}/{repo}/info/refs?service=git-receive-pack
POST   /{owner}/{repo}/git-receive-pack
```

## Handler Requirements

### Service Discovery Handler

- Parse `service` query parameter (`git-upload-pack` or `git-receive-pack`)
- Validate repository exists and user has access
- Return refs in packet-line format with capabilities

### Upload Pack Handler (fetch/clone)

- Stream request body to git-upload-pack process
- Stream output back to client
- Set Content-Type: `application/x-git-upload-pack-result`

### Receive Pack Handler (push)

- Stream request body to git-receive-pack process
- Stream output back to client
- Set Content-Type: `application/x-git-receive-pack-result`

## Common HTTP Headers

| Header          | Description                                                                         |
| --------------- | ----------------------------------------------------------------------------------- |
| `Content-Type`  | `application/x-git-upload-pack-request` or `application/x-git-receive-pack-request` |
| `Accept`        | `application/x-git-upload-pack-result`                                              |
| `Authorization` | Basic auth or Bearer token                                                          |

## Error Responses

| Status           | Meaning                 |
| ---------------- | ----------------------- |
| 200 OK           | Success                 |
| 401 Unauthorized | Authentication required |
| 403 Forbidden    | Access denied           |
| 404 Not Found    | Repository not found    |
| 500 Server Error | Internal error          |

## Implementation Notes

- **Streaming required**: Git protocol is bidirectional streaming. Do not buffer entire request/response.
- **Packet-line format**: All protocol messages use 4-hex-byte length prefix.
- **Side-band multiplexing**: Server may use bands for progress (2) and errors (3) alongside pack data (1).
