# Dungeon movement and exploration

Players navigate the dungeon, search for hidden features, inspect map information, and use stairs through named commands.

## Sub-features

- `move`: move in a direction.
- `search` and `findtrap`: search nearby or inspect a direction.
- `map`: reveal or inspect map information.
- `stairs`: change floors at a stair.

## How to get to it (user POV)

- At `gorogue>`, enter `move <direction>`, `search`, `findtrap <direction>`, `map info`, or `stairs up`/`stairs down`.

## Driving it with the CLI transcript harness

Preconditions:

- Follow `../SKILL.md` and `features/README.md`.
- Start a fresh process with a fixed seed and record the exact input.

- **Move.** Enter `move <valid-direction>`, then `status`; verify position changed. If blocked, choose an open direction.
- **Search.** Enter `search`; record whether a discovery occurred.
- **Map.** Enter `map info`; confirm map information appears.
- **Stairs.** Stand at a stair, enter `stairs down` or `stairs up`, then verify the floor change with `status` or map output. Do not use teleport as evidence of natural stair discovery.
- **Proof.** Capture the full user command sequence, transcript, and exit code.

## Gotchas

- Movement may be blocked by walls; confirm the position.
- Stairs work only at the matching stair.
- `map reveal` changes explored information and does not prove normal exploration.
