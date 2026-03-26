# AGENTS.md — Grafana Kiosk

## Project Overview

Grafana Kiosk is a Go application that renders Grafana dashboards/playlists in a
chromium-based kiosk mode using `chromedp`. It supports multiple login methods
(anonymous, local, OAuth, API key, AWS, ID token) and cross-compiles to 9
OS/arch targets. The build system uses [Mage](https://magefile.org/).

## Repository Layout

```text
pkg/
  cmd/
    grafana-kiosk/    # Main kiosk binary entrypoint
    migrate-config/   # Legacy config migration tool
  initialize/         # LXDE desktop environment setup
  kiosk/              # Core kiosk library (login flows, config, utils)
bin/                  # Build output (gitignored, per-arch subdirectories)
scripts/              # systemd service file for Linux deployment
testdata/             # YAML config fixtures for tests
Magefile.go           # Mage build definitions (build-tagged `//go:build mage`)
```

## Build Commands (Mage)

All builds use Mage. Do NOT use `go build` directly.

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

**IMPORTANT**: Always run `golangci-lint run --timeout 5m ./pkg/...` and
`gosec ./...` before committing and fix any issues. CI will reject code
with lint or security violations.

**IMPORTANT**: If `go.mod` or `go.sum` changes, always run `mage -v` to
verify the project builds successfully before committing.

## Code Style Guidelines

### Formatting

- Use `gofmt` for all formatting. Run `mage -v build:format` before committing.
- No `.editorconfig` exists; rely on `gofmt` defaults (tabs for indentation).
- Always run `npx markdownlint-cli <file>` when updating `.md` files and fix
  any issues before committing.

### Import Grouping

Use three groups separated by blank lines, in this order:

1. Standard library
2. External third-party packages
3. Internal packages (`github.com/grafana/grafana-kiosk/...`)

```go
import (
    "context"
    "fmt"
    "log"

    "github.com/chromedp/chromedp"
    "github.com/ilyakaznacheev/cleanenv"

    "github.com/grafana/grafana-kiosk/pkg/kiosk"
)
```

When only stdlib and one category of external packages are needed, two groups
are acceptable (stdlib, then external).

### Naming Conventions

- **Exported names**: `PascalCase` — `GenerateURL`, `GrafanaKioskLocal`, `ProcessArgs`
- **Unexported names**: `camelCase` — `generateExecutorOptions`, `listenChromeEvents`
- **Acronyms**: Keep fully capitalized — `URL`, `API`, `ID`, `LXDE`, `GCOM`
  - Exception in existing code: `Oauth` (should be `OAuth` in new code)
- **Receivers**: This codebase uses standalone functions with `*Config` parameters
  rather than methods on types. Follow this pattern.
- **Common short names**: `cfg` for config, `ctx` for context, `err` for errors,
  `dir` for directory paths
- **Struct names**: PascalCase, descriptive. Legacy types use `Legacy` prefix.

### Type Patterns

- Configuration structs use `yaml`, `env`, `env-default`, and `env-description`
  struct tags (via the `cleanenv` library).
- Structs are composed by value, not pointer.
- No interfaces are defined in this codebase. Do not introduce interfaces unless
  genuinely needed for testing or polymorphism.
- No methods on domain types; use standalone functions accepting `*Config`.

### Error Handling

This codebase uses an aggressive error handling style:

- **In `kiosk` package functions**: `panic(err)` on errors. Functions do not
  return errors — they panic. Follow this existing convention.
- **In `main` functions**: `log.Println(...)` followed by `os.Exit(-1)`.
- **Error wrapping**: Use `fmt.Errorf("context: %w", err)` when adding context
  before panicking. Only one instance currently exists but this is preferred
  over bare `panic(err)` for new code.
- Do not introduce custom error types or sentinel errors unless there is a
  clear need.

### Logging

- Use the standard library `log` package exclusively (`log.Println`, `log.Printf`).
- Do not introduce third-party logging libraries.
- No structured logging. No log levels beyond debug checks:

  ```go
  if cfg.ChromeDPFlags.DebugEnabled {
      log.Printf("debug info: %+v", data)
  }
  ```

### Comments

- Add Godoc comments on all exported types and functions, starting with the
  name being documented.
- Inline comments for non-obvious workflow steps (especially Chrome automation).
- No trailing period required on single-sentence Godoc comments (matching
  existing style).
- Use `//nolint:<linter> // <reason>` when suppressing lint warnings.

### Testing

- **Framework**: [GoConvey](https://github.com/smartystreets/goconvey) with
  dot-import (`import . "github.com/smartystreets/goconvey/convey"`).
- **Style**: BDD-style nested `Convey` blocks with `So`/`Should*` assertions:

  ```go
  func TestExample(t *testing.T) {
      Convey("Given some precondition", t, func() {
          Convey("When something happens", func() {
              result := DoSomething()
              So(result, ShouldEqual, expected)
          })
      })
  }
  ```

- **Test data**: Place YAML fixtures in the `testdata/` directory at the repo
  root. Reference them via relative paths from the test file.
- Do not use `testify`, table-driven tests, or `t.Run()` subtests — this
  project uses GoConvey exclusively.

### Configuration

- Configuration uses [`cleanenv`](https://github.com/ilyakaznacheev/cleanenv).
- Three-tier precedence: YAML file → environment variables → CLI flags.
- Environment variable naming: `KIOSK_<SECTION>_<FIELD>` (e.g.,
  `KIOSK_TARGET_URL`, `KIOSK_GRAFANA_AUTOFIT`).
- Provide `env-default` values in struct tags for all config fields.

## Updating README Usage Documentation

When CLI flags or environment variables change (e.g., new flags added, defaults
changed, descriptions updated), the README must be updated to match. Follow
this procedure:

1. **Fix source first** — Ensure the `env-description` tags in
   `pkg/kiosk/config.go` and the flag definitions in
   `pkg/cmd/grafana-kiosk/main.go` are accurate and consistent with each other.
   For example, if `-login-method` lists `aws` as an option, the corresponding
   `KIOSK_LOGIN_METHOD` env-description must also list `aws`.

2. **Build the binary** — Run `mage -v` to produce a fresh local binary.

3. **Capture the help output** — Run
   `./bin/<os>_<arch>/grafana-kiosk --help` and capture the full output.

4. **Update the CLI flags section** — Replace the code block under the
   `## Usage` heading in `README.md` with the flags portion of the `--help`
   output (everything from the first flag up to the `Environment variables:`
   line).

5. **Update the environment variables section** — Replace the code block under
   the "Environment variables can be set..." paragraph in `README.md` with the
   environment variables portion of the `--help` output.

6. **Lint the README** — Run `npx markdownlint-cli README.md` and fix any
   issues before committing.

7. **Run tests** — Run `mage -v test:default` to confirm no regressions.

## CI Pipeline

GitHub Actions workflows (`.github/workflows/`):

- **`ci.yml`**: Checkout → Go setup (version from `go.mod`) → `go get .` →
  golangci-lint → gosec → `mage -v build:ci` → coverage upload → release
  packaging on version tags (`v*`).
- **`osv-scanner-pr.yml`**: Vulnerability scanning on PRs to `main`.
- **`stale.yml`**: Auto-closes stale issues/PRs after 90+60 days.

All actions are pinned to commit SHAs with version comments (required by
zizmor). Current versions:

| Action | Version |
| --- | --- |
| `actions/checkout` | v6 |
| `actions/setup-go` | v6 (cache disabled) |
| `golangci/golangci-lint-action` | v9.2.0 |
| `securego/gosec` | v2.24.7 |
| `magefile/mage-action` | v3.1.0 |
| `jwalton/gh-find-current-pr` | v1.3.5 |
| `actions/upload-artifact` | v7 |
| `softprops/action-gh-release` | v2.4.1 |
| `actions/stale` | v10 |
| `google/osv-scanner-action` | v2.3.3 |

When updating actions, always pin to full commit SHA with a version comment:

```yaml
uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6
```

## Branching Policy

- **Never commit directly to `main`**. Always create a new branch for changes.
- Use descriptive branch names (e.g., `feat/add-feature`, `fix/bug-description`).
- When pushing new commits to a PR, always update the PR summary to reflect all
  changes.
- **Do not push automatically**. Only push when explicitly asked.

## Environment Notes

- `CGO_ENABLED=0` is set for all builds and tests.
- Cross-compilation targets: darwin (amd64, arm64), linux (386, amd64, arm64,
  armv5, armv6, armv7), windows (amd64).
- Version is derived from `git describe --tags` and injected via `-ldflags`.
