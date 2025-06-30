package kiosk

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func GetConfig() Config {
	return Config{
		Target: Target{
			URL: "https://play.grafana/com",
		},
		General: General{
			AutoFit: true,
		},
	}
}

// TestGenerateURL
func TestGenerateURL(t *testing.T) {
	Convey("Given URL params", t, func() {

		Convey("Fullscreen Anonymous Login", func() {
			conf := GetConfig()
			conf.General.Mode = "full"
			anURL := GenerateURL(&conf)
			So(anURL, ShouldEqual, "https://play.grafana/com?kiosk=1&autofitpanels")
		})
		Convey("TV Mode Anonymous Login", func() {
			conf := GetConfig()
			conf.General.Mode = "tv"
			anURL := GenerateURL(&conf)
			So(anURL, ShouldEqual, "https://play.grafana/com?kiosk=tv&autofitpanels")
		})
		Convey("Not Fullscreen Anonymous Login", func() {
			conf := GetConfig()
			conf.General.Mode = "disabled"
			anURL := GenerateURL(&conf)
			So(anURL, ShouldEqual, "https://play.grafana/com?autofitpanels")
		})
		Convey("Default Kiosk Anonymous Login", func() {
			conf := GetConfig()
			conf.General.Mode = ""
			conf.General.AutoFit = false
			anURL := GenerateURL(&conf)
			So(anURL, ShouldEqual, "https://play.grafana/com?kiosk=1")
		})
		Convey("Default Anonymous Login with autofit", func() {
			conf := GetConfig()
			conf.General.HideLinks = true
			conf.General.HideTimePicker = true
			conf.General.HideVariables = true
			anURL := GenerateURL(&conf)
			So(anURL, ShouldEqual, "https://play.grafana/com?_dash.hideLinks=&_dash.hideTimePicker=&_dash.hideVariables=&kiosk=1&autofitpanels")
		})
	})
}
