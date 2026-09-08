# PRD: stellaryard-core

## What We're Building

`stellaryard-core` is the orchestration engine and API server for StellarYard, a local development environment manager for the Stellar network. It runs as a background service on a developer's machine, controls Docker containers for a local Horizon (Stellar's API layer) and Soroban RPC instance, and exposes a single REST/WebSocket API that other StellarYard clients (the web dashboard, the CLI) consume.

Core does not have its own UI. It is infrastructure — the thing that makes the "click a button instead of running CLI commands" experience possible for the consumers built on top of it.

## Who It's For

- **Primary**: Stellar/Soroban smart contract developers who currently manage local test networks via raw CLI commands (`stellar network start`, manual `docker` invocations, curl calls to Horizon) and want a single controllable backend instead of juggling multiple tools.
- **Secondary**: Contributors to StellarYard itself — this repo needs to be legible enough that a Wave contributor can pick up an isolated issue (e.g., "add endpoint to fetch contract invocation history") without reverse-engineering the whole system.
- **Not for**: End users managing real funds in v1. Mainnet support is a stated future direction (see Architecture, Signer Interface) but is explicitly out of scope for initial releases.

## What The Product Actually Needs To Do

### V1 (local/testnet only)

1. **Container lifecycle management**: start, stop, restart, and health-check the local Horizon and Soroban RPC Docker containers via the Docker API.
2. **Account management (testnet/local only)**: create and fund test accounts using Friendbot or local network genesis funding; store generated test keypairs.
3. **Ledger inspection**: proxy/aggregate queries to local Horizon so consumers can list recent transactions, view account balances, and view ledger sequence/state without talking to Horizon directly.
4. **Contract deployment & invocation (local/testnet)**: deploy Soroban WASM contracts to the local network; invoke contract methods; return results and logs.
5. **Log streaming**: stream container logs (Horizon, Soroban RPC) to consumers in real time via WebSocket.
6. **A stable, versioned API contract**: since dashboard and CLI are separate repos built by potentially different contributors, core's API is the single source of truth both depend on. Breaking changes must be versioned, not silent.

### Explicitly NOT in V1

- Mainnet account management or fund custody.
- Any signing of transactions carrying real value.
- Multi-user / multi-tenant support (this is a local, single-developer tool).
- Authentication beyond a local-only trust boundary (see Architecture — this changes when remote/mainnet use is added).

### Future (not this repo's current scope, but architecturally anticipated)

- Mainnet signer support via external signer (hardware wallet, wallet-connect style flow) — see ARCHITECTURE.md's Signer Interface. This is the one piece of V1 design explicitly shaped to make this addition non-disruptive later.

## Success Criteria

- A contributor can run `docker-compose up` and have a working local Horizon + Soroban RPC pair managed entirely through core's API within one command.
- Dashboard and CLI can both be built against core's OpenAPI spec without needing to read core's source code.
- Adding a new local/testnet endpoint should be a Trivial-to-Medium Wave issue; it should not require touching the container orchestration layer.

## What Would Break

The highest-risk failure paths in V1:

- **Docker daemon goes away mid-operation**: Core has no retry/backoff logic. If Docker Desktop updates or crashes while containers are running, log streaming will hang and container status will return stale data. Consumers (CLI exit code 2, dashboard disconnect indicator) depend on core detecting this quickly — it currently doesn't.
- **SQLite corruption on crash**: A crash during account creation (which involves both a DB write and a Friendbot HTTP call) can leave partial state. There's no WAL recovery or reconciliation on startup. Lost accounts are unrecoverable.
- **Friendbot unreliability**: Testnet Friendbot has rate limits, can be slow, and sometimes returns errors. Account creation has no retry logic — a single failure means the user must retry manually. This is the most common user-facing failure in practice.
- **Container health-check false positives**: Horizon may report healthy before it's fully synced. Early queries return empty or stale data. The health check needs to verify actual Stellar API responsiveness, not just port availability.
- **Concurrent core instances**: Nothing prevents running two core processes against the same Docker containers and SQLite file. This corrupts state silently. A PID/lock file is needed but not specified.
- **Log streaming memory growth**: No backpressure or ring-buffer limit on WS log streams. High-throughput containers (debug-level Horizon logs) can exhaust memory over long sessions.

## Edge Cases Not Yet Addressed

- What happens if `docker-compose down` is run while core is managing containers? Core doesn't know the containers were removed — it will report stale status until the next health check.
- What if a WASM file is syntactically valid Soroban WASM but fails to deploy due to network issues? The deploy endpoint returns an error, but the WASM is lost — no re-upload or retry path.
- What if two accounts are created simultaneously with the same label? The data model has no unique constraint on label.
- What if the developer has an existing `docker-compose.yml` in their project that conflicts with StellarYard's? Port collisions on Horizon (port 8000) and Soroban RPC (port 8000) are likely.
- What if core's SQLite file is on a filesystem that doesn't support WAL mode (e.g., NFS)? Silent corruption or crashes.

## What's Overengineered

- **The `Signer` interface with one implementation**: Adds indirection and a boundary-test requirement for V1 where `LocalTestSigner` is the only path. Correct for future, but V1 contributors pay complexity cost for a feature that doesn't exist yet.
- **Hand-maintained OpenAPI spec**: Drifts from implementation without build-time validation. Should be generated or have a conformance test.
- **`LedgerSnapshot` cache**: Premature optimization. Local Horizon responds quickly — caching adds staleness risk and schema complexity for no measurable V1 benefit.
- **`local` vs `testnet` network distinction**: Two strings that behave identically in V1. Honest naming would be a boolean until actual behavioral difference exists.
