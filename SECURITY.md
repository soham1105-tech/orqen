# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| 0.5.x   | ✅        |
| < 0.5   | ❌        |

## Reporting a Vulnerability

**Please do not report security vulnerabilities via public GitHub issues.**

Instead, please report them privately through [GitHub Security Advisories](https://github.com/soham1105-tech/orqen/security/advisories/new).

We will acknowledge receipt within 48 hours and provide a more detailed response within 7 days.

## Security Features

### Secrets Management
- `.env`, `.env.*`, and the `orqen` binary are gitignored; only `.env.example` is committed
- ORQEN writes project env files with `0600` permissions and refuses group/other-writable files
- Credential-bearing URLs (e.g. `postgres://user:pass@host`) are masked in `orqen env`
- Git tokens are passed through `GIT_CONFIG_*` environment variables, never argv, and are scrubbed from errors

### Filesystem Boundary
- All agent file reads and writes go through `safefs.Jail`, which resolves symlinks on the longest existing prefix, rejects `..` and absolute escapes, and caps reads at 1 MiB
- The healer may only modify generated `Dockerfile.*`, `docker-compose.yml`, and `.dockerignore` inside the project workspace
- `nuke` only removes paths inside ORQEN-owned roots and refuses `/`, `.`, and the home directory

### Repair Invariants
LLM patches are rejected when they would:
- Disable type checking
- Remove a healthcheck
- Replace a command with `sleep`
- Drop install/build steps
- Use absolute compose paths
- Rewrite beyond a bounded scope

Applied patches are recorded with before/after SHA-256 in `repairs.json`.

### Test Execution Safety
- `orqen test` never spawns a shell: the test command is tokenized and handed to `docker compose exec -T`, so injected quoting cannot become a shell metacharacter
- Test commands run inside the service's container with the same filesystem, environment, and network as production

### Network Security
- Dynamic host port allocation prevents port conflicts and predictable port exposure
- Internal container communication uses standard container ports (e.g., `postgres:5432`)
- Cloud-managed services (Neon, Supabase, Upstash, RDS, Atlas, Confluent Cloud) are detected and marked external — no local duplicates are created

## Security Best Practices for Users

1. **Never commit `.env` files** — they are gitignored for a reason
2. **Use `orqen setup`** for interactive credential configuration with live verification
3. **Rotate API keys regularly** — especially after team changes
4. **Review generated manifests** — `orqen explain` shows the full dependency graph and port mapping
5. **Use `orqen doctor`** to verify your environment before deploying
6. **Enable `--yes` only in trusted automation** — it bypasses confirmation prompts

## Security Audit

The codebase is regularly audited for:
- Path traversal vulnerabilities in `safefs`
- Shell injection in `docker` and `testrun` packages
- LLM prompt injection in `agent` and `healer` packages
- Credential leakage in error messages and logs

## Contact

For security concerns, please use [GitHub Security Advisories](https://github.com/soham1105-tech/orqen/security/advisories/new).