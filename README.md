# asset-cache

A high-performance digital asset management system for media companies, game
studios, and design teams. Provides centralized storage, version control, and
fast retrieval of media files with intelligent multi-tier caching, role-based
access control, and signed-URL delivery.

## Stack

- **Language:** Go
- **Framework:** Gin (HTTP routing + middleware)
- **Storage:** PostgreSQL (metadata) + S3-compatible object store (binaries)
- **Cache:** Redis (metadata) + in-memory LRU (hot assets)

See [BRIEF.md](BRIEF.md) for the full specification.

## Install

```bash
git clone https://github.com/auto-forge-org/asset-cache.git
cd asset-cache
go mod download
go build ./cmd/...
```

## Usage

### Run the API server

```bash
./asset-cache serve --config ./config/config.yaml
```

### CLI

```bash
asset-cache upload --file=logo.png --tags=branding,logo
asset-cache search "logo" --tags=branding --limit=20
asset-cache version history --asset-id=abc123
```

### REST API

```
POST   /api/v1/assets              upload an asset (multipart/form-data)
GET    /api/v1/assets/{id}         fetch metadata + download URL
PUT    /api/v1/assets/{id}/version push a new revision
GET    /api/v1/assets/search       search with tag/date filters
POST   /api/v1/assets/{id}/sign    issue a time-limited signed URL
```

## Development

```bash
go test ./...           # run unit + integration tests
go vet ./...            # static analysis
gofmt -l .              # check formatting
```

## License

MIT © auto-forge-org
