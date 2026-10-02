# GoRogue verification map

This map covers the terminal CLI's user-facing gameplay paths. Read the matching feature file before driving. The SDL2 GUI is separate and is not covered by the CLI instructions.

## Baseline preconditions

- Build `/tmp/gorogue-verify-cli` from the current checkout using `../SKILL.md`.
- Use a fixed seed and save each run's input and transcript under `artifacts/verify-gorogue/<RUN_ID>/` so concurrent runs do not collide.
- Each CLI run is a fresh, short-lived process. End it with `quit` or EOF.
- Do not run save/load concurrently. All sessions use shared `saves/` by default.

## Driving and proof

- Follow Launch, Doctor, Drive, Evidence, and Cleanup in `../SKILL.md`.
- Record exact input, full transcript, exit code, seed, and source revision.
- Verify each result through user-visible state. For mutations, follow with a read-only command.

## Features

- [Character status and inventory](./character.md)
- [Dungeon movement and exploration](./exploration.md)
- [Items and equipment](./items.md)
- [Combat and monsters](./combat.md)
