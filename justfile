# List available commands.
default:
    @just --list

# Run every check.
check: check-fmt vet test race

# Fail when a file is not gofmt clean.
check-fmt:
    #!/usr/bin/env bash
    set -euo pipefail
    unformatted="$(gofmt -l .)"
    if [ -n "$unformatted" ]; then
        echo "not gofmt clean:" >&2
        echo "$unformatted" >&2
        exit 1
    fi

# Rewrite every Go file into gofmt form.
fmt:
    gofmt -w .

# Run go vet.
vet:
    go vet ./...

# Run all tests.
test:
    go test ./...

# Run the tests under the race detector. The ticker and End share a mutex.
race:
    go test -race ./...

# Print one of every rendered shape, for eyeballing a palette change.
demo:
    go run ./internal/demo
