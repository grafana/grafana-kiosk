# Add --incognito Flag Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a configurable `--incognito` flag (default `true`) so users can disable Chrome incognito mode.

**Architecture:** Add a boolean field to the `General` config struct, wire it through CLI args and the flag-to-config override map, and pass the config value to `chromedp.Flag("incognito", ...)` instead of the hardcoded `true`.

**Tech Stack:** Go, chromedp, goconvey (testing)

**Spec:** `docs/superpowers/specs/2026-04-11-incognito-flag-design.md`

**Issue:** [#127](https://github.com/grafana/grafana-kiosk/issues/127)

---

### Task 1: Add Incognito field to General config struct

**Files:**
- Modify: `pkg/kiosk/config.go:9-24` (General struct)

- [ ] **Step 1: Add the Incognito field**

In `pkg/kiosk/config.go`, add `Incognito` to the `General` struct. Insert it alphabetically between `HideVariables` and `LXDEEnabled`:

```go
Incognito       bool   `yaml:"incognito" env:"KIOSK_INCOGNITO" env-default:"true" env-description:"use incognito mode"`
```

The full line at insertion point (after line 12, before line 13):

```go
HideTimePicker  bool   `yaml:"hide-time-picker" env:"KIOSK_HIDE_TIME_PICKER" env-default:"false" env-description:"Hide time picker in the top nav bar"`
Incognito       bool   `yaml:"incognito" env:"KIOSK_INCOGNITO" env-default:"true" env-description:"use incognito mode"`
LXDEEnabled     bool   `yaml:"lxde" env:"KIOSK_LXDE_ENABLED" env-default:"false" env-description:"initialize LXDE for kiosk mode"`
```

- [ ] **Step 2: Verify it compiles**

Run: `CGO_ENABLED=0 go build ./...`
Expected: Clean build, no errors.

- [ ] **Step 3: Commit**

```bash
git add pkg/kiosk/config.go
git commit -m "feat: add Incognito field to General config struct (#127)"
```

---

### Task 2: Wire CLI flag and override map in main.go

**Files:**
- Modify: `pkg/cmd/grafana-kiosk/main.go:33-62` (Args struct)
- Modify: `pkg/cmd/grafana-kiosk/main.go:64-114` (ProcessArgs function)
- Modify: `pkg/cmd/grafana-kiosk/main.go:144-166` (summary function)
- Modify: `pkg/cmd/grafana-kiosk/main.go:200-233` (update map)

- [ ] **Step 1: Add Incognito to Args struct**

In `pkg/cmd/grafana-kiosk/main.go`, add `Incognito` field to the `Args` struct. Insert alphabetically between `HideVariables` and `IgnoreCertificateErrors` (after line 61, before line 35):

```go
type Args struct {
	AutoFit                              bool
	HideLinks                            bool
	HideTimePicker                       bool
	HideVariables                        bool
	Incognito                            bool
	IgnoreCertificateErrors              bool
```

- [ ] **Step 2: Register the CLI flag**

In the `ProcessArgs` function, add after the `--ignore-certificate-errors` flag registration (after line 87):

```go
flagSettings.BoolVar(&processedArgs.Incognito, "incognito", true, "Use incognito mode")
```

- [ ] **Step 3: Add to the flag-to-config override map**

In the `update` map inside `main()`, add in the General section (after the `"hide-variables"` entry, around line 220):

```go
"incognito":         func() { cfg.General.Incognito = args.Incognito },
```

- [ ] **Step 4: Add debug log line to summary function**

In the `summary` function, add after the `Mode` log line (after line 149):

```go
log.Println("Incognito:", cfg.General.Incognito)
```

- [ ] **Step 5: Verify it compiles**

Run: `CGO_ENABLED=0 go build ./...`
Expected: Clean build, no errors.

- [ ] **Step 6: Commit**

```bash
git add pkg/cmd/grafana-kiosk/main.go
git commit -m "feat: wire --incognito CLI flag and config override (#127)"
```

---

### Task 3: Use config value in browser flags (TDD)

**Files:**
- Modify: `pkg/kiosk/utils.go:86`
- Modify: `pkg/kiosk/utils_test.go`

- [ ] **Step 1: Write failing test for incognito=false**

In `pkg/kiosk/utils_test.go`, add a new `Convey` block inside the top-level `Convey("Given executor option generation", ...)` block, after the last existing test (after line 438, before the closing `})` on line 439):

```go
		Convey("When incognito is disabled", func() {
			cfg := &Config{
				BuildInfo: BuildInfo{Version: "v1.0.0"},
				General: General{
					Incognito:      false,
					WindowPosition: "0,0",
				},
			}
			opts := generateExecutorOptions("/tmp/test", cfg)
			flags := applyOptions(opts)

			Convey("Should set incognito to false", func() {
				So(flags["incognito"], ShouldEqual, false)
			})
		})
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `CGO_ENABLED=0 go test ./pkg/kiosk/ -v -run "TestGenerateExecutorOptions/When_incognito_is_disabled"`
Expected: FAIL — incognito is hardcoded to `true`, but test expects `false`.

- [ ] **Step 3: Update utils.go to use config value**

In `pkg/kiosk/utils.go`, change line 86 from:

```go
chromedp.Flag("incognito", true),
```

To:

```go
chromedp.Flag("incognito", cfg.General.Incognito),
```

- [ ] **Step 4: Run the new test to verify it passes**

Run: `CGO_ENABLED=0 go test ./pkg/kiosk/ -v -run "TestGenerateExecutorOptions/When_incognito_is_disabled"`
Expected: PASS

- [ ] **Step 5: Update existing test to set Incognito explicitly**

The existing default config test at line 102-145 asserts `So(flags["incognito"], ShouldEqual, true)` on line 126. The `General` struct in that test does not set `Incognito`, so it defaults to Go's zero value (`false`), which will now break.

Update the config in the `"When using default config with GPU disabled"` test block (line 103-111) to include `Incognito: true`:

```go
			Convey("When using default config with GPU disabled", func() {
				cfg := &Config{
					BuildInfo: BuildInfo{Version: "v1.0.0"},
					General: General{
						Incognito:      true,
						WindowPosition: "0,0",
					},
					Target: Target{
						IgnoreCertificateErrors: false,
					},
				}
```

- [ ] **Step 6: Run full test suite**

Run: `CGO_ENABLED=0 go test ./... -v`
Expected: All tests PASS.

- [ ] **Step 7: Commit**

```bash
git add pkg/kiosk/utils.go pkg/kiosk/utils_test.go
git commit -m "feat: make incognito mode configurable via config (#127)"
```

---

### Task 4: Update documentation

**Files:**
- Modify: `README.md` (usage section, config YAML example, environment variables)
- Modify: `CHANGELOG.md`

- [ ] **Step 1: Add --incognito to CLI usage block**

In `README.md`, the CLI flags are listed alphabetically in the code block starting around line 67. Add after the `-ignore-certificate-errors` line (after line 90):

```text
  -incognito
      Use incognito mode (default true)
```

- [ ] **Step 2: Add incognito to YAML config example**

In `README.md`, the YAML example starts around line 136. Add `incognito: true` after the `lxde-home` line (after line 140):

```yaml
general:
  kiosk-mode: full
  autofit: true
  incognito: true
  lxde: true
  lxde-home: /home/pi
  scale-factor: 1.0
```

- [ ] **Step 3: Add KIOSK_INCOGNITO to environment variables block**

In `README.md`, the env vars are listed alphabetically starting around line 157. Add after `KIOSK_IGNORE_CERTIFICATE_ERRORS` (after line 193):

```text
  KIOSK_INCOGNITO bool
      use incognito mode (default "true")
```

- [ ] **Step 4: Update CHANGELOG.md**

Add entry under `## [Unreleased]`:

```markdown
- Add `--incognito` flag to optionally disable Chrome incognito mode
  ([#127](https://github.com/grafana/grafana-kiosk/issues/127))
```

- [ ] **Step 5: Run markdownlint**

Run: `npx markdownlint-cli2 README.md CHANGELOG.md`
Expected: No errors (or only pre-existing ones).

- [ ] **Step 6: Commit**

```bash
git add README.md CHANGELOG.md
git commit -m "docs: add --incognito flag to README and CHANGELOG (#127)"
```

---

### Task 5: Final verification

- [ ] **Step 1: Run full test suite**

Run: `CGO_ENABLED=0 go test ./... -v`
Expected: All tests PASS.

- [ ] **Step 2: Run linters**

Run: `CGO_ENABLED=0 golangci-lint run ./...`
Expected: No new issues.

- [ ] **Step 3: Build all targets**

Run: `go run mage.go build:ci`
Expected: Clean build for all platforms.

- [ ] **Step 4: Verify flag appears in help output**

Run: `CGO_ENABLED=0 go run ./pkg/cmd/grafana-kiosk/ -h 2>&1 | grep incognito`
Expected: Shows `-incognito` with description and default `true`.
