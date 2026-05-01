package main

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/grafana/grafana-kiosk/pkg/kiosk"
	"github.com/ilyakaznacheev/cleanenv"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSanitize(t *testing.T) {
	Convey("Given a string to sanitize", t, func() {
		Convey("When string contains newlines", func() {
			result := sanitize("hello\nworld")
			So(result, ShouldEqual, "helloworld")
		})

		Convey("When string contains carriage returns", func() {
			result := sanitize("hello\rworld")
			So(result, ShouldEqual, "helloworld")
		})

		Convey("When string contains both newlines and carriage returns", func() {
			result := sanitize("line1\r\nline2\nline3\r")
			So(result, ShouldEqual, "line1line2line3")
		})

		Convey("When string has no control characters", func() {
			result := sanitize("clean string")
			So(result, ShouldEqual, "clean string")
		})

		Convey("When string is empty", func() {
			result := sanitize("")
			So(result, ShouldEqual, "")
		})
	})
}

// TestKiosk checks kiosk command.
func TestCLIFlagsOverrideConfigFile(t *testing.T) {
	Convey("Given a config file with specific values", t, func() {
		configContent := `
target:
  URL: https://example.com
  login-method: anon
  ignore-certificate-errors: false
general:
  kiosk-mode: full
  autofit: true
  incognito: true
  window-position: "0,0"
  scale-factor: "1.0"
`
		tmpFile, err := os.CreateTemp("", "kiosk-test-*.yaml")
		So(err, ShouldBeNil)
		defer func() { _ = os.Remove(tmpFile.Name()) }()
		_, err = tmpFile.WriteString(configContent)
		So(err, ShouldBeNil)
		_ = tmpFile.Close()

		Convey("CLI flag should override config file value", func() {
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = []string{
				"grafana-kiosk",
				"-c", tmpFile.Name(),
				"-ignore-certificate-errors",
			}
			var cfg kiosk.Config
			args, fs := ProcessArgs(&cfg)
			err := loadConfig(args, fs, &cfg)
			So(err, ShouldBeNil)
			So(cfg.Target.IgnoreCertificateErrors, ShouldBeTrue)
		})

		Convey("Config file value should be used when no CLI flag is passed", func() {
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = []string{
				"grafana-kiosk",
				"-c", tmpFile.Name(),
			}
			var cfg kiosk.Config
			args, fs := ProcessArgs(&cfg)
			err := loadConfig(args, fs, &cfg)
			So(err, ShouldBeNil)
			So(cfg.Target.IgnoreCertificateErrors, ShouldBeFalse)
		})

		Convey("Multiple CLI flags should override respective config values", func() {
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = []string{
				"grafana-kiosk",
				"-c", tmpFile.Name(),
				"-ignore-certificate-errors",
				"-kiosk-mode", "tv",
				"-incognito=false",
			}
			var cfg kiosk.Config
			args, fs := ProcessArgs(&cfg)
			err := loadConfig(args, fs, &cfg)
			So(err, ShouldBeNil)
			So(cfg.Target.IgnoreCertificateErrors, ShouldBeTrue)
			So(cfg.General.Mode, ShouldEqual, "tv")
			So(cfg.General.Incognito, ShouldBeFalse)
			// non-overridden values preserved from config file
			So(cfg.Target.URL, ShouldEqual, "https://example.com")
			So(cfg.General.AutoFit, ShouldBeTrue)
		})

		Convey("Invalid config path should return error", func() {
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = []string{
				"grafana-kiosk",
				"-c", "/nonexistent/config.yaml",
			}
			var cfg kiosk.Config
			args, fs := ProcessArgs(&cfg)
			err := loadConfig(args, fs, &cfg)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestLoadConfigEnvOnly(t *testing.T) {
	Convey("Given no config file and no CLI flags", t, func() {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{"grafana-kiosk"}

		Convey("Should load defaults from environment", func() {
			var cfg kiosk.Config
			args, fs := ProcessArgs(&cfg)
			err := loadConfig(args, fs, &cfg)
			So(err, ShouldBeNil)
			// env-defaults from struct tags
			So(cfg.Target.URL, ShouldEqual, "https://play.grafana.org")
			So(cfg.Target.LoginMethod, ShouldEqual, "anon")
			So(cfg.General.AutoFit, ShouldBeTrue)
			So(cfg.General.Mode, ShouldEqual, "full")
		})

		Convey("Should apply CLI flags over env defaults", func() {
			os.Args = []string{
				"grafana-kiosk",
				"-URL", "http://localhost:3000",
				"-kiosk-mode", "tv",
			}
			var cfg kiosk.Config
			args, fs := ProcessArgs(&cfg)
			err := loadConfig(args, fs, &cfg)
			So(err, ShouldBeNil)
			So(cfg.Target.URL, ShouldEqual, "http://localhost:3000")
			So(cfg.General.Mode, ShouldEqual, "tv")
			// non-overridden defaults preserved
			So(cfg.General.AutoFit, ShouldBeTrue)
			So(cfg.Target.LoginMethod, ShouldEqual, "anon")
		})
	})
}

func TestLoadConfigMalformedYAML(t *testing.T) {
	Convey("Given a malformed YAML config file", t, func() {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{
			"grafana-kiosk",
			"-c", "../../../testdata/config-malformed.yaml",
		}

		Convey("Should return error", func() {
			var cfg kiosk.Config
			args, fs := ProcessArgs(&cfg)
			err := loadConfig(args, fs, &cfg)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "error reading config file")
		})
	})
}

func TestLoadConfigFromTestdata(t *testing.T) {
	Convey("Given the anon testdata config", t, func() {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{
			"grafana-kiosk",
			"-c", "../../../testdata/config-anon.yaml",
		}

		Convey("Should load all values from config file", func() {
			var cfg kiosk.Config
			args, fs := ProcessArgs(&cfg)
			err := loadConfig(args, fs, &cfg)
			So(err, ShouldBeNil)
			So(cfg.General.Mode, ShouldEqual, "full")
			So(cfg.General.AutoFit, ShouldBeTrue)
			So(cfg.General.LXDEEnabled, ShouldBeTrue)
			So(cfg.General.LXDEHome, ShouldEqual, "/home/pi")
			So(cfg.Target.LoginMethod, ShouldEqual, "anon")
			So(cfg.Target.URL, ShouldEqual, "https://play.grafana.org")
			So(cfg.Target.Username, ShouldEqual, "user")
			So(cfg.Target.IsPlayList, ShouldBeFalse)
			So(cfg.Target.UseMFA, ShouldBeFalse)
			So(cfg.Target.IgnoreCertificateErrors, ShouldBeFalse)
			So(cfg.GoAuth.AutoLogin, ShouldBeFalse)
			So(cfg.GoAuth.UsernameField, ShouldEqual, "username")
			So(cfg.GoAuth.PasswordField, ShouldEqual, "password")
		})
	})

	Convey("Given the local testdata config", t, func() {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{
			"grafana-kiosk",
			"-c", "../../../testdata/config-local.yaml",
		}

		Convey("Should load login-method as local", func() {
			var cfg kiosk.Config
			args, fs := ProcessArgs(&cfg)
			err := loadConfig(args, fs, &cfg)
			So(err, ShouldBeNil)
			So(cfg.Target.LoginMethod, ShouldEqual, "local")
		})
	})
}

func TestLoadConfigEnvOverridesFile(t *testing.T) {
	Convey("Given a config file and an environment variable override", t, func() {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{
			"grafana-kiosk",
			"-c", "../../../testdata/config-anon.yaml",
		}

		Convey("Env var should override config file value", func() {
			oldVal := os.Getenv("KIOSK_URL")
			defer func() { _ = os.Setenv("KIOSK_URL", oldVal) }()
			_ = os.Setenv("KIOSK_URL", "https://env-override.example.com")

			var cfg kiosk.Config
			args, fs := ProcessArgs(&cfg)
			err := loadConfig(args, fs, &cfg)
			So(err, ShouldBeNil)
			So(cfg.Target.URL, ShouldEqual, "https://env-override.example.com")
		})
	})
}

func TestProcessArgsAllFlags(t *testing.T) {
	Convey("Given all CLI flags", t, func() {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{
			"grafana-kiosk",
			"-URL", "https://test.grafana.org",
			"-login-method", "gcom",
			"-username", "testuser",
			"-password", "testpass",
			"-use-mfa",
			"-kiosk-mode", "tv",
			"-window-position", "100,200",
			"-window-size", "1280,720",
			"-scale-factor", "1.5",
			"-browser", "edge",
			"-browser-path", "/opt/edge/msedge",
			"-page-load-delay-ms", "5000",
				"-restart-delay-ms", "10000",
			"-playlists",
			"-autofit=false",
			"-hide-links",
			"-hide-logo",
			"-hide-playlist-nav",
			"-hide-time-picker",
			"-hide-variables",
			"-lxde",
			"-lxde-home", "/home/test",
			"-ignore-certificate-errors",
			"-incognito=false",
			"-auto-login",
			"-wait-for-password-field",
			"-wait-for-password-field-class", "ignore-me",
			"-wait-for-stay-signed-in-prompt",
			"-field-username", "email",
			"-field-password", "passwd",
			"-audience", "my-client-id",
			"-keyfile", "/tmp/creds.json",
			"-apikey", "secret123",
		}

		var cfg kiosk.Config
		result, _ := ProcessArgs(&cfg)

		Convey("Should parse target flags", func() {
			So(result.URL, ShouldEqual, "https://test.grafana.org")
			So(result.LoginMethod, ShouldEqual, "gcom")
			So(result.Username, ShouldEqual, "testuser")
			So(result.Password, ShouldEqual, "testpass")
			So(result.UseMFA, ShouldBeTrue)
			So(result.IsPlayList, ShouldBeTrue)
			So(result.IgnoreCertificateErrors, ShouldBeTrue)
		})

		Convey("Should parse general flags", func() {
			So(result.Mode, ShouldEqual, "tv")
			So(result.AutoFit, ShouldBeFalse)
			So(result.Incognito, ShouldBeFalse)
			So(result.WindowPosition, ShouldEqual, "100,200")
			So(result.WindowSize, ShouldEqual, "1280,720")
			So(result.ScaleFactor, ShouldEqual, "1.5")
			So(result.Browser, ShouldEqual, "edge")
			So(result.BrowserPath, ShouldEqual, "/opt/edge/msedge")
			So(result.PageLoadDelayMS, ShouldEqual, 5000)
			So(result.RestartDelayMS, ShouldEqual, 10000)
			So(result.HideLinks, ShouldBeTrue)
			So(result.HideLogo, ShouldBeTrue)
			So(result.HidePlaylistNav, ShouldBeTrue)
			So(result.HideTimePicker, ShouldBeTrue)
			So(result.HideVariables, ShouldBeTrue)
			So(result.LXDEEnabled, ShouldBeTrue)
			So(result.LXDEHome, ShouldEqual, "/home/test")
		})

		Convey("Should parse OAuth flags", func() {
			So(result.OauthAutoLogin, ShouldBeTrue)
			So(result.OauthWaitForPasswordField, ShouldBeTrue)
			So(result.OauthWaitForPasswordFieldIgnoreClass, ShouldEqual, "ignore-me")
			So(result.OauthWaitForStaySignedInPrompt, ShouldBeTrue)
			So(result.UsernameField, ShouldEqual, "email")
			So(result.PasswordField, ShouldEqual, "passwd")
		})

		Convey("Should parse ID token flags", func() {
			So(result.Audience, ShouldEqual, "my-client-id")
			So(result.KeyFile, ShouldEqual, "/tmp/creds.json")
		})

		Convey("Should parse API key flag", func() {
			So(result.APIKey, ShouldEqual, "secret123")
		})
	})
}

func TestLoadConfigAllFlagsOverride(t *testing.T) {
	Convey("Given a config file with all CLI flags overriding", t, func() {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{
			"grafana-kiosk",
			"-c", "../../../testdata/config-anon.yaml",
			"-URL", "https://override.example.com",
			"-login-method", "local",
			"-username", "overrideuser",
			"-password", "overridepass",
			"-kiosk-mode", "disabled",
			"-window-position", "50,50",
			"-window-size", "800,600",
			"-scale-factor", "2.0",
			"-page-load-delay-ms", "3000",
			"-playlists",
			"-autofit=false",
			"-hide-links",
			"-hide-logo",
			"-hide-playlist-nav",
			"-hide-time-picker",
			"-hide-variables",
			"-incognito=false",
			"-auto-login",
			"-field-username", "email",
			"-field-password", "pw",
			"-audience", "aud123",
			"-keyfile", "/tmp/k.json",
			"-apikey", "key456",
		}

		Convey("All CLI flags should override config file", func() {
			var cfg kiosk.Config
			args, fs := ProcessArgs(&cfg)
			err := loadConfig(args, fs, &cfg)
			So(err, ShouldBeNil)
			So(cfg.Target.URL, ShouldEqual, "https://override.example.com")
			So(cfg.Target.LoginMethod, ShouldEqual, "local")
			So(cfg.Target.Username, ShouldEqual, "overrideuser")
			So(cfg.Target.Password, ShouldEqual, "overridepass")
			So(cfg.Target.IsPlayList, ShouldBeTrue)
			So(cfg.General.Mode, ShouldEqual, "disabled")
			So(cfg.General.AutoFit, ShouldBeFalse)
			So(cfg.General.Incognito, ShouldBeFalse)
			So(cfg.General.WindowPosition, ShouldEqual, "50,50")
			So(cfg.General.WindowSize, ShouldEqual, "800,600")
			So(cfg.General.ScaleFactor, ShouldEqual, "2.0")
			So(cfg.General.PageLoadDelayMS, ShouldEqual, int64(3000))
			So(cfg.General.HideLinks, ShouldBeTrue)
			So(cfg.General.HideLogo, ShouldBeTrue)
			So(cfg.General.HidePlaylistNav, ShouldBeTrue)
			So(cfg.General.HideTimePicker, ShouldBeTrue)
			So(cfg.General.HideVariables, ShouldBeTrue)
			So(cfg.GoAuth.AutoLogin, ShouldBeTrue)
			So(cfg.GoAuth.UsernameField, ShouldEqual, "email")
			So(cfg.GoAuth.PasswordField, ShouldEqual, "pw")
			So(cfg.IDToken.Audience, ShouldEqual, "aud123")
			So(cfg.IDToken.KeyFile, ShouldEqual, "/tmp/k.json")
			So(cfg.APIKey.APIKey, ShouldEqual, "key456")
		})
	})
}

func TestMain(t *testing.T) {
	Convey("Given Default Configuration", t, func() {
		cfg := kiosk.Config{
			BuildInfo: kiosk.BuildInfo{
				Version: "1.0.0",
			},
			General: kiosk.General{
				AutoFit:        true,
				LXDEEnabled:    true,
				LXDEHome:       "/home/pi",
				Mode:           "full",
				WindowPosition: "0,0",
				WindowSize:     "1920,1080",
				ScaleFactor:    "1.0",
			},
			Target: kiosk.Target{
				IgnoreCertificateErrors: false,
				IsPlayList:              false,
				UseMFA:                  false,
				LoginMethod:             "local",
				Password:                "admin",
				URL:                     "http://localhost:3000",
				Username:                "admin",
			},
			GoAuth: kiosk.GoAuth{
				AutoLogin:     false,
				UsernameField: "user",
				PasswordField: "password",
			},
			IDToken: kiosk.IDToken{
				KeyFile:  "/tmp/key.json",
				Audience: "clientid",
			},
			APIKey: kiosk.APIKey{
				APIKey: "abc",
			},
		}
		Convey("General Options", func() {
			Convey("Parameter - autofit", func() {
				oldArgs := os.Args
				defer func() { os.Args = oldArgs }()
				os.Args = []string{"grafana-kiosk", ""}
				// starts out default true
				result, _ := ProcessArgs(cfg)
				So(result.AutoFit, ShouldBeTrue)
				// flag to set it false
				os.Args = []string{
					"grafana-kiosk",
					"--autofit=false",
				}
				result, _ = ProcessArgs(cfg)
				So(result.AutoFit, ShouldBeFalse)
			})

			Convey("Environment - autofit", func() {
				oldArgs := os.Args
				defer func() { os.Args = oldArgs }()
				os.Args = []string{"grafana-kiosk", ""}
				err := os.Setenv("KIOSK_AUTOFIT", "false")
				if err != nil {
					log.Println("Error setting environment KIOSK_AUTOFIT", err)
				}
				cfg := kiosk.Config{}
				if err := cleanenv.ReadEnv(&cfg); err != nil {
					log.Println("Error reading config from environment", err)
				}
				So(cfg.General.AutoFit, ShouldBeFalse)
			})
		})
		// end of general options

		Convey("Anonymous Login", func() {
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = []string{"grafana-kiosk", ""}
			result, _ := ProcessArgs(cfg)
			So(result.LoginMethod, ShouldEqual, "anon")
			So(result.URL, ShouldEqual, "https://play.grafana.org")
			So(result.AutoFit, ShouldBeTrue)
		})
		Convey("Local Login", func() {
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = []string{"grafana-kiosk", "-login-method", "local"}
			result, _ := ProcessArgs(cfg)
			So(result.LoginMethod, ShouldEqual, "local")
			So(result.URL, ShouldEqual, "https://play.grafana.org")
			So(result.AutoFit, ShouldBeTrue)
		})
	})
}

// captureLogOutput redirects log output to a buffer for the duration of fn.
func captureLogOutput(fn func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)
	fn()
	return buf.String()
}

func TestLogGeneralSettings(t *testing.T) {
	Convey("Given a config with general settings", t, func() {
		cfg := &kiosk.Config{
			General: kiosk.General{
				AutoFit:         true,
				Mode:            "full",
				Incognito:       true,
				WindowPosition:  "0,0",
				WindowSize:      "1920,1080",
				ScaleFactor:     "1.5",
				PageLoadDelayMS: 3000,
				HideLinks:       true,
				HideLogo:        true,
				HidePlaylistNav: true,
				HideTimePicker:  true,
				HideVariables:   true,
			},
		}

		Convey("Should log all general fields", func() {
			output := captureLogOutput(func() { logGeneralSettings(cfg) })
			So(output, ShouldContainSubstring, "--- General ----")
			So(output, ShouldContainSubstring, "AutoFit: true")
			So(output, ShouldContainSubstring, "Mode: full")
			So(output, ShouldContainSubstring, "Incognito: true")
			So(output, ShouldContainSubstring, "WindowSize: 1920,1080")
			So(output, ShouldContainSubstring, "ScaleFactor: 1.5")
			So(output, ShouldContainSubstring, "PageLoadDelayMS: 3000")
			So(output, ShouldContainSubstring, "HideLinks: true")
			So(output, ShouldContainSubstring, "HideLogo: true")
			So(output, ShouldContainSubstring, "HidePlaylistNav: true")
			So(output, ShouldContainSubstring, "HideTimePicker: true")
			So(output, ShouldContainSubstring, "HideVariables: true")
		})
	})
}

func TestLogTargetSettings(t *testing.T) {
	Convey("Given a config with target settings", t, func() {
		cfg := &kiosk.Config{
			Target: kiosk.Target{
				URL:         "https://grafana.example.com",
				LoginMethod: "local",
				Username:    "admin",
				Password:    "supersecret",
				IsPlayList:  true,
				UseMFA:      true,
			},
		}

		Convey("Should log target fields and redact password", func() {
			output := captureLogOutput(func() { logTargetSettings(cfg) })
			So(output, ShouldContainSubstring, "--- Target ---")
			So(output, ShouldContainSubstring, "URL: https://grafana.example.com")
			So(output, ShouldContainSubstring, "LoginMethod: local")
			So(output, ShouldContainSubstring, "Username: admin")
			So(output, ShouldContainSubstring, "Password: *redacted*")
			So(output, ShouldNotContainSubstring, "supersecret")
			So(output, ShouldContainSubstring, "IsPlayList: true")
			So(output, ShouldContainSubstring, "UseMFA: true")
		})
	})
}

func TestLogGoAuthSettings(t *testing.T) {
	Convey("Given a config with GoAuth settings", t, func() {
		cfg := &kiosk.Config{
			GoAuth: kiosk.GoAuth{
				AutoLogin:     true,
				UsernameField: "email",
				PasswordField: "passwd",
			},
		}

		Convey("Should log GoAuth fields", func() {
			output := captureLogOutput(func() { logGoAuthSettings(cfg) })
			So(output, ShouldContainSubstring, "--- GoAuth ---")
			So(output, ShouldContainSubstring, "Fieldname AutoLogin: true")
			So(output, ShouldContainSubstring, "Fieldname Username: email")
			So(output, ShouldContainSubstring, "Fieldname Password: passwd")
		})
	})
}
