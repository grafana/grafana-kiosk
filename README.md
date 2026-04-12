# Grafana Kiosk

[![Build Status](https://img.shields.io/endpoint.svg?url=https%3A%2F%2Factions-badge.atrox.dev%2Fgrafana%2Fgrafana-kiosk%2Fbadge%3Fref%3Dmain&style=flat)](https://actions-badge.atrox.dev/grafana/grafana-kiosk/goto?ref=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/grafana/grafana-kiosk)](https://goreportcard.com/report/github.com/grafana/grafana-kiosk)
![coverage](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/grafana/grafana-kiosk/badges/badges/coverage.json)

A very useful feature of Grafana is the ability to display dashboards and playlists on a large TV.

This provides a utility to quickly standup a kiosk on devices like a Raspberry Pi or NUC.

The utility provides these options:

- Login
  - to a Grafana server (local account or bypass OAuth)
  - to a Grafana server with anonymous-mode enabled (same method used on [play.grafana.org](https://play.grafana.org))
  - to a Grafana Cloud instance
  - to a Grafana server with OAuth enabled
  - to an AWS Managed Grafana instance (both with and without MFA)
- Switch to kiosk or kiosk-tv mode
- Display the default home page set for the user
- Display a specified dashboard
- Start a playlist immediately (inactive mode enable)
- Can specify where to start kiosk for multiple displays

Additionally, an initialize option is provided to configure LXDE for Raspberry Pi Desktop.

## Installing on Linux

Download the zip or tar file from [releases](https://github.com/grafana/grafana-kiosk/releases)

The release file includes pre-built binaries. See table below for the types available.

| OS      | Architecture | Description  | Executable                      |
| ------- | ------------ | ------------ | ------------------------------- |
| linux   | amd64        | 64bit        | grafana-kiosk.linux.amd64       |
| linux   | 386          | 32bit        | grafana-kiosk.linux.386         |
| linux   | arm64        | 64bit Arm v7 | grafana-kiosk.linux.arm64       |
| linux   | arm          | ARM v5       | grafana-kiosk.linux.armv5       |
| linux   | arm          | ARM v6       | grafana-kiosk.linux.armv6       |
| linux   | arm          | ARM v7       | grafana-kiosk.linux.armv7       |
| darwin  | amd64        | 64bit        | grafana-kiosk.darwin.amd64      |
| windows | amd64        | 64bit        | grafana-kiosk.windows.amd64.exe |

Extract the zip or tar file, and copy the appropriate binary to /usr/bin/grafana-kiosk:

```bash
# sudo cp -p grafana-kiosk.linux.armv7 /usr/bin/grafana-kiosk
# sudo chmod 755 /usr/bin/grafana-kiosk
```

## Dependencies/Suggestion Packages

This application can run on most operating systems, but for linux some additional
binaries are suggested for full support.

Suggesting Packages:

`unclutter` (for hiding mouse/cursor)
`rng-tools` (for entropy issues)

## Usage

NOTE: Flags with parameters should use an "equals"
  `-autofit=true`
  `-URL=https://play.grafana.org` when used with any boolean flags.

```TEXT
  -URL string
      URL to Grafana server (default "https://play.grafana.org")
  -apikey string
      apikey
  -audience string
      idtoken audience
  -auto-login
      oauth_auto_login is enabled in grafana config
  -autofit
      Fit panels to screen (default true)
  -c string
      Path to configuration file (config.yaml)
  -field-password string
      Fieldname for the password (default "password")
  -field-username string
      Fieldname for the username (default "username")
  -hide-links
      Hide links in the top nav bar
  -hide-time-picker
      Hide time picker in the top nav bar
  -hide-variables
      Hide variables in the top nav bar
  -ignore-certificate-errors
      Ignore SSL/TLS certificate error
  -incognito
      Use incognito mode (default true)
  -keyfile string
      idtoken json credentials (default "key.json")
  -kiosk-mode string
      Kiosk Display Mode [full|tv|disabled]
      full = No TOPNAV and No SIDEBAR
      tv = No SIDEBAR
      disabled = omit option
       (default "full")
  -login-method string
      [anon|local|gcom|goauth|idtoken|apikey|aws|azuread] (default "anon")
  -lxde
      Initialize LXDE for kiosk mode
  -lxde-home string
      Path to home directory of LXDE user running X Server (default "/home/pi")
  -page-load-delay-ms int
      Delay in milliseconds before navigating to URL (default 2000)
  -password string
      password (default "guest")
  -playlists
      URL is a playlist
  -scale-factor string
      Scale factor, sort of zoom (default "1.0")
  -use-mfa
      password
  -username string
      username (default "guest")
  -wait-for-password-field
      oauth_auto_login is enabled in grafana config
  -wait-for-password-field-class string
      oauth_auto_login is enabled in grafana config
  -wait-for-stay-signed-in-prompt
      oauth_auto_login is enabled in grafana config
  -window-position string
      Top Left Position of Kiosk (default "0,0")
  -window-size string
      Size of Kiosk in pixels (width,height)
```

### Using a configuration file

The kiosk can also be started using a configuration file, along with environment variables.
When using this option, all other arguments passed are ignored.

```YAML
general:
  kiosk-mode: full
  autofit: true
  incognito: true
  lxde: true
  lxde-home: /home/pi
  scale-factor: 1.0

target:
  login-method: anon
  username: user
  password: changeme
  playlist: false
  URL: https://play.grafana.org
  ignore-certificate-errors: false
```

```bash
grafana-kiosk -c config.yaml
```

Environment variables can be set and will override the configuration file.
They can also be used instead of a configuration file.

```TEXT
  KIOSK_APIKEY_APIKEY string
      APIKEY
  KIOSK_AUTOFIT bool
      fit panels to screen (default "true")
  KIOSK_DEBUG bool
      enables debug output (default "false")
  KIOSK_GOAUTH_AUTO_LOGIN bool
      [false|true]
  KIOSK_GOAUTH_FIELD_PASSWORD string
      Password html input name value
  KIOSK_GOAUTH_FIELD_USER string
      Username html input name value
  KIOSK_GOAUTH_WAIT_FOR_PASSWORD_FIELD bool
      Indicate that it's necessary to wait for the password field
  KIOSK_GOAUTH_WAIT_FOR_PASSWORD_FIELD_CLASS string
      Ignore this password field when waiting for it being visible
  KIOSK_GOAUTH_WAIT_FOR_STAY_SIGNED_IN_PROMPT bool
      Indicate that it's necessary to wait for the stay signed in prompt,
      and will then click yes
  KIOSK_GPU_ENABLED bool
      disable GPU support (default "false")
  KIOSK_HIDE_LINKS bool
      Hide links in the top nav bar (default "false")
  KIOSK_HIDE_TIME_PICKER bool
      Hide time picker in the top nav bar (default "false")
  KIOSK_HIDE_VARIABLES bool
      Hide variables in the top nav bar (default "false")
  KIOSK_IDTOKEN_AUDIENCE string
      Audience for idtoken, typically your oauth client id
  KIOSK_IDTOKEN_KEYFILE string
      JSON Credentials for idtoken (default "key.json")
  KIOSK_IGNORE_CERTIFICATE_ERRORS bool
      ignore SSL/TLS certificate errors (default "false")
  KIOSK_INCOGNITO bool
      use incognito mode (default "true")
  KIOSK_IS_PLAYLIST bool
      URL is a playlist (default "false")
  KIOSK_LOGIN_METHOD string
      [anon|local|gcom|goauth|idtoken|apikey|aws|azuread] (default "anon")
  KIOSK_LOGIN_PASSWORD string
      password (default "guest")
  KIOSK_LOGIN_USER string
      username (default "guest")
  KIOSK_LXDE_ENABLED bool
      initialize LXDE for kiosk mode (default "false")
  KIOSK_LXDE_HOME string
      path to home directory of LXDE user running X Server (default "/home/pi")
  KIOSK_MODE string
      [full|tv|disabled] (default "full")
  KIOSK_OZONE_PLATFORM string
      Set ozone-platform option (wayland|cast|drm|wayland|x11) (default "")
  KIOSK_PAGE_LOAD_DELAY_MS int64
      milliseconds to wait before expecting page load (default "2000")
  KIOSK_SCALE_FACTOR string
      Scale factor, like zoom (default "1.0")
  KIOSK_URL string
      URL to Grafana server (default "https://play.grafana.org")
  KIOSK_USE_MFA bool
      MFA is enabled for given account (default "false")
  KIOSK_WINDOW_POSITION string
      Top Left Position of Kiosk (default "0,0")
  KIOSK_WINDOW_SIZE string
      Size of Kiosk in pixels (width,height) (default "")
```

### Window size and kiosk mode

By default, the kiosk launches Chrome in fullscreen kiosk mode (`--kiosk --start-fullscreen`). The `-window-size` and
`-kiosk-mode` flags interact as follows:

| `-window-size` | `-kiosk-mode` | Behavior |
| --- | --- | --- |
| not set | `full` (default) | Fullscreen kiosk, panels autofit to screen |
| not set | `tv` or `disabled` | Fullscreen kiosk, no panel autofit |
| set (e.g. `1920,1080`) | `full` (default) | App window at specified size, then cycles to fullscreen via CDP |
| set (e.g. `1920,1080`) | `tv` or `disabled` | App window at specified size, no fullscreen |

When `-window-size` is specified, Chrome launches in app mode (no address bar or controls). Before navigating to the
dashboard, the kiosk cycles the browser window state via the Chrome DevTools Protocol (CDP) to force Chrome to properly
register viewport dimensions. This ensures Grafana correctly applies the `autofitpanels` parameter on initial page load.

### Hosted Grafana using grafana.com authentication

This will login to a Hosted Grafana instance and take the browser to the default dashboard in fullscreen kiosk mode:

```bash
./bin/grafana-kiosk \
  -URL=https://bkgann3.grafana.net \
  -login-method=gcom -username=bkgann -password=abc123 \
  -kiosk-mode=full
```

This will login to a Hosted Grafana instance and take the browser to a
specific dashboard in tv kiosk mode:

```bash
./bin/grafana-kiosk \
  -URL=https://bkgann3.grafana.net/dashboard/db/sensu-summary \
  -login-method=gcom -username=bkgann -password=abc123 \
  -kiosk-mode tv
```

This will login to a Hosted Grafana instance and take the browser to a
playlist in fullscreen kiosk mode, and autofit the panels to fill the
display.

```bash
./bin/grafana-kiosk \
  -URL=https://bkgann3.grafana.net/playlists/play/1 \
  -login-method=gcom -username=bkgann -password=abc123 \
  -kiosk-mode=full -playlist -autofit=true
```

### Grafana Server with Local Accounts

This will login to a grafana server that uses local accounts:

```bash
./bin/grafana-kiosk \
  -URL=https://localhost:3000 -login-method=local \
  -username=admin -password=admin -kiosk-mode=tv
```

If you are using a self-signed certificate, you can remove the
certificate error with `-ignore-certificate-errors`

```bash
./bin/grafana-kiosk \
  -URL=https://localhost:3000 -login-method=local \
  -username=admin -password=admin \
  -kiosk-mode=tv -ignore-certificate-errors
```

This will login to a grafana server, configured for AzureAD OAuth and
has Oauth_auto_login is enabled, bypassing OAuth and using a manually
setup local username and password.

```bash
./bin/grafana-kiosk \
  -URL=https://localhost:3000 -login-method=local \
  -username=admin -password=admin \
  -auto-login=true -kiosk-mode=tv
```

### Grafana Server with Anonymous access enabled

This will take the browser to the default dashboard on play.grafana.org in fullscreen kiosk mode (no login needed):

```bash
./bin/grafana-kiosk -URL=https://play.grafana.org -login-method=anon -kiosk-mode=tv
```

This will take the browser to a playlist on play.grafana.org in fullscreen kiosk mode (no login needed):

```bash
./bin/grafana-kiosk -URL=https://play.grafana.org/playlists/play/1 -login-method=anon -kiosk-mode=tv
```

### Grafana Server with Api Key

This will take the browser to the default dashboard on play.grafana.org in fullscreen kiosk mode:

```bash
./bin/grafana-kiosk \
  -URL=https://play.grafana.org -login-method apikey \
  --apikey "xxxxxxxxxxxxxxx" -kiosk-mode=tv
```

### Grafana Server with Generic Oauth

This will login to a Generic Oauth service, configured on Grafana.
Oauth_auto_login is disabled. As Oauth provider is Keycloak used.

```bash
go run pkg/cmd/grafana-kiosk/main.go \
  -URL=https://my.grafana.oauth/playlists/play/1 \
  -login-method=goauth -username=test -password=test
```

This will login to a Generic Oauth service, configured on Grafana.
Oauth_auto_login is disabled. As Oauth provider is Keycloak used and
also the login and password html input name is set.

```bash
go run pkg/cmd/grafana-kiosk/main.go \
  -URL=https://my.grafana.oauth/playlists/play/1 \
  -login-method=goauth -username=test -password=test \
  -field-username=username -field-password=password
```

This will login to a Generic Oauth service, configured on Grafana.
Oauth_auto_login is enabled. As Oauth provider is Keycloak used and
also the login and password html input name is set.

```bash
go run pkg/cmd/grafana-kiosk/main.go \
  -URL=https://my.grafana.oauth/playlists/play/1 \
  -login-method=goauth -username=test -password=test \
  -field-username=username -field-password=password \
  -auto-login=true
```

### Google Idtoken authentication

This allows you to log in through Google Identity Aware Proxy using a
service account through injecting authorization headers with bearer
tokens into each request. The idtoken library will generate new tokens
as needed on expiry, allowing grafana kiosk mode without exposing a
fully privileged google user on your kiosk device.

```bash
./bin/grafana-kiosk \
  -URL=https://play.grafana.org/playlists/play/1 \
  -login-method=idtoken -keyfile /tmp/foo.json \
  -audience myoauthid.apps.googleusercontent.com
```

## LXDE Options

The `-lxde` option initializes settings for the desktop.

Actions Performed:

- sets profile via lxpanel to LXDE
- sets pcmanfs profile to LXDE
- runs `xset s off` to disable screensaver
- runs `xset -dpms` to disable power-saving (prevents screen from turning off)
- runs `xset s noblank` disables blank mode for screensaver (maybe not needed)
- runs `unclutter` to hide the mouse

The `-lxde-home` option allows you to specify a different $HOME directory where the lxde configuration files can be found.

## Automatic Startup

### Session-based

LXDE can start the kiosk automatically by creating this file:

Create/edit the file: `/home/pi/.config/lxsession/LXDE-pi/autostart`

```bash
/usr/bin/grafana-kiosk \
  -URL=https://bkgann3.grafana.net/dashboard/db/sensu-summary \
  -login-method=gcom -username=bkgann -password=abc123 \
  -kiosk-mode=full -lxde
```

#### Session with disconnected "screen"

Alternatively you can run grafana-kiosk under screen, which can be
very useful for debugging.

Create/edit the file: `/home/pi/.config/lxsession/LXDE-pi/autostart`

```bash
screen -d -m bash -c "/usr/bin/grafana-kiosk \
  -URL=https://bkgann3.grafana.net/dashboard/db/sensu-summary \
  -login-method=gcom -username=bkgann -password=abc123 \
  -kiosk-mode=full -lxde"
```

### Desktop Link

Create/edit the file: `/home/pi/.config/autostart/grafana-kiosk.desktop`

```ini
[Desktop Entry]
Type=Application
Exec=/usr/bin/grafana-kiosk -URL=https://bkgann3.grafana.net/dashboard/db/sensu-summary -login-method=gcom -username=bkgann -password=abc123 -kiosk-mode=full -lxde
```

#### Desktop Link with disconnected "screen"

```ini
[Desktop Entry]
Type=Application
Exec=screen -d -m bash -c /usr/bin/grafana-kiosk -URL=https://bkgann3.grafana.net/dashboard/db/sensu-summary -login-method=gcom -username=bkgann -password=abc123 -kiosk-mode=full -lxde
```

## Systemd startup

```bash
# sudo touch /etc/systemd/system/grafana-kiosk.service
# sudo chmod 664 /etc/systemd/system/grafana-kiosk.service
```

```ini
[Unit]
Description=Grafana Kiosk
Documentation=https://github.com/grafana/grafana-kiosk
Documentation=https://grafana.com/blog/2019/05/02/grafana-tutorial-how-to-create-kiosks-to-display-dashboards-on-a-tv
After=network-online.target

[Service]
User=pi
Environment="DISPLAY=:0"
Environment="XAUTHORITY=/home/pi/.Xauthority"

# Disable screensaver and monitor standby
ExecStartPre=xset s off
ExecStartPre=xset -dpms
ExecStartPre=xset s noblank

ExecStart=/usr/bin/grafana-kiosk -URL=<url> -login-method=local -username=<username> -password=<password> -playlist=true

[Install]
WantedBy=graphical.target
```

Reload systemd:

```bash
# sudo systemctl daemon-reload
```

Enable, Start, Get Status, and logs:

```bash
# sudo systemctl enable grafana-kiosk
# sudo systemctl start grafana-kiosk
# sudo systemctl status grafana-kiosk
```

Logs:

```bash
journalctl -u grafana-kiosk
```

## Troubleshooting

### Timeout Launching

```log
2020/08/24 10:18:41 Launching local login kiosk
panic: websocket url timeout reached
```

Often this is due to lack of entropy, for linux you would need to install `rng-tools` (or an equivalent).

```bash
apt install rng-tools
```

## Building

A Magefile is provided for building the utility, you can install mage by following the instructions at <https://magefile.org/>

```bash
mage -v
```

This will generate executables in "bin" that can be run on a variety of platforms.

For full build and testing options use:

```bash
mage -l
```

## TODO

- RHEL/CentOS auto-startup
- Everything in issues!

## References

- [TV and Kiosk Mode](https://grafana.com/docs/guides/whats-new-in-v5-3/#tv-and-kiosk-mode)
- [Grafana Playlists](https://grafana.com/docs/reference/playlist)

## Thanks to our Contributors

- [Michael Pasqualone](https://github.com/michaelp85) for the session-based startup ideas!
- [Brendan Ball](https://github.com/BrendanBall) for contributing the desktop link startup example for LXDE!
- [Alex Heylin](https://github.com/AlexHeylin) for the v7 login fix - and also works with v6!
- [Xan Manning](https://github.com/xanmanning) for the ignore certificate option!
- [David Stäheli](https://github.com/mistadave) for the OAuth implementation!
- [Marcus Ramberg](https://github.com/marcusramberg) for the Google ID Token Auth implementation!
- [Ronan Salmon](https://github.com/ronansalmon) for API token authentication!

Any many others!
