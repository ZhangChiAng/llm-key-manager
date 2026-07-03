# Repository Guidelines

## Project Structure & Module Organization

This is a Windows-only Wails desktop app with a Go backend and Vue 3/TypeScript frontend. Do not add or maintain support paths for macOS, Linux desktop, or other platforms. `build/darwin/` is leftover Wails template material; do not modify or remove it unless explicitly requested.

- `main.go` and `app.go` configure the Wails window, embedded assets, backend bindings, and exported `App` methods callable from the frontend.
- `frontend/src/` contains Vue application code, components, styles, and assets.
- `frontend/wailsjs/` contains generated Wails bindings; regenerate these after changing exported Go methods.
- `build/` contains Windows packaging assets and generated output under `build/bin/`.

## Build, Test, and Development Commands

- `cd frontend && npm install`: install frontend dependencies when needed.
- `cd frontend && npm run dev -- --host 127.0.0.1`: start Vite for browser-based UI work.
- `cd frontend && npm run build`: run `vue-tsc --noEmit` and build the production frontend.
- `GOTOOLCHAIN=local wails build -clean -platform windows/amd64`: run a full Windows Wails build check.
- `wails generate module`: regenerate `frontend/wailsjs` after changing exported backend method signatures.
- `GOTOOLCHAIN=local wails build -clean -platform windows/amd64 -nsis`: build a Windows installer from Linux when cross-compilation tools are installed.

Avoid `wails dev` on headless remote Linux hosts because it tries to open a desktop window.

## Coding Style & Naming Conventions

Use `gofmt` for Go files. Keep exported backend methods on `App` in PascalCase so Wails can bind them. Keep unexported helpers camelCase.

Frontend code uses Vue single-file components with TypeScript. Name components in PascalCase, for example `KeyList.vue`, and keep shared assets under `frontend/src/assets/`. Prefer existing Vue, Vite, and TDesign patterns.

## Security & Configuration Tips

Do not commit real API keys, local secrets, or generated credentials. Keep machine-specific configuration out of source control, and review generated Wails bindings before committing them.
