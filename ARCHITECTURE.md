# Architecture: stellaryard-core

## Tech Stack

- **Language**: Go (1.22+)
- **HTTP**: standard library `net/http` + `chi` router (lightweight, avoids heavy framework lock-in — important since Wave contributors will have varying Go experience levels)
- **WebSocket**: `gorilla/websocket` for log streaming
- **Docker control**: official `docker/docker` Go SDK (talks to the Docker daemon via its API, not shell-exec'd `docker` commands — shelling out is fragile and harder to test)
- **Horizon/Soroban RPC clients**: Stellar's official Go SDK (`stellar/go`) for Horizon and Soroban RPC interaction
- **Storage**: SQLite (single-file, zero-ops — appropriate for a local single-developer tool; do not reach for Postgres here)
- **API spec**: OpenAPI 3.0, hand-maintained in `/api/openapi.yaml`, source of truth for dashboard/CLI contract

## Why This Stack

Go was chosen (consistent across all three repos) because Docker API interaction and concurrent container/log management is a natural fit for goroutines, and it keeps the whole project in one language — a contributor fixing a core bug can plausibly also touch CLI without a context switch. SQLite over Postgres is deliberate: this runs on a developer's laptop, not a server; adding a Postgres dependency for a local tool is unjustified operational weight.

## System Overview

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
           │  (this repo)        │
           │                     │
           │  ┌───────────────┐  │
           │  │ API layer     │  │
           │  ├───────────────┤  │
           │  │ Signer        │  │◄── future: external signer for mainnet
           │  │ Interface     │  │
           │  ├───────────────┤  │
           │  │ Docker        │  │──► Horizon container
           │  │ orchestrator  │  │──► Soroban RPC container
           │  ├───────────────┤  │
           │  │ SQLite (local │  │
           │  │ state, test   │  │
           │  │ keys)         │  │
           │  └───────────────┘  │
           └─────────────────────┘
```

## The Signer Interface (critical design decision)

This is the one piece of core architecture explicitly shaped around the stated future need (real-fund/mainnet support), even though V1 never uses that path. Getting this wrong means a rewrite later, not an extension.

**Rule: core never holds or transmits a raw secret key outside of the `Signer` interface boundary.**

```go
type Signer interface {
    // Sign returns a signed transaction envelope for the given unsigned
    // transaction, for the given public key. Implementations decide how.
    Sign(ctx context.Context, unsignedTxXDR string, publicKey string) (signedTxXDR string, err error)

    // PublicKeys returns the public keys this signer can sign for.
    PublicKeys(ctx context.Context) ([]string, error)
}
```

- **V1 implementation**: `LocalTestSigner` — generates and holds testnet-only keypairs in SQLite, signs directly in-process. This is fine *only* because these keys never hold real value.
- **Future implementation**: `ExternalSigner` — delegates signing to a hardware wallet, browser extension (Freighter-style), or wallet-connect-equivalent flow. Core sends an unsigned XDR envelope out, gets a signed one back. Core itself never sees the mainnet secret key.

Every code path that currently calls `LocalTestSigner` directly must go through the `Signer` interface, never around it. This is the single architectural rule that matters most in this repo — a Wave issue that bypasses it (e.g., "quick way to sign server-side for convenience") should be rejected in review regardless of how small it looks.

## Data Models

```go
// Account represents a local/testnet keypair core manages.
// NOTE: SecretKey field only ever populated for LocalTestSigner-managed
// accounts. This model does not extend to mainnet accounts — a future
// mainnet account model should NOT have a SecretKey field at all.
type Account struct {
    ID         string    // internal ID
    PublicKey  string
    SecretKey  string    // testnet only; empty/absent for external-signer accounts
    Label      string    // user-assigned friendly name
    Network    string    // "local" | "testnet" — never "mainnet" in v1
    CreatedAt  time.Time
}

type ContainerStatus struct {
    Name    string // "horizon" | "soroban-rpc"
    State   string // "running" | "stopped" | "error"
    Health  string
    Started time.Time
}

type ContractDeployment struct {
    ID          string
    WASMHash    string
    ContractID  string
    DeployedBy  string // Account.PublicKey
    Network     string
    CreatedAt   time.Time
}

type LedgerSnapshot struct {
    Sequence  uint32
    Timestamp time.Time
    TxCount   int
}
```

## API Contract (consumed by dashboard and CLI — treat as stable)

Base URL: `http://localhost:{port}/api/v1` (v1 is local-only; no auth required because it never leaves localhost — **this assumption must be revisited before any remote or mainnet deployment**)

| Method | Path | Purpose |
|---|---|---|
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

Full request/response schemas live in `/api/openapi.yaml` — that file, not this doc, is the byte-level source of truth. This table is for orientation only.

## Known Failure Modes

- **Docker daemon unavailability**: Core has no retry/backoff when Docker is unreachable. Container start/stop will return errors, but WS log streaming may hang indefinitely if the daemon drops mid-stream. A health-poll loop with exponential backoff and a circuit breaker is needed.
- **SQLite WAL corruption on crash**: A crash mid-write (e.g., during account creation + Friendbot funding) can leave the WAL in an inconsistent state. Go's `mattn/go-sqlite3` handles WAL recovery on open, but a corrupted page is unrecoverable. No reconciliation exists to verify DB state matches on-chain state after a restart.
- **Container health-check false positives**: Horizon's `/` endpoint may return 200 before the node is fully synced. The health check should verify Horizon is responsive to actual Stellar queries (e.g., a ledger query), not just that the HTTP port is open.
- **Log streaming unbounded memory**: WS log streaming has no backpressure or ring-buffer cap. A fast-producing container can exhaust core's memory over time. Each WS connection should have a bounded buffer with oldest-first eviction.
- **Concurrent core instances**: No PID file or lock mechanism. Two core processes against the same SQLite file and Docker containers will corrupt state silently.
- **Horizon sync lag**: After container start, Horizon may lag the ledger. Queries return incomplete data until caught up. No sync-status is exposed to consumers.

## What's Overengineered for V1

- **`Signer` interface with one implementation**: Every signing call goes through an interface with a single `LocalTestSigner`. Adds indirection and a boundary-test requirement. The abstraction is correct for future mainnet support, but V1 contributors pay complexity cost for a feature that doesn't exist yet.
- **Hand-maintained `openapi.yaml`**: Prone to drift from implementation. Should be generated from Go code annotations or validated by a spec-conformance test in CI.
- **`LedgerSnapshot` SQLite cache**: Premature optimization. Local Horizon responds quickly — caching adds staleness risk and schema complexity for negligible V1 benefit.
- **`local` vs `testnet` network distinction**: Two strings with identical behavior in V1. A boolean or enum with actual behavioral difference would be more honest until the distinction matters.
