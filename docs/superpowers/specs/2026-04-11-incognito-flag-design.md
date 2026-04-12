# Add --incognito flag to disable incognito mode

**Issue:** [#127](https://github.com/grafana/grafana-kiosk/issues/127)
**Date:** 2026-04-11

## Problem

Incognito mode is hardcoded to `true` in `pkg/kiosk/utils.go:86`. Some users
need to run without incognito to persist browser state (cookies, local storage)
across kiosk restarts.

## Design

Add a configurable `--incognito` flag following the existing pattern used by
`--gpu-enabled`, `--ignore-certificate-errors`, and other boolean flags.

### Config (`pkg/kiosk/config.go`)

Add `Incognito` field to the `General` struct:

```go
Incognito bool `yaml:"incognito" env:"KIOSK_INCOGNITO" env-default:"true" env-description:"use incognito mode"`
```

Default is `true` to preserve backward compatibility.

### CLI flag (`pkg/cmd/grafana-kiosk/main.go`)

- Add `Incognito bool` to the `processedArgs` struct
- Register with `flagSettings.BoolVar(&processedArgs.Incognito, "incognito", true, "Use incognito mode")`
- Add to the flag-to-config override map: `"incognito": func() { cfg.General.Incognito = args.Incognito }`
- Add debug log line: `log.Println("Incognito:", cfg.General.Incognito)`

### Browser flags (`pkg/kiosk/utils.go`)

Change:

```go
chromedp.Flag("incognito", true),
```

To:

```go
chromedp.Flag("incognito", cfg.General.Incognito),
```

### Documentation

Update README.md usage output to include `--incognito` flag.

### Tests (`pkg/kiosk/utils_test.go`)

Update existing test to verify incognito reflects config value. Add a test case
for `Incognito: false` that confirms the flag is set to `false`.

## Scope

Files modified:

- `pkg/kiosk/config.go`
- `pkg/cmd/grafana-kiosk/main.go`
- `pkg/kiosk/utils.go`
- `pkg/kiosk/utils_test.go`
- `README.md`
- `CHANGELOG.md`
