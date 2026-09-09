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

### API Authentication

The V1 API has no authentication (localhost only). If exposing the API beyond localhost, implement appropriate auth before doing so.

## Disclosure Policy

We follow responsible disclosure. We will:
- Credit reporters (unless they prefer anonymity)
- Not pursue legal action for good-faith security research
- Work with reporters on disclosure timing
