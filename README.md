# stellaryard-core

The orchestration engine and API server for **StellarYard** — a local development environment manager for the Stellar network. Core runs as a background service on a developer's machine, controls Docker containers for a local Horizon and Soroban RPC instance, and exposes a REST/WebSocket API consumed by the [dashboard](../stellaryard-dashboard) and [CLI](../stellaryard-cli).

Core has no UI. It is infrastructure — the thing that makes the "click a button instead of running CLI commands" experience possible.

## What StellarYard Does

StellarYard eliminates the need to juggle raw CLI commands (`stellar network start`, manual `docker` invocations, curl calls to Horizon) by providing a single controllable backend for local Stellar/Soroban development.

## V1 Features

| Feature | Description |
|---------|-------------|
| **Container Lifecycle** | Start, stop, restart, and health-check local Horizon and Soroban RPC Docker containers via the Docker API |
| **Account Management** | Create and fund test accounts using Friendbot or local network genesis funding; store generated test keypairs |
| **Ledger Inspection** | Proxy/aggregate queries to local Horizon — list recent transactions, view account balances, view ledger sequence/state |
| **Contract Deployment & Invocation** | Deploy Soroban WASM contracts to the local network; invoke contract methods; return results and logs |
| **Log Streaming** | Stream container logs (Horizon, Soroban RPC) to consumers in real time via WebSocket |
| **Stable Versioned API** | A versioned, documented API contract that dashboard and CLI are built against |

## Tech Stack

- **Language**: Go (1.22+)
- **HTTP**: `chi` router (lightweight, avoids heavy framework lock-in)
- **WebSocket**: `gorilla/websocket` for log streaming
- **Docker Control**: Official `docker/docker` Go SDK (talks to Docker daemon via API, not shell-exec'd commands)
- **Stellar Clients**: Stellar's official Go SDK (`stellar/go`) for Horizon and Soroban RPC
- **Storage**: SQLite (single-file, zero-ops — appropriate for a local single-developer tool)
- **API Spec**: OpenAPI 3.0, hand-maintained in `/api/openapi.yaml`

## System Architecture

```
┌─────────────────┐     ┌──────────────────┐
│ stellaryard-     │     │ stellaryard-cli  │
│ dashboard (web)  │     │ (terminal)       │
└────────┬─────────┘     └────────┬─────────┘
         │      REST + WS (localhost only, v1)
         └──────────┬───────────────┘
                     │
           ┌─────────▼──────────┐
           │  stellaryard-core   │
           │                     │
           │  ┌───────────────┐  │
           │  │ API layer     │  │
           │  ├───────────────┤  │
           │  │ Signer        │  │◄── future: external signer
           │  │ Interface     │  │
           │  ├───────────────┤  │
           │  │ Docker        │  │──► Horizon container
           │  │ orchestrator  │  │──► Soroban RPC container
           │  ├───────────────┤  │
           │  │ SQLite        │  │
           │  └───────────────┘  │
           └─────────────────────┘
```

## The Signer Interface

**Critical architectural rule**: Core never holds or transmits a raw secret key outside of the `Signer` interface boundary.

```go
type Signer interface {
    Sign(ctx context.Context, unsignedTxXDR string, publicKey string) (signedTxXDR string, err error)
    PublicKeys(ctx context.Context) ([]string, error)
}
```

- **V1**: `LocalTestSigner` — generates and holds testnet-only keypairs in SQLite, signs in-process
- **Future**: `ExternalSigner` — delegates signing to a hardware wallet, browser extension, or wallet-connect flow

Every code path that signs must go through this interface. Bypasses should be rejected in review regardless of stated complexity.

## API Surface

Base URL: `http://localhost:{port}/api/v1`

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/containers/{name}/start` | Start Horizon or Soroban RPC container |
| POST | `/containers/{name}/stop` | Stop a container |
| GET | `/containers` | List container statuses |
| WS | `/containers/{name}/logs` | Stream container logs |
| POST | `/accounts` | Create + fund a new test account |
| GET | `/accounts` | List managed accounts |
| GET | `/accounts/{publicKey}` | Get account detail + balance |
| POST | `/contracts/deploy` | Deploy a WASM contract |
| POST | `/contracts/{contractId}/invoke` | Invoke a contract method |
| GET | `/ledger/snapshot` | Current ledger state summary |
| GET | `/ledger/transactions` | Recent transactions (paginated) |

Full request/response schemas: [`/api/openapi.yaml`](./api/openapi.yaml)

## Data Models

```go
type Account struct {
    ID        string
    PublicKey string
    SecretKey string    // testnet only; never extends to mainnet
    Label     string
    Network   string    // "local" | "testnet" — never "mainnet" in v1
    CreatedAt time.Time
}

type ContainerStatus struct {
    Name    string // "horizon" | "soroban-rpc"
    State   string // "running" | "stopped" | "error"
    Health  string
    Started time.Time
}

type ContractDeployment struct {
    ID         string
    WASMHash   string
    ContractID string
    DeployedBy string
    Network    string
    CreatedAt  time.Time
}

type LedgerSnapshot struct {
    Sequence  uint32
    Timestamp time.Time
    TxCount   int
}
```

## Roadmap

Full roadmap: [`ROADMAP.md`](./ROADMAP.md) — **must be updated with every contribution** (see agent rules below).

| Phase | Status | Scope |
|-------|--------|-------|
| **0 — Foundation** | Not started | Go module scaffold, docker-compose, OpenAPI spec, SQLite schema, Signer interface |
| **1 — Containers** | Not started | Docker API wrapper, health checks, start/stop/status endpoints, WS log streaming |
| **2 — Accounts** | Not started | Account creation + funding, CRUD endpoints, balance lookup |
| **3 — Ledger** | Not started | Snapshot + transaction endpoints (XDR decoding depth still undecided) |
| **4 — Contracts** | Not started | WASM deploy, contract invocation, deployment persistence |
| **5 — Hardening** | Not started | Error handling audit, API versioning proof, Docker failure modes, data cleanup |

### Open Questions
- XDR decoding depth for transaction detail (Phase 3) — undecided
- Whether `ContainerStatus` needs persistence or can be live-queried from Docker — undecided, affects Phase 0 schema

### Explicitly Deferred (not v1)
- `ExternalSigner` for mainnet — do not start until `PRD.md` moves mainnet into current scope
- Auth layer, multi-tenancy

## Known Failure Modes

These are the things most likely to bite you in v1, in priority order:

1. **Docker daemon unavailable**: If Docker isn't running or is mid-restart, container start/stop/status will fail. Core currently has no retry or backoff — it will surface the error but may hang on WebSocket log streaming if the daemon drops mid-stream. The CLI exit code `2` (core unreachable) covers this at the client level, but core itself needs a graceful degradation path, not just raw errors.
2. **SQLite corruption or concurrent writes**: SQLite is single-writer. If core crashes mid-write (e.g., during account creation + Friendbot funding), the WAL may need recovery on next startup. No recovery logic exists yet. A corrupted DB means lost accounts and deployment records.
3. **Container health-check false positives**: Horizon may report "healthy" before it's fully synced. A query to `/accounts/{pk}/balance` right after container start may return stale or empty data. The health check needs to verify Horizon is actually responsive to Stellar queries, not just that the HTTP port is open.
4. **Friendbot rate limiting and transient failures**: Testnet Friendbot has rate limits and can be slow. Account creation could fail silently or hang. No retry logic is specified — this needs explicit handling with backoff and user-facing error messages.
5. **Log streaming memory growth**: The WS log stream has no backpressure or ring-buffer limit. A fast-producing container (debug-level Horizon logs) could exhaust memory on core's host process over time.
6. **Partial state on crash**: If core crashes mid-contract-deploy or mid-account-fund, the SQLite record may exist but the on-chain state may not match. No reconciliation logic exists — a future `docker-compose down` + restart could leave orphaned records.
7. **Multiple core instances**: Nothing prevents a developer from running two core processes pointing at the same Docker containers and SQLite file. This will corrupt state. There's no PID file or lock mechanism.
8. **Horizon sync lag**: After starting a local network, Horizon may lag behind the ledger. Queries to `/ledger/snapshot` or `/ledger/transactions` may return incomplete data until Horizon catches up. No sync-status indicator is exposed to consumers.

## What's Overengineered for V1

- **The `Signer` interface abstraction**: Necessary for future mainnet support, but in V1 with only `LocalTestSigner`, every call goes through an interface with one implementation. This adds indirection and a boundary test requirement that's disproportionate to V1 complexity — but it's the correct tradeoff because retrofitting this later is a rewrite.
- **Hand-maintained OpenAPI spec**: Writing `openapi.yaml` by hand means it can drift from actual implementation without a build-time check. Consider generating it from Go code annotations or adding a spec-conformance test in CI.
- **`LedgerSnapshot` as a cached model**: Caching ledger state in SQLite is premature if Horizon responds quickly locally. A live-query-through-Horizon approach would be simpler and always accurate. The cache adds schema complexity and staleness risk.
- **`local` vs `testnet` network distinction in `Account`**: For V1 this is two strings that behave identically. A bool or enum with actual behavioral difference would be more honest.

## V1 Non-Goals

- ❌ Mainnet account management or fund custody
- ❌ Transaction signing carrying real value
- ❌ Multi-user / multi-tenant support
- ❌ Authentication (localhost-only by design)
- ❌ Postgres or external database

## Getting Started

```bash
docker-compose up
```

This starts a working local Horizon + Soroban RPC pair managed entirely through core's API.

## Contributing

- Adding a new local/testnet endpoint should be a Trivial-to-Medium Wave issue — it should not require touching the container orchestration layer.
- **If an issue touches key handling, signing, or the account data model** — flag for maintainer review before merging. "It's just testnet" is not a sufficient reason to cut corners.
- Dashboard and CLI can both be built against core's OpenAPI spec without reading core's source code.
- **Every PR must update `ROADMAP.md`** — mark completed items, add new work, or note invalidated assumptions. An unupdated roadmap is an incomplete PR.

### Agent Instructions

This repo includes instructions for AI coding agents:
- [`AGENTS.md`](./AGENTS.md) — Non-negotiable rules for all agents (Signer interface boundary, OpenAPI spec as source of truth, no Postgres, roadmap updates required)
- [`CLAUDE.md`](./CLAUDE.md) — Claude-specific notes (points to AGENTS.md, adds check ROADMAP/ARCHITECTURE before coding)

## License

See [LICENSE](./LICENSE).
