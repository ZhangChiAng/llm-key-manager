# Repository Guidelines

## Project Structure & Module Organization

This is a Windows-only Wails desktop app with a Go backend and Vue 3/TypeScript frontend. Do not add support paths for macOS, Linux desktop, or other platforms. `build/darwin/` is leftover Wails template material; do not modify or remove it unless explicitly requested.

- `main.go` and `app.go` configure the Wails window, embedded assets, backend bindings, and exported `App` methods callable from the frontend.
- `frontend/src/` contains Vue code, components, styles, and assets.
- `frontend/wailsjs/` contains generated Wails bindings; regenerate these after changing exported Go methods.
- `build/` contains Windows packaging assets and generated output in `build/bin/`.

## Build, Test, and Development Commands

- `cd frontend && npm install`: install frontend dependencies when needed.
- `cd frontend && npm run build`: type-check and build the production frontend.
- `cd frontend && npm run lint`: run ESLint for Vue and TypeScript code.
- `cd frontend && npm run format`: check formatting for Vue, TypeScript, CSS, Markdown, and root Markdown docs.
- `GOCACHE=/tmp/go-build-cache GOLANGCI_LINT_CACHE=/tmp/golangci-lint-cache golangci-lint run ./...`: run Go lint checks from `.golangci.yml`.
- `GOCACHE=/tmp/go-build-cache GOTOOLCHAIN=local wails build -clean -platform windows/amd64`: run a full Windows Wails build check.
- `GOCACHE=/tmp/go-build-cache GOTOOLCHAIN=local wails generate module`: regenerate `frontend/wailsjs` after changing exported backend method signatures.

Avoid `wails dev` on headless remote Linux hosts because it opens a desktop window.

## Coding Style & Naming Conventions

Use `gofmt` for Go files. Keep exported backend methods on `App` in PascalCase so Wails can bind them, and keep unexported helpers camelCase. Add Go doc comments for exported types, functions, methods, constants, and variables. Comments should explain behavior, constraints, side effects, or non-obvious decisions.

Use the repository `.golangci.yml` configuration for Go static analysis, including `govet`, `staticcheck`, and `revive` comment rules.

Frontend code uses Vue single-file components with TypeScript. Name components in PascalCase, for example `KeyList.vue`, and keep shared assets under `frontend/src/assets/`. Prefer existing Vue, Vite, and TDesign patterns. Use JSDoc/TSDoc for exported APIs, shared helpers, persistence or security behavior, and other non-obvious frontend logic.

Use the frontend ESLint flat config in `frontend/eslint.config.js` and formatting tools from `frontend/package.json`. Do not require docstrings for every local helper.

## Security & Configuration Tips

Do not commit real API keys, local secrets, or generated credentials. Keep machine-specific configuration out of source control, and review generated Wails bindings before committing.
