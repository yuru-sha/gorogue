#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
sandbox=$(mktemp -d)
fake_lint="$sandbox/golangci-lint"
setup_marker="$sandbox/setup-check"
setup_dev_marker="$sandbox/setup-dev-check"
touch "$setup_marker" "$setup_dev_marker"
trap 'rm -rf "$sandbox"' EXIT

write_version() {
	version=$1
	printf '#!/bin/sh\nprintf "golangci-lint has version %s built with go1.24.5\\n"\n' "$version" >"$fake_lint"
	chmod +x "$fake_lint"
}

write_version 1.64.8
output=$(GOLANGCI_LINT="$fake_lint" SETUP_MARKER="$setup_marker" SETUP_DEV_MARKER="$setup_dev_marker" make -C "$repo_dir" -s setup-dev 2>&1)
test -z "$output"

write_version 2.13.2
if output=$(GOLANGCI_LINT="$fake_lint" SETUP_MARKER="$setup_marker" SETUP_DEV_MARKER="$setup_dev_marker" make -C "$repo_dir" -s setup-dev 2>&1); then
	echo "incompatible golangci-lint version was accepted" >&2
	exit 1
fi
printf '%s\n' "$output" | grep -F 'Unsupported golangci-lint version: 2.13.2 (expected v1.64.8).' >/dev/null
