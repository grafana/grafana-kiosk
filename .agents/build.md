# Build, Test, Lint — grafana-kiosk

All builds use [Mage](https://magefile.org/). Do NOT use `go build` directly.
`CGO_ENABLED=0` is set for all builds and tests.

## Build Commands

```sh
# Default build — clean + build local arch binary to bin/<os>_<arch>/grafana-kiosk
mage -v

# Explicit local build
mage -v build:local

# Full CI pipeline — lint + format + test + build all architectures
mage -v build:ci

# Build all 9 cross-compilation targets
mage -v build:all

# Format source code with gofmt
mage -v build:format

# Run linter (requires golangci-lint installed)
mage -v build:lint

# Clean build artifacts
mage -v clean
```

## Test Commands

```sh
# Run all tests (with coverage)
mage -v test:default

# Run all tests in verbose mode (with coverage)
mage -v test:verbose

# Run a single test by name using go test directly
CGO_ENABLED=0 go test -v -run TestGenerateURL ./pkg/kiosk/...

# Run a single test file's tests
CGO_ENABLED=0 go test -v ./pkg/kiosk/ -run 'TestGenerateURL'

# Run all tests in a specific package
CGO_ENABLED=0 go test -v ./pkg/initialize/...
```

Tests produce `coverage.out` and `coverage.html` in the project root.

## Lint and Security

```sh
# Lint (also available via mage -v build:lint)
golangci-lint run --timeout 5m ./pkg/...

# Security scanning (used in CI)
gosec ./...
```

There is no `.golangci.yml` config file. CI runs golangci-lint v2 and gosec
with default settings.

**Always run** `golangci-lint run --timeout 5m ./pkg/...` and `gosec ./...`
before committing when `.go` files are modified or added. Fix any issues
before creating the commit. CI will reject code with lint or security
violations.

**If `go.mod` or `go.sum` changes**, always run `mage -v` to verify the
project builds successfully before committing.

### Upgrading the Go version

`golangci-lint` and `gosec` analyze Go source, so each must be built with a
Go at least as new as the `go` directive in `go.mod`. A binary built with an
older toolchain refuses to run:

```text
can't load config: the Go language version (go1.26) used to build
golangci-lint is lower than the targeted Go version (1.27.0)
```

After bumping `go` in `mise.toml` and `go.mod`, rebuild both analyzers.
`mise` does not rebuild pinned tools on a Go upgrade:

```sh
mise install --force
```

`mage` is unaffected — it shells out to the `go` on `PATH` rather than
type-checking source, so an older `mage` binary keeps working.

## Cross-compilation Targets

darwin (amd64, arm64), linux (386, amd64, arm64, armv5, armv6, armv7),
windows (amd64). Version is derived from `git describe --tags` and injected
via `-ldflags`.
