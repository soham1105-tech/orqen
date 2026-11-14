# Changelog

All notable changes to ORQEN are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial ORQEN branding and identity
- Professional documentation suite
- GitHub issue and PR templates
- Contributing guidelines

## [v0.5.0] - 2026-11-14

### Added
- **Automated in-container testing** (`orqen test`)
  - Detects test commands for every service (npm/pnpm/yarn/bun test script, jest/vitest dependency, pytest/unittest for Python, `go test ./...`, `cargo test`)
  - Runs tests inside the service's running container via `docker compose exec -T` (no shell)
  - Reports per-service passed/failed/skipped counts parsed from runner output
  - Unknown output formats never crash and never count as success; failures show the last 30 lines of raw output
  - Exit code 1 when any suite fails, 0 when everything passes or nothing is detected
  - `--service <name>` restricts the run to one service

### Changed
- Complete rebrand from "Spool" to **ORQEN** ("Build. Test. Heal. Ship.")
- Go module path updated to `github.com/soham1105-tech/orqen`
- Repository URL updated to `https://github.com/soham1105-tech/orqen`
- State directory changed from `~/.spool` to `~/.orqen`
- Lock file renamed from `spool.lock.json` to `orqen.lock.json`
- Environment variable prefix changed from `SPOOL_*` to `ORQEN_*`
- Generated file markers changed from `# spool:managed-*` to `# orqen:managed-*`
- Ignore directory entries changed from `.spool` to `.orqen`
- Greeting triggers updated from "hi spool" to "hi orqen"

### Fixed
- POSIX-style 0600 permission assertions now skip on Windows instead of failing
- Cross-drive build-context fallback keeps compose paths valid when the state home and the repository live on different drives
- pytest summary parsing only trusts real pytest summary lines, so tracebacks containing count-like text can no longer inflate pass/fail counts
- Chat `onboard` recognizer accepts Windows drive-letter paths
- Jest "Test Suites" inflation in test output parsing
- cargo test regex missing multiline flag
- pytest parser `TrimSuffix(m[2],"s")` bug

## [v0.4.1] - 2026-10-15

### Added
- **Honest status & live progress**
  - `orqen init/up/restart/heal/update` now stream live build progress (BuildKit steps) and per-service health-check transitions
  - Final status waits for health checks to settle before rendering
  - Live endpoints are only printed for containers that are actually running — a stopped stack no longer advertises URLs (`orqen status` included)
  - `/api/health` reports the build-injected version instead of a hardcoded one

### Changed
- **Web console UX**
  - Greetings and small talk are answered deterministically — no LLM, no tools
  - Dynamic streaming states ("Thinking…", "Checking logs and health…", …) while the agent works, and a typewriter reveal for finished replies
  - Intent tags hidden; nerd stats moved to hover-only; tool pills no longer dump raw JSON args
  - Project selector addresses projects by unique id, so two projects sharing a name (same repo onboarded twice) stay separately selectable

### Added
- **Interactive environment editor** (`orqen env <project> config`)
  - Interactive panel to edit, add, unset, and reveal project variables (secrets masked)
  - Persists as you type and applies to running stacks on exit

## [v0.4.0] - 2026-09-20

### Added
- **Web control plane** (`orqen serve`)
  - Local chat-first web UI with project sidebar
  - Persisted chats in `~/.orqen/chats`
  - Streaming SSE agent events
  - Slash commands for all operations
  - Confirmation modals and a nerd-stats drawer

- **Nerd stats**
  - Intent, tool calls, args, outputs, durations, bytes, raw events

- **Agent-first repo intelligence**
  - Scanner collects README, docs, Makefile, scripts, existing Dockerfiles and compose files
  - Exposes them through `inspect_repository`
  - Existing Dockerfiles are reused by default; if the reused build fails, ORQEN regenerates managed Dockerfiles and retries automatically
  - Agent system prompt now explicitly prefers inspecting/reusing repo assets

- **Interactive config** (`orqen config` / `orqen config set`)
  - Proper terminal config panel: provider cycling, editable model/base URL, masked API key and GitHub PAT with reveal, and save
  - `config set KEY VALUE` still works for scripts

### Fixed
- **Bulletproofing**
  - Strict compose validation, rollback of failed patches, source-error detection
  - Real BuildKit error extraction, and fast crash-loop failure
  - Node 22 default / Node 24 for `node:sqlite`
  - Dockerfile-only and static repos
  - Nested-app safe build contexts

## [v0.3.0] - 2026-08-10

### Added
- **Agent-first architecture**
  - Agent-native onboarding: `onboard_repository` tool + `onboard` intent
  - From a clean slate, the agent clones/scans/generates/runs/heals/verifies a repo and returns live URLs
  - Chat auto-switches context to the new project
  - New `diagnose` intent + `diagnose_failure` tool: explains what failed, why, with evidence, container state, and next steps
  - Persisted `last-failure.log` so diagnosis works after failed builds
  - LLM healer now rolls back experimental patches when it cannot fix a repo
  - Strict compose validation (YAML, duplicate keys, service removal) with validation feedback retried by the model
  - Real error extraction from BuildKit output, including TypeScript/Next errors

### Changed
- **Compatibility**
  - Node 22 default; auto Node 24 when `node:sqlite` is detected
  - Dockerfile-only repos are supported (existing Dockerfile reused)
  - Static root sites are detected alongside nested backends
  - Nested independent Node apps build from their own contexts; root builds exclude them via a managed `.dockerignore` block
  - Cookiecutter/template trees are skipped instead of misdetected
  - Missing env vars are reported after init/update with an exact set command

- **Experience**
  - Modern chat TUI: centered layout, bottom-pinned rounded input, spinner, intent badges, streaming tool pills, confirmation boxes
  - Interactive `orqen setup` wizard with live LLM verification
  - Interactive `orqen env` prompts for missing variables (secrets masked)

### Added
- **Tests**
  - Added live heal proof (runtime repair, build repair, honest failure rollback)
  - Added 8-repo and 23-repo real-repo matrices and public-internet validation
  - Added scanner/generator/healer/chat tests for all new paths

## [v0.2.0] - 2026-06-15

### Added
- **Intent-aware agent**: deterministic parser + validated LLM classifier, clarification questions, direct read-only execution, and `orqen ask`
- **New commands**: `ask`, `env`, `config`, `restart`, `open`, `stats`, `update`, `list` alias, shell completion
- **Zero-config env import**: repository `.env` files are imported; secrets are masked in listings and written 0600
- **Broader detection**: static sites, hapi/restify/pola/sails/loopback-style Node servers, source-declared ports, `.python-version`, `requirements` include cycles, and dependency-only `pyproject.toml`
- **Hermetic builds**: public `NEXT_PUBLIC_*`/`VITE_*` build args, no unconditional PyTorch index, relative compose contexts, app healthchecks
- **Safety**: symlink-resolved repo jail, 0600 secrets, port probing on all interfaces, destructive-action confirmation, orphan cleanup
- **Release engineering**: GitHub Actions CI, GoReleaser, `install.sh`, end-to-end no-Docker pipeline tests, race-tested suite

### Fixed
- **Healer hardening** (live-verified)
  - LLM plans that target generic `Dockerfile` are normalized to the failing service's actual filename
  - Validation failures are fed back to the model and retried instead of aborting the heal loop
  - Runtime crash loops (`restarting` with no exit code) fail fast after a grace period instead of waiting the full health timeout
  - Entry files (`main.py`, `app.py`, `server.js`, ...) are included in the LLM repair context
  - `make heal-proof` proves a live runtime repair and a live build repair
- Port allocator now detects Docker's `0.0.0.0` bindings (3000→3001 remap)
- `orqen down` no longer panics from a `-v` flag collision
- Non-interactive destructive intents require `--yes`
- Env changes only recreate stacks that are already running

---

[Unreleased]: https://github.com/soham1105-tech/orqen/compare/v0.5.0...HEAD
[v0.5.0]: https://github.com/soham1105-tech/orqen/releases/tag/v0.5.0
[v0.4.1]: https://github.com/soham1105-tech/orqen/releases/tag/v0.4.1
[v0.4.0]: https://github.com/soham1105-tech/orqen/releases/tag/v0.4.0
[v0.3.0]: https://github.com/soham1105-tech/orqen/releases/tag/v0.3.0
[v0.2.0]: https://github.com/soham1105-tech/orqen/releases/tag/v0.2.0