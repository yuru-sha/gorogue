#!/bin/sh

set -eu

workflow=.github/workflows/quality.yml

test -f "$workflow"
rg -q '^  pull_request:$' "$workflow"
rg -q '^  push:$' "$workflow"
rg -q '^    branches: \[main\]$' "$workflow"
rg -q 'go-version-file: go.mod' "$workflow"
rg -q 'gofmt -l' "$workflow"
rg -q 'go test ./\.\.\.' "$workflow"
rg -q 'go vet ./\.\.\.' "$workflow"
rg -q 'golangci/golangci-lint-action@' "$workflow"
rg -q 'version: v1\.64\.8' "$workflow"
rg -q 'honnef\.co/go/tools/cmd/staticcheck@2025\.1\.1' "$workflow"
! rg -q 'make (setup-dev|ci-checks)' "$workflow"
! rg -q 'go fmt ./\.\.\.' "$workflow"

echo "quality workflow contract passed"
