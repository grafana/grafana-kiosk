package main

import (
	"log"
	"os"
	"testing"

	"github.com/grafana/grafana-kiosk/pkg/kiosk"
	"github.com/ilyakaznacheev/cleanenv"
	. "github.com/smartystreets/goconvey/convey"
)

// TestKiosk checks kiosk command.
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
				result := ProcessArgs(cfg)
				So(result.AutoFit, ShouldBeTrue)
				// flag to set it false
				os.Args = []string{
					"grafana-kiosk",
					"--autofit=false",
				}
				result = ProcessArgs(cfg)
				So(result.AutoFit, ShouldBeFalse)
			})

			Convey("Environment - autofit", func() {
				oldArgs := os.Args
				defer func() { os.Args = oldArgs }()
				os.Args = []string{"grafana-kiosk", ""}
				os.Setenv("KIOSK_AUTOFIT", "false")
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
			result := ProcessArgs(cfg)
			So(result.LoginMethod, ShouldEqual, "anon")
			So(result.URL, ShouldEqual, "https://play.grafana.org")
			So(result.AutoFit, ShouldBeTrue)
		})
		Convey("Local Login", func() {
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = []string{"grafana-kiosk", "-login-method", "local"}
			result := ProcessArgs(cfg)
			So(result.LoginMethod, ShouldEqual, "local")
			So(result.URL, ShouldEqual, "https://play.grafana.org")
			So(result.AutoFit, ShouldBeTrue)
		})
	})
}
