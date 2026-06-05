# DevMan Version Management Design

## Summary

DevMan will add a version management console for Go, Node.js, Bun, and Flutter.
The feature lets users inspect installed versions, query official releases,
download and install selected versions, switch the globally active version, and
uninstall or remove versions safely.

The selected architecture is a stable DevMan shim entry point with intelligent
installation root discovery. DevMan owns the command entry point, but it does
not force every runtime into a fixed DevMan-owned `runtimes` directory.

## Current State

The current Versions page lists managed environments and their detected
instances. It does not query official versions, download releases, install new
versions, switch active versions, or uninstall installed versions.

The existing scanner remains useful as the source of local environment
detection. The new feature should not turn the scanner into an installer. A new
version-management module will own official release lookup, install planning,
downloads, switching, and deletion policy.

## Goals

1. Manage multiple installed versions of Go, Node.js, Bun, and Flutter.
2. Switch the Windows global command entry point for a tool to a selected
   version.
3. Query official release metadata and download official builds.
4. Infer installation roots from current system environment variables, PATH
   entries, and known tool conventions.
5. Avoid taking over tools already managed by nvm, fnm, gvm, asdf, fvm, or
   similar version managers.
6. Allow safe uninstall of DevMan-installed versions.
7. Avoid unsafe deletion of externally installed versions unless the user
   explicitly force-confirms the operation.
8. Preserve Wails bridge JSON naming with PascalCase fields.

## Non-Goals

- Supporting every scanner-detected tool in the first release.
- Managing Python, Java, Docker, pnpm, Rust, or arbitrary SDKs in the first
  release.
- Guaranteeing that already-open terminal sessions immediately receive updated
  environment variables. The target guarantee is new terminals and DevMan-run
  verification commands.
- Using third-party mirrors as default download sources.
- Taking over version-manager-owned installations.

## Terms

### Tool

One supported development runtime or SDK: `go`, `node`, `bun`, or `flutter`.

### Local Version

A version found on disk. It can be DevMan-installed, externally installed, or
owned by another version manager.

### Official Version

A version returned by the official release source for a supported tool.

### DevMan Installed

A version downloaded and installed by DevMan. DevMan can uninstall it directly.

### External

A version found from PATH or environment variables but not installed by DevMan.
DevMan can show it and optionally switch to it if no version-manager conflict
exists, but physical deletion is restricted.

### Version Manager Owned

A version controlled by tools such as nvm, fnm, gvm, asdf, or fvm. DevMan shows
the conflict and does not download, switch, or delete that tool in the first
release.

## Architecture

Add `internal/versionmanager` as the feature boundary.

Responsibilities:

- Detect supported tool status.
- Detect conflicting version managers.
- Resolve local versions into normalized metadata.
- Query official version sources.
- Build install previews.
- Download and extract official releases.
- Generate and update shims.
- Update user-level environment variables and PATH.
- Validate the active command after switching.
- Enforce uninstall and delete policy.

`internal/scanner` continues to detect local environments and save general
environment metadata. `internal/versionmanager` may reuse scanner output, but it
does not rely on scanner data as the only source of truth. Direct command and
PATH inspection is required for switch validation.

`app.go` exposes Wails methods and delegates to `versionmanager`. It should not
contain download, PATH, shim, or delete policy logic.

## Shim Strategy

DevMan maintains one stable command entry directory under the DevMan executable
directory:

- `<DevMan.exe dir>\shims`

The user-level PATH entry uses an environment variable rather than a hard-coded
absolute path:

- `%DEVMAN_HOME%\shims`

On startup, DevMan refreshes `DEVMAN_HOME` to the current executable directory.
This keeps the entry point portable when the DevMan application directory moves.

Switching a tool updates the matching shim commands:

- Go: `go.cmd`
- Node.js: `node.cmd`, `npm.cmd`, `npx.cmd`
- Bun: `bun.cmd`
- Flutter: `flutter.cmd`, `dart.cmd`

Each shim calls the currently selected version using a fully resolved target path
inside the generated shim file. Paths must be quoted to handle spaces.

## PATH and Environment Safety

Switching updates the current tool's user-level environment variables when
appropriate:

- Go: `GOROOT`
- Node.js: `NODE_HOME`
- Bun: `BUN_INSTALL`
- Flutter: `FLUTTER_ROOT`

DevMan ensures `%DEVMAN_HOME%\shims` exists in the user-level PATH. If another
PATH entry resolves `go`, `node`, `bun`, or `flutter` before the DevMan shim,
DevMan reports a PATH priority conflict and offers a repair action that moves
the DevMan shim entry earlier in user PATH.

Before modifying PATH, environment variables, or shims, DevMan records a rollback
snapshot. If switching fails, DevMan restores the previous shim and environment
state.

After switching, DevMan verifies the result by resolving and running:

- `go version`
- `node --version`
- `bun --version`
- `flutter --version`

The UI displays the resolved command path and version output.

## Installation Root Discovery

Versions are not installed under `<DevMan.exe dir>\runtimes` by default.
Each tool gets an `InstallRootResolver` that returns a suggested target path with
a confidence level and reason.

### Go

Resolution order:

1. `GOROOT`
2. PATH entry containing `go.exe`
3. Common sibling directories near the detected Go SDK root

If the current version root is `D:\production\go1.26`, installing `1.25` should
suggest a sibling root such as `D:\production\go1.25`.

### Node.js

Resolution order:

1. `NODE_HOME`
2. PATH entry containing `node.exe`
3. npm prefix-derived root when available

If `NVM_HOME`, fnm, or asdf ownership is detected, DevMan does not produce an
install plan for Node.js in the first release.

### Bun

Resolution order:

1. `BUN_INSTALL`
2. PATH entry containing `bun.exe`

asdf or other detected manager ownership blocks DevMan install and switch
actions for Bun in the first release.

### Flutter

Resolution order:

1. `FLUTTER_ROOT`
2. PATH entry containing `flutter.bat`

If fvm or asdf ownership is detected, DevMan only displays the conflict.

### Saved Strategy

After the user confirms an installation root for a tool, DevMan saves that
tool's install-root strategy. Future installs for the same tool use the saved
strategy first while still showing the plan before installation.

## Official Version Sources

Default version lookup uses official sources only:

- Go: official Go download metadata from `go.dev/dl`
- Node.js: official Node.js distribution metadata from `nodejs.org/dist/index.json`
- Bun: official Bun release source
- Flutter: official Flutter release metadata

Third-party mirrors are out of scope for the first release. A later release can
add configurable mirrors.

Network parsing must be isolated behind provider interfaces so tests can use
fixtures instead of live network access.

## Data Model

New models should use PascalCase JSON tags for Wails compatibility.

### ToolVersionCatalog

Fields:

- `ToolKey`
- `Versions`
- `FetchedAt`
- `SourceUrl`

### AvailableVersion

Fields:

- `Version`
- `Stable`
- `ReleaseDate`
- `Arch`
- `DownloadUrl`
- `Checksum`

### ManagedVersion

Fields:

- `Id`
- `ToolKey`
- `Version`
- `InstallPath`
- `BinPath`
- `Source`
- `IsDefault`
- `IsActive`
- `CanDelete`
- `DeletePolicy`
- `DetectedAt`

`Source` values:

- `devman`
- `external`
- `version_manager`

### VersionInstallPlan

Fields:

- `ToolKey`
- `Version`
- `TargetDir`
- `DownloadUrl`
- `ArchiveName`
- `ExtractedDir`
- `WillOverwrite`
- `ResolverReason`
- `EnvironmentChanges`

### VersionOperationResult

Fields:

- `Success`
- `Message`
- `ToolKey`
- `Version`
- `AffectedPaths`
- `AffectedEnvironment`
- `RollbackAvailable`
- `VerificationCommand`
- `VerificationOutput`

## Persistence

Prefer new version-management tables over overloading scanner-only data:

- `tool_versions`
- `version_operations`
- `version_install_strategies`

The Versions page can merge scanner-discovered instances with
versionmanager-managed records for display. This preserves scanner semantics and
keeps install/delete ownership explicit.

## Wails API

Add Wails methods on `App` that delegate to `internal/versionmanager`:

- `ListToolVersions()`
- `FetchOfficialVersions(toolKey string)`
- `PreviewVersionInstall(toolKey string, version string)`
- `InstallVersion(toolKey string, version string, targetDir string)`
- `SwitchVersion(toolKey string, instanceId int64)`
- `UninstallVersion(instanceId int64, forceDeleteExternal bool)`
- `DetectVersionManager(toolKey string)`

Long-running operations emit progress events:

- `version:operation-progress`

Progress payload includes:

- `Operation`
- `ToolKey`
- `Version`
- `Step`
- `Percent`
- `Message`

## Deletion Policy

Deletion depends on source:

- `devman`: physical uninstall is allowed.
- `external`: default action is remove from DevMan tracking or hide from the
  DevMan list. Physical deletion requires force confirmation.
- `version_manager`: DevMan does not delete or switch it in the first release.

If a DevMan-installed version is the current default version, deletion requires
the user to switch to another version first or choose an automatic fallback
version.

Forced external deletion requires a second confirmation that shows:

- Full path
- Tool and version
- Whether it is current default or active
- Commands that may be affected
- Estimated delete size when available

## Frontend UX

The Versions page becomes a tool-focused console.

Top area:

- Tool filter or tabs for Go, Node.js, Bun, Flutter
- Current default version
- Active command path
- Managed status
- PATH conflict state
- Version manager conflict state

Tool detail area:

- Left side: local versions
- Right side: official versions

Local version actions:

- Set default
- Uninstall or remove tracking
- Open directory
- View verification details

Official version actions:

- Refresh official versions
- Search/filter stable versions
- Preview install
- Download and install

Installation preview:

- Shows target directory
- Shows resolver reason
- Shows download source
- Shows environment changes
- Allows target directory override before install

Failure states should provide the next useful action: retry, repair PATH,
rollback, open logs, or cancel.

## Version Manager Conflict Handling

If DevMan detects an existing version manager for a tool, first release behavior
is intentionally conservative:

- Show detected manager and evidence.
- Show current versions if they can be safely discovered.
- Disable DevMan download, switch, and physical deletion for that tool.
- Do not modify that tool's PATH or environment variables.

This avoids competing with existing tools that already own PATH behavior.

## Phased Delivery

### Phase 1: Model, Detection, and Install Planning

- Add versionmanager module boundary.
- Add local version status for Go, Node.js, Bun, Flutter.
- Add version-manager conflict detection.
- Add install root resolvers.
- Add install preview API.
- Update frontend to show status and install plans.

No download, switch, or uninstall yet.

### Phase 2: Shim Ownership and Version Switching

- Add shim generation.
- Add user-level PATH and environment update abstraction.
- Add rollback snapshots.
- Add switch API.
- Add command verification.
- Add PATH conflict display and repair.

### Phase 3: Official Version Lookup and Installation

- Add official source providers.
- Add fixture-backed provider tests.
- Add download and extraction flow.
- Add install progress event.
- Persist DevMan-installed versions.

### Phase 4: Uninstall and Delete Policy

- Add DevMan-installed version uninstall.
- Add external remove-tracking behavior.
- Add forced external delete with second confirmation.
- Add operation history.

## Testing

Backend tests:

- Install root resolver behavior with controlled env and PATH inputs.
- Version manager detection.
- Official source parsing with fixtures.
- Shim generation with quoted paths and spaces.
- PATH insertion, priority repair, and rollback through a mock Windows
  environment writer.
- Deletion policy for `devman`, `external`, and `version_manager` sources.

Frontend tests:

- Versions page renders local and official versions.
- Conflict states disable unsafe actions.
- Install preview shows resolver reason and target directory.
- Switch result displays verification command output.

Manual verification:

- Install two versions of each supported tool where feasible.
- Switch between versions and open a new terminal to verify command output.
- Validate that existing version manager installations are not modified.
- Move DevMan executable directory and confirm `DEVMAN_HOME` refreshes on
  startup.

## Risks and Mitigations

### PATH and environment mutation

Risk: DevMan could break global commands.

Mitigation: snapshot before mutation, rollback on failure, show conflict state,
and verify after switching.

### Tool-specific companion commands

Risk: switching Node.js without npm/npx, or Flutter without dart, creates
partial toolchains.

Mitigation: shim companion commands together per tool.

### Existing version managers

Risk: DevMan competing with nvm, fnm, asdf, gvm, or fvm creates confusing PATH
state.

Mitigation: detect and avoid takeover in the first release.

### Official source drift

Risk: metadata formats can change.

Mitigation: isolate providers, test with fixtures, and surface provider failures
without blocking local version switching.

### Immediate effect expectations

Risk: already-open terminals may keep old environment variables.

Mitigation: document and display that new terminals use the updated global
entry. DevMan verification runs in a fresh process environment.

## Success Criteria

1. The Versions page clearly shows local versions, official versions, current
   default, active command path, and conflicts.
2. DevMan can switch between supported local versions through shims and verify
   the result.
3. DevMan can fetch official versions for supported tools from official sources.
4. DevMan can install official versions into intelligently inferred directories.
5. DevMan can uninstall DevMan-installed versions safely.
6. DevMan does not take over tools owned by detected external version managers.
