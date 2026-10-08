<p align="center">
  <img src="https://img.shields.io/badge/stellar-yard-blue?style=for-the-badge&logo=stellar&logoColor=white" alt="StellarYard"/>
</p>

<h1 align="center">stellaryard-core</h1>

<p align="center">
  The orchestration engine and API server for StellarYard — a local development environment for the Stellar network.
</p>

<p align="center">
  <a href="https://github.com/StellarYard/stellaryard-core/blob/main/LICENSE"><img src="https://img.shields.io/github/license/StellarYard/stellaryard-core?style=flat-square" alt="License"/></a>
  <a href="https://github.com/StellarYard/stellaryard-core/actions"><img src="https://img.shields.io/github/actions/workflow/status/StellarYard/stellaryard-core/ci.yml?style=flat-square&label=CI" alt="CI"/></a>
  <a href="https://github.com/StellarYard/stellaryard-core/issues"><img src="https://img.shields.io/github/issues/StellarYard/stellaryard-core?style=flat-square" alt="Issues"/></a>
</p>

---

## What is StellarYard?

StellarYard is a self-contained local development environment for Stellar. It runs Horizon and Soroban RPC as managed Docker containers, with a REST/WebSocket API for programmatic control.

**Core** is the backend service that:
- Manages Docker containers (Horizon + Soroban RPC)
- Exposes a REST/WebSocket API consumed by the [dashboard](../stellaryard-dashboard) and [CLI](../stellaryard-cli)
- Handles account creation, contract deployment, and ledger queries
- Signs transactions via a pluggable `Signer` interface

## Quick Start

```bash
# Clone and run
git clone https://github.com/StellarYard/stellaryard-core.git
cd stellaryard-core
docker-compose up -d
go run cmd/server/main.go
```

The API is available at `http://localhost:8080/api/v1`.

## Architecture

```
┌─────────────────┐     ┌──────────────────┐
│ stellaryard-     │     │ stellaryard-cli  │
│ dashboard (web)  │     │ (terminal)       │
└────────┬─────────┘     └────────┬─────────┘
         │      REST + WS (localhost only)
         └──────────┬───────────────┘
                     │
           ┌─────────▼──────────┐
           │  stellaryard-core   │
           │  ┌───────────────┐  │
           │  │ API layer     │  │
           │  │ Signer        │  │
           │  │ Docker        │──► Horizon + Soroban RPC
           │  │ SQLite        │  │
           │  └───────────────┘  │
           └─────────────────────┘
```

## API Endpoints

Endpoints below reflect what is implemented today. Anything returning `501` or a
placeholder payload is tracked in the issue tracker and is not yet functional.

| Method | Path | Status | Purpose |
|--------|------|--------|---------|
| GET | `/health` | ✅ | Liveness probe |
| POST | `/containers/{name}/start` | ✅ | Start a container |
| POST | `/containers/{name}/stop` | ✅ | Stop a container |
| GET | `/containers` | ✅ | List container statuses |
| POST | `/accounts` | ✅ | Create a test account (placeholder keypair) |
| GET | `/accounts` | ✅ | List managed accounts |
| GET | `/accounts/{publicKey}` | ✅ | Fetch one account |
| POST | `/contracts/deploy` | 🚧 `501` | Deploy a WASM contract |
| POST | `/contracts/{id}/invoke` | 🚧 `501` | Invoke a contract method |
| GET | `/ledger/snapshot` | 🚧 placeholder | Current ledger state |
| GET | `/ledger/transactions` | 🚧 placeholder | Recent transactions |
| WS | `/containers/{name}/logs` | 📋 planned | Stream container logs |

Full spec: [`api/openapi.yaml`](./api/openapi.yaml)

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.23 |
| HTTP Router | [chi](https://github.com/go-chi/chi) |
| WebSocket | gorilla/websocket |
| Docker | Official Go SDK |
| Database | SQLite |
| API Spec | OpenAPI 3.0 |

## Roadmap

Status is tracked against the code on `main`, not intent.

| Phase | Scope | Status |
|-------|-------|--------|
| 0 — Foundation | Scaffold, OpenAPI, SQLite, Signer | ✅ Done |
| 1 — Containers | Docker start/stop/status | ✅ Done |
| 1 — Containers | Health check verification, WebSocket log streaming | 🚧 In progress |
| 2 — Accounts | Create/list/get against SQLite | ✅ Done (placeholder keypairs, no Friendbot) |
| 3 — Ledger | Snapshot, transactions | 🚧 Stub — returns placeholder data |
| 4 — Contracts | WASM deploy, invocation | 🚧 Stub — returns `501 Not Implemented` |
| 5 — Hardening | Tests, error handling, documentation | 🔜 Not started |

Full roadmap: [`ROADMAP.md`](./ROADMAP.md)

## Contributing

We welcome contributions! See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

- Check [open issues](https://github.com/StellarYard/stellaryard-core/issues) for `ready` tasks
- Issues labeled `good-first-issue` are ideal for first-time contributors
- Every PR must update `ROADMAP.md`

## Maintainers

| Name | GitHub | Contact |
|------|--------|---------|
| Adejumo-2 | [@Adejumo-2](https://github.com/Adejumo-2) | [Telegram](https://t.me/Adejumo-2) |

## Community

- [GitHub Discussions](https://github.com/StellarYard/stellaryard-core/discussions)

## License

[Apache 2.0](./LICENSE)

---

<p align="center">
  Built for the Stellar ecosystem
</p>
