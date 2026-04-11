# Change Log

All notable changes to this project will be documented in this file.

The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

- Update Go module dependencies: chromedp/cdproto, magefile/mage v1.17.1,
  google.golang.org/api v0.275.0, golang.org/x/{crypto,net,sys,text},
  google.golang.org/grpc v1.80.0, go.opentelemetry.io/otel v1.43.0,
  cloud.google.com/go/auth v0.20.0
- Add CLAUDE.md referencing AGENTS.md
- Update Docker base images: golang 1.26.2-alpine,
  dtcooper/raspberrypi-os latest digest
- Switch to markdownlint-cli2 with `.markdownlint-cli2.yaml` config
- Fix long lines in README.md for markdownlint compliance
- Add missing technical terms to cspell config
- Fix API key host prefix matching to prevent auth header leakage
  to hosts sharing a prefix with the target
- Add tests for `IsDataSourceQueryRequest` and `IsTargetHostRequest`
  in apikey login
- Add tests for `sanitize` in main and initialize packages
- Add tests for `GenerateURL` playlist mode
- Add tests for command allowlist in initialize package
- Update `actions/upload-artifact` from v7 to v7.0.1 in CI workflow
- Update AGENTS.md action version table to match current CI pinned versions
- Add octocov-action to CI for test coverage reporting on PRs
- Add `pull_request` trigger to Build CI workflow
- Save coverage report as CI artifact
- Disable octocov PR comment, use PR body insertion only
- Add octocov coverage badges to README
- Fix broken badge images: push shields.io JSON endpoint to `badges`
  branch from CI, use shields.io dynamic badge in README
- Restrict CI workflow permissions: global `permissions: {}` with
  job-level `contents: write` and `pull-requests: write`
- Add `actionlint` job to CI for GitHub Actions workflow linting
- Fix shellcheck warnings in CI workflow: quote variables, group redirects

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
