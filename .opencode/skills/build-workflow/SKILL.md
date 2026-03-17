---
name: build-workflow
description: Build, test, and lint the grafana-kiosk Go project using mise and mage
---

## Build System

This project uses [Mage](https://magefile.org/) as its build tool. All build
definitions are in `Magefile.go`. Do NOT use `go build` directly.

Before running any build or test command, activate the mise-managed Go toolchain:

```sh
eval "$(mise env)"
```

The project pins its Go version in `mise.toml`. Run `mise install` if the
required version is not yet installed.

## Commands

### Build

```sh
# Default build (clean + build local arch binary)
eval "$(mise env)" && mage -v

# Full CI pipeline (lint + format + test + build all targets)
eval "$(mise env)" && mage -v build:ci

# Build all 9 cross-compilation targets
eval "$(mise env)" && mage -v build:all

# Format source with gofmt
eval "$(mise env)" && mage -v build:format

# Lint (requires golangci-lint)
eval "$(mise env)" && mage -v build:lint

# Clean build artifacts
eval "$(mise env)" && mage -v clean
```

### Test

```sh
# Run all tests with coverage
eval "$(mise env)" && mage -v test:default

# Run all tests verbose with coverage
eval "$(mise env)" && mage -v test:verbose

# Run a single test by name
eval "$(mise env)" && CGO_ENABLED=0 go test -v -run TestGenerateURL ./pkg/kiosk/...

# Run all tests in a specific package
eval "$(mise env)" && CGO_ENABLED=0 go test -v ./pkg/initialize/...
```

### Update Dependencies

```sh
eval "$(mise env)" && go get -u ./... && go mod tidy
```

Then verify with `mage -v`.

## When to Use

Use this skill whenever you need to build, test, lint, or update dependencies
in this repository. Always prefix commands with `eval "$(mise env)"` to ensure
the correct Go toolchain is active.
