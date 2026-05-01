# Change Log

All notable changes to this project will be documented in this file.

The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.12] - 2026-04-29

### Features

- Add `-browser` flag (env `KIOSK_BROWSER`, default `chrome`) to choose between Chrome and Microsoft Edge as the
  launched browser
- Add `-browser-path` flag (env `KIOSK_BROWSER_PATH`) to point at an explicit Chromium-based browser executable;
  overrides `-browser`
- Add `--incognito` flag to optionally disable Chrome incognito mode ([#127](https://github.com/grafana/grafana-kiosk/issues/127))
- Add `-hide-logo` flag to hide Powered by Grafana logo ([#240](https://github.com/grafana/grafana-kiosk/issues/240))
- Add `-hide-playlist-nav` flag to hide playlist navigation controls ([#240](https://github.com/grafana/grafana-kiosk/issues/240))
- Update `_dash.hideLinks`, `_dash.hideTimePicker`, `_dash.hideVariables` query param values to match Grafana's native format
- Add hide flags to startup config summary logging with visual section separators
- Refactor `summary()` into `logGeneralSettings`, `logTargetSettings`, `logGoAuthSettings`
- Extract `browser.Browser` interface to decouple login providers from chromedp ([#257](https://github.com/grafana/grafana-kiosk/issues/257))

### Bug Fixes

- Fix `hideLogo` query parameter from `_dash.hideLogo` to `hideLogo` to match Grafana's native format
- Fix Grafana 12+ scenes viewport changes causing kiosk to not autofit panels ([#177](https://github.com/grafana/grafana-kiosk/issues/177))
- Fix `cycleWindowToSize` unconditionally cycling to fullscreen when kiosk mode is `tv` or `disabled` ([#177](https://github.com/grafana/grafana-kiosk/issues/177))
- Fix HTTP URLs blocked by Chromium 130+ HTTPS-First Mode ([#155](https://github.com/grafana/grafana-kiosk/issues/155))
- Fix CLI flags not overriding config file values when using `-c` ([#210](https://github.com/grafana/grafana-kiosk/issues/210))
- Fix API key host prefix matching to prevent auth header leakage to hosts sharing a prefix with the target
- Fix broken badge images: push shields.io JSON endpoint to `badges` branch from CI, use shields.io dynamic badge in README
- Fix query errors with newer Grafana API ([#254](https://github.com/grafana/grafana-kiosk/pull/254))

### Tests

- Add tests for `resolveBrowserExecPath` covering chrome default, custom path override, edge PATH lookup, and unknown browsers
- Add tests for `IsDataSourceQueryRequest` and `IsTargetHostRequest` in apikey login
- Add tests for `sanitize` in main and initialize packages
- Add tests for `GenerateURL` playlist mode
- Add tests for command allowlist in initialize package
- Add tests for `cycleWindowToSize`, `waitForPageLoad`, and `waitForBrowserStartup` utilities
- Add tests for `loadConfig`: malformed YAML, testdata fixtures, env var overrides,
  all CLI flag overrides (37.5% -> 82.5%)
- Add tests for `ProcessArgs` with all CLI flags
- Add unit tests for anonymous login mock browser interaction patterns

### CI/CD

- Update `actions/upload-artifact` from v7 to v7.0.1 in CI workflow
- Add octocov-action to CI for test coverage reporting on PRs
- Add `pull_request` trigger to Build CI workflow
- Save coverage report as CI artifact
- Disable octocov PR comment, use PR body insertion only
- Add octocov coverage badges to README
- Restrict CI workflow permissions: global `permissions: {}` with job-level `contents: write` and `pull-requests: write`
- Add `actionlint` job to CI for GitHub Actions workflow linting
- Fix shellcheck warnings in CI workflow: quote variables, group redirects
- Add `markdownlint.yml` workflow to lint `.md` files on push to `main` and PRs

### Dependencies

- Update Go module dependencies: chromedp/cdproto, magefile/mage v1.17.1, google.golang.org/api v0.275.0,
  golang.org/x/{crypto,net,sys,text}, google.golang.org/grpc v1.80.0, go.opentelemetry.io/otel v1.43.0,
  cloud.google.com/go/auth v0.20.0
- Update Docker base images: golang 1.26.2-alpine, dtcooper/raspberrypi-os latest digest
- Update `actions/setup-go` from v6.3.0 to v6.4.0 in CI workflow ([#236](https://github.com/grafana/grafana-kiosk/pull/236))
- Update `google/osv-scanner-action` to v2.3.5 in CI workflow ([#235](https://github.com/grafana/grafana-kiosk/pull/235))
- Update `softprops/action-gh-release` to v3 in CI workflow ([#259](https://github.com/grafana/grafana-kiosk/pull/259))

### Documentation

- Expand README window-size/kiosk-mode section with Chrome launch flags, CDP cycling,
  and Grafana query parameter tables
- Add 10 example YAML config files for dashboard/playlist, fullscreen/tv/windowed, and multi-monitor positioning
- Fix long lines in README.md for MD013 compliance
- Rewrap AGENTS.md around the [agents-md](https://github.com/TheRealSeanDonahoe/agents-md) behavioral template;
  split project-specific detail into topic files under `.agents/` (build, code-style, readme-policy, ci, release)

### Chores

- Add CLAUDE.md referencing AGENTS.md
- Update AGENTS.md action version table to match current CI pinned versions
- Add MD013 (line length) and MD060 (aligned tables) rules to markdownlint config
- Switch to markdownlint-cli2 with `.markdownlint-cli2.yaml` config
- Add missing technical terms to cspell config
- Replace `CLAUDE.md` include stub with symlink to `AGENTS.md`; add `GEMINI.md` symlink for cross-tool portability
- Ignore `CLAUDE.md` and `GEMINI.md` symlinks in markdownlint config to avoid duplicate linting

## 1.0.11

- Add AzureAD authentication support for kiosk mode ([#211](https://github.com/grafana/grafana-kiosk/pull/211))
- Fix gosec security violations: add command allowlist in LXDE
  init, sanitize log inputs
  ([#219](https://github.com/grafana/grafana-kiosk/pull/219))
- Update all Go module dependencies to latest versions ([#214](https://github.com/grafana/grafana-kiosk/pull/214))
- Disable kiosk/fullscreen when explicit window size is requested ([#194](https://github.com/grafana/grafana-kiosk/pull/194))
- Fix environment variables not working with command-line arguments ([#199](https://github.com/grafana/grafana-kiosk/pull/199))
- Fix temp dir cleanup not running when process receives a signal ([#212](https://github.com/grafana/grafana-kiosk/pull/212))
- Add Content-Type header to host requests ([#208](https://github.com/grafana/grafana-kiosk/pull/208))
- Upgrade Go version to 1.26.1
- Add AGENTS.md with build, test, lint, and code style guidelines
- Add branching policy and contribution guidelines to AGENTS.md ([#231](https://github.com/grafana/grafana-kiosk/pull/231))
- Update chromedp to v0.15.1 to fix build compatibility ([#231](https://github.com/grafana/grafana-kiosk/pull/231))
- Replace deprecated `idtoken.WithCredentialsFile` with
  `option.WithAuthCredentialsFile`
  ([#231](https://github.com/grafana/grafana-kiosk/pull/231))
- Update mage to v1.17.0 and google.golang.org/api to v0.273.0
  ([#232](https://github.com/grafana/grafana-kiosk/pull/232))
- Fix `KIOSK_LOGIN_METHOD` env-description to include `aws`
  ([#232](https://github.com/grafana/grafana-kiosk/pull/232))
- Update README usage documentation (CLI flags and environment
  variables) from actual `--help` output
  ([#232](https://github.com/grafana/grafana-kiosk/pull/232))
- Add README update procedure to AGENTS.md
  ([#232](https://github.com/grafana/grafana-kiosk/pull/232))
- Add changelog policy to AGENTS.md
  ([#232](https://github.com/grafana/grafana-kiosk/pull/232))
- Update `actions/setup-go` from v6 to v6.3.0 in CI workflow
  ([#232](https://github.com/grafana/grafana-kiosk/pull/232))
- Add GitHub Actions update procedure to AGENTS.md
  ([#232](https://github.com/grafana/grafana-kiosk/pull/232))
- Add `workflow_dispatch` trigger to Build CI workflow
- Add release process documentation to AGENTS.md
- Add explicit markdownlint requirement for AGENTS.md to formatting guidelines
- Add tests for `generateExecutorOptions` in utils.go
  ([#234](https://github.com/grafana/grafana-kiosk/pull/234))
- Add README review policy to AGENTS.md
  ([#234](https://github.com/grafana/grafana-kiosk/pull/234))
- Fix apikey login for Grafana Cloud: remove stale WaitVisible
  and incorrect Content-Type
  ([#222](https://github.com/grafana/grafana-kiosk/pull/222))

## 1.0.10

- Fix for issue [[[#187](https://github.com/grafana/grafana-kiosk/issues/187)]]
- Updates go packages

## 1.0.9

- Fix for issue [[#159](https://github.com/grafana/grafana-kiosk/issues/159)]
- Fix for issue [[#160](https://github.com/grafana/grafana-kiosk/issues/160)]
- Updates go packages
- Please note: When using a service account token, you may need
  to increase the delay in your configuration for playlists
  depending on the device being used
  (10000 ms for RPi4b appears stable)

## 1.0.8

- Fix for issue
  [#137](https://github.com/grafana/grafana-kiosk/issues/137)
  How to get rid of "Choose your search engine" window
- Fix for scale-factor parameter [#142](https://github.com/grafana/grafana-kiosk/pull/142)

## 1.0.7

- Fix for GCOM login Issue [#132](https://github.com/grafana/grafana-kiosk/issues/132)
- Update go modules

## 1.0.6

- Includes PR#93 - adds authentication with API Key!
- Reduces app CPU utilization to near zero while running
- Adds version to build based on git tag
- Adds user agent with kiosk version
- Switch to git workflow for builds
- Switch to Mage for building instead of make
- Update go modules

## 1.0.5

- Update go modules to fix continuous error messages
- Updated linters and circleci config for go 1.19
- Adds support for Google IAP Auth (idtoken)
- Fixes GCOM auth login hanging

## 1.0.4

- Fix startup issue with new flags

## 1.0.3

- OAuth merged
- Fix Grafana Cloud login
- Updated modules
- Added "window-position" option, allows running kiosk on different displays
- Added `--check-for-update-interval=31536000` to default flags
  sent to chromium to workaround update popup

## 1.0.2

- Also compatible with Grafana v7
- New flag to ignore SSL certificates for local login
- Updated chromedp and build with go 1.14.2
- New configuration file based startup

## 1.0.1

- Automated build
- Includes PR #15
- Compatible with Grafana v6.4.1+

## 1.0.0

- First Release
- Compatible with Grafana v6.3
