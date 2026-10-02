# Combat and monsters

Players encounter monsters, attack them, and observe combat outcomes and character state through the CLI.

## Sub-features

- `spawn`: create a monster for a combat scenario.
- `attack`: attack by direction or coordinates.
- `status`: inspect player HP and experience after combat.

## How to get to it (user POV)

- At `gorogue>`, use `help spawn` and `help attack` for exact syntax, then enter `spawn` and `attack` commands.
- Use `status` or `character` to inspect player state.

## Driving it with the CLI transcript harness

Preconditions:

- Follow `../SKILL.md` and `features/README.md`.
- Start a fresh fixed-seed process and note a valid monster position/direction.

- **Spawn.** Spawn a monster using the syntax from `help spawn`; confirm it using a user-facing map or look command.
- **Attack.** Use `attack` with its documented arguments; confirm a combat message and resulting monster/player state.
- **Check state.** Run `status`; inspect HP and experience where relevant.
- **Proof.** Capture full input, transcript, and exit code. Do not use `kill` as a substitute for combat.

## Gotchas

- Debug spawning does not prove natural monster generation.
- Misses are possible; distinguish a miss from damage or a kill.
- A command response alone does not prove the target was visible or in range.
