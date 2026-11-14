# Changelog

## v0.5.0

### Automated in-container testing
- New `orqen test [project] [--service <name>]`: detects a test command for
  every service (npm/pnpm/yarn/bun test script, jest/vitest dependency,
  pytest/unittest for Python, `go test ./...`, `cargo test`), runs it inside
  the service's running container via `docker compose exec -T` (no shell), and
  reports per-service passed/failed/skipped counts parsed from runner output.
- Unknown output formats never crash and never count as success; failures show
  the last 30 lines of raw output. Exit code 1 when any suite fails, 0 when
  everything passes or nothing is detected.

### Highlights
- Multi-language service detection (Node/TS monorepos, Python, Go, Rust),
  backing-infrastructure inference with managed-cloud suppression, and
  conflict-free dynamic host port allocation.
- Closed-loop healer: deterministic fast rules, then invariant-guarded LLM
  repair of generated manifests, with SHA-256 provenance in `repairs.json`.
- Agent-first architecture: typed tool registry, intent engine, jailed
  filesystem access via `safefs`, and a chat-first web control plane
  (`orqen serve`).
- Provider-agnostic LLM client (Gemini, OpenAI, Anthropic, Groq, Ollama) with
  native tool calling and zero SDKs.
- Ephemeral public previews via `orqen tunnel`, live three-tab dashboard, and
  full HTTP health verification for every stack.

### Fixes
- POSIX-style 0600 permission assertions skip on Windows; the chat `onboard`
  recognizer accepts Windows drive-letter paths.
- Cross-drive build-context fallback keeps compose paths valid when the state
  home and the repository live on different drives.
- pytest summary parsing only trusts real pytest summary lines, so tracebacks
  containing count-like text can no longer inflate pass/fail counts.
