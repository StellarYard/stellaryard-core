# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in StellarYard Core, please report it responsibly.

**Do NOT open a public GitHub issue for security vulnerabilities.**

Instead, please email: **security@stellaryard.dev**

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

## Response Timeline

- **Acknowledgment**: Within 48 hours
- **Initial assessment**: Within 1 week
- **Fix or mitigation**: Depends on severity, typically within 2 weeks

## Scope

This security policy applies to:
- The `stellaryard-core` Go application
- The REST/WebSocket API
- Docker container management logic
- SQLite storage layer

## Out of Scope

- Third-party Docker images (Horizon, Soroban RPC) — report issues to their respective maintainers
- Stellar network protocol issues — report to Stellar Development Foundation

## Key Security Considerations

### Signer Interface

The Signer interface is the most security-critical component. The interface boundary must never be bypassed — secret keys must never cross the interface except through `Sign()`.

### Docker Permissions

StellarYard Core requires Docker access. The application should run with minimum necessary permissions. Do not run as root in production.

### SQLite

SQLite files contain account data and contract deployments. Ensure proper file permissions on the database file.

### API Authentication & Deployment Security

- **API Key & Session Authentication**:
  - **Loopback Development Mode**: By default, StellarYard Core binds exclusively to `127.0.0.1:8080`. When `STELLARYARD_API_KEY` is unset on a loopback interface (`127.0.0.1`, `localhost`, `::1`), authentication is disabled to streamline local development.
  - **Non-Loopback Listener Hardening**: If configured to bind to a non-loopback host (such as `0.0.0.0` or a routable network interface via `STELLARYARD_HOST`), Core enforces a fail-closed startup policy: it refuses to start unless `STELLARYARD_API_KEY` is configured with a cryptographically strong secret of at least 32 characters.
  - **Browser WebSocket & Reverse Proxy Boundary**: Standard browser WebSocket APIs cannot send custom HTTP headers (`Authorization: Bearer <key>`). Query-string API keys are explicitly **rejected** to prevent secret leakage in server logs, browser history, and HTTP referrers. When Core runs in authenticated mode, browser-based web UIs (such as `stellaryard-dashboard`) must connect through a trusted same-origin reverse proxy (or BFF) that injects the Bearer token upstream, or via a secure `stellaryard_session` cookie validated under constant-time comparison and strict Origin enforcement.
- **Transport Security (TLS / Private Network)**: StellarYard Core provides a plain HTTP listener. **Never expose the plain HTTP listener directly to untrusted networks or the public internet.** Doing so risks exposing Bearer authentication tokens and sensitive operational commands to packet inspection and interception. For remote or containerized deployments, Core **must** be placed behind a trusted TLS-terminating reverse proxy (e.g., Caddy, NGINX, AWS ALB, Cloudflare) that enforces HTTPS, or accessed strictly over an encrypted private overlay network (e.g., WireGuard, Tailscale).
- **Strong Key Provisioning**: Generate `STELLARYARD_API_KEY` using a cryptographically secure random generator (e.g., `openssl rand -hex 32` or `head -c 32 /dev/urandom | base64`). Store and inject the key via secure secret management rather than hardcoding it.

## Disclosure Policy

We follow responsible disclosure. We will:
- Credit reporters (unless they prefer anonymity)
- Not pursue legal action for good-faith security research
- Work with reporters on disclosure timing
