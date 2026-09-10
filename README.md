# GoRogue

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/yuru-sha/gorogue)

[日本語版 README](README.ja.md)

GoRogue is a Go implementation of the classic Rogue-style dungeon crawler. It is inspired by [PyRogue](https://github.com/yuru-sha/pyrogue) and uses [gruid](https://github.com/anaseto/gruid) with SDL2 for the graphical frontend.

## Current scope

The product requirements and acceptance criteria are defined in [`SPEC.md`](SPEC.md). The current implementation includes:

- A 26-floor, procedurally generated dungeon with seeded generation.
- Turn-based movement and combat with permadeath.
- Weapons, armor, food, potions, scrolls, wands, rings, gold, and the Amulet of Yendor.
- Item identification and equipment management.
- Victory by returning the Amulet of Yendor to the surface.
- JSON save/load with save format version `1.3.0`, including runtime RNG state and identification appearances.
- SDL2 GUI, CLI, and seed-aware game APIs.

Some legacy design documents describe features that are not part of the current scope. See [`docs/architecture.md`](docs/architecture.md) and [`docs/development.md`](docs/development.md) for the current implementation boundaries.

## Requirements

- Go 1.24.5 or later
- `make`
- `pkg-config`
- SDL2 development libraries

On macOS with Homebrew:

```bash
brew install pkg-config sdl2
```

## Build and run

```bash
# Set up Go tools and verify SDL2
make setup-dev

# Build the SDL2 GUI
make build
make run

# Run the CLI with a reproducible seed
go run ./cmd/gorogue-cli --seed 12345
```

The GUI chooses a seed automatically. Passing the same seed and input sequence to the CLI or API reproduces the same generated game state where the current save/runtime boundaries permit.

## Controls

| Key | Action |
| --- | --- |
| `h`, `j`, `k`, `l` | Move west, south, north, east |
| `y`, `u`, `b`, `n` | Move diagonally |
| Arrow keys | Move in four directions |
| `.` or Space | Wait/rest |
| `i` | Open inventory |
| `,` or `g` | Pick up an item |
| `d` | Drop an item |
| `a` or `z` | Use/apply an item |
| `q` | Quaff a potion |
| `r` | Read a scroll |
| `w` / `t` | Wield/wear or take off an item |
| `e` | Eat food |
| `<` / `>` | Use stairs |
| `x` | Look around |
| `?` | Show help |
| `Q` or Escape | Quit or cancel |

Moving into an adjacent monster attacks it. The game also supports doors, searching, traps, wizard mode, and an in-game CLI debug mode.

## CLI examples

```bash
# Interactive mode
go run ./cmd/gorogue-cli --seed 12345

# Batch commands from stdin
printf 'status\nquit\n' | go run ./cmd/gorogue-cli --seed 12345 --interactive=false

# Show CLI options
go run ./cmd/gorogue-cli --help
```

## Development

```bash
go test ./...
go vet ./...
git diff --check
```

The repository-specific development commands and architecture notes are in [`docs/development.md`](docs/development.md). The full design source of truth is [`SPEC.md`](SPEC.md).

## Project layout

```text
cmd/gorogue/          SDL2 GUI entry point
cmd/gorogue-cli/      CLI entry point
internal/core/        Engine, commands, and CLI integration
internal/game/        Actors, dungeon, items, magic, and saves
internal/ui/           Screens and rendering
docs/                  Architecture and development notes
```

## License

MIT
