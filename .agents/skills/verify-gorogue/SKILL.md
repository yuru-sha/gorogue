---
name: verify-gorogue
description: Verify GoRogue's terminal CLI user experience and gameplay commands; use after changes to CLI startup, shared commands, game state, or terminal-facing behavior.
---

# Verify GoRogue

## Surface

The primary automatable surface is the interactive terminal CLI in `cmd/gorogue-cli`. The SDL2 GUI is a separate surface and is not covered here. The CLI accepts named gameplay commands and exits with `quit`, `exit`, or `q`. Its default save directory is `saves`; avoid save/load unless persistence is the target.

## Launch

From the repository root, build once:

```sh
rtk go build -o /tmp/gorogue-verify-cli ./cmd/gorogue-cli
```

The CLI is short-lived, with no server or readiness port. Start each drive in a fresh process with a fixed seed and supplied stdin:

```sh
/tmp/gorogue-verify-cli --seed 12345 < /tmp/gorogue-verify-input.txt
```

The banner and `gorogue>` prompt indicate startup. End input with `quit`. The process then exits. Builds use `/tmp`. All runs share the default `saves/` directory, so do not run save/load drives concurrently.

## Doctor

Run this read-only check whenever startup or build identity is uncertain:

```sh
test -x /tmp/gorogue-verify-cli && /tmp/gorogue-verify-cli --help
```

Require GoRogue CLI usage/options output and exit code 0. This checks the executable, not its source freshness. Rebuild after source changes and drive only the executable built from this checkout.

## Drive

Use a new input file for each run. Stable handles are the literal `gorogue>` prompt, command names, and output labels such as `Player Status:`. Example:

```sh
printf 'status\nhelp\nquit\n' > /tmp/gorogue-verify-input.txt
/tmp/gorogue-verify-cli --seed 12345 < /tmp/gorogue-verify-input.txt > /tmp/gorogue-verify-output.txt 2>&1
status=$?
cat /tmp/gorogue-verify-output.txt
printf 'exit=%s\n' "$status"
```

Require the banner, `Player Status:`, `HP:`, `Position:`, `Available commands:`, `Goodbye!`, and exit code 0. Extend the input with exact gameplay commands from the matching feature map. Exercise the real CLI command path; do not substitute internal setters or tests.

## Evidence

Store exact input, full stdout/stderr transcript, exit code, seed, and source revision under `/tmp/gorogue-evidence/<feature-id>/`. Capture the action and its resulting state. For mutations, issue a later read-only command such as `inventory` or `status` to verify the effect. Check persistent side effects separately when testing persistence; command output alone is insufficient. The local CLI path needs no external-service mocks.

## Cleanup

The CLI exits on `quit` or EOF. For an interactive run, send `quit` only to the process/session started by this verification run. Never kill by process name. Remove the temporary binary and input files when no longer needed. Preserve `/tmp/gorogue-evidence/`; do not delete or alter shared `saves/` state.

## Helpers

No scripts are shipped. The shell commands in this skill are the complete recipe.