# asset-cache

[![Go](https://img.shields.io/badge/go-1.25-blue)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A high-performance digital asset management service with built-in versioning,
role-aware access controls, smart caching, and signed URL delivery. Designed for
media companies, game studios, and design teams that need a fast, secure layer
between their object storage and their applications.

## Why asset-cache?

Storing files is easy; serving them quickly, safely, and with version history is
not. `asset-cache` is a single Go binary that bundles the pieces most teams end
up reinventing:

- **Upload + metadata** in one request
- **Content-addressable checksums** for duplicate detection
- **Version history** with diffs
- **LRU + memory caching** so hot assets never hit cold storage
- **Time-limited signed URLs** for CDN-safe delivery
- **Full-text + tag search** over asset metadata

## Quick start

### Run from source

```bash
git clone https://github.com/auto-forge-org/asset-cache.git
cd asset-cache
go run ./cmd/asset-cache
```

The server listens on `:8080` by default.

### Configuration

All configuration is via environment variables:

| Variable     | Default  | Description                                  |
|--------------|----------|----------------------------------------------|
| `PORT`       | `8080`   | HTTP listener port                           |
| `CACHE_SIZE` | `256`    | Max entries in the in-memory LRU cache       |
| `SIGNER_KEY` | _random_ | HMAC key for signing download URLs (hex/raw) |
| `LOG_LEVEL`  | `info`   | Log verbosity                                |

> In production, set `SIGNER_KEY` to a stable value — a random key on every
> restart invalidates previously signed URLs.

## API overview

Base path: `/api/v1`

| Method | Path                          | Purpose                                   |
|--------|-------------------------------|-------------------------------------------|
| `GET`  | `/healthz`                    | Liveness probe                            |
| `GET`  | `/metrics/cache`              | Cache hit/miss statistics                 |
| `POST` | `/api/v1/assets`              | Upload an asset (multipart, ≤5 GB)        |
| `GET`  | `/api/v1/assets`              | List assets                               |
| `GET`  | `/api/v1/assets/search`       | Search by query + tags                    |
| `GET`  | `/api/v1/assets/:id`          | Get asset metadata                        |
| `GET`  | `/api/v1/assets/:id/download` | Download (requires signed `exp` + `sig`)  |
| `PUT`  | `/api/v1/assets/:id/version`  | Push a new revision                       |
| `GET`  | `/api/v1/assets/:id/versions` | List version history                      |
| `POST` | `/api/v1/assets/:id/sign`     | Generate a time-limited signed URL        |

### Example: upload an asset

```bash
curl -X POST http://localhost:8080/api/v1/assets \
  -H "X-User-ID: user-42" \
  -F "file=@logo.png" \
  -F 'metadata={"tags":["branding","logo"]}'
```

### Example: get a signed download URL

```bash
curl -X POST http://localhost:8080/api/v1/assets/<asset-id>/sign \
  -H "Content-Type: application/json" \
  -d '{"expiration":"1h"}'
```

## Project layout

```
cmd/asset-cache/    # Binary entry point
config/             # Env-driven configuration
internal/api/       # Gin handlers + routing
internal/service/   # Asset lifecycle, signing, versioning
internal/storage/   # Storage backends (in-memory today; S3-compatible next)
internal/cache/     # LRU cache layer
internal/model/     # Asset / Version / AccessControl types
pkg/utils/          # Shared utilities (hashing, etc.)
docs/               # Specs and design notes
```

## Development

```bash
go test ./...        # Run all tests
go build ./...       # Build everything
go run ./cmd/asset-cache
```

## Roadmap

See [ROADMAP.md](./ROADMAP.md) for the full milestone plan and current status.

## License

MIT — see the `LICENSE` file (or the SPDX identifier in `go.mod`'s module
metadata) for details.
