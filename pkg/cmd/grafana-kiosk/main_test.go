package main

import (
	"os"
	"testing"

	"github.com/grafana/grafana-kiosk/pkg/kiosk"
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
			IdToken: kiosk.IdToken{
				KeyFile:  "/tmp/key.json",
				Audience: "clientid",
			},
			ApiKey: kiosk.ApiKey{
				Apikey: "abc",
			},
		}

		Convey("Anonymous Login", func() {
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = []string{"grafana-kiosk", ""}
			result := ProcessArgs(cfg)
			So(result.LoginMethod, ShouldEqual, "anon")
			So(result.URL, ShouldEqual, "https://play.grafana.org")
		})
		Convey("Local Login", func() {
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = []string{"grafana-kiosk", "-login-method", "local"}
			result := ProcessArgs(cfg)
			So(result.LoginMethod, ShouldEqual, "local")
			So(result.URL, ShouldEqual, "https://play.grafana.org")
		})
	})
}
