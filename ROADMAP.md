# asset-cache Roadmap

This roadmap describes where `asset-cache` is today and where it is heading.
Items inside a milestone are not strictly ordered; items between milestones are.

## Vision

A single, embeddable Go service that gives any team a production-grade asset
layer — upload, version, cache, search, and securely deliver binary assets —
without forcing them to assemble half a dozen libraries and a custom CDN
contract first.

## Status legend

- ✅ Shipped
- 🚧 In progress
- ⏳ Planned
- 💡 Exploring

## Milestones

| Milestone | Theme                          | Status |
|-----------|--------------------------------|--------|
| M1        | Core asset service             | ✅     |
| M2        | Persistent storage + database  | ⏳     |
| M3        | Access control & auth          | ⏳     |
| M4        | Production observability       | ⏳     |
| M5        | CDN integration & format ops   | ⏳     |
| M6        | Search & analytics depth       | 💡     |

---

### M1 — Core asset service ✅

The minimum lovable product: a Gin service that can ingest, version, cache,
search, and securely deliver assets out of an in-memory store.

- [x] Project scaffolding (`cmd`, `internal`, `pkg`, `config`)
- [x] Asset upload with checksum and metadata
- [x] In-memory storage backend
- [x] LRU cache layer with hit/miss metrics
- [x] Version history per asset
- [x] HMAC-signed time-limited download URLs
- [x] Tag + query search over metadata
- [x] Unit tests across cache, storage, service, and utils

### M2 — Persistent storage + database ⏳

Move from in-memory to real durability so the service survives restarts.

- [ ] PostgreSQL-backed metadata store (Asset, Version, AccessControl)
- [ ] S3-compatible object storage backend (AWS S3, MinIO)
- [ ] Migrations + schema versioning
- [ ] Duplicate detection via content hash across the store
- [ ] Backend selection via configuration (`STORAGE_DRIVER`, `DB_URL`)

### M3 — Access control & auth ⏳

Make the service safe to expose beyond a trusted network.

- [ ] JWT (RSA-256) authentication middleware
- [ ] Role-based permissions (view / edit / admin)
- [ ] Asset-level ACLs with folder inheritance
- [ ] Audit log of access events with user context
- [ ] Rate limiting (default 1000 req/min/user)

### M4 — Production observability ⏳

What you need before putting it in front of real traffic.

- [ ] Structured JSON logging with request IDs
- [ ] Prometheus `/metrics` endpoint (cache hit rate, request latency)
- [ ] Health + readiness probes split (`/healthz` vs `/readyz`)
- [ ] Graceful shutdown and connection draining
- [ ] OpenTelemetry tracing

### M5 — CDN integration & format ops ⏳

Make delivery cheap and fast.

- [ ] Signed-URL handoff to CloudFront / Cloudflare
- [ ] Domain whitelisting enforcement on signed URLs
- [ ] On-the-fly image transcoding (WebP, AVIF)
- [ ] Thumbnail / preview generation pipeline
- [ ] Tiered cache (memory → disk → origin)

### M6 — Search & analytics depth 💡

Beyond exact-tag matching.

- [ ] Full-text search backend (PostgreSQL FTS or Meilisearch)
- [ ] Fuzzy matching + relevance scoring
- [ ] Usage analytics dashboard endpoints
- [ ] Per-asset trend reports with retention policies
- [ ] CLI client with tab-completion (`asset-cache upload/search/version`)

---

## Non-goals (for now)

- A bundled web UI — `asset-cache` is API-first; UIs should consume the API.
- Multi-tenant SaaS billing / quota tooling.
- Acting as a primary system of record for non-asset data.

## Contributing to the roadmap

Open a GitHub issue with the `roadmap` label to propose a new item or to argue
that an existing one should move up. Concrete use cases beat abstract requests
— tell us what you would build with it.
