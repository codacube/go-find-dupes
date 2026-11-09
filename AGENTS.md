# AGENTS.md

This repository contains the Go command-line utility `go-find-dupes` in the `src/` directory. The project scans local folders and Google Drive for duplicate files, groups them by size/hash, and exposes a CLI plus a terminal UI for manual review and deletion.

## Working directory and repository layout

- The Go module root is `src/`.
- Run Go commands from that directory unless a task explicitly says otherwise.
- Entry point: `src/main.go`
- CLI wiring: `src/cmd/*.go`
- Core logic: `src/internal/processor/*.go`
- Storage interfaces and drivers: `src/internal/storage/`, `src/internal/drivers/`
- Domain model: `src/internal/model/`
- UI: `src/internal/tui/`
- Google auth: `src/internal/auth/`
- Generated/exported artifacts: `src/output/`, `src/docker/`, `src/*.exe`, and `src/token.json`
- Sensitive local files: `_no-sync/` and other local-only secrets; do not commit them.

## Project overview

The app follows a simple strategy + pipeline design:

- A source strategy exposes how files are discovered and handled.
- Concrete drivers implement the storage/drive behavior for local disk and Google Drive.
- The processor pipeline groups files by size, then by hash, and optionally deletes marked duplicates.
- Cobra provides the CLI, and the TUI uses Bubble Tea for interactive review.

## Typical commands

From `src/`:

```bash
go test ./...
# or targeted tests

go test -v ./internal/processor

go build -o go-find-dupes .

gofmt -w ./...
```

Useful debug patterns copied from the project notes:

```bash
watchexec -r -e go -- dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient --continue -- local <local_folder>

watchexec -r -e go -- dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient -- gdrive -c ./_no-sync/secrets/private-client_secret_oauth2.json -f <folder_id>
```

Remove `--continue` if the debugger starts immediately and does not stop on breakpoints.

## Important conventions for AI agents

- Prefer small, targeted changes that match the existing architecture.
- Keep the source strategy and processor pipeline patterns intact when changing file-source or duplicate-detection behavior.
- When a change affects file deletion, preserve the `MarkForDeletion` flow and the progress callback semantics used by the processor.
- Do not check in OAuth credentials, generated tokens, or any file inside `_no-sync/`.
- Treat `token.json` and similar runtime artifacts as local-only unless a task explicitly requires updating them.
- Use descriptive Go names and keep package responsibilities separated: CLI code in `cmd/`, domain logic in `internal/model` and `internal/processor`, drivers in `internal/drivers`, and UI behavior in `internal/tui`.

## Notes for debugging and validation

- Existing project documentation is in `src/README.md`; use it as the first check when confirming how a command is intended to run.
- If a task depends on Google Drive access, it usually needs credentials located under `_no-sync/` or a matching local secret setup.
- Do not assume a root-level Makefile exists for this repo; the repository’s current layout is code-first and command execution is expected to use `go` directly from `src/`.
