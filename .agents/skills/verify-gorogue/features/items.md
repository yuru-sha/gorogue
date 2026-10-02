# Items and equipment

Players collect items, use consumables, and equip weapons, armor, and rings.

## Sub-features

- `create` and `pickup`: create or pick up items.
- `inventory` and `drop`: inspect or change carried items.
- `eat`, `quaff`, `read`, `zap`: use food and magical items.
- `wield`, `wear`, `takeoff`, `ring-on`, `ring-off`: manage equipment.

## How to get to it (user POV)

- At `gorogue>`, use `create <item_type>`, `pickup`, `inventory`, `drop <item_letter>`, or the relevant consumption/equipment command.
- Item-taking commands use the letter shown in inventory as an argument, for example `drop a` or `wield a`.

## Driving it with the CLI transcript harness

Preconditions:

- Follow `../SKILL.md` and `features/README.md`.
- Start a fresh process with seed `12345`.

- **Acquire food.** Enter `create food`, then `pickup`. The first prints `Created food ration at (x, y)`; the second prints `You picked up food ration.`
- **Confirm inventory.** Enter `inventory`; require `Current inventory:` and `a) food ration`.
- **Drop food.** Enter `drop a` using the listed item letter. Require `Dropped food ration.`, then enter `inventory` and require `Your pack is empty.`
- **Other categories.** Use literal creation commands `create weapon`, `create armor`, `create ring`, `create scroll`, `create potion`, `create gold`, or `create amulet`, then `pickup`. Use the letter shown in inventory for commands such as `wield a`, `wear a`, `ring-on a`, `read a`, or `quaff a`. Follow each change with `inventory`, `character`, or `status`.
- **Proof.** Capture exact input, full transcript, and exit code. Creation alone does not prove pickup, use, or equipment.

## Gotchas

- `drop` and item-taking commands require the inventory letter as an argument; there is no selection prompt.
- Food and other randomly named items can vary by seed/build. Confirm the actual displayed name.
- Avoid save/load unless persistence is the target; `saves/` is shared.
