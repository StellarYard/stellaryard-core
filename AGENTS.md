# AGENTS.md — stellaryard-core

Instructions for AI coding agents working in this repository. Read this before making changes.

## What this repo is

Backend orchestration engine + API server for StellarYard. No UI. Consumed by `stellaryard-dashboard` and `stellaryard-cli`, which are separate repos and cannot be modified from here. See `PRD.md` and `ARCHITECTURE.md` for full context; `ARCHITECTURE_ESSENTIALS.md` for a fast-reference outline if you don't need the full document.

## Non-negotiable rules

1. **Never bypass the `Signer` interface to handle a secret key directly.** Any code path that reads, writes, logs, or transmits a raw `SecretKey` outside the `Signer` interface boundary is a rejected change, regardless of how small or "just testnet" it looks. If your task seems to require this, stop and flag it rather than implementing a workaround.
2. **`/api/openapi.yaml` is the source of truth for the API contract.** `stellaryard-dashboard` and `stellaryard-cli` are built against it. If your change adds, removes, or changes the shape of an endpoint, you must update `openapi.yaml` in the same PR. Do not make undocumented API changes.
3. **Breaking API changes require versioning, not silent replacement.** If you must change an existing endpoint's contract in a breaking way, version it (e.g. `/api/v2/...`) rather than mutating `/api/v1/...` behavior in place — dashboard and CLI in other repos may not update in lockstep with you.
4. **No Postgres, no heavyweight framework additions.** This is a local single-developer tool; SQLite and the existing minimal stack (`chi`, `docker/docker` SDK, `stellar/go` SDK) are deliberate choices, not gaps to fill in. Don't add infrastructure weight without a documented reason in `ARCHITECTURE.md`.
5. **Update `ROADMAP.md` in every contribution.** Before opening a PR, check `ROADMAP.md` and: (a) mark your completed item as done, or (b) add a new entry if you've surfaced work that wasn't previously tracked, or (c) note if your change invalidates an existing roadmap assumption. A PR that changes functionality without touching `ROADMAP.md` should be treated as incomplete — flag this yourself before requesting review, don't wait for a maintainer to catch it.

## Before starting any task

- Read `ARCHITECTURE_ESSENTIALS.md` first. Only open the full `ARCHITECTURE.md` if you need detail it doesn't cover.
- Check `ROADMAP.md` for whether your task is already scoped there and what phase it belongs to.
- Check `/api/openapi.yaml` if your task touches any HTTP-facing behavior.

## Testing expectations

- Any change to Docker orchestration logic should be tested against the local docker-compose setup, not mocked entirely — container lifecycle bugs are the most disruptive class of bug in this repo since both other repos depend on containers actually starting correctly.
- Any change to the `Signer` interface or its implementations requires explicit test coverage for the boundary itself (i.e., confirm a raw key never crosses it), not just happy-path functional tests.

## What you cannot do from this repo

- Modify `stellaryard-dashboard` or `stellaryard-cli` code. If your change requires a corresponding change there, note it in your PR description and in `ROADMAP.md` — don't attempt a cross-repo edit.
- Add mainnet signing, mainnet key custody, or mainnet endpoints. This is explicitly out of scope until stated otherwise in `PRD.md`.

## Known pitfalls

- **Docker daemon unavailability has no retry/backoff.** If you're working on container endpoints, add exponential backoff and a circuit breaker — don't just surface raw errors. The CLI depends on this for exit code 2 vs 3 distinction.
- **SQLite has no crash recovery beyond WAL replay.** If your change touches account or contract writes, ensure the write is atomic (single transaction) or that partial state is detectable and recoverable on restart.
- **Container health checks are not yet defined.** Horizon's HTTP 200 doesn't mean it's synced. If you're building status endpoints, verify actual Stellar API responsiveness, not just port availability.
- **WS log streaming has no memory bound.** Add a ring-buffer cap per connection. Unbounded streaming is the most likely path to a production OOM.
- **The `openapi.yaml` can drift from implementation.** If your change modifies an endpoint's response shape, update the spec in the same PR AND verify it matches what the code actually returns. A spec-conformance test is needed but doesn't exist yet — until it does, manual verification is your responsibility.
