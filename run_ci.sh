#!/usr/bin/env bash
set -e

echo "--- Running Linters"
if command -v golangci-lint &> /dev/null
then
    golangci-lint run --modules-download-mode=vendor ./...
else
    go vet -mod=vendor ./...
fi

echo "--- Running Tests"
go test -mod=vendor -failfast -v --race -covermode=atomic ./...
