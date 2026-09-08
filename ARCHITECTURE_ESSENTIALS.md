# Architecture Essentials: stellaryard-core

> Quick-reference outline. Full detail: ARCHITECTURE.md. This file exists so an agent working a single issue doesn't need to load the whole doc.

## Stack
- Go 1.22+, `chi` router, `gorilla/websocket`, `docker/docker` SDK, `stellar/go` SDK, SQLite

## Non-negotiable rule
**Never bypass the `Signer` interface to handle a secret key directly.** V1 uses `LocalTestSigner` (testnet-only keys in SQLite). Mainnet will plug in as a second `Signer` implementation later — do not build shortcuts around this boundary, even for "just testnet, just this once."

## Repo role
Backend + orchestrator only. No UI. Dashboard and CLI are separate repos that consume this repo's API. This repo's API contract (`/api/openapi.yaml`) is the source of truth both other repos build against — do not make breaking changes to it without versioning.

## Core components
- **API layer** — REST + WS, `chi` router, localhost-only in v1 (no auth yet — that's intentional for v1, not an oversight)
- **Signer Interface** — see rule above
- **Docker orchestrator** — controls Horizon + Soroban RPC containers via Docker API (not shell-exec)
- **SQLite** — local state: accounts (testnet keys), deployments, cached ledger data

## Key data models
- `Account` — has `SecretKey` field, **testnet-only**, never extend this shape to mainnet accounts
- `ContainerStatus`, `ContractDeployment`, `LedgerSnapshot`

## API surface (see ARCHITECTURE.md for full table)
`/containers/*`, `/accounts/*`, `/contracts/*`, `/ledger/*` — all under `/api/v1`

## Explicit v1 non-goals
- No mainnet signing/custody
- No auth (localhost-only assumption)
- No multi-tenancy
- No Postgres

## What would break
- Docker daemon unavailable → container ops fail, WS log streaming hangs (no backoff/retry)
- SQLite crash mid-write → lost accounts/deployments, no reconciliation on restart
- Container health-check returns false positive → Horizon reports healthy before synced, queries return empty data
- Multiple core instances → silent SQLite + Docker state corruption (no lock file)
- WS log streaming memory growth → no backpressure on fast-producing containers

## Edge cases missing
- `docker-compose down` while core is running → core reports stale container status until next health check
- Friendbot rate limiting → account creation fails with no retry logic
- WASM deploy fails mid-network → WASM is lost, no re-upload path
- Duplicate account labels → no unique constraint on `Account.Label`
- Port collisions with existing docker-compose files (Horizon port 8000)

## What's overengineered
- `Signer` interface with one implementation (correct for future, costly for V1)
- Hand-maintained OpenAPI spec (drifts without build-time validation)
- `LedgerSnapshot` cache (premature; local Horizon is fast enough)
- `local` vs `testnet` distinction (identical behavior, should be boolean)

## When in doubt
If an issue touches key handling, signing, or the account data model — flag for maintainer review before merging, regardless of stated complexity level. This is the one area where "it's just testnet" is not a sufficient reason to cut corners, because the shape set now is what mainnet support extends later.
