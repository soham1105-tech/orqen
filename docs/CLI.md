# ORQEN CLI Reference

This document provides the complete command-line interface specification for **ORQEN**.

---

## Global Options

These flags apply to all commands:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--verbose` | `-v` | `false` | Enable detailed debug output and execution traces |
| `--quiet` | `-q` | `false` | Suppress all non-error output (ideal for automation/scripts) |
| `--no-color` | | `false` | Disable ANSI terminal color codes and styling |
| `--yes` | | `false` | Automatically confirm interactive and destructive prompts |
| `--version` | | `false` | Display ORQEN version, build time, and commit hash |
| `--help` | `-h` | `false` | Display command help and usage syntax |

---

## Commands Summary

| Command | Syntax | Primary Purpose |
|---------|--------|-----------------|
| [`init`](#orqen-init) | `orqen init [path-or-url]` | Complete repository onboarding pipeline |
| [`test`](#orqen-test) | `orqen test [project]` | Automated in-container test suite execution |
| [`chat`](#orqen-chat) | `orqen chat [project]` | Interactive conversational DevOps REPL |
| [`ask`](#orqen-ask) | `orqen ask [project] <prompt>` | One-shot intent-aware task execution |
| [`serve`](#orqen-serve) | `orqen serve` | Local web control plane (SSE + HTTP UI) |
| [`up`](#orqen-up) | `orqen up [project]` | Start stack containers and verify health |
| [`down`](#orqen-down) | `orqen down [project]` | Stop running stack containers |
| [`restart`](#orqen-restart) | `orqen restart [project]` | Recreate containers and re-verify health |
| [`status`](#orqen-status) | `orqen status [project]` | List projects or inspect project health and ports |
| [`logs`](#orqen-logs) | `orqen logs [project] [service]` | View or stream service container logs |
| [`stats`](#orqen-stats) | `orqen stats [project]` | Real-time container resource utilization |
| [`open`](#orqen-open) | `orqen open [project]` | Open service web endpoint in default browser |
| [`explain`](#orqen-explain) | `orqen explain [project]` | View dependency DAG, topology, and ports |
| [`heal`](#orqen-heal) | `orqen heal [project]` | Run closed-loop deterministic + AI repair |
| [`update`](#orqen-update) | `orqen update [project]` | Re-scan and regenerate while preserving host ports |
| [`env`](#orqen-env) | `orqen env [project] [subcommand]` | Inspect, set, unset, or edit environment variables |
| [`config`](#orqen-config) | `orqen config [subcommand]` | View or modify global ORQEN configuration |
| [`setup`](#orqen-setup) | `orqen setup` | Interactive provider and API credentials setup |
| [`doctor`](#orqen-doctor) | `orqen doctor` | Preflight system and environment verification |
| [`dash`](#orqen-dash) | `orqen dash [project]` | Interactive 3-tab terminal UI dashboard |
| [`tunnel`](#orqen-tunnel) | `orqen tunnel [project]` | Expose project port via ephemeral public URL |
| [`nuke`](#orqen-nuke) | `orqen nuke <project>` | Teardown containers, volumes, repository, and state |
| [`completion`](#orqen-completion) | `orqen completion [shell]` | Generate shell autocompletion script |

---

## Command Details

### `orqen init`

Onboards a local directory or remote Git repository into a fully running, containerized stack.

```bash
orqen init [path-or-url] [flags]
```

**Flags:**
- `--no-run`: Generate manifests and project state without building or starting containers.
- `--no-agent`: Disable autonomous LLM healing if build or runtime failures occur.
- `--max-services <n>`: Limit number of discovered microservices (default: unlimited).

**Examples:**
```bash
# Onboard a remote repository
orqen init https://github.com/fastapi/fastapi

# Onboard the current local directory
orqen init .

# Generate manifests only (dry run for CI/review)
orqen init /path/to/repo --no-run
```

---

### `orqen test`

Detects each service's native test runner and executes tests **inside the running container** via `docker compose exec -T`.

```bash
orqen test [project] [flags]
```

**Flags:**
- `--service <name>`: Restrict test execution to a specific service.

**Detected Runners:**
- **Node / TypeScript**: `npm test`, `pnpm test`, `yarn test`, `bun test`, `npx jest`, `npx vitest run`
- **Python**: `python -m pytest -q`, fallback `python -m unittest discover -q`
- **Go**: `go test ./...`
- **Rust**: `cargo test`

**Exit Codes:**
- `0`: All service test suites passed (or no tests detected).
- `1`: At least one test failure occurred.
- `>1`: Operational or communication failure.

---

### `orqen chat`

Launches an interactive, conversational DevOps terminal console with real tool calling, streaming status updates, and confirmation prompts.

```bash
orqen chat [project]
```

**Built-in Slash Commands:**
- `/status` — Display current project container state and live ports.
- `/explain` — Render topology DAG and dependency map.
- `/heal` — Trigger closed-loop diagnosis and repair.
- `/tunnel` — Launch public ephemeral preview tunnel.
- `/nuke` — Request project teardown.
- `/nuke-confirm` — Confirm teardown of containers, volumes, and state.
- `/clear` — Clear the conversation buffer.
- `/onboard <path-or-url>` — Onboard a new repository from within chat.
- `/help` — Display list of available chat commands.

---

### `orqen ask`

Executes a single intent-aware request directly from the command line without opening a long-lived interactive session.

```bash
orqen ask [project] "<instruction>" [flags]
```

**Examples:**
```bash
orqen ask "why is the api service unhealthy?"
orqen ask "set REDIS_URL=redis://redis:6379 on backend"
orqen --yes ask "nuke demo-app"
```

---

### `orqen serve`

Starts the local web control plane, serving an interactive browser interface backed by Server-Sent Events (SSE) and REST APIs.

```bash
orqen serve [flags]
```

**Flags:**
- `--port <p>`: Port to bind the web server (default: `8080`).
- `--host <h>`: Host address to bind (default: `127.0.0.1`).

**Features:**
- Project selector sidebar.
- Conversational chat with tool execution pills and typewriter animations.
- Confirmation modals for destructive actions.
- Real-time "Nerd Stats" drawer with execution timelines and raw telemetry.

---

### `orqen up`

Starts all containers for a project and verifies their HTTP readiness and container health.

```bash
orqen up [project] [flags]
```

**Flags:**
- `--build`: Force image rebuilding before starting.
- `--recreate`: Recreate containers even if configuration hasn't changed.

---

### `orqen down`

Stops running containers for a project.

```bash
orqen down [project] [flags]
```

**Flags:**
- `--volumes`: Remove named volumes declared in compose.

---

### `orqen restart`

Stops, recreates, and re-probes health for all services in the specified project.

```bash
orqen restart [project]
```

---

### `orqen status`

Displays the health, image, and port allocation table for a project. When invoked without arguments, lists all onboarded projects.

```bash
orqen status [project]
```

---

### `orqen logs`

Streams or prints logs from stack containers.

```bash
orqen logs [project] [service] [flags]
```

**Flags:**
- `-f`, `--follow`: Stream live log output continuously.
- `-n`, `--tail <lines>`: Number of lines to show from the end of the logs (default: `100`).

---

### `orqen stats`

Displays live resource utilization metrics (CPU %, memory usage and limits, network I/O, PIDs) for all running containers.

```bash
orqen stats [project]
```

---

### `orqen open`

Opens the primary exposed service endpoint in your default web browser.

```bash
orqen open [project] [flags]
```

**Flags:**
- `--print`: Print the live URL to stdout instead of launching the browser.
- `--service <name>`: Open the URL of a specific service.

---

### `orqen explain`

Analyzes and renders the service dependency Directed Acyclic Graph (DAG), external infrastructure mapping, allocated host ports, and repair provenance history.

```bash
orqen explain [project]
```

---

### `orqen heal`

Initiates the two-stage self-healing workflow on an existing project: runs deterministic rule-based corrections first, then applies invariant-guarded LLM repairs if needed.

```bash
orqen heal [project] [flags]
```

**Flags:**
- `--tries <n>`: Maximum number of repair attempts (default: `3`).

---

### `orqen update`

Re-scans repository source files and regenerates Dockerfiles and `docker-compose.yml` while strictly preserving assigned host ports to avoid breaking client workflows.

```bash
orqen update [project] [flags]
```

**Flags:**
- `--no-run`: Regenerate manifests without rebuilding or restarting containers.

---

### `orqen env`

Inspects, sets, or unsets environment variables for a project. Overrides are written to `env-overrides.json` and automatically applied to running containers.

```bash
orqen env [project] [subcommand]
```

**Subcommands:**
- `orqen env <project>`: List all detected and overridden environment variables.
- `orqen env <project> set KEY=VALUE`: Set or override a variable and restart the stack.
- `orqen env <project> unset KEY`: Remove an override.
- `orqen env <project> config`: Open an interactive terminal UI for navigating and modifying environment variables with secret masking.

---

### `orqen config`

Inspects or modifies global ORQEN configuration stored in `~/.orqen/config.json`.

```bash
orqen config [subcommand]
```

**Subcommands:**
- `orqen config` or `orqen config edit`: Open interactive configuration UI.
- `orqen config show`: Display active configuration values, sources, and paths.
- `orqen config path`: Print the absolute path to `config.json`.
- `orqen config set <key> <value>`: Set a specific configuration property (script-friendly).

---

### `orqen setup`

Interactive setup wizard that configures LLM providers (Gemini, OpenAI, Anthropic, Groq, Ollama), models, and authentication keys with live validation.

```bash
orqen setup
```

---

### `orqen doctor`

Performs end-to-end preflight diagnostics on the local environment.

```bash
orqen doctor
```

Exits with code `0` when all core requirements are satisfied; exits non-zero if critical components (such as Docker or Git) are unavailable.

---

### `orqen dash`

Opens a full-screen, terminal-based dashboard powered by Bubble Tea with three live tabs:
1. **Logs**: Live streaming service logs with service filtering.
2. **Environment**: Interactive variable explorer and editor.
3. **Containers**: Process monitoring, status badges, and container lifecycle actions.

```bash
orqen dash [project]
```

---

### `orqen tunnel`

Creates an ephemeral, secure public tunnel to an exposed service port using `cloudflared` (Quick Tunnels) with a fallback to SSH `localhost.run`.

```bash
orqen tunnel [project] [flags]
```

**Flags:**
- `--port <p>`: Specific host port to tunnel (default: primary web service port).

---

### `orqen nuke`

Completely tears down a project, stopping containers, deleting docker volumes, removing shallow clones, and purging project metadata from `~/.orqen/`.

```bash
orqen nuke <project> [flags]
```

*Aliases:* `incinerate`, `remove`, `delete`

**Flags:**
- `--yes`: Skip confirmation prompt.

---

### `orqen completion`

Outputs shell autocompletion code.

```bash
orqen completion [bash|zsh|fish|powershell]
```
