# ORQEN

[![ci](https://github.com/soham1105-tech/orqen/actions/workflows/ci.yml/badge.svg)](https://github.com/soham1105-tech/orqen/actions/workflows/ci.yml)

**Build. Test. Heal. Ship.**
*Autonomous infrastructure control plane & self-healing DevOps agent.*

ORQEN takes any repository — a local path or a Git URL — and turns it into a
running, health-verified, containerized development stack. It detects services
across Node/TypeScript (including monorepos), Python, Go, and Rust, infers the
backing infrastructure it needs, generates portable multi-stage Dockerfiles and
a `docker-compose.yml`, allocates conflict-free host ports, starts the stack,
runs each service's test suite inside its own container, verifies everything
with real HTTP probes, and repairs build or runtime failures in a closed loop.

```
$ orqen init https://github.com/acme/app
[1/8] Resolving repository source
[2/8] Scanning services
[3/8] Discovering environment variables
[4/8] Inferring infrastructure
[5/8] Generating Dockerfiles and compose
[6/8] Writing environment and state
[7/8] Building and verifying the stack
[8/8] Stack healthy in 41.2s
```

ORQEN also ships a conversational agent:

```
$ orqen chat
› why is the api unhealthy?

⚡ inspect_logs_and_health (service=api)
✓ inspect_logs_and_health — 2 containers, 1 unhealthy
ORQEN  The api container is healthy but your web service cannot reach it:
       API_URL still points at localhost. I can reconcile the service URLs
       and restart the stack.
```

---

## Highlights

- **Multi-language detection** — Node/Next/Nuxt/Remix/SvelteKit/Nest/Vite/CRA/Astro/Express/Fastify/Koa, pnpm/npm/yarn/bun workspaces, Python (PEP 621, Poetry, uv, pipenv, requirements), Go modules, and Rust crates.
- **Automated in-container testing** — `orqen test` detects each service's test runner and runs it *inside the running container*, parsing pytest/unittest, `go test`, cargo, and Jest/Vitest output into honest pass/fail/skip counts.
- **Infrastructure inference** — Postgres, Redis, MySQL, Mongo, Kafka (KRaft), MinIO, RabbitMQ, Elasticsearch from dependencies and env-var values, with **managed-cloud suppression** (Neon, Supabase, RDS, Upstash, Atlas, Confluent Cloud, and more).
- **Dynamic port allocation for everything** — apps *and* infra containers get conflict-free host ports (e.g. `5439:5432`) while in-network URLs stay on the standard container port (`postgres:5432`).
- **Dockerfile-only repos & static sites** — a root `Dockerfile` or `index.html` is enough; no manifest required.
- **Nested-app safe builds** — per-service build contexts plus managed `.dockerignore` exclusions stop a nested app from poisoning a root build.
- **Modern Node support** — Node 22 default, auto-upgrades to Node 24 when `node:sqlite` is detected.
- **Honest failure handling** — real error extraction, persisted failure records, manifest rollback, and actionable next steps.
- **Agent console & interactive setup** — modern centered chat TUI, a `diagnose` tool, `orqen setup` wizard, and interactive missing-env prompts.
- **Web control plane** — `orqen serve` opens a local, chat-first agent console with project-aware conversations, slash commands, confirmation modals, and nerd stats.
- **Repo intelligence** — reads README/docs/Makefile and existing Dockerfiles, reuses them when viable, and auto-falls back to generated manifests if the existing build fails.
- **Nerd stats** — full tool timeline with args, outputs, durations, bytes, intent confidence, and raw event JSON.
- **Hermetic generated images** — multi-stage Dockerfiles with non-root-friendly runtimes and curl for health probes. **No unconditional PyTorch index**: `download.pytorch.org` is added only when `torch` is actually a dependency.
- **Real health verification** — container state *and* HTTP readiness are checked; failures are never discarded or reported as success.
- **Closed-loop healer** — deterministic fast rules first, then an invariant-guarded LLM repair that can only touch generated `Dockerfile.*` / `docker-compose.yml`, with a stale-context guard and SHA-256 provenance in `repairs.json`.
- **Jailed agent** — every agent file read/write goes through `safefs`, a symlink-resolved, repo-root-jailed boundary. Path traversal (`..`, absolute paths, symlink escapes) is rejected.
- **Three-tab dashboard** — live logs, env overrides, and container controls/stats in a Bubble Tea TUI.
- **Ephemeral public previews** — `orqen tunnel` uses cloudflared (`trycloudflare.com`) with a localhost.run SSH fallback.
- **Provider-agnostic LLM** — zero SDKs, plain `net/http`: Gemini, OpenAI, Anthropic, Groq, and Ollama, with native tool calling.

---

## Install

```bash
git clone https://github.com/soham1105-tech/orqen.git
cd orqen
make install          # installs to $(go env GOPATH)/bin
# or
make build            # produces ./orqen
# or
./install.sh          # builds and installs to ~/.local/bin
```

Requires Go 1.27+ and Docker with Compose v2.

---

## Configuration

ORQEN reads credentials from, in order of precedence:

1. OS environment variables
2. dotenv files (`~/.orqen/.env`, `$ORQEN_ENV_FILE`, `./.env`)
3. `~/.orqen/config.json`

Supported variables:

| Variable | Purpose |
|---|---|
| `ORQEN_LLM_PROVIDER` | `gemini`, `openai`, `anthropic`, `groq`, `ollama` |
| `ORQEN_LLM_MODEL` | model id (defaults per provider) |
| `ORQEN_LLM_BASE_URL` | override endpoint (Ollama / self-hosted) |
| `GEMINI_API_KEY` / `OPENAI_API_KEY` / `ANTHROPIC_API_KEY` / `GROQ_API_KEY` | provider keys |
| `ORQEN_LLM_API_KEY` / `LLM_API_KEY` | generic key fallback |
| `GITHUB_PAT` / `GH_TOKEN` | private repo clones |
| `ORQEN_HOME` | state root (default `~/.orqen`) |

Copy `.env.example` to `.env` and fill in a key. **`.env` is gitignored and must
never be committed.**

Run `orqen doctor` to verify git, Docker, Compose v2, LLM credentials, and
tunnel tooling.

---

## Commands

| Command | Description |
|---|---|
| `orqen init [path-or-url]` | Full pipeline: clone → scan → infer → generate → build → heal → verify |
| `orqen test [project] [--service <name>]` | Detect and run each service's tests inside its running container |
| `orqen chat [project]` | Interactive, intent-aware agent REPL |
| `orqen serve` | Local web control plane (chat-first agent console + nerd stats) |
| `orqen ask [project] <request>` | One-shot intent-aware request (prints the parsed intent first) |
| `orqen up [project]` | Start a stack and wait for real health (`--build`, `--recreate`) |
| `orqen down [project]` | Stop a stack (`--volumes` removes volumes) |
| `orqen restart [project]` | Recreate and health-verify a stack |
| `orqen open [project]` | Open the live URL (`--print`, `--service`) |
| `orqen nuke <project>` | Destroy containers, volumes, clone, and state (aliases: `incinerate`, `remove`) |
| `orqen update [project]` | Re-scan + regenerate while preserving host ports (`--no-run`) |
| `orqen heal [project]` | Deterministic + LLM closed-loop repair (`--tries`) |
| `orqen explain [project]` | Dependency DAG, port map, repair provenance, required config |
| `orqen dash [project]` | Three-tab live dashboard |
| `orqen tunnel [project]` | Ephemeral public preview URL (`--port`) |
| `orqen status [project]` | Container/health/port table; no arg lists projects (`list`) |
| `orqen stats [project]` | Live CPU, memory, network, and PID stats |
| `orqen logs [project] [service]` | Tail or stream logs (`-f`, `-n`) |
| `orqen env [project] [set KEY=VALUE \| unset KEY \| config]` | List/required vars or edit overrides (auto-applies); `config` opens the interactive editor |
| `orqen setup` | Interactive provider/model/key/GitHub wizard with live verification |
| `orqen config [show\|path\|edit\|set <key> <value>]` | Interactive config panel (or show/set for scripts) |
| `orqen doctor` | Environment preflight (non-zero exit on critical failures) |
| `orqen completion [bash\|zsh\|fish]` | Shell completion scripts |

Global flags: `--verbose/-v`, `--quiet/-q`, `--no-color`, `--yes`.

In `orqen chat`, mutating tools prompt for confirmation unless `--yes` is set.
Slash commands: `/status`, `/explain`, `/heal`, `/tunnel`, `/nuke`,
`/nuke-confirm`, `/clear`, `/help`.

See [docs/DEMO.md](docs/DEMO.md) for a complete from-scratch demo runbook,
including a live `orqen test` walkthrough.

---

## Testing (`orqen test`)

`orqen test` runs each service's test suite **inside that service's running
container** — the same filesystem, environment, and network the container runs
with in production. Nothing is executed on the host, and no shell is involved:
the command is tokenized and passed directly to `docker compose exec -T <svc>`,
so quoted arguments can never become shell metacharacters.

### What is detected

| Language | Detection | Command |
|---|---|---|
| Node/TS | `test` script in `package.json` | `npm test` / `pnpm test` / `yarn test` / `bun test` (per detected package manager) |
| Node/TS | `jest` or `vitest` dependency, no test script | `npx jest` / `npx vitest run` |
| Python | `pytest` in dependencies or a `tests/` directory | `python -m pytest -q` |
| Python | neither (stdlib fallback) | `python -m unittest discover -q` |
| Go | any Go service | `go test ./...` |
| Rust | any Rust service | `cargo test` |
| other | no recognizable runner | reported as `no tests detected` (not an error) |

Detection happens at scan time and is persisted in the project lock file;
`orqen test` re-detects on the fly for older locks.

### What the output means

```
$ orqen test
... banner ...

RUNNING SERVICE TESTS

api            2 passed, 1 failed (1.4s)
web            3 passed (2.1s)
worker         no tests detected

--------------------------------
TOTAL: 5 passed, 1 failed
RESULT: FAILED
```

- Counts are parsed best-effort from the runner's summary output (pytest,
  unittest, `go test`, cargo, Jest/Vitest). Unknown formats never crash and
  never count as success: a non-zero exit always fails the run.
- Failing services show the last 30 lines of raw output.
- Services are visited in dependency (topological) order.
- Exit codes: `0` all passed, `0` no tests detected, `1` at least one failure,
  non-zero for operational errors (Docker missing, project not found, ...).
- `--service <name>` restricts the run to one service; the name must exist in
  the project or the command errors with the available services listed.

---

## Architecture

```
cmd/                 Cobra commands and the chat/dashboard entrypoints
internal/
  domain/            Pure types, interfaces, typed error taxonomy
  safefs/            Symlink-resolved jail, atomic writes, guarded deletes
  config/            config.json + dotenv + OS env resolution
  git/               URL/branch/subdir parsing, shallow clone, PAT via env
  scanner/           Node (incl. monorepo), Python, Go, Rust + test-cmd detection
  envvar/            Source + dotenv + Pydantic discovery, crypto/rand secrets
  infra/             Backing-service inference + cloud suppression + ports
  portprobe/         Host socket prober + concurrency-safe allocator
  reconciler/        App port assignment + start-command/URL wiring
  generator/         Multi-stage Dockerfiles + portable compose + dockerignore
  graph/             Deterministic dependency DAG + topological start order
  docker/            Single os/exec boundary for compose (incl. exec)
  testrun/           In-container test execution + runner-output parsing
  httpcheck/         Real HTTP readiness probes
  llm/               Zero-SDK multi-provider client + native tool calling
  healer/            Failure analysis, deterministic rules, invariants, provenance
  agent/             Typed tool registry, state machine, guardrails
  tunnel/            cloudflared / localhost.run manager
  dashboard/         Bubble Tea TUI
  state/             Atomic per-project locks, env, overrides, provenance
```

Design rules:

- **One boundary per external system**: `docker/` for containers, `git/` for
  clones, `llm/` for providers, `safefs/` for the filesystem. `orqen test`
  goes through the same `docker/` boundary (`Compose.Exec`), never `os/exec`
  directly.
- **Determinism**: map keys are sorted before rendering or validation; compose
  and `.env` output are stable and diff-friendly.
- **Honest results**: health failures surface as errors; the healer never
  prints a success banner for a broken stack; `orqen test` never reports a
  suite as passed on a non-zero exit it cannot parse.

### Agent-first architecture

ORQEN is not a Docker wrapper with a chatbot bolted on. The agent harness is
the core:

- **Typed tool registry** — `onboard_repository`, `scan_repository`,
  `infer_infrastructure`, `synthesize_manifests`, `orchestrate_stack`,
  `inspect_logs_and_health`, `diagnose_failure`, `heal_service`,
  `read_repo_file`, `explain_topology`, `start_public_tunnel`.
- **Agent-native onboarding** — from a clean slate (even with zero projects),
  `orqen ask --yes "onboard <url>"` or `/onboard <url>` in chat clones, scans,
  generates, runs, heals, verifies, and returns live URLs.
- **Intent engine** — every request is parsed into a structured intent
  (`status`, `diagnose`, `heal`, `env_set`, ...) with project/service
  validation. High-confidence intents execute tools directly; ambiguous ones
  ask one clarifying question; open-ended asks fall back to the LLM tool loop.
- **Grounding** — the agent reads real logs, health probes, topology, env,
  README/docs, existing Dockerfiles/compose, and jailed repo files. It restates
  the intent (`🎯`) before acting, and prefers reusing viable Docker assets.
- **Guardrails** — mutating tools require confirmation; destructive tools are
  explicitly marked; file access is jailed; patches are invariant-checked.
- **Closed loop** — diagnose → deterministic fix → LLM repair → rebuild →
  verify → provenance (and rollback if repair fails).

`orqen chat` is a modern, centered, bottom-input agent console with streaming
tool pills, spinner, confirmation boxes, markdown responses, and slash commands.
`orqen serve` adds the same agent in a local web control plane: chat-first,
project sidebar, persisted conversations, slash commands for every operation,
confirmation modals, and a nerd-stats drawer with full tool telemetry.

`orqen env` and `orqen setup` are interactive: missing variables are prompted
one-by-one (secrets masked), and the setup wizard verifies the LLM live.

### Intent-aware agent

`orqen chat` and `orqen ask` do not blindly hand text to a model. Every request
is first parsed into a structured intent (`status`, `health`, `explain`, `logs`,
`env_set`, `heal`, `nuke`, `tunnel`, ...). High-confidence intents execute
directly (fast and deterministic); ambiguous or unknown requests fall back to an
LLM classifier that must return the same schema. The agent:

- restates the intent (`🎯 ...`) before acting,
- asks one clarifying question when the project/service is ambiguous,
- validates targets against real projects/services (never invents them),
- requires confirmation for mutating/destructive actions,
- executes read-only intents without any LLM call.

### Proving the AI healer is real

`make heal-proof` runs two adversarial fixtures with Docker and a live LLM:

1. **Runtime failure** — a FastAPI app defines `application` instead of `app`,
   so the generated `uvicorn main:app` crashes. The agent reads the logs and
   the entry file, patches the Dockerfile to `main:application`, rebuilds, and
   verifies HTTP health.
2. **Build failure** — a `git+https://...` requirement needs `git`, which the
   base image omits. The agent adds `git` to the builder's package install,
   rebuilds, and verifies health.

Each repair is recorded in `repairs.json` with the LLM's diagnosis and SHA-256
before/after hashes. The script is `scripts/live-heal-test.sh`.

`orqen ask` is ideal for demos and scripts:

```bash
orqen ask "what is running?"
orqen ask "explain the topology"
orqen ask "show environment"
orqen ask "set JWT_SECRET=dev123 on certify-v2"
orqen --yes ask "nuke certify-v2"
```

### Security model

- The repository clone is the only directory the agent can read, and it is
  resolved through `safefs.Jail` (symlink-aware, `..`-proof, 1 MiB read cap).
- The healer can only write `Dockerfile.*`, `docker-compose.yml`, and
  `.dockerignore` inside the generated workspace.
- LLM patches are validated against invariants: no disabling type checks, no
  removing healthchecks, no `CMD sleep`, no removing install steps, no absolute
  compose paths, and no unbounded rewrites.
- Every applied patch is recorded in `repairs.json` with before/after SHA-256.
- Git tokens are passed via `GIT_CONFIG_*` environment variables, never argv,
  and are scrubbed from error output.
- `orqen test` never spawns a shell: the test command is tokenized and handed
  to `docker compose exec -T`, so injected quoting cannot become a shell.

---

## Development

```bash
make fmt
make check      # gofmt check + vet + tests + race
make e2e        # full no-Docker init + update pipeline tests
make heal-proof # live LLM repair of a runtime and a build failure
bash scripts/repo-matrix.sh  # broad real-repo generation matrix
make build
```

CI runs `gofmt`, `go vet`, `go test -race`, and a build on every push. Tagged
releases (`v*`) are built with GoReleaser for linux/darwin/windows on amd64 and
arm64 (see `.github/workflows/` and `.goreleaser.yaml`).

State lives under `~/.orqen/`:

```
~/.orqen/
├── config.json          # 0600
├── repos/<id>/          # shallow clones
└── projects/<id>/
    ├── orqen.lock.json  # 0600
    ├── .env             # 0600
    ├── env-overrides.json
    ├── repairs.json
    ├── docker-compose.yml
    └── Dockerfile.<svc>
```

## License

MIT
