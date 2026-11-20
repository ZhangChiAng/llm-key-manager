# Repository Guidelines

## Project Structure

This is a Wails desktop app for Windows and macOS with a Go backend and Vue 3/TypeScript frontend. Do not add Linux desktop or other platform support paths. `build/windows/` and `build/darwin/` hold each platform's packaging assets.

- `main.go` and `app.go` configure the Wails window, embedded assets, backend bindings, and exported `App` methods.
- `key_store.go` contains platform-neutral private helpers for storage file I/O, path assembly, duplicate detection, and key masking.
- `platform_windows.go` (DPAPI, `%LOCALAPPDATA%`) and `platform_darwin.go` (AES-256-GCM with a login-Keychain master key read via `/usr/bin/security`, `~/Library/Application Support`) implement `protectKeyValue`, `unprotectKeyValue`, and `keyStoreBaseDir`. `platform_unsupported.go` only lets the package type-check and lint on the Linux development host. Keep the Windows storage path, error text, and DPAPI format compatible with existing `keys.json` files.
- `frontend/src/` contains Vue components, styles, and assets.
- `frontend/wailsjs/` contains generated Wails bindings; regenerate it after exported Go method changes.
- `build/` contains Windows and macOS packaging assets and generated output in `build/bin/`.

## Required Verification

This is a personal application with the frontend as its only user interaction entry point. Keep form validation at the frontend and avoid duplicate backend input checks, concurrency controls, or abstractions for hypothetical use cases. Preserve business rules and error handling for storage, Windows system calls, and macOS Keychain calls.

Do not add or maintain unit tests, test frameworks, test scripts, or test-only dependency injection. Verification consists of compilation, lint, and user end-to-end testing on Windows and macOS. Report completed tool checks separately from user acceptance that has not yet been performed.

After every code edit, run formatting first, then run both frontend and backend checks before reporting completion. Do not skip one side because a change looks isolated; Go, Wails bindings, and Vue build behavior are coupled.

- Frontend: `cd frontend && npm run lint:fix && npm run format:fix && npm run lint && npm run format && npm run build`.
- Backend: `gofmt -w <edited-go-files>`, then lint for each target. The Linux host's default GOOS compiles neither platform file, so run all three:

  ```bash
  GOCACHE=/tmp/go-build-cache GOLANGCI_LINT_CACHE=/tmp/golangci-lint-cache golangci-lint run ./...
  GOOS=windows GOCACHE=/tmp/go-build-cache GOLANGCI_LINT_CACHE=/tmp/golangci-lint-cache golangci-lint run ./...
  GOOS=darwin CGO_ENABLED=0 GOCACHE=/tmp/go-build-cache GOLANGCI_LINT_CACHE=/tmp/golangci-lint-cache golangci-lint run ./...
  ```

- Full integration when feasible: Windows on the Linux host with `GOCACHE=/tmp/go-build-cache GOTOOLCHAIN=local wails build -clean -platform windows/amd64`. macOS with `GOCACHE=/tmp/go-build-cache GOTOOLCHAIN=local wails build -clean -platform darwin/universal` only runs on a Mac with Xcode Command Line Tools; when working on Linux, report it as not performed.
- Wails bindings: when exported `App` methods are added, removed, renamed, or have signature changes, run `GOCACHE=/tmp/go-build-cache GOTOOLCHAIN=local wails generate module` before frontend checks.
- If a required command cannot run, state the exact command, failure reason, and completed partial verification.

Avoid `wails dev` on headless remote Linux hosts because it opens a desktop window.

## Coding Style

Use `gofmt` for Go. Keep exported backend methods on `App` in PascalCase and unexported helpers camelCase. Add Go doc comments for exported types, functions, methods, constants, and variables; explain behavior, constraints, side effects, or non-obvious decisions.

Frontend code uses Vue single-file components with TypeScript. Name components in PascalCase, keep shared assets under `frontend/src/assets/`, and follow existing Vue, Vite, TDesign, ESLint, and Prettier patterns. Use JSDoc/TSDoc for exported APIs, shared helpers, persistence or security behavior, and other non-obvious frontend logic.

## Security

Do not commit real API keys, local secrets, generated credentials, or machine-specific configuration. Review generated Wails bindings before committing.
