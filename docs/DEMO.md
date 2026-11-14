# ORQEN — Judge Demo Runbook

A 4–5 minute, from-scratch demonstration. All commands use the installed `orqen`
binary (fish/zsh/bash). Docker Desktop must be running.

## Pre-flight (10s)

```bash
docker info >/dev/null && echo "Docker ready"
orqen --version
orqen doctor
orqen status          # clean slate: "No projects yet"
```

Talking point: *“One binary, zero state, five provider integrations, and a
deterministic-first intent engine.”*

## Act 1 — Product surface (30s)

```bash
orqen --help
orqen config show      # provider/model/key source/paths, key masked
orqen doctor           # git, Docker, Compose v2, LLM, tunnel tooling
```

## Act 2 — From scratch: first app (~2 min; narrate while it builds)

```bash
orqen init /path/to/certify-v2
```

Call out the 8 steps:
1. resolve source → 2. scan services → 3. discover env → 4. infer infra →
5. generate manifests → 6. persist state → 7. build + heal + verify → 8. healthy.

Key points:
- Detects **Next.js** automatically and imports the repo’s real `.env`.
- Detects **Neon Postgres** and marks it **external** — no duplicate local DB.
- Health is verified with a real HTTP probe (never discarded).
- Live URL: **http://localhost:3000**

## Act 3 — Second app + port intelligence (~30s)

```bash
orqen init /path/to/portfolio-website --max-services 1
```

- Port 3000 is taken, so ORQEN automatically remaps to **3001**.
- `.env` values (Gemini, Upstash) are imported; **Upstash Redis** is external.
- Live URL: **http://localhost:3001**

Open both URLs side-by-side.

## Act 3.5 — Automated in-container testing (~60s)

Run the demo fixture (2 tests pass, 1 fails on purpose):

```bash
orqen init test/fixtures/test-sample
orqen test
```

Expected output:

```text
ORQEN TEST

RUNNING SERVICE TESTS

test-sample    2 passed, 1 failed (3.1s)

--------------------------------
TOTAL: 2 passed, 1 failed
RESULT: FAILED
```

With the failure tail showing exactly which assertion broke:

```text
--- test-sample (last 30 lines) ---
...
    assert r.json() == {"sum": 8}
E   assert 7 == 8
...
```

Note the exit code is `1` — a broken test is never dressed up as a success.

Now fix the intentional failure (open `test/fixtures/test-sample/tests/test_app.py`
and change the expected sum from `8` to `7` in `test_sum_wrong`), then rerun:

```bash
orqen test
```

Expected output:

```text
test-sample    3 passed (3.0s)

--------------------------------
TOTAL: 3 passed
RESULT: PASSED
```

Talking points:
- The tests run **inside the running container** — the same filesystem, env,
  and network as production — via `docker compose exec -T`, with no shell.
- Detection is automatic: `pytest` is a dependency, so ORQEN picked
  `python -m pytest -q`.
- Unknown output formats never count as success; a non-zero exit always fails.
- `--service <name>` runs a single service; visit order is dependency-aware.

## Act 4 — Intent-aware agent (60s)

```bash
orqen ask "what is running?"
orqen ask "explain the topology"
orqen ask "show environment for portfolio-website"
orqen ask "check the health"
orqen ask "set GEMINI_MODEL=gemini-3.5-flash-lite on portfolio-website"
orqen ask "why did it fail?"          # diagnosis with evidence + next steps
orqen ask --yes "onboard https://github.com/vercel/nextjs-portfolio-starter"
orqen --yes ask "restart portfolio-website"
```

Every request prints its parsed intent first (`🎯 ...`). Ambiguous requests ask
one clarifying question and execute nothing.

Then the interactive REPL:

```bash
orqen chat
```

```text
diagnose
why is the api unhealthy?
logs for web
use project certify-v2
heal
```

## Act 4.5 — Web control plane (60s)

```bash
orqen serve
```

Open the printed URL. Show:
- chat sidebar with persisted conversations
- project selector with live running/healthy counts
- streaming tool pills and intent chips
- `/onboard <url>` with a confirmation modal
- **Nerd stats** drawer: tool args, outputs, durations, bytes, raw events

## Act 5 — Show the engine room (60s)

```bash
orqen status
orqen explain certify-v2

set P (ls -d ~/.orqen/projects/certify-v2-* | head -1)   # fish
ls -la $P
sed -E 's/=.*/=<masked>/' $P/.env | head
cat $P/docker-compose.yml
cat $P/Dockerfile.certify-next
```

Call out:
- Portable relative build context + Dockerfile.
- Infra host-port mapping vs internal container ports (`5439:5432` style).
- 0600 `.env`, lock file, repairs.json provenance (SHA-256).
- `.dockerignore` is generated automatically.

```bash
orqen stats certify-v2
orqen logs certify-v2
orqen dash portfolio-website
```

## Act 6 — Reset / teardown (20s)

```bash
orqen down certify-v2
orqen --yes nuke portfolio-website
```

`nuke` removes containers, volumes, state, and (for remote clones) the clone,
guarded so it only deletes inside ORQEN-owned roots.

## Reset to a clean slate

```bash
orqen --yes nuke certify-v2
orqen --yes nuke portfolio-website
orqen --yes nuke test-sample
orqen status          # No projects yet
```

## Fallback if a build is slow or fails

```bash
orqen heal certify-v2          # deterministic rules first, then guarded LLM repair
orqen logs certify-v2          # inspect the failure
```

## Judge talking points

- **Dynamic port allocation** for apps *and* infra; 3000→3001 handled live.
- **Cloud suppression**: Neon/Upstash external, not duplicated locally.
- **Real verification**: HTTP + container health, no false success.
- **Automated in-container testing**: `orqen test` detects and runs each
  service's tests in its own container, with honest pass/fail/skip counts.
- **Intent-first agent**: deterministic parser → validated LLM classifier →
  clarify → confirm → act. Read-only intents need no LLM.
- **Security**: symlink-resolved repo jail, 0600 secrets, PAT via env (not argv),
  invariant-guarded patches, SHA-256 provenance.
- **Determinism**: sorted ports/graph/env; compose validated before running.
- **Zero-config env**: repo `.env` imported; `orqen env` shows/edits live.
