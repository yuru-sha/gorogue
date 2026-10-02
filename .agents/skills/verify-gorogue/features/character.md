# Character status and inventory

Players inspect character statistics and carried items through named CLI commands.

## Sub-features

- `status`: level, HP, attributes, hunger, position, and inventory count.
- `character`: level, experience, strength, HP, gold, and food.
- `inventory`: carried items.

## How to get to it (user POV)

- At `gorogue>`, enter `status`, `character`, or `inventory`.

## Driving it with the CLI transcript harness

Preconditions:

- Follow `../SKILL.md` and `features/README.md`.
- Start a fresh process with a fixed seed.

- **Status.** Enter `status`; require `Player Status:`, `Level:`, `HP:`, `Position:`, and `Inventory:`.
- **Character.** Enter `character`; require level, experience, strength, HP, gold, and food values.
- **Inventory.** Enter `inventory`; an empty fresh pack prints `Your pack is empty.`
- **Proof.** Capture the input, full transcript, and exit code. Report only the commands exercised.

## Gotchas

- Position depends on seed and build; do not hard-code it.
- Use bare `inventory` to inspect. Arguments can mutate inventory.
