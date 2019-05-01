# Grafana Kiosk
[![CircleCI](https://circleci.com/gh/grafana/grafana-kiosk.svg?style=svg)](https://circleci.com/gh/grafana/grafana-kiosk)
[![Go Report Card](https://goreportcard.com/badge/github.com/grafana/grafana-kiosk)](https://goreportcard.com/report/github.com/grafana/grafana-kiosk) [![codecov](https://codecov.io/gh/grafana/grafana-kiosk/branch/master/graph/badge.svg)](https://codecov.io/gh/grafana/grafana-kiosk)

A very useful feature of Grafana is the ability to display dashboards and playlists on a large TV.

This provides a utility to quickly standup a kiosk on devices like a Raspberry Pi or NUC.

The utitilty provides these options:

- login
  - to a Grafana server (local account)
  - to a grafana server with anonymous-mode enabled (same method used on https://play.grafana.org)
  - to a hosted grafana instance
- switch to kiosk or kiosk-tv mode
- display the default home page set for the user
- display a specified dashboard
- start a playlist immediately (inactive mode enable)


## Usage

`--URL`
  - URL to a Grafana server

`--playlist`
  - designates the URL as a playlist, allowing instant "inactive" user vs waiting for the timeout

`--login-method` (default anon)
  - anon (anonymous)
  - local (local user)
  - gcom (Hosted Grafana)

`--username` used with local and gcom login methods

`--password` used with local and gcom login methods

`--kiosk`
  - default  (no sidebar, top navigation disabled)
  - tv (no sidebar, top navigation enabled)
  - false (sidebar and top navigation enabled)

`--autofit` scales panels to fit the display (default is true)
  - true
  - false 

### Hosted Grafana using grafana.com authentication

This will login to a Hosted Grafana instance and take the browser to the default dashboard in fullscreen kiosk mode:

```BASH
./bin/grafana-kiosk --URL https://bkgann3.grafana.net --login-method gcom --user bkgann --password abc123 --kiosk-mode tv
```

This will login to a Hosted Grafana instance and take the browser to a specific dashboard in fullscreen kiosk mode:

```BASH
./bin/grafana-kiosk --URL https://bkgann3.grafana.net/dashboard/db/sensu-summary --login-method gcom --user bkgann --password abc123 --kiosk-mode tv
```

This will login to a Hosted Grafana instance and take the browser to a playlist in fullscreen kiosk mode, and autofit the panels to fill the display.

```BASH
./bin/grafana-kiosk --URL https://bkgann3.grafana.net/playlists/play/1 --login-method gcom --user bkgann --password abc123 --kiosk-mode tv --playlist --autofit
```

### Grafana Server with Local Accounts

This will login to a grafana server that uses local accounts:

```BASH
./bin/grafana-kiosk --URL https://localhost:3000 --login-method local --user admin --password admin --kiosk-mode tv
```

```BASH
./bin/grafana-kiosk --URL https://localhost:3000 --login-method local --user admin --password admin --kiosk-mode tv
```

### Grafana Server with Anonymous access enabled

This will take the browser to the default dashboard on play.grafana.org in fullscreen kiosk mode (no login needed):

```BASH
./bin/grafana-kiosk --URL https://play.grafana.org --login-method anon --kiosk-mode tv
```


This will take the browser to a playlist on play.grafana.org in fullscreen kiosk mode (no login needed):

```BASH
./bin/grafana-kiosk --URL https://play.grafana.org/playlists/play/1 --login-method anon --kiosk-mode tv
```

## Building

A Makefile is provided for building the utility.

```BASH
make
```

This will generate executables in "bin" that can be run on a variety of platforms.

## TODO
- Support for OAuth2 logins

## References

https://grafana.com/docs/reference/playlist/