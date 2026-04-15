# Add hide-logo and hide-playlist-nav config options

Issue: [#240](https://github.com/grafana/grafana-kiosk/issues/240)

## Problem

Grafana 12.4 introduced a "Powered by Grafana" logo in kiosk mode that reserves
vertical space across the full screen width. Grafana now supports query parameters
to hide this logo and playlist navigation controls, but grafana-kiosk has no way
to set them.

Additionally, the existing `_dash.*` query parameters (`hideLinks`, `hideTimePicker`,
`hideVariables`) currently emit empty values (`_dash.hideLinks=`). These should be
updated to use explicit values matching Grafana's native URL format.

## Design

### New config options

| Config field | CLI flag | YAML key | Env var | Default |
|---|---|---|---|---|
| `HideLogo` | `-hide-logo` | `hide-logo` | `KIOSK_HIDE_LOGO` | `false` |
| `HidePlaylistNav` | `-hide-playlist-nav` | `hide-playlist-nav` | `KIOSK_HIDE_PLAYLIST_NAV` | `false` |

Both are boolean fields on the `General` struct. When `false` (default), the
query parameter is omitted entirely.

### Query parameter values

All `_dash.*` parameters follow Grafana's native URL format:

| Config field | Query param | Value when enabled |
|---|---|---|
| `HideLinks` | `_dash.hideLinks` | `"true"` |
| `HideTimePicker` | `_dash.hideTimePicker` | `"true"` |
| `HideVariables` | `_dash.hideVariables` | `"true"` |
| `HideLogo` | `_dash.hideLogo` | `"1"` |
| `HidePlaylistNav` | `_dash.hidePlaylistNav` | `"true"` |

The first three are existing parameters updated from empty values to `"true"`.
`hideLogo` uses `"1"` to match Grafana's native format.

### Changes by file

**`pkg/kiosk/config.go`** -- Add to `General` struct:

```go
HideLogo        bool `yaml:"hide-logo" env:"KIOSK_HIDE_LOGO" env-default:"false"`
HidePlaylistNav bool `yaml:"hide-playlist-nav" env:"KIOSK_HIDE_PLAYLIST_NAV" env-default:"false"`
```

**`pkg/cmd/grafana-kiosk/main.go`** -- Add to `Args` struct, `ProcessArgs` flags,
and `loadConfig` override map:

```go
// Args struct
HideLogo        bool
HidePlaylistNav bool

// ProcessArgs
flagSettings.BoolVar(&processedArgs.HideLogo, "hide-logo", false, "Hide Powered by Grafana logo")
flagSettings.BoolVar(&processedArgs.HidePlaylistNav, "hide-playlist-nav", false, "Hide playlist navigation controls")

// loadConfig map
"hide-logo":         func() { cfg.General.HideLogo = args.HideLogo },
"hide-playlist-nav": func() { cfg.General.HidePlaylistNav = args.HidePlaylistNav },
```

**`pkg/kiosk/utils.go`** -- Update `GenerateURL`:

```go
// Update existing params from empty to explicit values
parsedQuery.Set("_dash.hideLinks", "true")       // was ""
parsedQuery.Set("_dash.hideTimePicker", "true")   // was ""
parsedQuery.Set("_dash.hideVariables", "true")    // was ""

// Add new params
if cfg.General.HideLogo {
    parsedQuery.Set("_dash.hideLogo", "1")
}
if cfg.General.HidePlaylistNav {
    parsedQuery.Set("_dash.hidePlaylistNav", "true")
}
```

### Tests

- Update existing `GenerateURL` tests to expect `"true"` instead of `""`
- Add `GenerateURL` test cases for `hideLogo` and `hidePlaylistNav`
- Update `TestProcessArgsAllFlags` and `TestLoadConfigAllFlagsOverride` with new flags

### Documentation

- Update CHANGELOG.md with new feature entries
- Add example YAML configs or update existing ones if appropriate

## Example usage

YAML:

```yaml
general:
  kiosk-mode: full
  autofit: true
  hide-logo: true
  hide-playlist-nav: true
```

CLI:

```bash
grafana-kiosk -hide-logo -hide-playlist-nav -URL https://grafana.example.com/d/abc/dashboard
```

Resulting URL query string:

```
?kiosk=1&_dash.hideLogo=1&_dash.hidePlaylistNav=true&autofitpanels
```
