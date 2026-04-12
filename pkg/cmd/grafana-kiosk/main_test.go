package main

import (
	"flag"
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
		// create a temp config file with ignore-certificate-errors: false
		configContent := `
target:
  URL: https://example.com
  login-method: anon
  ignore-certificate-errors: false
general:
  kiosk-mode: full
  autofit: true
  incognito: true
`
		tmpFile, err := os.CreateTemp("", "kiosk-test-*.yaml")
		So(err, ShouldBeNil)
		defer os.Remove(tmpFile.Name())
		_, err = tmpFile.WriteString(configContent)
		So(err, ShouldBeNil)
		tmpFile.Close()

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
			// load config file
			So(args.ConfigPath, ShouldNotBeEmpty)
			err := cleanenv.ReadConfig(args.ConfigPath, &cfg)
			So(err, ShouldBeNil)
			// config file has it false
			So(cfg.Target.IgnoreCertificateErrors, ShouldBeFalse)

			// apply CLI overrides (the fixed behavior)
			update := map[string]func(){
				"ignore-certificate-errors": func() { cfg.Target.IgnoreCertificateErrors = args.IgnoreCertificateErrors },
				"incognito":                 func() { cfg.General.Incognito = args.Incognito },
			}
			fs.Visit(func(f *flag.Flag) {
				if do, ok := update[f.Name]; ok {
					do()
				}
			})
			// CLI flag should win
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
			err := cleanenv.ReadConfig(args.ConfigPath, &cfg)
			So(err, ShouldBeNil)

			update := map[string]func(){
				"ignore-certificate-errors": func() { cfg.Target.IgnoreCertificateErrors = args.IgnoreCertificateErrors },
			}
			fs.Visit(func(f *flag.Flag) {
				if do, ok := update[f.Name]; ok {
					do()
				}
			})
			// no CLI flag passed, config value (false) should remain
			So(cfg.Target.IgnoreCertificateErrors, ShouldBeFalse)
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
