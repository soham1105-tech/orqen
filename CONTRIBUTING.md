# Contributing to ORQEN

Thank you for your interest in contributing to ORQEN! This document provides guidelines for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Environment](#development-environment)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Code Style](#code-style)
- [Architecture Overview](#architecture-overview)
- [Reporting Issues](#reporting-issues)

---

## Code of Conduct

This project adheres to a [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to the maintainers.

---

## Getting Started

### Prerequisites

- **Go 1.27+** — [Download](https://go.dev/dl/)
- **Docker** with Compose v2 — [Install](https://docs.docker.com/get-docker/)
- **Git** — [Download](https://git-scm.com/)
- **Make** — Standard on Linux/macOS; `choco install make` or `scoop install make` on Windows

### Fork and Clone

```bash
# Fork the repository on GitHub, then clone your fork
git clone https://github.com/YOUR-USERNAME/orqen.git
cd orqen

# Add upstream remote
git remote add upstream https://github.com/soham1105-tech/orqen.git
```

### Build and Test

```bash
# Build the binary
make build

# Run all tests (including race detector)
make check

# Run end-to-end tests (no Docker required)
make e2e

# Run live heal proof (requires Docker + LLM API key)
make heal-proof
```

---

## Development Environment

### Editor Setup

We recommend **VS Code** with the **Go extension** (golang.go). The repository includes `.vscode/settings.json` with recommended settings.

### Useful Commands

```bash
make fmt            # Format all Go files
make vet            # Run go vet
make test           # Run unit tests
make race           # Run tests with race detector
make tidy           # Tidy module dependencies
make clean          # Remove build artifacts
```

### Running the Binary

```bash
# From the repo root
./orqen --help

# Or install to GOBIN
make install
orqen --help
```

---

## Making Changes

### Branch Naming

Use descriptive branch names:

- `feat/<short-description>` — new features
- `fix/<short-description>` — bug fixes
- `docs/<short-description>` — documentation updates
- `refactor/<short-description>` — code refactoring
- `test/<short-description>` — test additions/updates

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `perf`, `ci`

Examples:

```
feat(test): add support for vitest test runner
fix(scanner): handle pnpm workspaces correctly
docs(readme): update installation instructions
refactor(agent): extract intent parser to separate package
test(generator): add cross-drive build context test
```

### Keep Changes Focused

- One logical change per PR
- Keep PRs small and reviewable
- Separate refactoring from functional changes

---

## Testing

### Test Requirements

- All new code must have tests
- Tests must pass `go test -race ./...`
- No flaky tests — fix or remove them
- Integration tests should be in `_test.go` files alongside the code they test

### Running Tests

```bash
# All tests with race detector (CI standard)
make check

# Specific package
go test ./internal/scanner/... -v

# Specific test
go test ./internal/testrun -run TestParsePytest -v

# End-to-end tests
make e2e

# Live heal proof (requires Docker + LLM key)
make heal-proof
```

### Test Conventions

- Use `testing.T` helpers for common setup
- Table-driven tests for multiple cases
- Mock external dependencies (Docker, LLM, Git)
- Test fixtures in `test/fixtures/`

---

## Pull Request Process

### Before Submitting

1. **Sync with upstream**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run the full verification suite**
   ```bash
   make check
   ```

3. **Update documentation** if your change affects user-facing behavior

4. **Add tests** for new functionality

### PR Template

When opening a PR, please fill out the template with:

- **Description**: What does this PR do?
- **Motivation**: Why is this change needed?
- **Testing**: How was this tested?
- **Breaking Changes**: Any breaking changes? Migration path?
- **Screenshots**: For UI changes

### Review Process

- All PRs require at least one maintainer review
- CI must pass (gofmt, vet, tests, race, build)
- Maintainers may request changes
- Once approved, a maintainer will merge

---

## Code Style

### Go

- Follow [Effective Go](https://go.dev/doc/effective_go) and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- `gofmt` is mandatory — `make fmt` before committing
- `go vet` must pass — `make vet`
- No lint warnings from `golangci-lint` (if configured)

### Naming Conventions

| Type | Convention |
|------|------------|
| Packages | Lowercase, single word (`scanner`, `testrun`) |
| Exported types | PascalCase (`Project`, `Lock`) |
| Unexported types | PascalCase (`config`, `lock`) |
| Exported functions | PascalCase (`Resolve`, `Home`) |
| Unexported functions | PascalCase (`resolve`, `home`) |
| Interfaces | PascalCase, often ending in `-er` (`ComposeController`) |
| Error types | PascalCase, ending in `Error` (`MissingLockError`) |
| Constants | PascalCase (`DefaultModel`, `lockFile`) |
| Variables | camelCase (`projectID`, `repoPath`) |

### Error Handling

- Use the typed error taxonomy in `internal/domain/errors.go`
- Wrap errors with context: `domain.E(kind, op, msg, err)`
- Return `domain.ErrNotConfigured`, `domain.ErrNotFound`, etc. for common cases
- Never panic in library code

### Documentation

- Export comments on all public types, functions, and constants
- Keep comments concise and accurate
- Update comments when behavior changes

---

## Architecture Overview

### Package Structure

```
cmd/                 CLI commands (Cobra)
internal/
  agent/             Agent: intent engine, tool registry, state machine
  config/            Configuration: env, dotenv, config.json
  dashboard/         Bubble Tea TUI dashboard
  docker/            Docker Compose boundary (single os/exec entry)
  domain/            Pure types, interfaces, error taxonomy
  envvar/            Environment variable discovery
  generator/         Dockerfile + docker-compose generation
  git/               Git operations (clone, URL parsing)
  graph/             Dependency DAG, topological ordering
  healer/            Failure analysis, deterministic + LLM repair
  httpcheck/         HTTP readiness probes
  infra/             Infrastructure inference, port allocation
  llm/               Multi-provider LLM client, tool calling
  portprobe/         Host port availability checking
  reconciler/        Port assignment, URL wiring
  safefs/            Symlink-resolved filesystem jail
  scanner/           Language/framework detection + test cmd detection
  state/             Per-project atomic state (locks, env, repairs)
  testrun/           In-container test execution + output parsing
  tunnel/            Public tunnel management (cloudflared/ssh)
  web/               Web control plane (SSE + static assets)
```

### Key Design Principles

1. **One boundary per external system** — Each external system has a single package that owns all interaction
2. **Determinism** — Sorted maps, stable output, reproducible builds
3. **Honest results** — Failures surface as errors; no false success
4. **Jailed filesystem** — All agent file access through `safefs.Jail`
5. **Agent-first** — The agent harness is the core; CLI commands are thin wrappers

---

## Reporting Issues

### Bug Reports

Use the [Bug Report template](.github/ISSUE_TEMPLATE/bug_report.md) and include:

- ORQEN version (`orqen --version`)
- Go version (`go version`)
- Docker version (`docker version`)
- OS and architecture
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs (sanitize secrets!)
- Minimal reproduction case if possible

### Feature Requests

Use the [Feature Request template](.github/ISSUE_TEMPLATE/feature_request.md) and include:

- Problem statement
- Proposed solution
- Alternatives considered
- Use cases

### Security Issues

**Do not open public issues for security vulnerabilities.** Report them privately via [GitHub Security Advisories](https://github.com/soham1105-tech/orqen/security/advisories/new).

---

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).

---

## Questions?

- Open a [Discussion](https://github.com/soham1105-tech/orqen/discussions)
- Check existing [Issues](https://github.com/soham1105-tech/orqen/issues)
- Review the [Documentation](docs/)

Thank you for contributing to ORQEN! 🚀