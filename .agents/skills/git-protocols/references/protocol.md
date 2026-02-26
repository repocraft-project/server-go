# Git Smart HTTP Protocol Specification

## Packet Line Format

All Git protocol communication uses "packet lines":

```
<4 hex bytes> <payload>
```

- First 4 chars: length of entire packet (including these 4 bytes)
- Length is in hex (e.g., `0032` = 50 bytes)
- `0000` indicates end of data

Example:
```
0032want 0a53e9ddeaddad63ad106860237bbf53411d11a7
```

## Service Discovery

### Request

```
GET /{repo}/info/refs?service=git-upload-pack
```

### Response

```
001f# service=git-upload-pack
00ab<capabilities>\0<refs>
0000
```

Capabilities (null-terminated, space-separated):
- `multi_ack` - Server can handle multiple ack messages
- `thin-pack` - Server may send "thin" packs (deltas against local objects)
- `side-band` - Server supports multiplexed output
- `side-band-64k` - Side band with 64k max packet size
- `ofs-delta` - Server can send deltas as offsets
- `shallow` - Supports shallow clones
- `no-progress` - Client doesn't want progress messages
- `include-tag` - Server will send tags for transferred objects
- `multi_ack_detailed` - Extended multi-ack

## Fetch (git-upload-pack)

### Phase 1: Capability Negotiation

Client:
```
POST /{repo}/git-upload-pack
Content-Type: application/x-git-upload-pack-request

<want lines>
<have lines>
done
```

Want line:
```
want <sha1> <capabilities>
```

Have line:
```
have <sha1>
```

### Phase 2: Packfile Transfer

Server sends packfile using side-band multiplexing:

- Band 1: Pack data
- Band 2: Progress info
- Band 3: Error messages

## Push (git-receive-pack)

### Phase 1: Reference Discovery

Client connects, server sends:
```
<ref> <sha1> <capabilities>
...
0000
```

### Phase 2: Push Request

Client sends:
```
<old-sha1> <new-sha1> <refname>\n
<packfile>
```

Example:
```
0076ca82a6dff817ec66f44342007202690a93763949 15027957951b64cf874c3557a0f3547bd83b3ff6 refs/heads/master
0000
<packfile data>
```

### Phase 3: Report

Server responds:
```
unpack ok
<ref> ok <refname>
```

Or error:
```
unpack <error>
<ref> ng <refname> <error message>
```

## Common HTTP Headers

| Header         | Description                                                                         |
| -------------- | ----------------------------------------------------------------------------------- |
| `Content-Type` | `application/x-git-upload-pack-request` or `application/x-git-receive-pack-request` |
| `Accept`       | `application/x-git-upload-pack-result`                                              |
| `Git-Protocol` | Version indicator                                                                   |

## Error Responses

| Status           | Meaning                 |
| ---------------- | ----------------------- |
| 200 OK           | Success                 |
| 401 Unauthorized | Authentication required |
| 403 Forbidden    | Access denied           |
| 404 Not Found    | Repository not found    |
| 500 Server Error | Internal error          |
