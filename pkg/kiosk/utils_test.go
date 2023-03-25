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
			So(anURL, ShouldEqual, "https://play.grafana/com?autofitpanels=&kiosk=1")
		})
		Convey("TV Mode Anonymous Login", func() {
			anURL := GenerateURL("https://play.grafana/com", "tv", true, false)
			So(anURL, ShouldEqual, "https://play.grafana/com?autofitpanels=&kiosk=tv")
		})
		Convey("Not Fullscreen Anonymous Login", func() {
			anURL := GenerateURL("https://play.grafana/com", "disabled", true, false)
			So(anURL, ShouldEqual, "https://play.grafana/com?autofitpanels=")
		})
		Convey("Default Kiosk Anonymous Login", func() {
			anURL := GenerateURL("https://play.grafana/com", "", false, false)
			So(anURL, ShouldEqual, "https://play.grafana/com?kiosk=1")
		})
	})
}
