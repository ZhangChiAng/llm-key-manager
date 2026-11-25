# Repository Guidelines

Setup, commands, and the binding workflow are in the development guide of [README.md](README.md). Every rule below carries its reason. Add a rule only for a problem that has actually occurred or a project fact that makes an agent's default behaviour wrong, and delete it once the reason no longer holds.

- **Locate code with the map.** Read [docs/architecture.md](docs/architecture.md) before searching the repository; it maps features and UI terms to files and symbols.
- **Verify with `scripts/check.sh`.** Run it after each change, and with `--build` when the change touches Go code, bindings, or build configuration. The repository has no tests, so this is the only automated check.
- **Do not run the application or touch real data.** That covers launching the built app, running `wails dev`, reading the application data directory, querying the Keychain item, and sending probe requests. All of them operate on the user's real API keys, and every probe is a billed request.
- **Report only what the script verified.** You cannot observe the native window or real data, so say what `scripts/check.sh` covered and leave UI and real-data behaviour for the user to confirm.
- **Do not add tests.** No unit tests, test frameworks, test scripts, or test-only dependency injection: the user has decided that verification is the script plus their own acceptance.
