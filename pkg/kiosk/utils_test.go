package kiosk

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TestGenerateURL
func TestGenerateURL(t *testing.T) {
	Convey("Given URL params", t, func() {
		Convey("Fullscreen Anonymous Login", func() {
			anURL := GenerateURL("https://play.grafana/com", "full", true, false)
			So(anURL, ShouldEqual, "https://play.grafana/com?kiosk=1&autofitpanels")
		})
		Convey("TV Mode Anonymous Login", func() {
			anURL := GenerateURL("https://play.grafana/com", "tv", true, false)
			So(anURL, ShouldEqual, "https://play.grafana/com?kiosk=tv&autofitpanels")
		})
		Convey("Not Fullscreen Anonymous Login", func() {
			anURL := GenerateURL("https://play.grafana/com", "disabled", true, false)
			So(anURL, ShouldEqual, "https://play.grafana/com?autofitpanels")
		})
		Convey("Default Kiosk Anonymous Login", func() {
			anURL := GenerateURL("https://play.grafana/com", "", false, false)
			So(anURL, ShouldEqual, "https://play.grafana/com?kiosk=1")
		})
		Convey("Default Anonymous Login with autofit", func() {
			anURL := GenerateURL("https://play.grafana/com", "disabled", true, false)
			So(anURL, ShouldEqual, "https://play.grafana/com?autofitpanels")
		})
	})
}
