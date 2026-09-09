# stellaryard-core — Phase 1 Issues

---

# [Phase 1] Docker client initialization and connection

> **Repository:** stellaryard-core
> **Phase:** Phase 1 — Container Orchestration
> **Complexity:** Trivial
> **Points:** 100
> **Estimated time:** 1-2 hours

---

## Repo Context

> ⚠️ **Before starting this issue, read the following files in the repo for full context:**
> - `AGENTS.md` — Non-negotiable rules for all contributors
> - `ARCHITECTURE_ESSENTIALS.md` — Quick reference for architecture decisions
> - `ARCHITECTURE.md` — Full architecture document (read if essentials don't cover your question)
> - `ROADMAP.md` — Check that your task is scoped and update it when done
>
> **If this issue touches signing, keys, or the data model → flag for maintainer review before merging.**

stellaryard-core is the orchestration engine for StellarYard. It manages Docker containers for Horizon (Stellar's API layer) and Soroban RPC. This issue initializes the Docker client that connects to the local Docker daemon — the foundation for all container operations.

---

## Problem

All container operations (start, stop, status, logs) require a connection to the Docker daemon. Without a properly initialized Docker client, no container management is possible. The client must handle the case where Docker is not running (e.g., developer hasn't started Docker Desktop) gracefully.

---

## Scope

**In scope:**
- Initialize Docker SDK client using `github.com/docker/docker/client`
- Connect via environment variable or default socket
- Return a `*Client` type wrapping the Docker SDK
- Handle connection errors when Docker daemon is unavailable
- Add `Close()` method for cleanup
- Define container name constants ("horizon", "soroban-rpc")

**Out of scope:**
- Starting, stopping, or inspecting containers (separate issues)
- Health checks (separate issues)
- WebSocket log streaming (separate issue)
- Retry logic (separate issue: C-1.07)

---

## What "Done" Looks Like

> A contributor should be able to read this section and know exactly when to stop.

When this issue is complete:

1. `internal/docker/client.go` has a `NewClient() (*Client, error)` function
2. The function connects to Docker via `docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())`
3. If Docker is unreachable, returns a descriptive error (not a panic)
4. `Client` struct wraps the Docker SDK client
5. `Close()` method cleans up the connection
6. Constants `ContainerHorizon = "horizon"` and `ContainerSorobanRPC = "soroban-rpc"` are defined
7. Tests verify: client initializes successfully, client fails gracefully when Docker is down

---

## Implementation Guidelines

> Direction, not handcuffs. Use your judgment, but stay within these boundaries.

### Key files to modify or create
- `internal/docker/client.go` — Main implementation (replace skeleton)

### Architecture constraints
- Use `docker/docker` Go SDK — do NOT shell out to `docker` CLI
- Use `docker.WithAPIVersionNegotiation()` for cross-version compatibility
- Client must be usable by all subsequent Docker operations in this repo

### Suggested approach
1. Add Docker SDK dependency: `go get github.com/docker/docker`
2. Define `Client` struct with embedded `*client.Client`
3. Implement `NewClient()`:
   - Call `docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())`
   - Ping Docker daemon to verify connection: `client.Ping(ctx)`
   - Return error if ping fails
4. Implement `Close()` to close the underlying client
5. Add container name constants

### Testing approach
- Test `NewClient()` returns non-nil client when Docker is available
- Test `NewClient()` returns error when Docker daemon is unavailable (mock or use `DOCKER_HOST` env var pointing to invalid socket)
- Test `Close()` doesn't panic

---

## Acceptance Criteria

> PRs that don't meet ALL of these criteria will be sent back for revision.

- [ ] `NewClient() (*Client, error)` exists in `internal/docker/client.go`
- [ ] Uses `docker/docker` Go SDK, not shell exec
- [ ] Uses `docker.WithAPIVersionNegotiation()`
- [ ] Returns error when Docker daemon is unreachable (not a panic)
- [ ] `Close()` method exists and doesn't panic
- [ ] Container name constants are defined
- [ ] Tests exist for both success and failure cases
- [ ] All tests pass: `go test ./internal/docker/... -v`
- [ ] `ROADMAP.md` is updated to mark this item as done
- [ ] No violations of `AGENTS.md` rules

---

## Checklist

> Tick each item in your PR description to confirm completion.

- [ ] I have read `AGENTS.md`, `ARCHITECTURE_ESSENTIALS.md`, and `ROADMAP.md`
- [ ] I understand the non-negotiable rules for this repo
- [ ] I have implemented `NewClient()` as described
- [ ] I have written tests for success and failure cases
- [ ] All existing tests still pass
- [ ] I have updated `ROADMAP.md` to reflect this work
- [ ] I have included a clear commit message
- [ ] I have not violated any rules in `AGENTS.md`

---

## Example Commit Message

`feat: initialize Docker client with connection handling`

---

## References

- `ARCHITECTURE.md` → "Tech Stack" → Docker control
- `ARCHITECTURE.md` → "Known Failure Modes" → Docker daemon unavailability
- `docker-compose.yml` — containers this client will manage

---

# [Phase 1] Docker container start method

> **Repository:** stellaryard-core
> **Phase:** Phase 1 — Container Orchestration
> **Complexity:** Trivial
> **Points:** 100
> **Estimated time:** 1-2 hours

---

## Repo Context

> ⚠️ **Before starting this issue, read the following files in the repo for full context:**
> - `AGENTS.md` — Non-negotiable rules for all contributors
> - `ARCHITECTURE_ESSENTIALS.md` — Quick reference for architecture decisions
> - `ARCHITECTURE.md` — Full architecture document (read if essentials don't cover your question)
> - `ROADMAP.md` — Check that your task is scoped and update it when done

stellaryard-core orchestrates Docker containers for Stellar development. This issue adds the ability to start a named container (Horizon or Soroban RPC) via the Docker SDK.

---

## Problem

Developers need to start containers programmatically through core's API instead of running `docker start` manually. The start method is a prerequisite for the POST /containers/{name}/start endpoint.

---

## Scope

**In scope:**
- `Start(name string) error` method on the Docker client
- Use `docker.ContainerStart()` from the Docker SDK
- Validate container name against managed list
- Return error for invalid names or Docker failures

**Out of scope:**
- Creating containers (done by docker-compose)
- Stop, status, or log methods (separate issues)
- Retry logic (separate issue: C-1.07)
- API endpoint (separate issue: C-1.11)

---

## What "Done" Looks Like

> A contributor should be able to read this section and know exactly when to stop.

1. `Start(name string) error` exists in `internal/docker/client.go`
2. Calling `Start("horizon")` starts the Horizon container via Docker SDK
3. Calling `Start("invalid-name")` returns an error
4. Calling `Start()` when Docker is unreachable returns a classified error
5. Tests verify: start succeeds, invalid name fails, Docker unreachable fails

---

## Implementation Guidelines

### Key files to modify or create
- `internal/docker/client.go` — Add `Start()` method

### Architecture constraints
- Do NOT create containers — only start existing ones (created by docker-compose)
- Use `context.Context` for all Docker SDK calls
- Map container name to Docker container ID using `docker.ContainerInspect()`

### Suggested approach
1. In `internal/docker/client.go`, implement `Start(name string) error`
2. Validate name against managed container list
3. Inspect container to get Docker container ID
4. Call `client.ContainerStart(ctx, containerID, types.ContainerStartOptions{})`
5. Return nil on success, error on failure

### Testing approach
- Test `Start("horizon")` succeeds when container exists and is stopped
- Test `Start("invalid-name")` returns error
- Test `Start()` when Docker is unreachable returns error
- Use mocked Docker client for unit tests

---

## Acceptance Criteria

> PRs that don't meet ALL of these criteria will be sent back for revision.

- [ ] `Start(name string) error` method exists
- [ ] Starts container via Docker SDK (not shell exec)
- [ ] Returns error for invalid container names
- [ ] Returns error when Docker daemon is unreachable
- [ ] Uses `context.Context` for SDK calls
- [ ] Tests cover success, invalid name, and Docker failure
- [ ] All tests pass
- [ ] `ROADMAP.md` is updated
- [ ] No violations of `AGENTS.md` rules

---

## Checklist

> Tick each item in your PR description to confirm completion.

- [ ] I have read `AGENTS.md`, `ARCHITECTURE_ESSENTIALS.md`, and `ROADMAP.md`
- [ ] I understand the non-negotiable rules for this repo
- [ ] I have implemented `Start()` as described
- [ ] I have written tests for all cases
- [ ] All existing tests still pass
- [ ] I have updated `ROADMAP.md`
- [ ] I have included a clear commit message
- [ ] I have not violated any rules in `AGENTS.md`

---

## Example Commit Message

`feat: implement Docker container start method`

---

## References

- `ARCHITECTURE.md` → "API Contract" → POST /containers/{name}/start
- `docker-compose.yml` — the containers this method will start

---

# [Phase 1] Docker container stop method

> **Repository:** stellaryard-core
> **Phase:** Phase 1 — Container Orchestration
> **Complexity:** Trivial
> **Points:** 100
> **Estimated time:** 1-2 hours

---

## Repo Context

> ⚠️ **Before starting this issue, read the following files in the repo for full context:**
> - `AGENTS.md` — Non-negotiable rules for all contributors
> - `ARCHITECTURE_ESSENTIALS.md` — Quick reference for architecture decisions
> - `ARCHITECTURE.md` — Full architecture document (read if essentials don't cover your question)
> - `ROADMAP.md` — Check that your task is scoped and update it when done

stellaryard-core orchestrates Docker containers. This issue adds the ability to stop a named container via the Docker SDK.

---

## Problem

Developers need to stop containers programmatically. Stopping an already-stopped container must NOT be treated as an error — this is a common edge case that breaks CLI scripts.

---

## Scope

**In scope:**
- `Stop(name string) error` method on the Docker client
- Use `docker.ContainerStop()` with a 10-second timeout
- Return nil if container is already stopped (not an error)
- Return error for invalid names or Docker failures

**Out of scope:**
- Creating or starting containers (separate issues)
- Force-killing containers (V1 uses graceful stop with timeout)
- API endpoint (separate issue: C-1.12)

---

## What "Done" Looks Like

> A contributor should be able to read this section and know exactly when to stop.

1. `Stop(name string) error` exists in `internal/docker/client.go`
2. Calling `Stop("horizon")` stops the Horizon container
3. Calling `Stop("horizon")` when already stopped returns nil (not error)
4. Calling `Stop("invalid-name")` returns error
5. Uses 10-second timeout before force-killing
6. Tests verify: stop succeeds, already-stopped returns nil, invalid name fails

---

## Implementation Guidelines

### Key files to modify or create
- `internal/docker/client.go` — Add `Stop()` method

### Architecture constraints
- Stopping an already-stopped container returns nil, not an error
- Use a 10-second stop timeout
- Use `context.Context` for all Docker SDK calls

### Suggested approach
1. Implement `Stop(name string) error`
2. Inspect container to check current state
3. If already stopped, return nil
4. Call `client.ContainerStop(ctx, containerID, &timeout)` with 10s timeout
5. Return nil on success, error on failure

### Testing approach
- Test `Stop("horizon")` succeeds when running
- Test `Stop("horizon")` succeeds when already stopped (returns nil)
- Test `Stop("invalid-name")` returns error
- Use mocked Docker client

---

## Acceptance Criteria

> PRs that don't meet ALL of these criteria will be sent back for revision.

- [ ] `Stop(name string) error` method exists
- [ ] Stops container via Docker SDK
- [ ] Returns nil when container is already stopped
- [ ] Returns error for invalid container names
- [ ] Uses 10-second stop timeout
- [ ] Tests cover: running, already-stopped, invalid name
- [ ] All tests pass
- [ ] `ROADMAP.md` is updated
- [ ] No violations of `AGENTS.md` rules

---

## Checklist

> Tick each item in your PR description to confirm completion.

- [ ] I have read `AGENTS.md`, `ARCHITECTURE_ESSENTIALS.md`, and `ROADMAP.md`
- [ ] I understand the non-negotiable rules for this repo
- [ ] I have implemented `Stop()` as described
- [ ] I have written tests for all cases
- [ ] All existing tests still pass
- [ ] I have updated `ROADMAP.md`
- [ ] I have included a clear commit message
- [ ] I have not violated any rules in `AGENTS.md`

---

## Example Commit Message

`feat: implement Docker container stop method`

---

## References

- `ARCHITECTURE.md` → "API Contract" → POST /containers/{name}/stop

---

# [Phase 1] Docker container status inspection

> **Repository:** stellaryard-core
> **Phase:** Phase 1 — Container Orchestration
> **Complexity:** Trivial
> **Points:** 100
> **Estimated time:** 1-2 hours

---

## Repo Context

> ⚠️ **Before starting this issue, read the following files in the repo for full context:**
> - `AGENTS.md` — Non-negotiable rules for all contributors
> - `ARCHITECTURE_ESSENTIALS.md` — Quick reference for architecture decisions
> - `ARCHITECTURE.md` — Full architecture document (read if essentials don't cover your question)
> - `ROADMAP.md` — Check that your task is scoped and update it when done

stellaryard-core needs to report the current state of each container. This issue adds a status inspection method that returns state, health, and start time.

---

## Problem

Consumers (dashboard, CLI) need to know whether containers are running, stopped, or unhealthy. Without status inspection, the dashboard can't show container state and the CLI can't implement `stellaryard containers status`.

---

## Scope

**In scope:**
- `Status(name string) (*models.ContainerStatus, error)` method
- Use `docker.ContainerInspect()` to get container state
- Map Docker state to `models.ContainerStatus` fields
- Return error if container doesn't exist

**Out of scope:**
- Listing all container statuses (separate issue: C-1.05)
- Health checks (separate issues: C-1.08, C-1.09)
- API endpoint (separate issue: C-1.13)

---

## What "Done" Looks Like

> A contributor should be able to read this section and know exactly when to stop.

1. `Status(name string) (*models.ContainerStatus, error)` exists
2. Returns correct state: "running", "stopped", or "error"
3. Returns correct health: "healthy", "unhealthy", "starting", or "" (no healthcheck)
4. Returns correct start time from Docker
5. Returns error for invalid container names
6. Tests verify all state mappings

---

## Implementation Guidelines

### Key files to modify or create
- `internal/docker/client.go` — Add `Status()` method
- `internal/models/container.go` — Verify ContainerStatus model (already exists)

### Architecture constraints
- Map Docker's internal state strings to our simpler model:
  - Docker "running" → our "running"
  - Docker "exited" → our "stopped"
  - Docker "dead" → our "error"
- If container has no healthcheck configured, Health should be empty string, not "healthy"

### Suggested approach
1. Implement `Status(name string) (*models.ContainerStatus, error)`
2. Call `client.ContainerInspect(ctx, containerID)`
3. Map `container.State.Status` to our State string
4. Map `container.State.Health.Status` to our Health string
5. Parse `container.State.StartedAt` for start time

### Testing approach
- Test with mocked inspect responses for each state
- Test health mapping for healthy, unhealthy, starting, and no-healthcheck cases
- Test error for nonexistent container

---

## Acceptance Criteria

> PRs that don't meet ALL of these criteria will be sent back for revision.

- [ ] `Status()` method exists and returns `*models.ContainerStatus`
- [ ] Correctly maps Docker "running" → "running"
- [ ] Correctly maps Docker "exited" → "stopped"
- [ ] Correctly maps Docker "dead" → "error"
- [ ] Health field is empty when no healthcheck is configured
- [ ] Returns error for invalid container names
- [ ] Tests cover all state and health mappings
- [ ] All tests pass
- [ ] `ROADMAP.md` is updated
- [ ] No violations of `AGENTS.md` rules

---

## Checklist

> Tick each item in your PR description to confirm completion.

- [ ] I have read `AGENTS.md`, `ARCHITECTURE_ESSENTIALS.md`, and `ROADMAP.md`
- [ ] I understand the non-negotiable rules for this repo
- [ ] I have implemented `Status()` as described
- [ ] I have written tests for all state mappings
- [ ] All existing tests still pass
- [ ] I have updated `ROADMAP.md`
- [ ] I have included a clear commit message
- [ ] I have not violated any rules in `AGENTS.md`

---

## Example Commit Message

`feat: implement Docker container status inspection`

---

## References

- `ARCHITECTURE.md` → "Data Models" → ContainerStatus
- `internal/models/container.go` — ContainerStatus struct

---

# [Phase 1] Docker container list all statuses

> **Repository:** stellaryard-core
> **Phase:** Phase 1 — Container Orchestration
> **Complexity:** Trivial
> **Points:** 100
> **Estimated time:** 1-2 hours

---

## Repo Context

> ⚠️ **Before starting this issue, read the following files in the repo for full context:**
> - `AGENTS.md` — Non-negotiable rules for all contributors
> - `ARCHITECTURE_ESSENTIALS.md` — Quick reference for architecture decisions
> - `ARCHITECTURE.md` — Full architecture document (read if essentials don't cover your question)
> - `ROADMAP.md` — Check that your task is scoped and update it when done

stellaryard-core manages two containers: Horizon and Soroban RPC. This issue adds a method to get the status of all managed containers in a single call.

---

## Problem

The GET /containers endpoint and the dashboard need to show all container statuses at once. Without a batch status method, consumers would need to make two separate API calls and handle partial failures manually.

---

## Scope

**In scope:**
- `ListStatus() ([]models.ContainerStatus, error)` method
- Returns status for both "horizon" and "soroban-rpc"
- If one container's status fails, still return the other's status
- Order: horizon first, then soroban-rpc

**Out of scope:**
- Health checks (separate issues)
- API endpoint (separate issue: C-1.13)
- Filtering or sorting containers

---

## What "Done" Looks Like

> A contributor should be able to read this section and know exactly when to stop.

1. `ListStatus() ([]models.ContainerStatus, error)` exists
2. Returns status for both containers when both are running
3. Returns partial results when one container is unavailable (doesn't fail the whole list)
4. Order is consistent: horizon first, soroban-rpc second
5. Tests verify: both containers, partial failure, empty list

---

## Implementation Guidelines

### Key files to modify or create
- `internal/docker/client.go` — Add `ListStatus()` method

### Architecture constraints
- Partial failure is acceptable — return what you can, log what failed
- Do NOT fail the entire list because one container is unreachable
- Use the `Status()` method from C-1.04 internally

### Suggested approach
1. Define managed container names as a slice: `["horizon", "soroban-rpc"]`
2. Iterate over names, call `Status()` for each
3. Collect successful results, log failures
4. Return collected results (even if partial)

### Testing approach
- Test both containers running → returns 2 statuses
- Test one container unavailable → returns 1 status, no error
- Test both unavailable → returns empty slice, no error

---

## Acceptance Criteria

> PRs that don't meet ALL of these criteria will be sent back for revision.

- [ ] `ListStatus()` method exists
- [ ] Returns both container statuses when available
- [ ] Returns partial results when one container fails
- [ ] Does NOT return error when one container is unavailable
- [ ] Order is consistent (horizon first)
- [ ] Tests cover: both, partial, empty cases
- [ ] All tests pass
- [ ] `ROADMAP.md` is updated
- [ ] No violations of `AGENTS.md` rules

---

## Checklist

> Tick each item in your PR description to confirm completion.

- [ ] I have read `AGENTS.md`, `ARCHITECTURE_ESSENTIALS.md`, and `ROADMAP.md`
- [ ] I understand the non-negotiable rules for this repo
- [ ] I have implemented `ListStatus()` as described
- [ ] I have written tests for all cases
- [ ] All existing tests still pass
- [ ] I have updated `ROADMAP.md`
- [ ] I have included a clear commit message
- [ ] I have not violated any rules in `AGENTS.md`

---

## Example Commit Message

`feat: implement Docker container list all statuses`

---

## References

- `ARCHITECTURE.md` → "API Contract" → GET /containers

---

# [Phase 1] Docker error classification (transient vs permanent)

> **Repository:** stellaryard-core
> **Phase:** Phase 1 — Container Orchestration
> **Complexity:** Trivial
> **Points:** 100
> **Estimated time:** 1-2 hours

---

## Repo Context

> ⚠️ **Before starting this issue, read the following files in the repo for full context:**
> - `AGENTS.md` — Non-negotiable rules for all contributors
> - `ARCHITECTURE_ESSENTIALS.md` — Quick reference for architecture decisions
> - `ARCHITECTURE.md` — Full architecture document (read if essentials don't cover your question)
> - `ROADMAP.md` — Check that your task is scoped and update it when done

stellaryard-core's Docker operations can fail for different reasons. Some failures are temporary (Docker daemon restarting) and some are permanent (container doesn't exist). This issue adds error classification so consumers can decide whether to retry.

---

## Problem

The CLI uses exit code 2 for "core unreachable" (retryable) and exit code 3 for "core error" (not retryable). Without error classification in core, the CLI can't make this distinction. This is documented as a known failure mode in ARCHITECTURE.md.

---

## Scope

**In scope:**
- Define `TransientError` and `PermanentError` types
- Add `IsTransient(err error) bool` helper
- Classify errors in Start, Stop, Status, ListStatus methods

**Out of scope:**
- Retry logic (separate issue: C-1.07)
- API error responses (separate issue: C-1.17)
- CLI exit codes (separate repo issue)

---

## What "Done" Looks Like

> A contributor should be able to read this section and know exactly when to stop.

1. `internal/docker/errors.go` defines TransientError and PermanentError types
2. `IsTransient(err) bool` returns true for transient errors
3. Connection refused → TransientError
4. Container not found → PermanentError
5. All Docker methods return classified errors
6. Tests verify classification for each error type

---

## Implementation Guidelines

### Key files to modify or create
- `internal/docker/errors.go` — New file with error types
- `internal/docker/client.go` — Update methods to wrap errors

### Architecture constraints
- Keep error types simple — just transient vs permanent, no hierarchy
- Error messages should be human-readable and include container name
- Preserve original error (use `fmt.Errorf("...: %w", err)`)

### Suggested approach
1. Create `internal/docker/errors.go`
2. Define `TransientError` struct with `Err error` and `Error() string`
3. Define `PermanentError` struct similarly
4. Implement `IsTransient(err error) bool` that checks error chain
5. In Start/Stop/Status/ListStatus, wrap Docker SDK errors:
   - `docker.IsErrNotFound(err)` → PermanentError
   - Connection errors → TransientError

### Testing approach
- Test IsTransient returns true for connection errors
- Test IsTransient returns false for container-not-found errors
- Test error wrapping preserves original message

---

## Acceptance Criteria

> PRs that don't meet ALL of these criteria will be sent back for revision.

- [ ] `TransientError` and `PermanentError` types exist
- [ ] `IsTransient(err) bool` helper exists
- [ ] Connection errors classified as TransientError
- [ ] Container-not-found errors classified as PermanentError
- [ ] Original error message preserved in wrapped error
- [ ] Tests verify classification for each type
- [ ] All tests pass
- [ ] `ROADMAP.md` is updated
- [ ] No violations of `AGENTS.md` rules

---

## Checklist

> Tick each item in your PR description to confirm completion.

- [ ] I have read `AGENTS.md`, `ARCHITECTURE_ESSENTIALS.md`, and `ROADMAP.md`
- [ ] I understand the non-negotiable rules for this repo
- [ ] I have implemented error types as described
- [ ] I have written tests for all classification cases
- [ ] All existing tests still pass
- [ ] I have updated `ROADMAP.md`
- [ ] I have included a clear commit message
- [ ] I have not violated any rules in `AGENTS.md`

---

## Example Commit Message

`feat: classify Docker errors as transient or permanent`

---

## References

- `ARCHITECTURE.md` → "Known Failure Modes" → Docker daemon unavailability
- `ARCHITECTURE_ESSENTIALS.md` → "What would break"

---

# [Phase 1] Docker retry with exponential backoff

> **Repository:** stellaryard-core
> **Phase:** Phase 1 — Container Orchestration
> **Complexity:** Small
> **Points:** 150
> **Estimated time:** 2-4 hours

---

## Repo Context

> ⚠️ **Before starting this issue, read the following files in the repo for full context:**
> - `AGENTS.md` — Non-negotiable rules for all contributors
> - `ARCHITECTURE_ESSENTIALS.md` — Quick reference for architecture decisions
> - `ARCHITECTURE.md` — Full architecture document (read if essentials don't cover your question)
> - `ROADMAP.md` — Check that your task is scoped and update it when done

stellaryard-core's Docker operations can fail transiently when Docker daemon is restarting or temporarily unavailable. This issue adds retry logic so operations recover automatically.

---

## Problem

When Docker Desktop updates or restarts, container operations fail immediately. Users see errors and must manually retry. This is documented as the #1 known failure mode in ARCHITECTURE.md. Retry with backoff allows automatic recovery without user intervention.

---

## Scope

**In scope:**
- `RetryOnTransient()` helper function
- Retry only transient errors (using IsTransient from C-1.06)
- Exponential backoff: 100ms, 200ms, 400ms (max 3 retries)
- Apply to Start() and Stop() methods

**Out of scope:**
- Retrying Status() or ListStatus() — these should fail fast
- Circuit breaker pattern (future work)
- Retry configuration (hardcoded for V1)

---

## What "Done" Looks Like

> A contributor should be able to read this section and know exactly when to stop.

1. `internal/docker/retry.go` has `RetryOnTransient(ctx, fn, maxRetries) error`
2. Retries up to 3 times with exponential backoff (100ms, 200ms, 400ms)
3. Only retries transient errors — permanent errors fail immediately
4. Start() and Stop() methods use retry wrapper
5. Status() and ListStatus() do NOT use retry (fail fast)
6. Each retry is logged at INFO level
7. Tests verify: retry succeeds on second attempt, permanent errors not retried, gives up after max

---

## Implementation Guidelines

### Key files to modify or create
- `internal/docker/retry.go` — New file with retry logic
- `internal/docker/client.go` — Apply retry to Start() and Stop()

### Architecture constraints
- Max 3 retries — don't retry forever
- Log each retry at INFO level
- Use `time.Sleep` for backoff delays (simple, no external dependency)
- Check `ctx.Done()` between retries for cancellation

### Suggested approach
1. Create `internal/docker/retry.go`
2. Implement `RetryOnTransient(ctx, fn func() error, maxRetries int) error`
3. Loop: call fn(), if error and IsTransient, sleep and retry
4. If error and not transient, return immediately
5. If all retries exhausted, return last error
6. In client.go, wrap Start() and Stop() with RetryOnTransient

### Testing approach
- Mock: first call returns transient error, second succeeds → test passes
- Mock: call returns permanent error → no retry, returns immediately
- Mock: all retries return transient error → gives up after 3

---

## Acceptance Criteria

> PRs that don't meet ALL of these criteria will be sent back for revision.

- [ ] `RetryOnTransient()` exists in `internal/docker/retry.go`
- [ ] Retries up to 3 times
- [ ] Backoff is exponential: 100ms, 200ms, 400ms
- [ ] Only retries transient errors (uses IsTransient)
- [ ] Start() uses retry wrapper
- [ ] Stop() uses retry wrapper
- [ ] Status() does NOT use retry
- [ ] Each retry is logged
- [ ] Tests verify retry behavior
- [ ] All tests pass
- [ ] `ROADMAP.md` is updated
- [ ] No violations of `AGENTS.md` rules

---

## Checklist

> Tick each item in your PR description to confirm completion.

- [ ] I have read `AGENTS.md`, `ARCHITECTURE_ESSENTIALS.md`, and `ROADMAP.md`
- [ ] I understand the non-negotiable rules for this repo
- [ ] I have implemented retry logic as described
- [ ] I have written tests for retry behavior
- [ ] All existing tests still pass
- [ ] I have updated `ROADMAP.md`
- [ ] I have included a clear commit message
- [ ] I have not violated any rules in `AGENTS.md`

---

## Example Commit Message

`feat: add retry with exponential backoff for transient Docker errors`

---

## References

- `ARCHITECTURE.md` → "Known Failure Modes" → Docker daemon unavailability
- Depends on: C-1.06 (error classification)

---

# [Phase 1] Horizon health check with actual API verification

> **Repository:** stellaryard-core
> **Phase:** Phase 1 — Container Orchestration
> **Complexity:** Small
> **Points:** 150
> **Estimated time:** 2-4 hours

---

## Repo Context

> ⚠️ **Before starting this issue, read the following files in the repo for full context:**
> - `AGENTS.md` — Non-negotiable rules for all contributors
> - `ARCHITECTURE_ESSENTIALS.md` — Quick reference for architecture decisions
> - `ARCHITECTURE.md` — Full architecture document (read if essentials don't cover your question)
> - `ROADMAP.md` — Check that your task is scoped and update it when done

stellaryard-core manages Horizon containers. This issue adds a health check that verifies Horizon is actually responsive to Stellar API queries — not just that its HTTP port is open.

---

## Problem

Horizon can return HTTP 200 before it's fully synced. A health check that only checks port availability gives false positives. Consumers see "healthy" but queries return empty data. This is documented as a known failure mode in ARCHITECTURE.md.

---

## Scope

**In scope:**
- `HealthCheckHorizon() (string, error)` method
- HTTP GET to Horizon root endpoint `/`
- Verify `horizon_version` field exists in response
- Return "healthy", "unhealthy", or "starting"

**Out of scope:**
- Soroban RPC health check (C-1.09)
- Health check dispatcher (C-1.10)
- Docker-level healthcheck configuration

---

## What "Done" Looks Like

> A contributor should be able to read this section and know exactly when to stop.

1. `HealthCheckHorizon() (string, error)` exists in `internal/docker/client.go`
2. Makes HTTP GET to `http://localhost:8000/`
3. If response has `horizon_version` → "healthy"
4. If response missing `horizon_version` → "unhealthy"
5. If connection refused → "starting"
6. 5-second HTTP timeout
7. Tests use mocked HTTP server

---

## Implementation Guidelines

### Key files to modify or create
- `internal/docker/client.go` — Add `HealthCheckHorizon()` method

### Architecture constraints
- Do NOT use Docker's built-in healthcheck — verify at Stellar API level
- Do NOT shell out to `curl` — use Go's `net/http`
- Horizon root returns: `{"_links": {...}, "horizon_version": "v20.0.0", ...}`

### Suggested approach
1. Create `http.Client` with 5-second timeout
2. GET `http://localhost:8000/`
3. Parse JSON response body
4. Check for `horizon_version` key in parsed map
5. Return appropriate health status

### Testing approach
- Use `httptest.NewServer` to mock Horizon
- Mock 1: valid response with `horizon_version` → "healthy"
- Mock 2: response without `horizon_version` → "unhealthy"
- Mock 3: server not running → "starting"

---

## Acceptance Criteria

> PRs that don't meet ALL of these criteria will be sent back for revision.

- [ ] `HealthCheckHorizon()` method exists
- [ ] Returns "healthy" when Horizon has valid metadata
- [ ] Returns "unhealthy" when response missing `horizon_version`
- [ ] Returns "starting" when Horizon unreachable
- [ ] Has 5-second HTTP timeout
- [ ] Tests use mocked HTTP server (not real Horizon)
- [ ] All tests pass
- [ ] `ROADMAP.md` is updated
- [ ] No violations of `AGENTS.md` rules

---

## Checklist

> Tick each item in your PR description to confirm completion.

- [ ] I have read `AGENTS.md`, `ARCHITECTURE_ESSENTIALS.md`, and `ROADMAP.md`
- [ ] I understand the non-negotiable rules for this repo
- [ ] I have implemented `HealthCheckHorizon()` as described
- [ ] I have written tests for all three health states
- [ ] All existing tests still pass
- [ ] I have updated `ROADMAP.md`
- [ ] I have included a clear commit message
- [ ] I have not violated any rules in `AGENTS.md`

---

## Example Commit Message

`feat: implement Horizon health check with actual API verification`

---

## References

- `ARCHITECTURE.md` → "Known Failure Modes" → Container health-check false positives
- `ARCHITECTURE.md` → "Data Models" → ContainerStatus.Health

---

# [Phase 1] Soroban RPC health check

> **Repository:** stellaryard-core
> **Phase:** Phase 1 — Container Orchestration
> **Complexity:** Trivial
> **Points:** 100
> **Estimated time:** 1-2 hours

---

## Repo Context

> ⚠️ **Before starting this issue, read the following files in the repo for full context:**
> - `AGENTS.md` — Non-negotiable rules for all contributors
> - `ARCHITECTURE_ESSENTIALS.md` — Quick reference for architecture decisions
> - `ARCHITECTURE.md` — Full architecture document (read if essentials don't cover your question)
> - `ROADMAP.md` — Check that your task is scoped and update it when done

stellaryard-core manages Soroban RPC containers. This issue adds a health check for Soroban RPC, following the same pattern as the Horizon health check.

---

## Problem

Soroban RPC containers need health verification. Without it, the dashboard and CLI can't report accurate container health.

---

## Scope

**In scope:**
- `HealthCheckSorobanRPC() (string, error)` method
- Query Soroban RPC health endpoint
- Return "healthy", "unhealthy", or "starting"

**Out of scope:**
- Horizon health check (C-1.08)
- Health check dispatcher (C-1.10)

---

## What "Done" Looks Like

> A contributor should be able to read this section and know exactly when to stop.

1. `HealthCheckSorobanRPC() (string, error)` exists
2. Queries Soroban RPC health endpoint
3. Returns correct health status
4. Tests verify all three states

---

## Implementation Guidelines

### Key files to modify or create
- `internal/docker/client.go` — Add `HealthCheckSorobanRPC()` method

### Architecture constraints
- Follow the same pattern as C-1.08 (Horizon health check)
- Soroban RPC health endpoint may differ — check the Soroban RPC docs

### Suggested approach
1. Implement `HealthCheckSorobanRPC()` following C-1.08 pattern
2. HTTP GET to Soroban RPC health endpoint
3. Parse response, determine health status

### Testing approach
- Use `httptest.NewServer` to mock Soroban RPC
- Test all three health states

---

## Acceptance Criteria

> PRs that don't meet ALL of these criteria will be sent back for revision.

- [ ] `HealthCheckSorobanRPC()` method exists
- [ ] Returns correct health status for all states
- [ ] Tests use mocked HTTP server
- [ ] All tests pass
- [ ] `ROADMAP.md` is updated
- [ ] No violations of `AGENTS.md` rules

---

## Checklist

> Tick each item in your PR description to confirm completion.

- [ ] I have read `AGENTS.md`, `ARCHITECTURE_ESSENTIALS.md`, and `ROADMAP.md`
- [ ] I understand the non-negotiable rules for this repo
- [ ] I have implemented `HealthCheckSorobanRPC()` as described
- [ ] I have written tests for all cases
- [ ] All existing tests still pass
- [ ] I have updated `ROADMAP.md`
- [ ] I have included a clear commit message
- [ ] I have not violated any rules in `AGENTS.md`

---

## Example Commit Message

`feat: implement Soroban RPC health check`

---

## References

- `ARCHITECTURE.md` → "Data Models" → ContainerStatus.Health
- Depends on: C-1.08 (pattern to follow)

---

# [Phase 1] Unified container health check dispatcher

> **Repository:** stellaryard-core
> **Phase:** Phase 1 — Container Orchestration
> **Complexity:** Trivial
> **Points:** 100
> **Estimated time:** 1-2 hours

---

## Repo Context

> ⚠️ **Before starting this issue, read the following files in the repo for full context:**
> - `AGENTS.md` — Non-negotiable rules for all contributors
> - `ARCHITECTURE_ESSENTIALS.md` — Quick reference for architecture decisions
> - `ARCHITECTURE.md` — Full architecture document (read if essentials don't cover your question)
> - `ROADMAP.md` — Check that your task is scoped and update it when done

stellaryard-core has container-specific health checks. This issue adds a unified dispatcher that routes to the correct health check based on container name.

---

## Problem

API handlers and consumers need a single `HealthCheck(name)` method, not separate Horizon and Soroban RPC methods. A dispatcher simplifies the API surface.

---

## Scope

**In scope:**
- `HealthCheck(name string) (string, error)` method
- Dispatch to HealthCheckHorizon() for "horizon"
- Dispatch to HealthCheckSorobanRPC() for "soroban-rpc"
- Return error for invalid names

**Out of scope:**
- Implementing the individual health checks (C-1.08, C-1.09)
- API endpoint (separate issue)

---

## What "Done" Looks Like

> A contributor should be able to read this section and know exactly when to stop.

1. `HealthCheck(name string) (string, error)` exists
2. Routes to correct health check based on name
3. Returns error for unknown names
4. Tests verify routing for each container and error for invalid names

---

## Implementation Guidelines

### Key files to modify or create
- `internal/docker/client.go` — Add `HealthCheck()` method

### Architecture constraints
- Thin dispatcher — no logic beyond routing
- Depends on C-1.08 and C-1.09 being merged, but can be developed in parallel with mock tests

### Suggested approach
1. Implement `HealthCheck(name string) (string, error)`
2. Switch on name, call appropriate health check
3. Return error for unknown names

### Testing approach
- Test with mock health checks (not real ones)
- Verify correct dispatch for each container name
- Verify error for invalid names

---

## Acceptance Criteria

> PRs that don't meet ALL of these criteria will be sent back for revision.

- [ ] `HealthCheck(name)` method exists
- [ ] Routes to Horizon check for "horizon"
- [ ] Routes to Soroban RPC check for "soroban-rpc"
- [ ] Returns error for unknown names
- [ ] Tests verify routing and error cases
- [ ] All tests pass
- [ ] `ROADMAP.md` is updated
- [ ] No violations of `AGENTS.md` rules

---

## Checklist

> Tick each item in your PR description to confirm completion.

- [ ] I have read `AGENTS.md`, `ARCHITECTURE_ESSENTIALS.md`, and `ROADMAP.md`
- [ ] I understand the non-negotiable rules for this repo
- [ ] I have implemented `HealthCheck()` as described
- [ ] I have written tests for all cases
- [ ] All existing tests still pass
- [ ] I have updated `ROADMAP.md`
- [ ] I have included a clear commit message
- [ ] I have not violated any rules in `AGENTS.md`

---

## Example Commit Message

`feat: implement unified container health check dispatcher`

---

## References

- Depends on: C-1.08 (Horizon), C-1.09 (Soroban RPC)

---

# [Phase 1] POST /containers/{name}/start endpoint

> **Repository:** stellaryard-core
> **Phase:** Phase 1 — Container Orchestration
> **Complexity:** Small
> **Points:** 150
> **Estimated time:** 2-4 hours

---

## Repo Context

> ⚠️ **Before starting this issue, read the following files in the repo for full context:**
> - `AGENTS.md` — Non-negotiable rules for all contributors
> - `ARCHITECTURE_ESSENTIALS.md` — Quick reference for architecture decisions
> - `ARCHITECTURE.md` — Full architecture document (read if essentials don't cover your question)
> - `ROADMAP.md` — Check that your task is scoped and update it when done

stellaryard-core exposes a REST API for consumers. This issue implements the first container API endpoint: starting a named container.

---

## Problem

The dashboard and CLI need an HTTP endpoint to start containers. Without this, there's no way to manage containers remotely (even on localhost).

---

## Scope

**In scope:**
- POST /containers/{name}/start handler
- Extract name from URL path using chi
- Call Docker client Start(name)
- Return appropriate HTTP status codes and JSON responses
- Wire to router

**Out of scope:**
- Stop endpoint (C-1.12)
- List endpoint (C-1.13)
- WebSocket endpoints (C-1.14)

---

## What "Done" Looks Like

> A contributor should be able to read this section and know exactly when to stop.

1. `handleContainerStart(dc)` handler exists in `internal/api/handlers.go`
2. Route registered in `internal/api/router.go`
3. Returns 200 with `{"status": "started"}` on success
4. Returns 404 with `{"error": "container not found"}` for invalid names
5. Returns 503 for transient Docker errors
6. Returns 500 for other errors
7. All responses are JSON with correct Content-Type
8. Tests verify all status codes and response shapes

---

## Implementation Guidelines

### Key files to modify or create
- `internal/api/handlers.go` — Implement handler
- `internal/api/router.go` — Register route

### Architecture constraints
- All responses must be JSON with `Content-Type: application/json`
- Use error classification from C-1.06 to determine status codes
- Response shape must match `api/openapi.yaml`

### Suggested approach
1. Implement `handleContainerStart(dc)` in handlers.go
2. Use `chi.URLParam(r, "name")` to extract container name
3. Call `dc.Start(name)`
4. On success: return 200 with `{"status": "started"}`
5. On PermanentError: return 404
6. On TransientError: return 503
7. On other error: return 500
8. Register route in router.go

### Testing approach
- Test 200 on successful start
- Test 404 for invalid container name
- Test 503 when Docker unreachable
- Test response is valid JSON with correct shape

---

## Acceptance Criteria

> PRs that don't meet ALL of these criteria will be sent back for revision.

- [ ] POST /containers/{name}/start handler exists
- [ ] Returns 200 with `{"status": "started"}` on success
- [ ] Returns 404 for invalid container names
- [ ] Returns 503 for transient Docker errors
- [ ] All responses are JSON
- [ ] Route is registered in router.go
- [ ] Tests verify all status codes and response shapes
- [ ] All tests pass
- [ ] `ROADMAP.md` is updated
- [ ] No violations of `AGENTS.md` rules

---

## Checklist

> Tick each item in your PR description to confirm completion.

- [ ] I have read `AGENTS.md`, `ARCHITECTURE_ESSENTIALS.md`, and `ROADMAP.md`
- [ ] I understand the non-negotiable rules for this repo
- [ ] I have implemented the handler as described
- [ ] I have written tests for all status codes
- [ ] All existing tests still pass
- [ ] I have updated `ROADMAP.md`
- [ ] I have included a clear commit message
- [ ] I have not violated any rules in `AGENTS.md`

---

## Example Commit Message

`feat: implement POST /containers/{name}/start endpoint`

---

## References

- `ARCHITECTURE.md` → "API Contract" → POST /containers/{name}/start
- `api/openapi.yaml` → /containers/{name}/start
- Depends on: C-1.02 (Start method), C-1.06 (error classification)

---

# [Phase 1] POST /containers/{name}/stop endpoint

> **Repository:** stellaryard-core
> **Phase:** Phase 1 — Container Orchestration
> **Complexity:** Small
> **Points:** 150
> **Estimated time:** 2-4 hours

---

## Repo Context

> ⚠️ **Before starting this issue, read the following files in the repo for full context:**
> - `AGENTS.md` — Non-negotiable rules for all contributors
> - `ARCHITECTURE_ESSENTIALS.md` — Quick reference for architecture decisions
> - `ARCHITECTURE.md` — Full architecture document (read if essentials don't cover your question)
> - `ROADMAP.md` — Check that your task is scoped and update it when done

stellaryard-core's API needs a stop endpoint. This issue implements POST /containers/{name}/stop.

---

## Problem

The dashboard and CLI need an HTTP endpoint to stop containers. Stopping an already-stopped container must return 200, not 400 or 409.

---

## Scope

**In scope:**
- POST /containers/{name}/stop handler
- Return 200 on success (including already-stopped)
- Return 404 for invalid names
- Return 503 for transient errors

**Out of scope:**
- Start endpoint (C-1.11)
- Force-kill functionality

---

## What "Done" Looks Like

> A contributor should be able to read this section and know exactly when to stop.

1. `handleContainerStop(dc)` handler exists
2. Returns 200 with `{"status": "stopped"}` on success
3. Returns 200 when container is already stopped (not error)
4. Returns 404 for invalid names
5. Returns 503 for transient errors
6. Tests verify all cases

---

## Implementation Guidelines

### Key files to modify or create
- `internal/api/handlers.go` — Implement handler
- `internal/api/router.go` — Register route

### Architecture constraints
- Follow the same pattern as C-1.11
- Stopping an already-stopped container returns 200

### Suggested approach
1. Implement `handleContainerStop(dc)` following C-1.11 pattern
2. Call `dc.Stop(name)`
3. Return appropriate status codes

### Testing approach
- Test 200 on successful stop
- Test 200 when already stopped
- Test 404 for invalid name
- Test 503 for Docker unreachable

---

## Acceptance Criteria

> PRs that don't meet ALL of these criteria will be sent back for revision.

- [ ] POST /containers/{name}/stop handler exists
- [ ] Returns 200 on success
- [ ] Returns 200 when already stopped
- [ ] Returns 404 for invalid names
- [ ] Returns 503 for transient errors
- [ ] Tests verify all cases
- [ ] All tests pass
- [ ] `ROADMAP.md` is updated
- [ ] No violations of `AGENTS.md` rules

---

## Checklist

> Tick each item in your PR description to confirm completion.

- [ ] I have read `AGENTS.md`, `ARCHITECTURE_ESSENTIALS.md`, and `ROADMAP.md`
- [ ] I understand the non-negotiable rules for this repo
- [ ] I have implemented the handler as described
- [ ] I have written tests for all cases
- [ ] All existing tests still pass
- [ ] I have updated `ROADMAP.md`
- [ ] I have included a clear commit message
- [ ] I have not violated any rules in `AGENTS.md`

---

## Example Commit Message

`feat: implement POST /containers/{name}/stop endpoint`

---

## References

- `ARCHITECTURE.md` → "API Contract" → POST /containers/{name}/stop
- Depends on: C-1.03 (Stop method), C-1.06 (error classification)

---

# [Phase 1] GET /containers endpoint

> **Repository:** stellaryard-core
> **Phase:** Phase 1 — Container Orchestration
> **Complexity:** Small
> **Points:** 150
> **Estimated time:** 2-4 hours

---

## Repo Context

> ⚠️ **Before starting this issue, read the following files in the repo for full context:**
> - `AGENTS.md` — Non-negotiable rules for all contributors
> - `ARCHITECTURE_ESSENTIALS.md` — Quick reference for architecture decisions
> - `ARCHITECTURE.md` — Full architecture document (read if essentials don't cover your question)
> - `ROADMAP.md` — Check that your task is scoped and update it when done

stellaryard-core's API needs a list endpoint for container statuses. This is polled by the dashboard every 3 seconds.

---

## Problem

The dashboard needs to display all container statuses at once. Without a list endpoint, the dashboard would need to make separate calls for each container.

---

## Scope

**In scope:**
- GET /containers handler
- Return JSON array of ContainerStatus objects
- Return 503 if Docker unreachable

**Out of scope:**
- Filtering or pagination
- Individual container endpoints (C-1.11, C-1.12)

---

## What "Done" Looks Like

> A contributor should be able to read this section and know exactly when to stop.

1. `handleContainerList(dc)` handler exists
2. Returns 200 with JSON array of container statuses
3. Returns `[]` (empty array), not `null`, when no containers
4. Returns 503 when Docker unreachable
5. Response matches openapi.yaml schema

---

## Implementation Guidelines

### Key files to modify or create
- `internal/api/handlers.go` — Implement handler
- `internal/api/router.go` — Register route

### Architecture constraints
- Return `[]`, not `null`, for empty list
- This endpoint is polled frequently — keep it fast
- Response shape must match openapi.yaml

### Suggested approach
1. Implement `handleContainerList(dc)`
2. Call `dc.ListStatus()`
3. Marshal result to JSON
4. Return 200 with array

### Testing approach
- Test returns array with both containers
- Test returns empty array when no containers
- Test returns 503 when Docker unreachable

---

## Acceptance Criteria

> PRs that don't meet ALL of these criteria will be sent back for revision.

- [ ] GET /containers handler exists
- [ ] Returns 200 with JSON array
- [ ] Returns `[]` for empty list (not `null`)
- [ ] Returns 503 when Docker unreachable
- [ ] Response matches openapi.yaml
- [ ] Tests verify all cases
- [ ] All tests pass
- [ ] `ROADMAP.md` is updated
- [ ] No violations of `AGENTS.md` rules

---

## Checklist

> Tick each item in your PR description to confirm completion.

- [ ] I have read `AGENTS.md`, `ARCHITECTURE_ESSENTIALS.md`, and `ROADMAP.md`
- [ ] I understand the non-negotiable rules for this repo
- [ ] I have implemented the handler as described
- [ ] I have written tests for all cases
- [ ] All existing tests still pass
- [ ] I have updated `ROADMAP.md`
- [ ] I have included a clear commit message
- [ ] I have not violated any rules in `AGENTS.md`

---

## Example Commit Message

`feat: implement GET /containers endpoint`

---

## References

- `ARCHITECTURE.md` → "API Contract" → GET /containers
- `api/openapi.yaml` → /containers
- Depends on: C-1.05 (ListStatus method)

---

# [Phase 1] WebSocket /containers/{name}/logs connection handler

> **Repository:** stellaryard-core
> **Phase:** Phase 1 — Container Orchestration
> **Complexity:** Small
> **Points:** 150
> **Estimated time:** 2-4 hours

---

## Repo Context

> ⚠️ **Before starting this issue, read the following files in the repo for full context:**
> - `AGENTS.md` — Non-negotiable rules for all contributors
> - `ARCHITECTURE_ESSENTIALS.md` — Quick reference for architecture decisions
> - `ARCHITECTURE.md` — Full architecture document (read if essentials don't cover your question)
> - `ROADMAP.md` — Check that your task is scoped and update it when done

stellaryard-core streams container logs via WebSocket. This issue handles the WebSocket upgrade and connection establishment.

---

## Problem

The dashboard and CLI need real-time log streaming. WebSocket is the transport — this issue establishes the connection (streaming logic is a separate issue).

---

## Scope

**In scope:**
- WebSocket upgrade handler using gorilla/websocket
- Validate container name before upgrade
- Return 404 for invalid names
- Return 101 on successful upgrade

**Out of scope:**
- Actual log streaming (C-1.15)
- Ring buffer (C-1.15)
- Client-side WS handling

---

## What "Done" Looks Like

> A contributor should be able to read this section and know exactly when to stop.

1. `handleContainerLogs(dc)` handler upgrades HTTP to WebSocket
2. Returns 404 for invalid container names (before upgrade)
3. Returns 101 on successful upgrade
4. `CheckOrigin` allows localhost origins
5. Tests verify upgrade and validation

---

## Implementation Guidelines

### Key files to modify or create
- `internal/api/handlers.go` — Implement handler

### Architecture constraints
- Use `gorilla/websocket` for WS upgrade
- `CheckOrigin` must return true for localhost
- Validate container name BEFORE attempting upgrade

### Suggested approach
1. Create `websocket.Upgrader` with `CheckOrigin: func(r) bool { return true }`
2. Validate container name, return 404 if invalid
3. Upgrade connection with `upgrader.Upgrade(w, r, nil)`
4. On success, begin log streaming (stub for now)

### Testing approach
- Test returns 404 for invalid name (before upgrade)
- Test upgrades to WebSocket successfully
- Test CheckOrigin allows localhost

---

## Acceptance Criteria

> PRs that don't meet ALL of these criteria will be sent back for revision.

- [ ] WebSocket upgrade handler exists
- [ ] Returns 404 for invalid container names
- [ ] Returns 101 on successful upgrade
- [ ] CheckOrigin allows localhost
- [ ] Tests verify upgrade and validation
- [ ] All tests pass
- [ ] `ROADMAP.md` is updated
- [ ] No violations of `AGENTS.md` rules

---

## Checklist

> Tick each item in your PR description to confirm completion.

- [ ] I have read `AGENTS.md`, `ARCHITECTURE_ESSENTIALS.md`, and `ROADMAP.md`
- [ ] I understand the non-negotiable rules for this repo
- [ ] I have implemented the handler as described
- [ ] I have written tests for upgrade and validation
- [ ] All existing tests still pass
- [ ] I have updated `ROADMAP.md`
- [ ] I have included a clear commit message
- [ ] I have not violated any rules in `AGENTS.md`

---

## Example Commit Message

`feat: implement WebSocket container logs connection handler`

---

## References

- `ARCHITECTURE.md` → "API Contract" → WS /containers/{name}/logs
- `api/openapi.yaml` → /containers/{name}/logs

---

# [Phase 1] Docker log streaming with ring buffer

> **Repository:** stellaryard-core
> **Phase:** Phase 1 — Container Orchestration
> **Complexity:** High
> **Points:** 200
> **Estimated time:** 4-8 hours

---

## Repo Context

> ⚠️ **Before starting this issue, read the following files in the repo for full context:**
> - `AGENTS.md` — Non-negotiable rules for all contributors
> - `ARCHITECTURE_ESSENTIALS.md` — Quick reference for architecture decisions
> - `ARCHITECTURE.md` — Full architecture document (read if essentials don't cover your question)
> - `ROADMAP.md` — Check that your task is scoped and update it when done

stellaryard-core streams container logs to consumers via WebSocket. This issue implements the actual log streaming from Docker to WebSocket clients, with a ring buffer to prevent memory exhaustion.

---

## Problem

Without a ring buffer, fast-producing containers (debug-level Horizon logs) can exhaust core's memory. This is documented as a known failure mode in ARCHITECTURE.md. The ring buffer caps memory usage at 1000 lines per connection.

---

## Scope

**In scope:**
- `Logs(name string) (<-chan string, error)` method on Docker client
- Stream logs using Docker SDK's `ContainerLogs()` with Follow: true
- Ring buffer: max 1000 lines per connection, oldest evicted first
- Send log lines to WebSocket clients
- Handle WebSocket close gracefully
- Handle Docker log stream errors

**Out of scope:**
- WebSocket connection handler (C-1.14 — already exists)
- Log filtering or search
- Multiple consumer support (one WS per connection)

---

## What "Done" Looks Like

> A contributor should be able to read this section and know exactly when to stop.

1. `Logs(name string) (<-chan string, error)` exists in Docker client
2. Streams logs using Docker SDK `ContainerLogs()` with Follow: true
3. Ring buffer caps at 1000 lines per connection
4. Oldest lines evicted when buffer full
5. WebSocket handler reads from channel and sends to client
6. WS close is handled gracefully
7. Channel closes when container stops or WS disconnects
8. Tests verify: streaming, ring buffer, WS close handling

---

## Implementation Guidelines

### Key files to modify or create
- `internal/docker/client.go` — Add `Logs()` method
- `internal/api/handlers.go` — Update WS handler to stream logs

### Architecture constraints
- Docker log output has an 8-byte header frame — strip it before sending
- Ring buffer is critical — without it, memory exhaustion is guaranteed
- Each WS connection gets its own ring buffer (not shared)
- Channel must be closed when WS connection closes

### Suggested approach
1. Implement `Logs(name string) (<-chan string, error)` in client.go
2. Call `client.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{Follow: true})`
3. Parse Docker log output (strip 8-byte header frame)
4. Send lines to buffered channel (cap 1000)
5. When channel full, drop oldest line (ring buffer behavior)
6. In handlers.go, read from channel and write to WebSocket
7. On WS close, close the log channel

### Testing approach
- Test Logs() returns channel that receives lines
- Test ring buffer evicts oldest when full
- Test channel closes when container stops
- Test WS close is handled gracefully
- Use mock Docker client for unit tests

---

## Acceptance Criteria

> PRs that don't meet ALL of these criteria will be sent back for revision.

- [ ] `Logs()` method exists and returns `<-chan string`
- [ ] Streams logs using Docker SDK (not shell exec)
- [ ] Ring buffer caps at 1000 lines
- [ ] Oldest lines evicted when buffer full
- [ ] Docker log header frame is stripped
- [ ] WS handler reads from channel
- [ ] WS close closes the log channel
- [ ] Tests verify streaming, buffer, and close behavior
- [ ] All tests pass
- [ ] `ROADMAP.md` is updated
- [ ] No violations of `AGENTS.md` rules

---

## Checklist

> Tick each item in your PR description to confirm completion.

- [ ] I have read `AGENTS.md`, `ARCHITECTURE_ESSENTIALS.md`, and `ROADMAP.md`
- [ ] I understand the non-negotiable rules for this repo
- [ ] I have implemented `Logs()` with ring buffer as described
- [ ] I have updated the WS handler to stream logs
- [ ] I have written tests for all cases
- [ ] All existing tests still pass
- [ ] I have updated `ROADMAP.md`
- [ ] I have included a clear commit message
- [ ] I have not violated any rules in `AGENTS.md`

---

## Example Commit Message

`feat: implement Docker log streaming with ring buffer`

---

## References

- `ARCHITECTURE.md` → "Known Failure Modes" → Log streaming unbounded memory
- `ARCHITECTURE.md` → "Known Failure Modes" → WS log streaming memory growth
- Depends on: C-1.14 (WS connection handler)

---

# [Phase 1] API CORS middleware for localhost

> **Repository:** stellaryard-core
> **Phase:** Phase 1 — Container Orchestration
> **Complexity:** Trivial
> **Points:** 100
> **Estimated time:** 1-2 hours

---

## Repo Context

> ⚠️ **Before starting this issue, read the following files in the repo for full context:**
> - `AGENTS.md` — Non-negotiable rules for all contributors
> - `ARCHITECTURE_ESSENTIALS.md` — Quick reference for architecture decisions
> - `ARCHITECTURE.md` — Full architecture document (read if essentials don't cover your question)
> - `ROADMAP.md` — Check that your task is scoped and update it when done

stellaryard-core serves API on port 8080. The dashboard runs on port 3000. Without CORS headers, the dashboard can't make API calls from the browser.

---

## Problem

CORS misconfiguration is the #1 "works in curl, not in browser" bug. All dashboard API calls fail silently without proper CORS headers. This is documented as a known failure mode in ARCHITECTURE.md.

---

## Scope

**In scope:**
- CORS middleware for chi router
- Allow origin: localhost (all ports)
- Allow methods: GET, POST, OPTIONS
- Allow headers: Content-Type
- Handle preflight OPTIONS requests

**Out of scope:**
- Production CORS configuration (V1 is localhost-only)
- Authentication headers (no auth in V1)

---

## What "Done" Looks Like

> A contributor should be able to read this section and know exactly when to stop.

1. CORS middleware exists in `internal/api/middleware.go`
2. Sets `Access-Control-Allow-Origin: http://localhost:*`
3. Sets `Access-Control-Allow-Methods: GET, POST, OPTIONS`
4. Sets `Access-Control-Allow-Headers: Content-Type`
5. Handles OPTIONS preflight requests
6. Middleware is applied to chi router
7. Tests verify CORS headers on responses

---

## Implementation Guidelines

### Key files to modify or create
- `internal/api/middleware.go` — Add CORS middleware
- `internal/api/router.go` — Apply middleware

### Architecture constraints
- Allow all localhost ports (not just 3000) — port may change
- For V1, this is localhost-only — no need for production CORS config

### Suggested approach
1. Create CORS middleware function in middleware.go
2. Set CORS headers on every response
3. Handle OPTIONS preflight by returning 200 with headers
4. Apply middleware to chi router in router.go

### Testing approach
- Test preflight OPTIONS returns correct headers
- Test GET request includes CORS headers
- Test POST request includes CORS headers

---

## Acceptance Criteria

> PRs that don't meet ALL of these criteria will be sent back for revision.

- [ ] CORS middleware exists
- [ ] Allows localhost origins
- [ ] Allows GET, POST, OPTIONS methods
- [ ] Allows Content-Type header
- [ ] Handles OPTIONS preflight
- [ ] Middleware applied to router
- [ ] Tests verify CORS headers
- [ ] All tests pass
- [ ] `ROADMAP.md` is updated
- [ ] No violations of `AGENTS.md` rules

---

## Checklist

> Tick each item in your PR description to confirm completion.

- [ ] I have read `AGENTS.md`, `ARCHITECTURE_ESSENTIALS.md`, and `ROADMAP.md`
- [ ] I understand the non-negotiable rules for this repo
- [ ] I have implemented CORS middleware as described
- [ ] I have written tests for CORS headers
- [ ] All existing tests still pass
- [ ] I have updated `ROADMAP.md`
- [ ] I have included a clear commit message
- [ ] I have not violated any rules in `AGENTS.md`

---

## Example Commit Message

`feat: add CORS middleware for localhost dashboard access`

---

## References

- `ARCHITECTURE.md` → "Known Failure Modes" → CORS

---

# [Phase 1] Consistent JSON error response shape

> **Repository:** stellaryard-core
> **Phase:** Phase 1 — Container Orchestration
> **Complexity:** Small
> **Points:** 150
> **Estimated time:** 2-4 hours

---

## Repo Context

> ⚠️ **Before starting this issue, read the following files in the repo for full context:**
> - `AGENTS.md` — Non-negotiable rules for all contributors
> - `ARCHITECTURE_ESSENTIALS.md` — Quick reference for architecture decisions
> - `ARCHITECTURE.md` — Full architecture document (read if essentials don't cover your question)
> - `ROADMAP.md` — Check that your task is scoped and update it when done

stellaryard-core's API must return consistent error responses. Dashboard and CLI depend on this for error display and exit code classification.

---

## Problem

Without a consistent error shape, each endpoint returns errors differently. The CLI can't classify errors for exit codes, and the dashboard can't display meaningful error messages.

---

## Scope

**In scope:**
- Error response shape: `{"error": "message", "code": "ERROR_CODE"}`
- Error codes: CONTAINER_NOT_FOUND, DOCKER_UNREACHABLE, INTERNAL_ERROR
- Helper functions: `writeError()`, `writeJSON()`
- Apply to all container endpoints

**Out of scope:**
- Error handling for non-container endpoints (Phase 2-4)
- Error logging or monitoring

---

## What "Done" Looks Like

> A contributor should be able to read this section and know exactly when to stop.

1. `writeError(w, status, code, message)` helper exists
2. `writeJSON(w, status, data)` helper exists
3. Error response shape is `{"error": "...", "code": "..."}`
4. All container endpoints use these helpers
5. Error codes are defined as constants
6. Tests verify error response shape

---

## Implementation Guidelines

### Key files to modify or create
- `internal/api/handlers.go` — Add helpers, update handlers

### Architecture constraints
- Error shape must match openapi.yaml
- Every error response must include both `error` (human-readable) and `code` (machine-parseable)
- This must be done BEFORE dashboard and CLI start building error handling

### Suggested approach
1. Define error code constants
2. Implement `writeError(w, status, code, message)` helper
3. Implement `writeJSON(w, status, data)` helper
4. Update all container handlers to use these helpers

### Testing approach
- Test writeError returns correct JSON shape
- Test writeJSON returns correct Content-Type header
- Test all container endpoints return consistent error shape

---

## Acceptance Criteria

> PRs that don't meet ALL of these criteria will be sent back for revision.

- [ ] `writeError()` helper exists
- [ ] `writeJSON()` helper exists
- [ ] Error shape is `{"error": "...", "code": "..."}`
- [ ] Error codes are defined as constants
- [ ] All container endpoints use helpers
- [ ] Tests verify error shape consistency
- [ ] All tests pass
- [ ] `ROADMAP.md` is updated
- [ ] No violations of `AGENTS.md` rules

---

## Checklist

> Tick each item in your PR description to confirm completion.

- [ ] I have read `AGENTS.md`, `ARCHITECTURE_ESSENTIALS.md`, and `ROADMAP.md`
- [ ] I understand the non-negotiable rules for this repo
- [ ] I have implemented error helpers as described
- [ ] I have updated all container handlers
- [ ] I have written tests for error shape
- [ ] All existing tests still pass
- [ ] I have updated `ROADMAP.md`
- [ ] I have included a clear commit message
- [ ] I have not violated any rules in `AGENTS.md`

---

## Example Commit Message

`feat: define consistent JSON error response shape`

---

## References

- `api/openapi.yaml` — error response schemas
- Must be done before: Dashboard and CLI error handling

---

# [Phase 1] Docker client unit tests with mocked Docker SDK

> **Repository:** stellaryard-core
> **Phase:** Phase 1 — Container Orchestration
> **Complexity:** Medium
> **Points:** 150
> **Estimated time:** 4-8 hours

---

## Repo Context

> ⚠️ **Before starting this issue, read the following files in the repo for full context:**
> - `AGENTS.md` — Non-negotiable rules for all contributors
> - `ARCHITECTURE_ESSENTIALS.md` — Quick reference for architecture decisions
> - `ARCHITECTURE.md` — Full architecture document (read if essentials don't cover your question)
> - `ROADMAP.md` — Check that your task is scoped and update it when done

stellaryard-core's Docker client has several methods. This issue adds comprehensive unit tests using a mocked Docker SDK.

---

## Problem

Docker client methods interact with the Docker daemon, which isn't available in CI. Mocked tests ensure the logic is correct without requiring Docker.

---

## Scope

**In scope:**
- Create `internal/docker/client_test.go`
- Mock Docker SDK interface
- Test all client methods: Start, Stop, Status, ListStatus, HealthCheck
- Test error cases: daemon unreachable, container not found, invalid name
- Test retry behavior with mocked transient errors

**Out of scope:**
- Integration tests against real Docker (Phase 5)
- API handler tests (separate issue: C-1.19)

---

## What "Done" Looks Like

> A contributor should be able to read this section and know exactly when to stop.

1. `internal/docker/client_test.go` exists
2. Mock Docker client implements SDK interface
3. Table-driven tests for each method
4. Tests cover success, error, and edge cases
5. Coverage > 80% for docker package
6. All tests pass

---

## Implementation Guidelines

### Key files to modify or create
- `internal/docker/client_test.go` — New test file

### Architecture constraints
- Do NOT test against real Docker daemon — use mocks
- Table-driven tests preferred (idiomatic Go)
- Cover retry behavior from C-1.07

### Suggested approach
1. Create mock Docker client implementing the SDK interface
2. Write table-driven tests for each method
3. Use `t.Run()` for subtests
4. Test success, error, and edge cases

### Testing approach
- Table-driven tests for each method
- Subtests for each scenario
- Mock returns controlled responses for each case

---

## Acceptance Criteria

> PRs that don't meet ALL of these criteria will be sent back for revision.

- [ ] `client_test.go` exists with comprehensive tests
- [ ] Mock Docker client is implemented
- [ ] Table-driven tests for each method
- [ ] Tests cover success, error, and edge cases
- [ ] Coverage > 80% for docker package
- [ ] All tests pass
- [ ] `ROADMAP.md` is updated
- [ ] No violations of `AGENTS.md` rules

---

## Checklist

> Tick each item in your PR description to confirm completion.

- [ ] I have read `AGENTS.md`, `ARCHITECTURE_ESSENTIALS.md`, and `ROADMAP.md`
- [ ] I understand the non-negotiable rules for this repo
- [ ] I have implemented comprehensive tests as described
- [ ] Tests use mocked Docker client (not real)
- [ ] Coverage > 80% for docker package
- [ ] All existing tests still pass
- [ ] I have updated `ROADMAP.md`
- [ ] I have included a clear commit message
- [ ] I have not violated any rules in `AGENTS.md`

---

## Example Commit Message

`test: add comprehensive Docker client unit tests with mocks`

---

## References

- `ARCHITECTURE.md` → "Testing expectations"

---

# [Phase 1] API handler integration tests

> **Repository:** stellaryard-core
> **Phase:** Phase 1 — Container Orchestration
> **Complexity:** Medium
> **Points:** 150
> **Estimated time:** 4-8 hours

---

## Repo Context

> ⚠️ **Before starting this issue, read the following files in the repo for full context:**
> - `AGENTS.md` — Non-negotiable rules for all contributors
> - `ARCHITECTURE_ESSENTIALS.md` — Quick reference for architecture decisions
> - `ARCHITECTURE.md` — Full architecture document (read if essentials don't cover your question)
> - `ROADMAP.md` — Check that your task is scoped and update it when done

stellaryard-core's API handlers need integration tests. This issue tests all container endpoints using Go's `httptest` package.

---

## Problem

Without integration tests, API handlers may return wrong status codes, wrong response shapes, or missing CORS headers. Integration tests catch these issues before they reach consumers.

---

## Scope

**In scope:**
- Create `internal/api/handlers_test.go`
- Use `httptest.NewServer` with chi router
- Mock Docker client
- Test all container endpoints: start, stop, list
- Test error responses match shape from C-1.17
- Test CORS headers are present

**Out of scope:**
- WebSocket testing (requires special setup)
- Docker client testing (C-1.18)
- Performance testing

---

## What "Done" Looks Like

> A contributor should be able to read this section and know exactly when to stop.

1. `internal/api/handlers_test.go` exists
2. Tests use httptest.NewServer with chi router
3. Mock Docker client returns controlled responses
4. Tests verify HTTP status codes, response bodies, headers
5. Tests verify error response shape
6. Tests verify CORS headers
7. All tests pass

---

## Implementation Guidelines

### Key files to modify or create
- `internal/api/handlers_test.go` — New test file

### Architecture constraints
- Use httptest — not a real server
- Mock the Docker client, not the HTTP layer
- Verify JSON response shape matches openapi.yaml

### Suggested approach
1. Create mock Docker client for API tests
2. Set up httptest.NewServer with chi router
3. Write tests for each endpoint
4. Test HTTP status codes, response bodies, headers

### Testing approach
- Table-driven tests for each endpoint
- Verify response shape with JSON parsing
- Verify CORS headers

---

## Acceptance Criteria

> PRs that don't meet ALL of these criteria will be sent back for revision.

- [ ] `handlers_test.go` exists
- [ ] Tests use httptest.NewServer
- [ ] Tests verify HTTP status codes
- [ ] Tests verify response bodies
- [ ] Tests verify error response shape
- [ ] Tests verify CORS headers
- [ ] All tests pass
- [ ] `ROADMAP.md` is updated
- [ ] No violations of `AGENTS.md` rules

---

## Checklist

> Tick each item in your PR description to confirm completion.

- [ ] I have read `AGENTS.md`, `ARCHITECTURE_ESSENTIALS.md`, and `ROADMAP.md`
- [ ] I understand the non-negotiable rules for this repo
- [ ] I have implemented integration tests as described
- [ ] Tests use httptest, not real server
- [ ] All existing tests still pass
- [ ] I have updated `ROADMAP.md`
- [ ] I have included a clear commit message
- [ ] I have not violated any rules in `AGENTS.md`

---

## Example Commit Message

`test: add API handler integration tests for container endpoints`

---

## References

- `ARCHITECTURE.md` → "Testing expectations"
- `api/openapi.yaml` — response shapes to verify

---

# [Phase 1] OpenAPI spec update for container endpoints

> **Repository:** stellaryard-core
> **Phase:** Phase 1 — Container Orchestration
> **Complexity:** Small
> **Points:** 150
> **Estimated time:** 2-4 hours

---

## Repo Context

> ⚠️ **Before starting this issue, read the following files in the repo for full context:**
> - `AGENTS.md` — Non-negotiable rules for all contributors
> - `ARCHITECTURE_ESSENTIALS.md` — Quick reference for architecture decisions
> - `ARCHITECTURE.md` — Full architecture document (read if essentials don't cover your question)
> - `ROADMAP.md` — Check that your task is scoped and update it when done

stellaryard-core's OpenAPI spec is the source of truth for API consumers. This issue updates the spec to fully document all container endpoints.

---

## Problem

The OpenAPI spec has skeleton definitions for container endpoints. Without complete schemas, consumers (dashboard, CLI) can't generate accurate typed clients.

---

## Scope

**In scope:**
- Update container endpoint schemas in openapi.yaml
- Add error response schemas with codes
- Add WebSocket documentation
- Add CORS headers documentation
- Add response examples
- Validate spec with linter

**Out of scope:**
- Non-container endpoint schemas (Phase 2-4)
- Auto-generation from Go code

---

## What "Done" Looks Like

> A contributor should be able to read this section and know exactly when to stop.

1. All container endpoints have full request/response schemas
2. Error responses documented with code field
3. WebSocket endpoint documented
4. Response examples provided
5. Spec validates with OpenAPI linter
6. Spec matches actual handler implementation

---

## Implementation Guidelines

### Key files to modify or create
- `api/openapi.yaml` — Update spec

### Architecture constraints
- Spec must match actual implementation
- CLI and dashboard generate clients from this file
- Include examples for every response shape

### Suggested approach
1. Update container endpoint schemas
2. Add error response schemas
3. Add WebSocket documentation
4. Add examples
5. Validate with linter

### Testing approach
- Run OpenAPI linter
- Verify spec matches handler responses

---

## Acceptance Criteria

> PRs that don't meet ALL of these criteria will be sent back for revision.

- [ ] All container endpoints have full schemas
- [ ] Error responses documented with codes
- [ ] WebSocket endpoint documented
- [ ] Response examples provided
- [ ] Spec validates with linter
- [ ] Spec matches actual implementation
- [ ] `ROADMAP.md` is updated
- [ ] No violations of `AGENTS.md` rules

---

## Checklist

> Tick each item in your PR description to confirm completion.

- [ ] I have read `AGENTS.md`, `ARCHITECTURE_ESSENTIALS.md`, and `ROADMAP.md`
- [ ] I understand the non-negotiable rules for this repo
- [ ] I have updated the OpenAPI spec as described
- [ ] Spec validates with linter
- [ ] Spec matches actual implementation
- [ ] All existing tests still pass
- [ ] I have updated `ROADMAP.md`
- [ ] I have included a clear commit message
- [ ] I have not violated any rules in `AGENTS.md`

---

## Example Commit Message

`docs: update OpenAPI spec with container endpoint schemas`

---

## References

- `api/openapi.yaml` — the file to update
- Must match: All container handlers (C-1.11 through C-1.17)
