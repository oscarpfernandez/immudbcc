#!/usr/bin/env bash
set -e

echo "--- Running Linters"
if command -v golangci-lint &> /dev/null
then
    golangci-lint  run ./...
else
    go vet ./...
fi

echo "--- Running Tests"
go test -failfast  --race -covermode=atomic ./...

