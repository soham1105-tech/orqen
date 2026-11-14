# ORQEN Architecture

This document describes the high-level architecture of ORQEN, its core components, data flow, and design principles.

---

## Overview

ORQEN is an **autonomous infrastructure control plane** that transforms any repository into a running, health-verified, containerized development stack. It combines:

1. **Declarative infrastructure generation** — from source code to Dockerfiles + docker-compose
2. **Automated in-container testing** — runs each service's tests inside its own container
3. **Intent-aware DevOps agent** — conversational interface with real tools and guardrails
4. **Closed-loop healing** — deterministic rules + LLM repair with provenance

---

## System Context

```
┌─────────────────────────────────────────────────────────────────────┐
│                         ORQEN CONTROL PLANE                         │
├─────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌───────────┐  │
│  │   CLI       │  │   Web UI    │  │   Agent     │  │   API     │  │
│  │   (Cobra)   │  │   (SSE+WS)  │  │   (Tools)   │  │   (REST)  │  │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └─────┬─────┘  │
│         │                │                │                │        │
│         └────────────────┼────────────────┼────────────────┘        │
│                          ▼                ▼                         │
│              ┌─────────────────────────────────────┐               │
│              │         ORCHESTRATION CORE          │               │
│              │  ┌────────┐ ┌────────┐ ┌────────┐  │               │
│              │  │ Scanner│ │ Generator│ │ Healer │  │               │
│              │  └────────┘ └────────┘ └────────┘  │               │
│              └──────────────┬─────────────────────┘               │
│                             ▼                                     │
│              ┌─────────────────────────────────────┐               │
│              │         EXTERNAL BOUNDARIES         │               │
│              │  ┌──────┐ ┌─────┐ ┌────┐ ┌───────┐ │               │
│              │  │Docker│ │ Git │ LLM │ │ FS    │ │               │
│              │  └──────┘ └─────┘ └────┘ └───────┘ │               │
│              └─────────────────────────────────────┘               │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Package Architecture

### `cmd/` — Command Line Interface

Cobra-based CLI commands. Thin wrappers that delegate to internal packages.

| Command | Purpose |
|---------|---------|
| `init` | Full onboarding pipeline |
| `test` | In-container test execution |
| `chat` / `ask` | Agent REPL / one-shot |
| `serve` | Web control plane |
| `up`/`down`/`restart` | Stack lifecycle |
| `heal` | Closed-loop repair |
| `explain` | Topology & dependency analysis |
| `dash` | TUI dashboard |
| `env`/`config`/`setup` | Configuration |
| `doctor`/`status`/`logs`/`stats` | Observability |

### `internal/domain` — Pure Types & Interfaces

The **kernel** of ORQEN. No external dependencies.

- **Types**: `Project`, `Service`, `Lock`, `Intent`, `Repair`, `Error`
- **Interfaces**: `ComposeController`, `Scanner`, `LLMClient`, `GitClient`, `FS`
- **Error Taxonomy**: Typed errors with kinds (`Config`, `IO`, `Validation`, `UnsafePath`, `NotConfigured`, `NotFound`, `Conflict`, `External`)

### `internal/safefs` — Filesystem Jail

Symlink-resolved, repo-root-jailed filesystem boundary.

- Resolves symlinks on longest existing prefix
- Rejects `..`, absolute paths, symlink escapes
- 1 MiB read cap per operation
- Atomic writes (`WriteFileAtomic`) with 0600 permissions
- Guarded deletes (only within owned roots)

### `internal/config` — Configuration Resolution

Priority order: OS env → dotenv files → `~/.orqen/config.json`

- `ORQEN_HOME`, `ORQEN_LLM_PROVIDER`, `ORQEN_LLM_MODEL`, `ORQEN_LLM_API_KEY`, `ORQEN_LLM_BASE_URL`, `ORQEN_ENV_FILE`
- Provider key fallbacks: `GEMINI_API_KEY`, `OPENAI_API_KEY`, etc.
- `config.json` persisted at 0600

### `internal/git` — Git Operations

- URL/branch/subdir parsing (`git@`, `https://`, `ssh://`)
- Shallow clone with depth=1
- PAT via `GIT_CONFIG_*` env vars (never argv)
- Submodule support (optional)

### `internal/scanner` — Language/Framework Detection

| Language | Frameworks Detected | Test Command Detection |
|----------|---------------------|------------------------|
| Node/TS | Next, Nuxt, Remix, SvelteKit, Nest, Vite, CRA, Astro, Express, Fastify, Koa | `package.json` test script, jest, vitest |
| Python | PEP 621, Poetry, uv, pipenv, requirements.txt | pytest, unittest |
| Go | Go modules | `go test ./...` |
| Rust | Cargo crates | `cargo test` |

Monorepo support: pnpm, npm, yarn, bun workspaces.

### `internal/envvar` — Environment Variable Discovery

- Source code scanning (`.env`, `dotenv`, `pydantic-settings`)
- Pydantic model field extraction
- Crypto/rand secret generation for missing values
- Masking in output

### `internal/infra` — Infrastructure Inference

- Backing services from dependencies: Postgres, Redis, MySQL, Mongo, Kafka (KRaft), MinIO, RabbitMQ, Elasticsearch
- Cloud suppression: Neon, Supabase, RDS, Upstash, Atlas, Confluent Cloud, etc.
- Dynamic host port allocation (apps + infra)
- Port conflict resolution via `portprobe`

### `internal/portprobe` — Port Availability

- Concurrent socket probing on all interfaces
- Ephemeral port range awareness
- Deterministic allocation (sorted service order)

### `internal/reconciler` — Port Assignment & Wiring

- App port → host port mapping
- Start command / health URL wiring
- Environment variable injection for service discovery

### `internal/generator` — Manifest Generation

**Dockerfiles:**
- Multi-stage builds (builder → runner)
- Non-root user by default
- `curl` for health probes
- Framework-specific optimizations
- PyTorch index only when `torch` is a dependency

**docker-compose.yml:**
- Relative build contexts
- Service dependencies (topological order)
- Health checks (HTTP + container state)
- Managed `.dockerignore` with `# orqen:managed-*` markers

### `internal/graph` — Dependency DAG

- Service + infrastructure nodes
- Topological start order
- Cycle detection
- Deterministic iteration (sorted keys)

### `internal/docker` — Docker Compose Boundary

**Single `os/exec` entry point** for all Compose operations:
- `up`, `down`, `ps`, `logs`, `exec`, `config`, `build`
- Streaming BuildKit output parsing
- Structured error extraction
- `Compose.Exec` for `orqen test` (no shell)

### `internal/testrun` — In-Container Test Execution

- Runner detection (npm/pnpm/yarn/bun, jest/vitest, pytest/unittest, go test, cargo test)
- Tokenized command execution via `docker compose exec -T`
- Output parsing → pass/fail/skip counts
- Exit code semantics: 0=all pass, 0=no tests, 1=failure, >1=operational error

### `internal/httpcheck` — HTTP Readiness

- Real HTTP probes (not just container state)
- Configurable path, method, expected status
- Retry with backoff
- Streaming status updates

### `internal/llm` — Multi-Provider LLM Client

| Provider | Models | Tool Calling |
|----------|--------|--------------|
| Gemini | gemini-3.5-flash-lite, ... | Native |
| OpenAI | gpt-4o-mini, gpt-4o, ... | Native |
| Anthropic | claude-3-5-sonnet, ... | Native |
| Groq | llama-3.3-70b-versatile, ... | Native |
| Ollama | llama3.1, ... | Native |

- Zero SDKs — plain `net/http`
- Native tool calling (no function-calling shim)
- Streaming responses
- Request/response logging for debugging

### `internal/healer` — Closed-Loop Repair

**Two-phase approach:**
1. **Deterministic rules** — fast fixes for common patterns (entry point, missing packages, port conflicts, etc.)
2. **LLM repair** — invariant-guarded patches for complex failures

**Invariants (never violated):**
- No disabling type checks
- No removing healthchecks
- No `CMD sleep`
- No removing install/build steps
- No absolute compose paths
- Bounded scope (only generated `Dockerfile.*`, `docker-compose.yml`, `.dockerignore`)

**Provenance:** Every patch recorded in `repairs.json` with:
- LLM diagnosis
- SHA-256 before/after hashes
- Timestamp, operation type, failure category

### `internal/agent` — Intent-Aware Agent

**Tools:**
- `onboard_repository` — clone, scan, generate, build, heal, verify
- `scan_repository` — detect services, infra, env, test commands
- `infer_infrastructure` — backing services + ports
- `synthesize_manifests` — Dockerfiles + compose
- `orchestrate_stack` — build, start, health verify
- `inspect_logs_and_health` — container state + HTTP probes
- `diagnose_failure` — evidence-based root cause
- `heal_service` — deterministic + LLM repair
- `read_repo_file` — jailed file read
- `explain_topology` — DAG, ports, env, repairs
- `start_public_tunnel` — cloudflared / localhost.run

**Intent Engine:**
- Deterministic parser → structured intent (`status`, `diagnose`, `heal`, `env_set`, `nuke`, `tunnel`, `onboard`, `explain`, `logs`, `open`, `restart`, `up`, `down`, `config`, `chat`)
- Confidence scoring
- Clarification questions for ambiguity
- LLM classifier fallback (same schema)
- Read-only intents execute without LLM

**Guardrails:**
- Mutating tools require confirmation
- Destructive tools explicitly marked
- File access jailed via `safefs`
- Patches invariant-checked

### `internal/tunnel` — Public Preview

- `cloudflared` (trycloudflare.com) — primary
- `ssh -R` (localhost.run) — fallback
- Ephemeral URLs, auto-cleanup

### `internal/dashboard` — Bubble Tea TUI

Three tabs:
1. **Logs** — streaming container logs
2. **Env** — live overrides, secret masking
3. **Containers** — controls (start/stop/restart), stats (CPU/mem/net/PID)

### `internal/state` — Per-Project Atomic State

```
~/.orqen/projects/<id>/
├── orqen.lock.json      # 0600 — project metadata, services, infra, test cmds
├── .env                 # 0600 — environment variables
├── env-overrides.json   # user overrides (applied at runtime)
├── repairs.json         # repair provenance (SHA-256)
├── docker-compose.yml   # generated compose
└── Dockerfile.<svc>     # generated Dockerfiles
```

- Atomic writes via `safefs.WriteFileAtomic`
- Lock file tracks: project ID, repo path, work dir, services, infra, network, ports, test commands, last init timestamp

### `internal/web` — Web Control Plane

- SSE for agent event streaming
- Static assets (HTML/JS/CSS) embedded
- Chat persistence (`~/.orqen/chats/`)
- Project-aware conversations
- Slash commands for all operations
- Nerd stats drawer (tool timeline, args, outputs, durations, bytes, raw events)

---

## Data Flow: `orqen init`

```
User Input (path or URL)
         │
         ▼
┌─────────────────────┐
│   git.Clone()       │ ──► ~/.orqen/repos/<id>/
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│   scanner.Scan()    │ ──► Services[], Infra[], EnvVars[], TestCmds[]
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│   infra.Infer()     │ ──► Infra[] (with cloud suppression + ports)
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│   generator.Gen()   │ ──► Dockerfile.*, docker-compose.yml, .dockerignore
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│   state.SaveLock()  │ ──► orqen.lock.json
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│   docker.Up()       │ ──► Build + Start (BuildKit streaming)
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│   httpcheck.Wait()  │ ──► Healthy / Unhealthy
└─────────┬───────────┘
          │
          ▼ (if unhealthy)
┌─────────────────────┐
│   healer.Heal()     │ ──► Rules → LLM → Rebuild → Verify
└─────────┬───────────┘
          │
          ▼
       Result
```

---

## Data Flow: `orqen test`

```
orqen test [project] [--service <name>]
         │
         ▼
┌─────────────────────┐
│   state.LoadLock()  │ ──► TestCmds per service
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│   For each service  │ (topological order)
│   testrun.Run()     │
│     ├── Tokenize cmd│
│     ├── docker.Exec │ ──► docker compose exec -T <svc> <cmd>
│     ├── Capture out │
│     └── Parse result│ ──► Pass/Fail/Skip counts
└─────────┬───────────┘
          │
          ▼
       Aggregate + Exit Code
```

---

## Design Principles

1. **One Boundary Per External System**
   - `docker/` → containers
   - `git/` → clones
   - `llm/` → providers
   - `safefs/` → filesystem
   - `testrun/` → in-container exec (via `docker/`)

2. **Determinism**
   - Sorted map keys before rendering/validation
   - Stable compose/.env output (diff-friendly)
   - Deterministic port allocation (sorted service order)

3. **Honest Results**
   - Health failures surface as errors
   - Healer never prints success for broken stack
   - `orqen test` never reports pass on non-zero exit it cannot parse

4. **Jailed Agent**
   - All file access through `safefs.Jail`
   - Healer writes only to generated manifests
   - 1 MiB read cap, symlink resolution, `..` rejection

5. **Agent-First**
   - CLI commands are thin wrappers
   - Agent tools are the primary API
   - Intent engine drives all interactions

---

## Security Model

| Layer | Protection |
|-------|------------|
| Filesystem | `safefs.Jail` — symlink-resolved, `..`-proof, 1 MiB cap |
| Credentials | `GIT_CONFIG_*` env, never argv; masked in output; 0600 files |
| Healer Scope | Only `Dockerfile.*`, `docker-compose.yml`, `.dockerignore` |
| Invariants | Type checks, healthchecks, no sleep, no install removal, bounded scope |
| Provenance | SHA-256 before/after in `repairs.json` |
| Test Exec | Tokenized `docker compose exec -T` — no shell |

---

## Extensibility Points

| Extension | Location |
|-----------|----------|
| New language detection | `internal/scanner/` + `internal/scanner/<lang>/` |
| New framework | `internal/generator/frameworks.go` |
| New infra service | `internal/infra/infra.go` |
| New LLM provider | `internal/llm/providers.go` |
| New agent tool | `internal/agent/registry.go` |
| New heal rule | `internal/healer/rules.go` |
| New intent | `internal/agent/intent.go` |

---

## Version Compatibility

| Component | Policy |
|-----------|--------|
| Go module | Semantic versioning (v0.x.x — breaking changes possible) |
| Lock file | Forward compatible; new fields ignored |
| Config | New env vars additive; old names never removed |
| CLI | Commands stable; flags additive |
| Web API | SSE event schema versioned in `event.type` |