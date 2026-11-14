---
name: Bug Report
about: Report a bug in ORQEN
title: "[BUG] "
labels: bug
assignees: ''
---

## Bug Description

A clear and concise description of what the bug is.

## Environment

| Component | Version |
|-----------|---------|
| ORQEN | `orqen --version` output |
| Go | `go version` output |
| Docker | `docker version` output |
| OS | (e.g., Ubuntu 24.04, macOS 15, Windows 11) |
| Arch | (e.g., amd64, arm64) |

## Steps to Reproduce

1. Step one
2. Step two
3. Step three
4. Observe the bug

## Expected Behavior

What you expected to happen.

## Actual Behavior

What actually happened. Include error messages, logs, or screenshots.

**Important**: Sanitize any secrets, API keys, or sensitive data from logs before posting!

## Minimal Reproduction

If possible, provide a minimal repository or code sample that reproduces the issue. You can use `orqen init` on a public repo or create a small test case.

## Additional Context

Any other information, configuration, or context that might help resolve the issue.

## Checklist

- [ ] I have searched existing issues for duplicates
- [ ] I have tested with the latest `main` branch
- [ ] I have run `make check` and it passes
- [ ] I have sanitized secrets from any logs/output