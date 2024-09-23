package main

import (
	"log"
	"os"
	"testing"

	"github.com/grafana/grafana-kiosk/pkg/kiosk"
	"github.com/ilyakaznacheev/cleanenv"
	. "github.com/smartystreets/goconvey/convey"
)

// TestMigration checks kiosk command.
func TestMigration(t *testing.T) {
	Convey("Given Previous YAML Configuration", t, func() {
		Convey("Migrate YAML Configuration", func() {
			Convey("General", func() {
				cfg := kiosk.ConfigLegacy{}
				pathToYAML := "../../../testdata/legacy-config-local.yaml"
				if err := cleanenv.ReadConfig(pathToYAML, &cfg); err != nil {
					log.Println("Error reading config file", err)
					os.Exit(-1)
				} else {
					log.Println("Using config from", pathToYAML)
				}
				So(cfg.General.AutoFit, ShouldBeTrue)
				So(cfg.General.Mode, ShouldEqual, "full")
				So(cfg.General.LXDEEnabled, ShouldBeTrue)
				So(cfg.General.LXDEEnabled, ShouldBeTrue)
				So(cfg.General.WindowSize, ShouldEqual, "1920,1080")
				//
				So(cfg.Target.LoginMethod, ShouldEqual, "local")
				So(cfg.Target.Username, ShouldEqual, "user1")
				So(cfg.Target.Password, ShouldEqual, "changeme")
				So(cfg.Target.IsPlayList, ShouldBeFalse)
				So(cfg.Target.URL, ShouldEqual, "https://notplay.grafana.com")
				So(cfg.Target.IgnoreCertificateErrors, ShouldBeFalse)

				// now migrate to new config
			})
		})
	})
}
