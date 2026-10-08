#!/usr/bin/env bash
# Creates the planned Phase 2 (accounts) and Phase 4 (contracts) backlog for
# stellaryard-core in one run.
#
#   ./scripts/create-issues.sh [owner/repo]
#
# Titles are commit-style, bodies carry Summary / Acceptance Criteria /
# Tech Stack. Re-running skips issues whose title already exists.
set -euo pipefail

REPO="${1:-StellarYard/stellaryard-core}"

create_issue() {
  local title="$1" labels="$2" body="$3"

  if gh issue list --repo "$REPO" --state all --limit 300 \
      --json title --jq '.[].title' | grep -Fxq "$title"; then
    echo "skip (exists): $title"
    return 0
  fi

  gh issue create --repo "$REPO" --title "$title" --label "$labels" --body "$body" >/dev/null
  echo "created: $title"
}

# --- Phase 2: accounts -------------------------------------------------------

create_issue \
  "[PHASE 2] feat: generate real Stellar keypairs instead of placeholder keys" \
  "feature,phase-2-core,critical-path,cross-repo" \
  "## Repo Context

Read \`AGENTS.md\`, \`ARCHITECTURE_ESSENTIALS.md\`, \`ROADMAP.md\` first.

\`CreateAccount\` writes a placeholder: a 56-character string shaped like a
Stellar public/secret key but not derived from any seed. The record is
usable for listing and lookup, but nothing can sign with it.

## Problem

Every downstream feature that needs to sign — funding, deploying,
invoking — is blocked until the stored key is real. This is the root
dependency for the rest of Phase 2.

## Scope

**In scope**
- Generate an ed25519 keypair per the Stellar derivation used by the network
- Persist the secret in SQLite as today (local/testnet only, documented)
- Keep the \`Signer\` interface as the only path to the secret — no direct reads outside \`Sign()\`
- Return the real \`publicKey\` in the API response

**Out of scope**
- Mainnet support, hardware wallets, external signers
- Encryption at rest (tracked separately)

## What "Done" Looks Like

1. A created account's \`publicKey\` is a valid Stellar address
2. The account can sign a payload that verifies against \`publicKey\`
3. Existing tests still pass; new tests cover generation and signing
4. \`ROADMAP.md\` updated

## Acceptance Criteria

- [ ] Keypairs are real and deterministic from a stored seed
- [ ] \`Sign()\` produces a signature that verifies against the stored public key
- [ ] The secret never appears in an API response
- [ ] Tests cover generation, signing, and verification
- [ ] \`ROADMAP.md\` updated
- [ ] No \`AGENTS.md\` rules violated

## Tech Stack

Go 1.23, \`github.com/stellar/go/txnbuild\` or \`keypair\`, SQLite"

create_issue \
  "[PHASE 2] feat: fund accounts through Friendbot with retry and clear errors" \
  "feature,phase-2-core,ready" \
  "## Repo Context

Read \`AGENTS.md\`, \`ARCHITECTURE_ESSENTIALS.md\`, \`ROADMAP.md\` first.

## Problem

Account creation currently stops at writing a row. The PRD calls out
Friendbot unreliability as the most common user-facing failure: it rate
limits, is slow, and returns opaque errors. There is no retry and no
distinction between "Friendbot is down" and "this request is malformed".

## Acceptance Criteria

- [ ] Creating an account funds it via Friendbot on testnet and via genesis on local
- [ ] Retries with backoff on transient failures (network, 5xx, 429)
- [ ] Does not retry on permanent failures (400) — surfaces the reason
- [ ] Failure response distinguishes unreachable Friendbot from rejected request
- [ ] Partial state is rolled back: a failed fund does not leave an unusable row
- [ ] Tests cover success, retry-then-success, and permanent failure
- [ ] \`ROADMAP.md\` updated

## Tech Stack

Go 1.23, Friendbot HTTP API, SQLite, context timeouts"

create_issue \
  "[PHASE 2] feat: unique constraint on account label" \
  "feature,phase-2-core,ready,good-first-issue" \
  "## Repo Context

Read \`AGENTS.md\`, \`ARCHITECTURE_ESSENTIALS.md\`, \`ROADMAP.md\` first.

## Problem

The PRD lists this under edge cases: two accounts created with the same
label are both accepted, so \`accounts list\` shows indistinguishable rows
and any label-keyed lookup becomes ambiguous.

## Acceptance Criteria

- [ ] Migration adds a uniqueness constraint on \`accounts.label\`
- [ ] Existing duplicate rows are handled (rename or reject with a clear migration note)
- [ ] \`POST /accounts\` with a duplicate label returns 409 with a specific error code
- [ ] Test covers duplicate rejection and the migration path
- [ ] \`ROADMAP.md\` updated

## Tech Stack

Go 1.23, SQLite migrations, chi"

# --- Phase 4: contracts ------------------------------------------------------

create_issue \
  "[PHASE 4] feat: implement POST /contracts/deploy for real (replace the 501 stub)" \
  "feature,phase-4-core,critical-path,cross-repo" \
  "## Repo Context

Read \`AGENTS.md\`, \`ARCHITECTURE_ESSENTIALS.md\`, \`ROADMAP.md\` first.

\`DeployContract\` currently returns 501 with a placeholder body. The README
documents this openly; the dashboard's Contracts page and the CLI's
\`contracts deploy\` both call it.

## Problem

This is the endpoint that makes StellarYard useful for its stated purpose —
deploying a Soroban contract to the local network without leaving the tool.

## Scope

**In scope**
- Accept a WASM upload (multipart or raw body), validate the magic bytes and size limit
- Deploy via Soroban RPC using the \`Signer\` interface — never bypass it
- Persist a \`contract_deployments\` row (table already exists)
- Return the contract ID and the transaction hash

**Out of scope**
- Contract verification/explorer links
- Deployment to mainnet

## Acceptance Criteria

- [ ] Valid WASM deploys and returns a contract ID
- [ ] Invalid or oversized input returns 4xx with a specific error code, not 500
- [ ] The secret key never crosses the \`Signer\` boundary
- [ ] A deployment row is written and readable via \`ListDeployments\`
- [ ] Tests cover success, invalid WASM, and rejected size
- [ ] \`ROADMAP.md\` updated

## Tech Stack

Go 1.23, Soroban RPC, stellar-xdr, SQLite"

create_issue \
  "[PHASE 4] feat: implement POST /contracts/{id}/invoke with typed argument decoding" \
  "feature,phase-4-core,critical-path" \
  "## Repo Context

Read \`AGENTS.md\`, \`ARCHITECTURE_ESSENTIALS.md\`, \`ROADMAP.md\` first.

\`InvokeContract\` returns 501. The CLI already sends \`contractId\`, \`method\`,
and a string slice of args; the dashboard sends a typed form.

## Problem

Invocation is the second half of the contract loop. Without it a developer
can deploy but cannot observe behavior, which makes the tool useless for the
actual development cycle.

## Acceptance Criteria

- [ ] Forwards to Soroban RPC and returns the decoded result
- [ ] Argument decoding handles string, i128, u32, bool, address, and BytesN — rejects anything else with a clear error
- [ ] Simulation failures surface the RPC's diagnostic, not a generic 500
- [ ] Events emitted by the call are returned with the result
- [ ] Tests cover each supported type and one rejection
- [ ] \`ROADMAP.md\` updated

## Tech Stack

Go 1.23, Soroban RPC, stellar-xdr, JSON"

create_issue \
  "[PHASE 4] test: contract deploy and invoke integration tests against a local network" \
  "testing,phase-4-core" \
  "## Repo Context

Read \`AGENTS.md\`, \`ARCHITECTURE_ESSENTIALS.md\`, \`ROADMAP.md\` first.

## Problem

Unit tests can cover argument parsing and error mapping, but the deploy and
invoke paths are integration-shaped: they need a running Soroban RPC. There
is currently no test that exercises the real flow end to end.

## Acceptance Criteria

- [ ] Integration suite starts (or connects to) a local network and skips cleanly when unavailable
- [ ] Deploys a known-good WASM and asserts the returned contract ID
- [ ] Invokes it and asserts the decoded result
- [ ] CI runs the unit suite always; integration runs when the environment provides a network
- [ ] \`ROADMAP.md\` updated

## Tech Stack

Go 1.23, Soroban RPC, Docker, build tags for integration suites"

echo
echo "Done. Open issues in $REPO:"
gh issue list --repo "$REPO" --state open --limit 100 --json number,title \
  --jq '.[] | "  #\(.number) \(.title)"'
