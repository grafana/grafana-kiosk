# Grafana Kiosk

[![CircleCI](https://circleci.com/gh/grafana/grafana-kiosk.svg?style=svg)](https://circleci.com/gh/grafana/grafana-kiosk)
[![Go Report Card](https://goreportcard.com/badge/github.com/grafana/grafana-kiosk)](https://goreportcard.com/report/github.com/grafana/grafana-kiosk) [![codecov](https://codecov.io/gh/grafana/grafana-kiosk/branch/master/graph/badge.svg)](https://codecov.io/gh/grafana/grafana-kiosk)

A very useful feature of Grafana is the ability to display dashboards and playlists on a large TV.

This provides a utility to quickly standup a kiosk on devices like a Raspberry Pi or NUC.

The utitilty provides these options:

- login
  - to a Grafana server (local account)
  - to a grafana server with anonymous-mode enabled (same method used on [play.grafana.org](https://play.grafana.org))
  - to a hosted grafana instance
- switch to kiosk or kiosk-tv mode
- display the default home page set for the user
- display a specified dashboard
- start a playlist immediately (inactive mode enable)

Additionally, an initialize option is provided to configure LXDE for Raspberry Pi Desktop.

## Installing on Linux

Download the zip or tar file from [releases](https://github.com/grafana/grafana-kiosk/releases)

The release file includes pre-built binaries. See table below for the types available.

|  OS    | Architecture | Description    | Executable                      |
| ------ | ------------ | -------------- | ------------------------------- |
| linux  | amd64        | 64bit          | grafana-kiosk.linux.amd64       |
| linux  | 386          | 32bit          | grafana-kiosk.linux.386         |
| linux  | arm64        | 64bit Arm v7   | grafana-kiosk.linux.arm64       |
| linux  | arm          | ARM v5         | grafana-kiosk.linux.armv5       |
| linux  | arm          | ARM v6         | grafana-kiosk.linux.armv6       |
| linux  | arm          | ARM v7         | grafana-kiosk.linux.armv7       |
| darwin | amd64        | 64bit          | grafana-kiosk.darwin.amd64      |
| windows| amd64        | 64bit          | grafana-kiosk.windows.amd64.exe |

Extract the zip or tar file, and copy the appropriate binary to /usr/bin/grafana-kiosk:

```BASH
$ sudo cp -p grafana-kiosk.linux.armv7 /usr/bin/grafana-kiosk
```

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

`--ignore-certificate-errors` used with local and anonymous login methods

`--kiosk-mode`

- full  (no sidebar, top navigation disabled)
- tv (no sidebar, top navigation enabled)
- disabled (sidebar and top navigation enabled)

`--autofit` scales panels to fit the display (default is true)

- true
- false

`--lxde` enables initialization of LXDE

`--lxde-home` specifies home directory of LXDE user (default $HOME)

### Hosted Grafana using grafana.com authentication

This will login to a Hosted Grafana instance and take the browser to the default dashboard in fullscreen kiosk mode:

```bash
./bin/grafana-kiosk --URL https://bkgann3.grafana.net --login-method gcom --username bkgann --password abc123 --kiosk-mode full
```

This will login to a Hosted Grafana instance and take the browser to a specific dashboard in tv kiosk mode:

```bash
./bin/grafana-kiosk --URL https://bkgann3.grafana.net/dashboard/db/sensu-summary --login-method gcom --username bkgann --password abc123 --kiosk-mode tv
```

This will login to a Hosted Grafana instance and take the browser to a playlist in fullscreen kiosk mode, and autofit the panels to fill the display.

```bash
./bin/grafana-kiosk --URL https://bkgann3.grafana.net/playlists/play/1 --login-method gcom --username bkgann --password abc123 --kiosk-mode full --playlist --autofit
```

### Grafana Server with Local Accounts

This will login to a grafana server that uses local accounts:

```bash
./bin/grafana-kiosk --URL https://localhost:3000 --login-method local --username admin --password admin --kiosk-mode tv
```

If you are using a self-signed certificate, you can remove the certificate error with `--ignore-certificate-errors`

```bash
./bin/grafana-kiosk --URL https://localhost:3000 --login-method local --username admin --password admin --kiosk-mode tv --ignore-certificate-errors
```

### Grafana Server with Anonymous access enabled

This will take the browser to the default dashboard on play.grafana.org in fullscreen kiosk mode (no login needed):

```bash
./bin/grafana-kiosk --URL https://play.grafana.org --login-method anon --kiosk-mode tv
```

This will take the browser to a playlist on play.grafana.org in fullscreen kiosk mode (no login needed):

```bash
./bin/grafana-kiosk --URL https://play.grafana.org/playlists/play/1 --login-method anon --kiosk-mode tv
```

## LXDE Options

The `--lxde` option initializes settings for the desktop.

Actions Performed:

- sets profile via lxpanel to LXDE
- sets pcmanfs profile to LXDE
- runs `xset s off` to disable screensaver
- runs `xset -dpms` to disable power-saving (prevents screen from turning off)
- runs `xset s noblank` disables blank mode for screensaver (maybe not needed)
- runs `unclutter` to hide the mouse

The `--lxde-home` option allows you to specify a different $HOME directory where the lxde configuration files can be found.

## Automatic Startup

### Session-based

LXDE can start the kiosk automatically by creating this file:

Create/edit the file: `/home/pi/.config/lxsession/LXDE-pi/autostart`

```BASH
/usr/bin/grafana-kiosk --URL https://bkgann3.grafana.net/dashboard/db/sensu-summary --login-method gcom --username bkgann --password abc123 --kiosk-mode full --lxde
```

#### Session with disconnected "screen"

Alternatively you can run grafana-kiosk under screen, which can very useful for debugging.

Create/edit the file: `/home/pi/.config/lxsession/LXDE-pi/autostart`

```BASH
screen -d -m bash -c "/usr/bin/grafana-kiosk --URL https://bkgann3.grafana.net/dashboard/db/sensu-summary --login-method gcom --username bkgann --password abc123 --kiosk-mode full --lxde"
```

### Desktop Link

Create/edit the file: `/home/pi/.config/autostart/grafana-kiosk.desktop`

```INI
[Desktop Entry]
Type=Application
Exec=/usr/bin/grafana-kiosk --URL https://bkgann3.grafana.net/dashboard/db/sensu-summary --login-method gcom --username bkgann --password abc123 --kiosk-mode full --lxde
```

#### Desktop Link with disconnected "screen"

```INI
[Desktop Entry]
Type=Application
Exec=screen -d -m bash -c /usr/bin/grafana-kiosk --URL https://bkgann3.grafana.net/dashboard/db/sensu-summary --login-method gcom --username bkgann --password abc123 --kiosk-mode full --lxde
```

## Systemd startup

```BASH
$ sudo touch /etc/systemd/system/grafana-kiosk.service
$ sudo chmod 664 /etc/systemd/system/grafana-kiosk.service
```

```INI
[Unit]
Description=Grafana Kiosk
Documentation=https://github.com/grafana/grafana-kiosk
Documentation=https://grafana.com/blog/2019/05/02/grafana-tutorial-how-to-create-kiosks-to-display-dashboards-on-a-tv
After=network.target

[Service]
User=pi
Environment="DISPLAY=:0"
Environment="XAUTHORITY=/home/pi/.Xauthority"
ExecStart=/usr/bin/grafana-kiosk --URL <url>  -login-method local -username <username> -password <password> -playlist true

[Install]
WantedBy=graphical.target
```

Reload systemd:

```BASH
$ sudo systemctl daemon-reload
```

Enable, Start, Get Status, and logs:

```BASH
$ sudo systemctl enable grafana-kiosk
$ sudo systemctl start grafana-kiosk
$ sudo systemctl status grafana-kiosk
```

Logs:

```BASH
journalctl -u grafana-kiosk
```

## Building

A Makefile is provided for building the utility.

```bash
make
```

This will generate executables in "bin" that can be run on a variety of platforms.

## TODO

- Support for OAuth2 logins
- RHEL/CentOS auto-startup
- Everything in issues!

## References

- [TV and Kiosk Mode](https://grafana.com/docs/guides/whats-new-in-v5-3/#tv-and-kiosk-mode)
- [Grafana Playlists](https://grafana.com/docs/reference/playlist)

## Thanks to our Contributors

- [Michael Pasqualone](https://github.com/michaelp85) for the session-based startup ideas!
- [Brendan Ball](https://github.com/BrendanBall) for contributing the desktop link startup example for LXDE!
