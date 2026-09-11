#!/bin/sh

set -eu

workflow=.github/workflows/quality.yml

test -f "$workflow"
grep -Eq '^  pull_request:$' "$workflow"
grep -Eq '^  push:$' "$workflow"
grep -Eq '^    branches: \[main\]$' "$workflow"
grep -Eq 'go-version-file: go.mod' "$workflow"
grep -Eq 'actions/checkout@v5' "$workflow"
grep -Eq 'actions/setup-go@v6' "$workflow"
grep -Eq 'gofmt -l' "$workflow"
grep -Eq 'go test ./\.\.\.' "$workflow"
grep -Eq 'go vet ./\.\.\.' "$workflow"
grep -Eq 'github.com/golangci/golangci-lint/cmd/golangci-lint@v1\.64\.8' "$workflow"
! grep -Eq 'golangci/golangci-lint-action@' "$workflow"
grep -Eq 'honnef\.co/go/tools/cmd/staticcheck@2025\.1\.1' "$workflow"
! grep -Eq 'make (setup-dev|ci-checks)' "$workflow"
! grep -Eq 'go fmt ./\.\.\.' "$workflow"

echo "quality workflow contract passed"
