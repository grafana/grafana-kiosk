# Change Log

## 1.0.9

- Fix for issue [[#159](https://github.com/grafana/grafana-kiosk/issues/159)]
- Fix for issue [[#160](https://github.com/grafana/grafana-kiosk/issues/160)]
- Updates go packages
- Please note: When using a service account token, you may need to increase the delay in your configuration for playlists depending on the device being used (10000 ms for RPi4b appears stable)

## 1.0.8

- Fix for issue [#137](https://github.com/grafana/grafana-kiosk/issues/137) How to get rid of "Choose your search engine" window
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
- Added `--check-for-update-interval=31536000` to default flags sent to chromium to workaround update popup

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
