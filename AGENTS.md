# AGENTS.md

DevMan — Windows development environment manager. Wails v2 desktop app: Go backend + React/TypeScript frontend.

## Commands

| Command | Where | Purpose |
|---|---|---|
| `wails dev` | repo root | Live dev with hot reload (Go + frontend) |
| `wails build` | repo root | Production build, outputs to `build/bin/` |
| `go test ./...` | repo root | Run all Go tests |
| `go test ./internal/registry` | repo root | Single-package test |
| `npm install` | `frontend/` | Install frontend deps |
| `npm run build` | `frontend/` | `tsc && vite build` → `frontend/dist/` |

**Build order**: `wails build` handles both frontend and backend. For focused dev, run `wails dev` and it auto-reloads on changes.

## Architecture

```
main.go          → Wails entrypoint, window config, binds App struct
app.go           → API surface exposed to frontend (ScanAll, GetEnvs, Migrate, ...)
internal/
  models/        → Shared data types (PascalCase JSON tags for Wails bridge)
  registry/      → SQLite persistence (go-sqlite3). Auto-creates schema on Open().
  scanner/       → Env detectors: NodeScanner, PythonScanner, JavaScanner, ...
                  To add a new scanner: implement the Scanner interface, register in NewEngine()
  migrator/      → Dir migration with 10-step pipeline (snapshot→copy→verify→commit→junction)
  utils/         → Disk info (platform-specific: disk.go dispatches via build tags)
frontend/
  src/
    App.tsx      → SPA router (6 pages: dashboard, environments, migration, cleaner, versions, settings)
    components/  → Shared UI: Sidebar, Panel
    pages/       → One file per page
    bindings/go/ → Wails auto-generated bridge — do NOT edit
  wailsjs/       → Wails runtime bridge — auto-generated, do NOT edit
build/           → Platform build assets (icons, manifests)
```

**Data flow**: Frontend calls Go methods on the bound `App` struct via Wails bridge → Go reads/writes SQLite → returns JSON with PascalCase keys matching frontend types in `devman-types.ts`.

## Key conventions

### JSON naming (critical — Wails bridge requirement)
Go struct JSON tags use **PascalCase**: `json:"Id"`, `json:"EnvId"`, `json:"TotalSize"`. Frontend types mirror this exactly. Do NOT change to camelCase — it will break the Wails binding serialization.

### Platform-specific code
Use Go build tags for OS branching:
- `//go:build windows` — Windows implementation
- `//go:build !windows` — Linux stub
- Files: `windows.go` / `linux.go` or `disk_windows.go` / `disk_stub.go`

### SQLite database
- Portable mode: stored as `devman.db` next to the executable (via `os.Executable()`)
- Tests redirect via `registry.DbPathOverride` global var — tests use `/tmp/test_devman_*.db`
- Schema is created in `registry.migrate()` with `CREATE TABLE IF NOT EXISTS` — no migration framework, just idempotent DDL
- Integer booleans: `is_managed`, `is_default`, `is_active`, `success` stored as 0/1, mapped to Go bool via `!= 0`

### Expanding scanners
1. Create a scanner struct in `internal/scanner/` implementing the `Scanner` interface
2. Add an env model entry in `modelsForScanner()` in `scanner.go`
3. Register in `NewEngine()` scanner list

### Wails bridge types
Frontend types in `frontend/src/devman-types.ts` must stay in sync with `internal/models/models.go` struct field names (PascalCase). The Wails `Bind` in `main.go` auto-exposes all public methods on the `App` struct.

### Tailwind
Uses Tailwind CSS v4 with `@tailwindcss/postcss` plugin (not the standalone CLI). Custom design tokens in `tailwind.config.js` under the `devman` color namespace. Dark theme only — no light mode.

## Gotchas

- **`frontend/wailsjs/` and `frontend/src/bindings/go/` are auto-generated** by Wails. Never edit these files.
- **`frontend/dist/`** is the Vite build output, gitignored.
- **Scanner `readVersion()` methods return `"detected"`** — real version parsing not yet implemented. This is a known TODO.
- **Linux is mostly stubs** — the project targets Windows primarily. `disk.go` returns fake data on Linux, `migrator/linux.go` no-ops.
- **Migration pipeline must not be interrupted mid-step** — it creates `.devman_tmp` staging dirs, then renames+deletes. Failed migrations leave cleanup to manual handling.
- **Tests use `/tmp/` paths** which work on WSL/MinGW but not on native Windows `cmd`. Run tests inside a bash-like environment or WSL.
- **The `devman-test.exe` binary** in repo root is a test artifact — don't commit it (it's gitignored via `build/bin` pattern).
- **Single initial commit** `feat: initial DevMan v0.1.0` — the repo is early-stage, patterns may evolve.
