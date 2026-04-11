package kiosk

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLocalLoginBypassURL(t *testing.T) {
	Convey("Given a URL for local login bypass", t, func() {
		Convey("When URL has a path", func() {
			result := LocalLoginBypassURL("https://grafana.example.com/dashboard/db/test")
			So(result, ShouldEqual, "https://grafana.example.com/login/local")
		})

		Convey("When URL has no path", func() {
			result := LocalLoginBypassURL("https://grafana.example.com")
			So(result, ShouldEqual, "https://grafana.example.com/login/local")
		})

		Convey("When URL has a port", func() {
			result := LocalLoginBypassURL("https://localhost:3000/d/abc123")
			So(result, ShouldEqual, "https://localhost:3000/login/local")
		})

		Convey("When URL uses http", func() {
			result := LocalLoginBypassURL("http://grafana.local/playlists/play/1")
			So(result, ShouldEqual, "http://grafana.local/login/local")
		})

		Convey("When URL has a deep path", func() {
			result := LocalLoginBypassURL("https://bkgann3.grafana.net/dashboard/db/sensu-summary")
			So(result, ShouldEqual, "https://bkgann3.grafana.net/login/local")
		})
	})
}
