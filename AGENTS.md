# Repository Guidelines

## Project Structure

This is a Windows-only Wails desktop app with a Go backend and Vue 3/TypeScript frontend. Do not add macOS, Linux desktop, or other platform support paths. `build/darwin/` is leftover Wails template material; do not modify or remove it unless explicitly requested.

- `main.go` and `app.go` configure the Wails window, embedded assets, backend bindings, and exported `App` methods.
- `key_store.go` contains private helpers for encrypted storage, duplicate detection, and key masking.
- `frontend/src/` contains Vue components, styles, and assets.
- `frontend/wailsjs/` contains generated Wails bindings; regenerate it after exported Go method changes.
- `build/` contains Windows packaging assets and generated output in `build/bin/`.

## Required Verification

This is a personal application with the frontend as its only user interaction entry point. Keep form validation at the frontend and avoid duplicate backend input checks, concurrency controls, or abstractions for hypothetical use cases. Preserve business rules and error handling for storage and Windows system calls.

Do not add or maintain unit tests, test frameworks, test scripts, or test-only dependency injection. Verification consists of compilation, lint, and user end-to-end testing on Windows. Report completed tool checks separately from user acceptance that has not yet been performed.

After every code edit, run formatting first, then run both frontend and backend checks before reporting completion. Do not skip one side because a change looks isolated; Go, Wails bindings, and Vue build behavior are coupled.

- Frontend: `cd frontend && npm run lint:fix && npm run format:fix && npm run lint && npm run format && npm run build`.
- Backend: `gofmt -w <edited-go-files>` and `GOCACHE=/tmp/go-build-cache GOLANGCI_LINT_CACHE=/tmp/golangci-lint-cache golangci-lint run ./...`.
- Full integration when feasible: `GOCACHE=/tmp/go-build-cache GOTOOLCHAIN=local wails build -clean -platform windows/amd64`.
- Wails bindings: when exported `App` methods are added, removed, renamed, or have signature changes, run `GOCACHE=/tmp/go-build-cache GOTOOLCHAIN=local wails generate module` before frontend checks.
- If a required command cannot run, state the exact command, failure reason, and completed partial verification.

Avoid `wails dev` on headless remote Linux hosts because it opens a desktop window.

## Personality and writing style

Default to using clear, concise paragraphs, each developing one main idea. Use lists only when the information is genuinely parallel, sequential, or easier to compare, and avoid nested lists unless the hierarchy cannot be expressed clearly in prose. Use plain, simple language: familiar words, concrete examples, and precise verbs. Prefer active voice and direct statements.

Make sure to state the main point clearly and early, then develop it with the explanation and detail the reader needs. Let each sentence build on what came before. Develop the points that matter and provide enough support to be useful.

Use plain language over jargon, and reference technical details only to the degree that it helps illustrate an idea or your work to the user. Communicate complex concepts in a clear and cohesive manner, and calibrate your writing to the level of background knowledge assumed from the user's prompt and context.

Avoid using slop words or phrases like "Bottom Line:" in conclusions, "delve," "foster," "leverage," "it's worth noting," "importantly," "Question? Answer." or "This isn't about X. It's about Y.", "genuinely" or hyphenated compound descriptions and adjectives. Do not use concluding summary statements such as "In short:..", "The simplest mental model is:...".

State the intended action directly. Avoid adding what you won't do, what will remain unchanged, or how you'll separate or categorize results. Do not use contrastive framing such as "X, not Y" or "X—not Y" that introduces an unprompted alternative that the user didn't ask about. Avoid invented compound labels like "exact-head checks" and "editorial-row layouts", vague qualifiers, and canned transitions; use plain verbs and prepositions to state the actual relationship directly.

## Coding Style

Use `gofmt` for Go. Keep exported backend methods on `App` in PascalCase and unexported helpers camelCase. Add Go doc comments for exported types, functions, methods, constants, and variables; explain behavior, constraints, side effects, or non-obvious decisions.

Frontend code uses Vue single-file components with TypeScript. Name components in PascalCase, keep shared assets under `frontend/src/assets/`, and follow existing Vue, Vite, TDesign, ESLint, and Prettier patterns. Use JSDoc/TSDoc for exported APIs, shared helpers, persistence or security behavior, and other non-obvious frontend logic.

## Security

Do not commit real API keys, local secrets, generated credentials, or machine-specific configuration. Review generated Wails bindings before committing.
