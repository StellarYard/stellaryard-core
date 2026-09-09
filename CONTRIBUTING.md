# Contributing to StellarYard Core

Thank you for your interest in contributing to StellarYard Core! This document provides guidelines and instructions for contributing.

## Code of Conduct

Be respectful, constructive, and professional. We're building tools for the Stellar ecosystem together.

## Getting Started

### Prerequisites

- Go 1.23+
- Docker and Docker Compose
- Git

### Setup

```bash
# Clone the repo
git clone https://github.com/StellarYard/stellaryard-core.git
cd stellaryard-core

# Install dependencies
go mod tidy

# Start local development environment
docker-compose up -d

# Run the server
go run cmd/server/main.go
```

### Running Tests

```bash
go test ./...
go vet ./...
```

## How to Contribute

### Finding Issues

1. Check the [open issues](https://github.com/StellarYard/stellaryard-core/issues) for tasks labeled `ready`
2. Issues labeled `good-first-issue` are ideal for first-time contributors
3. Read the issue description carefully — each issue includes acceptance criteria and implementation guidelines

### Submitting Changes

1. **Fork** the repository
2. **Create a branch** from `main`:
   ```bash
   git checkout -b feat/your-feature-name
   ```
3. **Make your changes** following the coding standards below
4. **Write or update tests** for your changes
5. **Update ROADMAP.md** as part of your PR — this is required
6. **Commit** with a descriptive message:
   ```bash
   git commit -m "feat: add Docker client initialization"
   ```
7. **Push** your branch:
   ```bash
   git push origin feat/your-feature-name
   ```
8. **Open a Pull Request** against `main`

### Commit Message Format

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

[optional body]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring without behavior change
- `chore`: Maintenance tasks

Examples:
```
feat(docker): add container health check with exponential backoff
fix(api): handle WebSocket disconnection gracefully
docs: update README with setup instructions
test(storage): add SQLite migration rollback tests
```

### Pull Request Guidelines

- **One logical change per PR** — don't bundle unrelated changes
- **Include tests** — PRs without test coverage will be sent back
- **Update ROADMAP.md** — every PR must update the roadmap to reflect what was done
- **Keep PRs small** — ideally under 500 lines of diff
- **Describe what and why** — not just what changed, but why

## Coding Standards

### Go Style

- Follow [Effective Go](https://go.dev/doc/effective_go) conventions
- Use `gofmt` and `go vet` — no manual formatting
- Error messages should be lowercase, no punctuation: `fmt.Errorf("container %s not found", name)`
- Prefer table-driven tests
- Use `context.Context` as the first parameter for functions that do I/O
- Interface names should describe behavior: `Signer`, `Storage`, not `SignerInterface`

### Project Structure

```
cmd/server/          — Entry point
internal/api/        — HTTP handlers and routing
internal/docker/     — Docker SDK integration
internal/signer/     — Signing interface and implementations
internal/storage/    — SQLite and migrations
internal/models/     — Data models
internal/account/    — Account service
internal/contract/   — Contract service
internal/ledger/     — Ledger service
api/                 — OpenAPI spec
```

### Architecture Rules (Non-Negotiable)

1. **The Signer interface must never be bypassed** — no direct key access outside `Sign()`
2. **Core's OpenAPI spec is the single source of truth** — never hand-write API clients
3. **No cross-repo modifications** — note dependencies in PR descriptions
4. **ROADMAP.md must be updated in every PR**

## Reporting Issues

- Use GitHub Issues for bug reports and feature requests
- Include steps to reproduce for bugs
- Include your Go version and OS for environment-related issues

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
