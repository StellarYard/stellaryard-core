# ROADMAP: stellaryard-core

> **This file must be updated with every contribution.** Before opening a PR: mark completed items done, add newly-surfaced work, or note if your change invalidates an assumption below. A PR that changes functionality without a corresponding `ROADMAP.md` update should be considered incomplete — see `AGENTS.md` rule 5.

Status legend: `[ ]` not started · `[~]` in progress · `[x]` done

## Phase 0 — Foundation (blocks everything downstream)

- [ ] Scaffold Go module, repo structure, CI (lint + test on PR)
- [ ] `docker-compose.yml` for local Horizon + Soroban RPC
- [ ] `/api/openapi.yaml` — initial version covering all v1 endpoints listed in `ARCHITECTURE.md`. **This must exist and be merged before any `stellaryard-dashboard` or `stellaryard-cli` issue is opened** — those repos generate their clients from this file. Opening consumer-repo issues before this lands recreates the contract-drift problem this architecture was designed to avoid.
- [ ] SQLite schema + migrations for `Account`, `ContainerStatus` (if persisted), `ContractDeployment`, `LedgerSnapshot` cache
- [ ] `Signer` interface + `LocalTestSigner` implementation, with explicit boundary test (raw key never leaves the interface)

## Phase 1 — Container orchestration

- [ ] Docker API client wrapper (start/stop/status for named containers)
- [ ] Health check logic for Horizon + Soroban RPC containers
- [ ] `POST /containers/{name}/start`, `/stop`
- [ ] `GET /containers` (list statuses)
- [ ] WS `/containers/{name}/logs` streaming

## Phase 2 — Accounts

- [ ] Account creation + Friendbot/local-genesis funding
- [ ] `POST /accounts`, `GET /accounts`, `GET /accounts/{publicKey}`
- [ ] Balance lookup via Horizon proxy

## Phase 3 — Ledger

- [ ] `GET /ledger/snapshot`
- [ ] `GET /ledger/transactions` (paginated)
- [ ] Transaction detail enrichment (decode XDR into a usable response shape — decide how much decoding belongs in core vs. left raw for consumers; **this decision isn't made yet and should be resolved before Phase 3 issues open**, not discovered mid-implementation)

## Phase 4 — Contracts

- [ ] WASM upload/deploy flow (`POST /contracts/deploy`)
- [ ] Contract invocation (`POST /contracts/{contractId}/invoke`)
- [ ] Deployment record persistence

## Phase 5 — Hardening (required before calling this "100% ready," not optional polish)

- [ ] Error handling audit — consistent error response shape across all endpoints (dashboard/CLI exit-code and error-display logic depend on this being consistent, not ad hoc per-endpoint)
- [ ] API versioning mechanism actually exercised once (prove `/api/v2/` pattern works before it's needed under pressure)
- [ ] Load/soundness check: what happens if Docker daemon is unreachable, mid-restart, or containers crash — core should degrade predictably, not hang
- [ ] Documented, tested account/contract data cleanup (what happens on `docker-compose down` — is SQLite state stale garbage or intentionally persisted?)

## Explicitly deferred (not v1, tracked so it isn't forgotten or silently assumed done)

- [ ] `ExternalSigner` implementation for mainnet (hardware wallet / wallet-connect-style flow) — **do not start this until PRD.md is updated to move mainnet out of "future" and into current scope**
- [ ] Auth layer (currently localhost-only, no-auth by design — revisit if remote access is ever needed)
- [ ] Multi-tenancy

## What would break

- **Docker daemon unavailability**: No retry/backoff in container ops. WS log streaming hangs if daemon drops mid-stream. Affects all downstream repos.
- **SQLite crash mid-write**: Account creation (DB write + Friendbot HTTP) can leave partial state. No reconciliation on restart. Lost accounts are unrecoverable.
- **Container health-check false positives**: Horizon HTTP 200 ≠ synced. Early queries return empty data. Health check needs actual Stellar API verification.
- **Concurrent core instances**: No PID/lock file. Two processes against same DB + Docker = silent corruption.
- **WS log streaming memory growth**: No backpressure or ring-buffer. High-throughput containers exhaust memory over long sessions.
- **Friendbot rate limiting**: Account creation fails with no retry. Most common user-facing failure.

## Edge cases not yet addressed

- `docker-compose down` while core is running → stale container status until next health check
- WASM deploy fails mid-network → WASM lost, no re-upload path
- Duplicate account labels → no unique constraint
- Port collisions with existing docker-compose files (port 8000)
- SQLite on NFS or non-WAL-compatible filesystem → silent corruption
- Multiple simultaneous account creations → race condition on Friendbot funding

## What's overengineered for V1

- `Signer` interface with one implementation (correct future-proofing, costly for V1)
- Hand-maintained `openapi.yaml` (drifts without build-time validation)
- `LedgerSnapshot` SQLite cache (premature; local Horizon is fast)
- `local` vs `testnet` distinction (identical behavior, should be boolean)

## Open questions blocking full readiness

- XDR decoding depth for transaction detail (Phase 3) — undecided
- Whether `ContainerStatus` needs persistence or can be purely live-queried from Docker each time — undecided, affects SQLite schema in Phase 0
- PID/lock mechanism to prevent concurrent core instances — needed but not scoped
- Spec-conformance test for `openapi.yaml` — needed to prevent drift, not yet built
- Log streaming ring-buffer design — needed to prevent memory exhaustion
