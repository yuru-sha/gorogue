#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
sandbox=$(mktemp -d)
fake_lint="$sandbox/golangci-lint"
fake_staticcheck="$sandbox/staticcheck"
fake_goimports="$sandbox/goimports"
real_go=$(command -v go)
real_modcache=$(go env GOMODCACHE)
fake_go="$sandbox/go"
setup_marker="$sandbox/setup-check"
setup_dev_marker="$sandbox/setup-dev-check"
trap 'rm -rf "$sandbox"' EXIT

write_version() {
	version=$1
	printf '#!/bin/sh\nprintf "golangci-lint has version %s built with go1.24.5\\n"\n' "$version" >"$fake_lint"
	chmod +x "$fake_lint"
}

write_staticcheck_version() {
	version=$1
	printf '#!/bin/sh\nprintf "staticcheck %s (0.6.1)\\n"\n' "$version" >"$fake_staticcheck"
	chmod +x "$fake_staticcheck"
}

printf '#!/bin/sh\ncase "$1 $2" in\n  "mod download"|"mod tidy"|"get "*) exit 0 ;;\nesac\nexec %s "$@"\n' "$real_go" >"$fake_go"
printf '#!/bin/sh\nexit 0\n' >"$fake_staticcheck"
cp "$fake_staticcheck" "$fake_goimports"
chmod +x "$fake_go" "$fake_staticcheck" "$fake_goimports"

write_version 1.64.8
write_staticcheck_version 2025.1.1
GOPATH="$sandbox/gopath" GOMODCACHE="$real_modcache" PATH="$sandbox:$PATH" GOLANGCI_LINT="$fake_lint" make -C "$repo_dir" -s SETUP_MARKER="$setup_marker" SETUP_DEV_MARKER="$setup_dev_marker" BUILD_DIR="$sandbox/bin" LOG_DIR="$sandbox/logs" STATICCHECK="$fake_staticcheck" GOIMPORTS="$fake_goimports" setup-dev
test -f "$setup_marker" && test -f "$setup_dev_marker"

write_staticcheck_version 2024.1.1
if output=$(GOPATH="$sandbox/gopath" GOMODCACHE="$real_modcache" PATH="$sandbox:$PATH" GOLANGCI_LINT="$fake_lint" make -C "$repo_dir" -s SETUP_MARKER="$setup_marker" SETUP_DEV_MARKER="$setup_dev_marker" BUILD_DIR="$sandbox/bin" LOG_DIR="$sandbox/logs" STATICCHECK="$fake_staticcheck" GOIMPORTS="$fake_goimports" setup-dev 2>&1); then
	echo "incompatible staticcheck version was accepted" >&2
	exit 1
fi
printf '%s\n' "$output" | grep -F 'Unsupported staticcheck version: 2024.1.1 (expected 2025.1.1).' >/dev/null

write_staticcheck_version 2025.1.1

write_version 2.13.2
mv "$setup_dev_marker" "$sandbox/setup-dev-check.saved"
if output=$(GOPATH="$sandbox/gopath" GOMODCACHE="$real_modcache" PATH="$sandbox:$PATH" GOLANGCI_LINT="$fake_lint" make -C "$repo_dir" -s SETUP_MARKER="$setup_marker" SETUP_DEV_MARKER="$setup_dev_marker" BUILD_DIR="$sandbox/bin" LOG_DIR="$sandbox/logs" STATICCHECK="$fake_staticcheck" GOIMPORTS="$fake_goimports" setup-dev 2>&1); then
	echo "incompatible golangci-lint version was accepted" >&2
	exit 1
fi
printf '%s\n' "$output" | grep -F 'Unsupported golangci-lint version: 2.13.2 (expected v1.64.8).' >/dev/null

write_version 1.64.8
GOPATH="$sandbox/gopath" GOMODCACHE="$real_modcache" PATH="$sandbox:$PATH" GOLANGCI_LINT="$fake_lint" make -C "$repo_dir" -s SETUP_MARKER="$setup_marker" SETUP_DEV_MARKER="$setup_dev_marker" BUILD_DIR="$sandbox/bin" LOG_DIR="$sandbox/logs" STATICCHECK="$fake_staticcheck" GOIMPORTS="$fake_goimports" setup-dev
GOPATH="$sandbox/gopath" GOMODCACHE="$real_modcache" PATH="$sandbox:$PATH" GOLANGCI_LINT="$fake_lint" make -C "$repo_dir" -s SETUP_MARKER="$setup_marker" SETUP_DEV_MARKER="$setup_dev_marker" BUILD_DIR="$sandbox/bin" LOG_DIR="$sandbox/logs" STATICCHECK="$fake_staticcheck" GOIMPORTS="$fake_goimports" ci-checks >/dev/null
