# StellarYard — Ecosystem Reconnaissance & Critical Review

## Phase 1: Ecosystem Landscape (September 2026)

### What the Stellar Stack Offers Right Now

- **Consensus layer**: Stellar Consensus Protocol (SCP), live since 2015
- **Smart contracts**: Soroban (Rust-based), live on mainnet since Protocol 20 (Feb 2024)
- **DEX**: Built-in decentralized exchange, one of the oldest in crypto
- **SEP standards**: 40+ standards covering payments, anchors, identity, contracts
- **Stellar CLI**: `stellar` command-line tool for contract deployment and interaction
- **Soroban RPC**: Local and testnet RPC endpoints for contract invocation
- **Horizon API**: REST API for ledger queries, account management, transaction submission

### Drips Wave Program — Current State

| Metric | Value |
|--------|-------|
| Approved repos | 737 |
| Organizations | 442 |
| Completed Waves | 8 (Jan–Aug 2026) |
| Budget per wave | $60K–$75K |
| Total disbursed (Jan–Apr) | $255K |
| Contributors (latest wave) | 782 |

### Approved Repo Categories (by domain)

| Domain | Examples | Saturation |
|--------|----------|------------|
| **Escrow** | Trustless Work, SafeTrust, KindFi | High — 3+ major projects |
| **Payments** | Stellopay, Stellar-MarketPay, Stellar-MicroPay | High — multiple payroll/payment tools |
| **RWA / Real Estate** | akkuea, StellarRent | Medium — growing, 4x points |
| **Freelance / Marketplace** | OFFER-HUB | Medium |
| **DeFi** | stellar-portfolio-rebalancer | Medium |
| **Developer Tools** | soroban-cost-linter, soroban-cost-estimator, soroban-budget-assert, astroid-sdk, vellar-sdk | Medium — tooling is fragmented |
| **Cross-chain** | OverSync (ETH-Stellar bridge) | Low |
| **AI / Agent** | Stellar Agentic Hackathon projects | New — 2x points |
| **Infrastructure** | Stellar-K8s | Low |

### SDF Funding Priorities

- **Soroban adoption**: $100M Soroban Adoption Fund
- **RWA tokenization**: Stellar is the leading RWA chain ($3B+ in tokenized assets)
- **Stablecoin infrastructure**: USDC on Stellar is a primary use case
- **Developer tooling**: SCF Build Track provides $15K–$150K grants
- **Open source**: Wave program specifically incentivizes open-source maintenance

### Where StellarYard Fits

**White space identified:**
- **Local development environment** — No approved Wave repo provides a Docker-based local dev environment for Stellar. Developers currently use `stellar-cli` + manual Docker setup or remote testnet.
- **Multi-component orchestration** — Most tools are single-purpose (SDK, linter, estimator). Nothing orchestrates Horizon + Soroban RPC as a unified local stack.
- **Developer experience tooling** — CLI + dashboard for managing local Stellar development is genuinely missing.

**Competition:**
- `stellar-cli` (official) — handles contract deployment but not local infrastructure
- Manual `docker-compose` setups — scattered, undocumented, not maintained
- Soroban docs examples — reference docker-compose but don't provide a managed solution

---

## Phase 3: Critical Review

### Idea: StellarYard — Local Development Environment for Stellar

**One-paragraph description:** StellarYard is a self-contained local development environment for Stellar that runs Horizon and Soroban RPC as managed Docker containers, with a CLI for automation and a web dashboard for visual management.

### Weak Spots (Specific)

1. **No Stellar primitive dependency** — StellarYard doesn't use Soroban contracts, SEP standards, or Stellar-specific APIs in a way that creates lock-in. It wraps Docker containers that run Stellar software. A developer could replicate this with a `docker-compose.yml` file and 20 lines of shell script. The "Stellar-specific" value is thin.

2. **Cold start problem** — The target users (Stellar developers) already have working workflows. `stellar-cli` + local RPC is already documented. You're asking developers to switch from "works fine" to "works better" — and "better" needs to be dramatically better, not marginally.

3. **Maintenance burden** — Horizon and Soroban RPC versions change with Stellar protocol upgrades. Every protocol upgrade requires testing and updating StellarYard. This is ongoing maintenance work that doesn't generate new value — it just keeps existing value from breaking.

4. **Low ceiling** — The maximum value StellarYard can deliver is "saves 30 minutes of setup." That's a one-time benefit per developer. Once they've set up their environment, StellarYard adds no ongoing value unless it also provides monitoring, log analysis, or development workflow improvements.

5. **Three repos for a wrapper** — The 3-repo split (core/cli/dashboard) is overengineered for what amounts to a Docker orchestrator with a REST API and a web UI. A single repo with packages would be simpler to maintain and contribute to.

6. **No on-chain component** — Wave programs reward projects that advance the Stellar ecosystem. A Docker wrapper doesn't advance Soroban adoption, doesn't create new SEP standards, doesn't enable new use cases. It's infrastructure plumbing.

### Does It Depend on Infrastructure That Doesn't Exist?

No — all components (Horizon, Soroban RPC, Docker, Go, React) exist and are stable. The project can be built with current technology.

### Regulatory / Compliance Walls

None — this is a developer tool, not a financial product.

### Stellar Fit Rating

**WEAK** — The project is useful but doesn't leverage Stellar-specific primitives in a meaningful way. It's a developer convenience tool, not an ecosystem advancement.

### MVP Feasibility

**HIGH** — The project can be built quickly because it's wrapping existing Docker images. The core value proposition (managed containers + CLI + dashboard) is straightforward.

### Verdict: **CONDITIONAL**

StellarYard is a legitimate developer tool that solves a real (if small) problem. However:

- It does NOT justify 3 separate repos for Wave submission (overhead > value)
- It does NOT advance Stellar ecosystem capabilities (no new primitives, standards, or use cases)
- It DOES fill a genuine gap in local development tooling
- It CAN work as a Wave submission IF:
  1. Consolidated into 1-2 repos (core+cli in one, dashboard in another OR everything in one)
  2. Positioned as "developer tooling" not "infrastructure"
  3. Includes genuinely useful features beyond docker-compose (log streaming, health monitoring, account management, contract deployment helpers)
  4. Has a compelling demo showing time savings vs. manual setup

### Recommendation

**Consider pivoting the value proposition.** Instead of "local dev environment," position StellarYard as:

- **"Stellar Development Hub"** — a unified tool that combines local dev, testnet access, contract deployment, account management, and monitoring in one place
- **Add Soroban-specific features** — contract cost estimation, storage simulation, test account funding, ABI management — things that only make sense in a Stellar context
- **Consolidate repos** — merge core+cli into one repo, keep dashboard separate. This doubles your Wave surface area without tripling maintenance.

If you proceed with the current design, the 3-repo split will make Wave submission harder (3x the hygiene work, 3x the branch protection, 3x the CI). The project's value doesn't justify that complexity.
