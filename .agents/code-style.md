# Code Style — grafana-kiosk

## Formatting

- Use `gofmt` for all formatting. Run `mage -v build:format` before committing.
- No `.editorconfig` exists; rely on `gofmt` defaults (tabs for indentation).
- Always run `npx markdownlint-cli2 <file>` when updating `.md` files and fix
  any issues before committing. This includes `AGENTS.md`, `README.md`,
  `CHANGELOG.md`, and files under `.agents/`.

## Import Grouping

Three groups separated by blank lines, in this order:

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

## Naming Conventions

- **Exported names**: `PascalCase` — `GenerateURL`, `GrafanaKioskLocal`, `ProcessArgs`
- **Unexported names**: `camelCase` — `generateExecutorOptions`, `listenChromeEvents`
- **Acronyms**: Keep fully capitalized — `URL`, `API`, `ID`, `LXDE`, `GCOM`
  - Exception in existing code: `Oauth` (should be `OAuth` in new code)
- **Receivers**: This codebase uses standalone functions with `*Config`
  parameters rather than methods on types. Follow this pattern.
- **Common short names**: `cfg` for config, `ctx` for context, `err` for errors,
  `dir` for directory paths
- **Struct names**: PascalCase, descriptive. Legacy types use `Legacy` prefix.

## Type Patterns

- Configuration structs use `yaml`, `env`, `env-default`, and `env-description`
  struct tags (via the `cleanenv` library).
- Structs are composed by value, not pointer.
- No interfaces are defined in this codebase unless genuinely needed for
  testing or polymorphism (see `pkg/browser/`).
- No methods on domain types; use standalone functions accepting `*Config`.

## Error Handling

Aggressive error handling style:

- **In `kiosk` package functions**: `panic(err)` on errors. Functions do not
  return errors — they panic. Follow this existing convention.
- **In `main` functions**: `log.Println(...)` followed by `os.Exit(-1)`.
- **Error wrapping**: Use `fmt.Errorf("context: %w", err)` when adding context
  before panicking. Preferred over bare `panic(err)` for new code.
- Do not introduce custom error types or sentinel errors unless there is a
  clear need.

## Logging

- Use the standard library `log` package exclusively (`log.Println`,
  `log.Printf`).
- Do not introduce third-party logging libraries.
- No structured logging. No log levels beyond debug checks:

  ```go
  if cfg.ChromeDPFlags.DebugEnabled {
      log.Printf("debug info: %+v", data)
  }
  ```

## Comments

- Add Godoc comments on all exported types and functions, starting with the
  name being documented.
- Inline comments for non-obvious workflow steps (especially Chrome
  automation).
- No trailing period required on single-sentence Godoc comments (matching
  existing style).
- Use `//nolint:<linter> // <reason>` when suppressing lint warnings.

## Testing

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

## Configuration

- Configuration uses [`cleanenv`](https://github.com/ilyakaznacheev/cleanenv).
- Three-tier precedence: YAML file → environment variables → CLI flags.
- Environment variable naming: `KIOSK_<SECTION>_<FIELD>` (e.g.,
  `KIOSK_TARGET_URL`, `KIOSK_GRAFANA_AUTOFIT`).
- Provide `env-default` values in struct tags for all config fields.
