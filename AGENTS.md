# Repository Guidelines

## Project Structure

This is a Windows-only Wails desktop app with a Go backend and Vue 3/TypeScript frontend. Do not add macOS, Linux desktop, or other platform support paths. `build/darwin/` is leftover Wails template material; do not modify or remove it unless explicitly requested.

- `main.go` and `app.go` configure the Wails window, embedded assets, backend bindings, and exported `App` methods.
- `frontend/src/` contains Vue components, styles, and assets.
- `frontend/wailsjs/` contains generated Wails bindings; regenerate it after exported Go method changes.
- `build/` contains Windows packaging assets and generated output in `build/bin/`.

## Required Verification

After every code edit, run formatting first, then run both frontend and backend checks before reporting completion. Do not skip one side because a change looks isolated; Go, Wails bindings, and Vue build behavior are coupled.

- Frontend: `cd frontend && npm run lint:fix && npm run format:fix && npm run lint && npm run format && npm run build`.
- Backend: `gofmt -w <edited-go-files>` and `GOCACHE=/tmp/go-build-cache GOLANGCI_LINT_CACHE=/tmp/golangci-lint-cache golangci-lint run ./...`.
- Full integration when feasible: `GOCACHE=/tmp/go-build-cache GOTOOLCHAIN=local wails build -clean -platform windows/amd64`.
- Wails bindings: when exported `App` methods are added, removed, renamed, or have signature changes, run `GOCACHE=/tmp/go-build-cache GOTOOLCHAIN=local wails generate module` before frontend checks.
- If a required command cannot run, state the exact command, failure reason, and completed partial verification.

Avoid `wails dev` on headless remote Linux hosts because it opens a desktop window.

## Coding Style

Use `gofmt` for Go. Keep exported backend methods on `App` in PascalCase and unexported helpers camelCase. Add Go doc comments for exported types, functions, methods, constants, and variables; explain behavior, constraints, side effects, or non-obvious decisions.

Frontend code uses Vue single-file components with TypeScript. Name components in PascalCase, keep shared assets under `frontend/src/assets/`, and follow existing Vue, Vite, TDesign, ESLint, and Prettier patterns. Use JSDoc/TSDoc for exported APIs, shared helpers, persistence or security behavior, and other non-obvious frontend logic.

## Security

Do not commit real API keys, local secrets, generated credentials, or machine-specific configuration. Review generated Wails bindings before committing.
