# Repository Guidelines

## Project Overview

GoRogue is a Go and gruid implementation of the Rogue 5.4.4 text roguelike. Requirements and acceptance criteria are defined in [`docs/SPEC.md`](docs/SPEC.md); it describes a seeded, 26-floor game with permanent death and victory by returning the Amulet of Yendor to the surface. Read the relevant specification before changing game behavior. When current behavior is unclear, use code and tests as evidence; update the specification first if the requirement itself must change.

## Architecture & Data Flow

GUI and CLI inputs converge on shared command representations and execution. The GUI flow is `core.Engine` → `state.StateManager` → active screen; the screen parses Rogue key inputs and executes through `internal/core/command/`. CLI gameplay commands also use the shared command/session path. Commands invoke game rules in `internal/game/`; results carry state changes, messages, and turn/death/victory outcomes.

Keep the game model independent of presentation: do not put gruid `Grid`/`Cell`, glyph, or color data in game-state types. `internal/ui/screen/` converts game state into prioritized logical display cells and renders text. Do not duplicate game rules in GUI, CLI, or screen code. Use the game-managed random source so identical versions, seeds, and input sequences remain reproducible.

## Key Directories

- `cmd/gorogue/`: SDL2 GUI entry point.
- `cmd/gorogue-cli/`: interactive and batch CLI entry point.
- `internal/core/`: engine and screen-state coordination.
- `internal/core/command/`: shared key parsing, command/session state, and execution.
- `internal/core/cli/`: CLI command interface using shared game commands.
- `internal/game/actor/`: player, monsters, combat, and actor state.
- `internal/game/dungeon/`: seeded dungeon generation, levels, visibility, traps, stairs, and transitions.
- `internal/game/item/`, `inventory/`, `identification/`, `magic/`: items and their lifecycle/effects.
- `internal/game/save/`: JSON save/load, validation, conversion, and recovery.
- `internal/ui/screen/`: screens, display-cell conversion, and text renderer.
- `internal/config/`: `ROGUEOPTS` parsing and runtime options.
- `docs/`: specification, architecture, development, feature, and QA guidance.
- `scripts/`: checks for CI workflow and development-tool version configuration.

## Code and Comment Guidance

コード・テスト・コメントの指針は [`docs/development.md`](docs/development.md#コードテストコメント) を参照してください。コミットログの指針は [`CONTRIBUTING.md`](CONTRIBUTING.md#commit-messages) を参照してください。

## Development Commands

Go version and module dependencies are declared in [`go.mod`](go.mod). Useful commands:

```sh
# CLI (no SDL2 window); fixed seed for reproducible play
rtk go run ./cmd/gorogue-cli --seed 12345

# SDL2 GUI setup and run
rtk make setup-dev
rtk make run

# Checks
rtk go test ./...
rtk go vet ./...
rtk make ci-checks
```

GUI builds/runs require SDL2 development libraries and `pkg-config`; on macOS, `brew install pkg-config sdl2` provides them. `make setup` downloads modules and runs `go mod tidy`; `make setup-dev` installs/checks development tools and SDL2, so avoid setup targets unless needed. `make lint` runs `go fmt ./...` (which may modify Go files), `go vet`, golangci-lint, and staticcheck. `make ci-checks` runs lint and tests. The CI workflow is `.github/workflows/quality.yml`; headless CI does not verify launching the GUI.

Run shell commands through `rtk`. Do not run setup targets unless dependencies or development tools need setup; `make setup` runs `go mod tidy`. Use `rtk make lint` cautiously because its `go fmt ./...` step can modify Go files.

## Repository Operations

- Work on a feature branch; do not edit, commit, or push directly to `main`.
- Preserve existing staged, unstaged, and untracked user changes. Do not reset, stash, or remove them.
- Do not create commits, tags, branches, or pull requests unless explicitly requested. Do not merge, publish releases, rewrite history, or change repository settings unless explicitly requested.
- Keep changes within this repository. External services, credentials, browsers, and physical devices require an explicit request. Never expose secrets or personal data.
- Use surgical edits; do not modify generated files, lockfiles, migrations, vendored code, binaries, coverage artifacts, or untracked files without confirming they are intended to change.

## Code Conventions & Common Patterns

- Follow existing Go package boundaries and `gofmt`; keep changes focused and use standard library/existing dependencies unless a new dependency is justified.
- Keep Rogue rules in `internal/game/` and the shared command path in `internal/core/command/`, not in interface-specific code.
- Maintain seeded random-stream determinism for dungeon generation, placement, combat, and item effects.
- Save format version is defined by `SaveVersion` in `internal/game/save/`. A format change requires coordinated updates to versioning, compatibility checks, conversion, and save/load tests.
- Configuration comes from `ROGUEOPTS` in the environment or project-root `.env`; see `docs/development.md` and `docs/features.md` for syntax and options.
- Do not add adjacent game features beyond `docs/SPEC.md` requirements. Update relevant docs when behavior, commands, configuration, or public APIs change.

## Important Files

- `docs/SPEC.md`: authoritative game requirements and exclusions.
- `docs/architecture.md`: package responsibilities and input/rendering flow.
- `docs/development.md`: setup, commands, options, and save lifecycle.
- `docs/quality_assurance.md`: expected verification and boundaries.
- `Makefile`: build, run, setup, lint, and test targets.
- `go.mod`: Go toolchain version and dependencies.
- `.github/workflows/quality.yml`: CI checks.
- `internal/game/save/save_data.go`: save schema and `SaveVersion`.
- `internal/core/engine.go`, `cmd/gorogue/main.go`, `cmd/gorogue-cli/main.go`: engine and application entrypoints.

## Runtime/Tooling Preferences

Use the Go toolchain from `go.mod`; this is not a Node/Bun project. Use Go modules and the existing Make targets. GUI-specific development needs SDL2 and `pkg-config`; CLI development does not require launching the GUI. `make setup-dev` pins golangci-lint and staticcheck versions in the Makefile and installs goimports. Use the game-owned seeded RNG rather than introducing other random sources into gameplay.

## Testing & QA

Tests use Go's standard `testing` package and live beside implementation as `*_test.go`. Follow local table-driven `t.Run` patterns, direct `t.Fatalf`/`t.Errorf` assertions, `t.TempDir` for filesystem isolation, `t.Setenv` for environment overrides, and explicit seeds for deterministic gameplay tests. Coverage includes gameplay rules, dungeon generation, save/load and recovery, shared command behavior, and screen flow/parity.

Run the relevant package test first, then `go test ./...` and `go vet ./...`; run `make ci-checks` when its tools are available. `make test-coverage` generates coverage artifacts. Check seeded behavior, save/load, win/loss, and shared GUI/CLI command boundaries when affected. Report SDL2/GUI, OS-specific, or external-CI behavior that was not exercised. `scripts/test-quality-workflow.sh` and `scripts/test-golangci-lint-version.sh` validate workflow/tool-version assumptions; consult `docs/cli_test_scenarios.md` for manual CLI scenarios.
