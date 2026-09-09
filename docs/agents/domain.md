# Domain Docs

How the engineering skills should consume this repo's domain documentation.

## Before exploring, read these

- `CONTEXT.md` at the repo root, or `CONTEXT-MAP.md` if it exists.
- `docs/adr/`: read ADRs relevant to the area being changed.

If these files do not exist, proceed silently. The `/domain-modeling` skill creates them lazily when needed.

## File structure

This is a single-context repository:

- `CONTEXT.md` — domain glossary and concepts
- `docs/adr/` — architecture decision records

## Use the glossary's vocabulary

When naming domain concepts in issues, refactor proposals, or tests, use the terminology defined in `CONTEXT.md`.

## Flag ADR conflicts

If output contradicts an existing ADR, surface the conflict explicitly rather than silently overriding it.
