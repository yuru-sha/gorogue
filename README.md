# GoRogue

[English](README.md) | [日本語](README.ja.md)

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/yuru-sha/gorogue)

GoRogue is a Go and gruid implementation of the Rogue 5.4.4 text roguelike. [`SPEC.md`](SPEC.md) is the source of truth for requirements and acceptance criteria.

## Current game

- Seeded 26-floor dungeons with Rogue rooms, corridors, dark and treasure rooms, maze floors, hidden passages/doors, stairs, and traps.
- Rogue movement, combat, progression, hunger, monsters, equipment, items, effects, and identification.
- Shared command execution for the GUI and CLI.
- Victory requires returning the Amulet of Yendor to the surface; death is permanent.
- JSON save format `1.5.0` stores the state required to resume a game, including per-floor hallucination stair knowledge. Older save formats are not automatically migrated.

## Controls

| Key | Action |
| --- | --- |
| `h j k l y u b n` | Move in eight directions |
| `H J K L Y U B N` | Run in eight directions |
| Arrow keys | Move in four directions |
| `.` | Rest |
| `f` | Fight in a direction |
| `,` / `d` | Pick up / drop an item |
| `e` | Eat food |
| `w` / `W` / `T` | Wield / wear / take off armor |
| `P` / `R` | Put on / remove a ring |
| `t` | Throw an item |
| `q` / `r` / `z` | Quaff / read / zap |
| `s` / `^` | Search nearby / search in a direction |
| `i` / `@` | Inventory / character information |
| `D` / `c` | Discovered items / name an unidentified item |
| `<` / `>` | Go up / down stairs |
| `a` | Repeat the last command |
| `?` / `/` | Help / explain a symbol |
| `Q` / Escape | Quit / cancel |

GoRogue conveniences: `^W` toggles wizard mode and `:` enters the CLI debug prompt.

## Build and run

Requirements: Go version in [`go.mod`](go.mod), `make`, `pkg-config`, and SDL2 development libraries for the GUI. On macOS:

```sh
brew install pkg-config sdl2
```

```sh
# Set up development tools and verify SDL2
make setup-dev

# Build and run the GUI
make build
make run

# Run the CLI with a reproducible seed
go run ./cmd/gorogue-cli --seed 12345
```

The GUI chooses a seed automatically. The CLI and game APIs accept an explicit seed; the same version, seed, and input sequence reproduce the same game random stream.

## Development checks

```sh
go test ./...
go vet ./...
make ci-checks
git diff --check
```

See [`docs/development.md`](docs/development.md) for development commands and [`docs/architecture.md`](docs/architecture.md) for package boundaries and state/rendering flow.

## Project layout

- `cmd/gorogue/`: SDL2 GUI entry point.
- `cmd/gorogue-cli/`: CLI entry point.
- `internal/core/command/`: shared command parsing and execution.
- `internal/core/`, `internal/game/`: game flow and rules.
- `internal/ui/screen/`: screens, logical display-cell conversion, and text rendering.

## License

MIT

See [docs/agents/release.md](docs/agents/release.md) for the release note format and creation procedure. The shared body template is [.github/release-notes-template.md](.github/release-notes-template.md), and the generated-note categories are managed in [.github/release.yml](.github/release.yml).
