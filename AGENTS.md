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
- Always run `npx markdownlint-cli2 <file>` when updating `.md` files and fix
  any issues before committing. This includes `AGENTS.md`, `README.md`, and
  `CHANGELOG.md`.

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

## README Review Policy

**When any code changes are made, check if `README.md` needs to be updated.**
This includes changes to CLI flags, environment variables, configuration
options, default values, login methods, build targets, or any user-facing
behavior. If the change does not affect user-facing behavior (e.g., tests,
internal refactoring, CI-only changes), no README update is needed.

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

6. **Lint the README** — Run `npx markdownlint-cli2 README.md` and fix any
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
| `actions/checkout` | v6.0.2 |
| `actions/setup-go` | v6.4.0 (cache disabled) |
| `golangci/golangci-lint-action` | v9.2.0 |
| `securego/gosec` | v2.25.0 |
| `magefile/mage-action` | v4.0.0 |
| `jwalton/gh-find-current-pr` | v1.3.5 |
| `actions/upload-artifact` | v7.0.1 |
| `softprops/action-gh-release` | v2.6.1 |
| `actions/stale` | v10.2.0 |
| `google/osv-scanner-action` | v2.3.5 |

When updating actions, always pin to full commit SHA with a version comment:

```yaml
uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6.0.2
```

### Checking for Action Updates

Follow this procedure to check for and apply GitHub Actions updates:

1. **List all actions** — Extract every `uses:` line from the workflow files
   in `.github/workflows/`.

2. **Check latest releases** — For each action, run:

   ```sh
   gh api repos/<owner>/<repo>/releases/latest --jq '.tag_name'
   ```

3. **Compare SHAs** — If a newer version exists, get its commit SHA:

   ```sh
   gh api repos/<owner>/<repo>/git/ref/tags/<tag> --jq '.object.sha'
   ```

   Compare against the SHA currently pinned in the workflow file.

4. **Update the workflow file** — Replace the old SHA and version comment
   with the new SHA and version tag. Always use the full 40-character
   commit SHA, never a tag reference.

5. **Update the version table** — Update the action version table in
   this file (AGENTS.md) to reflect the new version.

6. **Update the changelog** — Add an entry to `CHANGELOG.md`.

## Changelog Policy

**Always update `CHANGELOG.md` when making changes.** Every commit that
modifies code, documentation, dependencies, or configuration must have a
corresponding entry in the changelog under the current unreleased version
section. Add entries as part of the same commit or as a follow-up commit
before pushing.

## Release Process

This project uses [semantic versioning](https://semver.org/) with tags in the
form `vX.X.X`. The CI workflow in `.github/workflows/ci.yml` handles building,
packaging, and publishing releases automatically when a version tag is pushed.

### Determining the New Version

1. **Get the current version** from git tags:

   ```sh
   git tag --list 'v*' --sort=-v:refname | head -1
   ```

2. **Choose the version bump** — the default is patch:

   | Bump | When to use | Example |
   | --- | --- | --- |
   | Patch | Bug fixes, dependency updates, docs | v1.0.10 → v1.0.11 |
   | Minor | New features, non-breaking changes | v1.0.10 → v1.1.0 |
   | Major | Breaking changes to CLI, config, login | v1.0.10 → v2.0.0 |

### Pre-release Checklist

1. **Switch to `main` and pull latest changes**:

   ```sh
   git checkout main
   git pull origin main
   ```

2. **Merge any feature branches** that should be included in the release.

3. **Verify the changelog** — `CHANGELOG.md` must have a section header
   matching the new version number (without the `v` prefix). Cross-reference
   all entries against the commits since the last tag:

   ```sh
   git log <last-tag>..HEAD --oneline
   ```

   Every user-facing change must have a corresponding changelog entry.

4. **Run the full verification suite**:

   ```sh
   mage -v                                       # build
   mage -v test:default                           # tests
   golangci-lint run --timeout 5m ./pkg/...       # lint
   gosec ./...                                    # security scan
   ```

5. **Lint the changelog**:

   ```sh
   npx markdownlint-cli2 CHANGELOG.md
   ```

### Tagging and Pushing

Create the tag on `main` and push both the branch and the tag:

```sh
git tag vX.X.X
git push origin main
git push origin vX.X.X
```

### What CI Does Automatically

Pushing a `v*` tag triggers the CI workflow which:

1. Runs lint, security scan, and tests.
2. Builds all 9 OS/arch binaries via `mage -v build:ci`.
3. Packages binaries into a flat directory with platform-specific names
   (e.g., `grafana-kiosk.linux.amd64`, `grafana-kiosk.darwin.arm64`).
4. Creates `.zip` and `.tar.gz` archives.
5. Publishes a GitHub **prerelease** with auto-generated release notes
   and the archives attached.

### Post-release

1. **Verify the release** — Go to the
   [GitHub Releases](https://github.com/grafana/grafana-kiosk/releases) page
   and confirm the artifacts are attached and downloadable.
2. **Promote the release** — Edit the release on GitHub and uncheck
   "Set as a pre-release" to make it a full release.

## Branching Policy

- **Never commit directly to `main`**. Always create a new branch for changes.
- Use descriptive branch names (e.g., `feat/add-feature`, `fix/bug-description`).
- When pushing new commits to a PR, always update the PR summary to reflect all
  changes. Categorize by type with Bug fixes listed first (e.g., Bug fixes,
  Dependencies, Tests, Documentation & tooling).
- **Do not commit automatically**. Only commit when explicitly asked.
- **Do not push automatically**. Only push when explicitly asked.

## Environment Notes

- `CGO_ENABLED=0` is set for all builds and tests.
- Cross-compilation targets: darwin (amd64, arm64), linux (386, amd64, arm64,
  armv5, armv6, armv7), windows (amd64).
- Version is derived from `git describe --tags` and injected via `-ldflags`.
