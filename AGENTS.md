# AGENTS.md

Operational guidelines for agents working on GoRogue. Game requirements are defined in [SPEC.md](SPEC.md); refer to the code and related documentation for implementation details.

## Sources of truth

At the start of a task, read the following as relevant:

- Requirements, scope, exclusions, and acceptance criteria: [SPEC.md](SPEC.md)
- Current architecture: [docs/architecture.md](docs/architecture.md)
- Development commands: [docs/development.md](docs/development.md), [Makefile](Makefile), and [go.mod](go.mod)
- Actual behavior: `cmd/`, `internal/`, and `*_test.go`

Treat `SPEC.md` as the source of truth for game requirements. If the documentation under `docs/` differs from the implementation or tests, use the code and tests as the authority for current behavior. If the requirements need to change, update `SPEC.md` and the acceptance criteria before changing the implementation.

## Workflow

1. Check `rtk git status --short --branch` and the diff, preserving existing user changes.
2. Read the relevant sections of `SPEC.md`, callers, and related tests.
3. Reuse existing types, helpers, and dependencies; make only the smallest change required by the specification.
4. After making changes, run the target tests and `rtk go test ./...`. If the development tools are available, also run `rtk make ci-checks`.
5. Check the final diff, worktree, verification results, and remaining risks before reporting completion.

## Safety boundaries

- Keep changes within this repository. Operate external services, GitHub, credentials, browsers, or physical devices only when explicitly requested.
- Do not create commits, push, create branches, open pull requests, merge, or release until explicitly requested.
- Do not perform difficult-to-recover operations such as `rm -rf`, force-pushing, rewriting history, or resetting a database.
- Do not print, commit, or transmit secrets, tokens, credentials, or personal information. `.env` may be used without displaying its values.
- Confirm the intended scope before modifying generated files, binaries, coverage data, lockfiles, or untracked files.
- Run shell commands through `rtk` and use `apply_patch` for file edits.

## Project constraints

- Follow the Go version and existing Go Modules dependencies specified by `go.mod`. Explain the need and alternatives before adding a new dependency or framework.
- Follow the GoRogue requirements and exclusions in `SPEC.md`; do not add adjacent features that were not requested.
- Keep game rules in `internal/core/` and `internal/game/`. Do not duplicate CLI- or GUI-specific rules in `cmd/` or `internal/ui/`; prefer the existing shared command path in `internal/core/command/`.
- Use the game-managed random source for dungeon generation, combat, and placement, preserving reproducibility for the same seed and input sequence.
- When changing the save format, update `SaveVersion`, compatibility checks, conversion logic, and tests in `internal/game/save/` together.
- Follow `gofmt` and the existing Go package structure. Do not break public APIs unless necessary.

## Definition of done

Do not report completion until all of the following are satisfied:

- The diff meets the requirements and exclusions in `SPEC.md`.
- The target tests and, where possible, `rtk go test ./...`, `rtk go vet ./...`, and `rtk make ci-checks` pass.
- The affected seed, save/load, win/loss, and shared CLI/GUI rule boundaries have been checked.
- Related documentation is updated when code, commands, configuration, or public APIs change.
- `rtk git diff --check` passes, with no secrets, unintended changes, or untracked files.
- Any unverified GUI, SDL2, OS-specific, or external CI behavior is explicitly reported.

## Verification entry point

```bash
rtk go test ./...
rtk go vet ./...
rtk make ci-checks  # when the development tools are installed
```

`make ci-checks` runs `make lint` and `make test`. Run `make setup` or `make setup-dev` only when necessary, because they change dependencies, development tools, or the SDL2 environment.

## Git

Do not reorganize, delete, or stash staged, unstaged, or untracked user changes. Add permanent fixes for review comments or verification failures to the smallest appropriate place among the tests, verification commands, and documentation.

## Agent skills

### Issue tracker

Issues live in GitHub Issues; use the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

Use the default five canonical triage labels. See `docs/agents/triage-labels.md`.

### Domain docs

This is a single-context repository. See `docs/agents/domain.md`.
