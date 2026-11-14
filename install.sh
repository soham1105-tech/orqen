#!/usr/bin/env sh
# ORQEN installer — builds from source with the local Go toolchain.
set -eu

MODULE="github.com/soham1105-tech/orqen"
REPO="https://github.com/soham1105-tech/orqen.git"
BIN_DIR="${BIN_DIR:-$HOME/.local/bin}"
VERSION="${VERSION:-v0.5.0}"

say() { printf '%s\n' "$*"; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }

command -v go >/dev/null 2>&1 || die "Go 1.27+ is required (https://go.dev/dl/)"
command -v git >/dev/null 2>&1 || die "git is required"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

say "Cloning ORQEN..."
git clone --depth=1 "$REPO" "$tmp/orqen" >/dev/null 2>&1 || die "clone failed"

say "Building..."
(
  cd "$tmp/orqen"
  go build -trimpath \
    -ldflags "-s -w -X ${MODULE}/cmd.Version=${VERSION} -X ${MODULE}/cmd.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o orqen .
)

mkdir -p "$BIN_DIR"
install -m 755 "$tmp/orqen/orqen" "$BIN_DIR/orqen"
say "Installed orqen to $BIN_DIR/orqen"

case ":${PATH}:" in
  *":${BIN_DIR}:"*) ;;
  *) say "Add $BIN_DIR to your PATH to run 'orqen' from anywhere." ;;
esac