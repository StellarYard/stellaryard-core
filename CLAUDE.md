# CLAUDE.md — stellaryard-core

This project's agent instructions live in [`AGENTS.md`](./AGENTS.md). Read that file in full before making changes — it is the source of truth for both Claude Code and any other coding agent working in this repo, and this file intentionally does not duplicate it.

Claude-specific notes:

- When asked to implement a Wave issue, check `ROADMAP.md` and `ARCHITECTURE_ESSENTIALS.md` before writing code, in that order.
- Update `ROADMAP.md` as part of the same turn/commit as your functional change, not as an afterthought — see `AGENTS.md` rule 5. If you're about to end a session or hand off without having touched `ROADMAP.md`, say so explicitly rather than silently skipping it.
- If a task appears to require violating any rule in `AGENTS.md` (especially the `Signer` interface boundary), stop and surface that conflict to the user instead of proceeding.
- **Docker operations have no retry logic.** When implementing container endpoints, always add backoff. Don't just surface raw Docker SDK errors — classify them (transient vs permanent) so consumers can distinguish retryable from hard failures.
- **SQLite writes should be atomic.** Wrap account/contract writes in transactions. A crash mid-write with no transaction means orphaned partial state that's unrecoverable.
- **The OpenAPI spec will drift from code.** Always verify your implementation matches `openapi.yaml` bidirectionally — the spec says what you return, and your code returns what the spec says. Until a conformance test exists, this is manual.
