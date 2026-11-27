#!/usr/bin/env bash
# Format, check, and optionally package on macOS. Never launches the application.
# --build cleans build/bin once, then preserves both the universal app and amd64 exe.
set -euo pipefail

build=false
case "$#:${1:-}" in
  0:) ;;
  1:--build) build=true ;;
  1:--help|1:-h) echo "Usage: scripts/check.sh [--build]"; exit 0 ;;
  *) echo "Usage: scripts/check.sh [--build]" >&2; exit 2 ;;
esac

cd "$(dirname "$0")/.."
export GOTOOLCHAIN=local
if [ "$(uname -s)" != Darwin ]; then
  echo "Run this verification script on macOS with Xcode Command Line Tools." >&2
  exit 1
fi
# Ignore cross-compilation overrides: the first lint and bindings generation are native.
unset GOOS GOARCH CGO_ENABLED
if ! node -e 'const [major] = process.versions.node.split(".").map(Number); process.exit(major >= 22 ? 0 : 1)'; then
  echo "Node >=22 required (found $(node -v))" >&2
  exit 1
fi
if [ ! -d frontend/node_modules ]; then
  echo "Install frontend dependencies first: (cd frontend && npm ci)" >&2
  exit 1
fi

# Include root files and future internal packages without assuming directories exist.
find . -type d \( -name .git -o -name node_modules -o -name build \) -prune -o \
  -type f -name '*.go' -exec gofmt -l -w {} +
# The frontend must build before Go lint: main.go embeds frontend/dist.
(cd frontend && npm run fix && npm run check)
golangci-lint run ./...
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 golangci-lint run ./...

if [ "$build" = true ]; then
  snapshot=$(mktemp -d "${TMPDIR:-/tmp}/lkm-bindings.XXXXXX")
  trap 'rm -rf "$snapshot"' EXIT
  cp -R frontend/wailsjs "$snapshot/wailsjs"
  # Frontend already checked/built; -s avoids implicit dependency installation.
  # Separate targets retain the standard output names; only the first build cleans.
  wails build -clean -s -platform darwin/universal
  wails build -s -platform windows/amd64
  if ! diff -qr "$snapshot/wailsjs" frontend/wailsjs; then
    echo "Bindings changed during build. Review them, regenerate before frontend checks, and rerun." >&2
    exit 1
  fi
  lipo -info build/bin/llm-key-manager.app/Contents/MacOS/llm-key-manager
  file build/bin/llm-key-manager.exe
fi
