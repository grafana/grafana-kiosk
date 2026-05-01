package kiosk

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/grafana/grafana-kiosk/pkg/kiosk/config"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/login/shared"
)

func TestResolveBrowserExecPath(t *testing.T) {
	Convey("Given ResolveBrowserExecPath", t, func() {
		origLookPath := shared.LookPath
		defer func() { shared.LookPath = origLookPath }()

		Convey("When BrowserPath is set, it wins regardless of Browser", func() {
			shared.LookPath = func(string) (string, error) { return "/should/not/be/used", nil }
			cfg := &config.Config{General: config.General{Browser: "edge", BrowserPath: "/custom/path/to/browser"}}
			So(shared.ResolveBrowserExecPath(cfg), ShouldEqual, "/custom/path/to/browser")
		})

		Convey("When Browser is empty, returns empty (chromedp default)", func() {
			cfg := &config.Config{General: config.General{Browser: ""}}
			So(shared.ResolveBrowserExecPath(cfg), ShouldEqual, "")
		})

		Convey("When Browser is chrome, returns empty (chromedp default)", func() {
			cfg := &config.Config{General: config.General{Browser: "chrome"}}
			So(shared.ResolveBrowserExecPath(cfg), ShouldEqual, "")
		})

		Convey("When Browser is CHROME (case-insensitive), returns empty", func() {
			cfg := &config.Config{General: config.General{Browser: "CHROME"}}
			So(shared.ResolveBrowserExecPath(cfg), ShouldEqual, "")
		})

		Convey("When Browser is edge and msedge is on PATH, returns its path", func() {
			shared.LookPath = func(name string) (string, error) {
				if name == "msedge" {
					return "/usr/local/bin/msedge", nil
				}
				return "", fmt.Errorf("not found")
			}
			cfg := &config.Config{General: config.General{Browser: "edge"}}
			So(shared.ResolveBrowserExecPath(cfg), ShouldEqual, "/usr/local/bin/msedge")
		})

		Convey("When Browser is edge and only microsoft-edge is on PATH, returns it", func() {
			shared.LookPath = func(name string) (string, error) {
				if name == "microsoft-edge" {
					return "/usr/bin/microsoft-edge", nil
				}
				return "", fmt.Errorf("not found")
			}
			cfg := &config.Config{General: config.General{Browser: "edge"}}
			So(shared.ResolveBrowserExecPath(cfg), ShouldEqual, "/usr/bin/microsoft-edge")
		})

		Convey("When Browser is edge and nothing is on PATH, returns empty", func() {
			shared.LookPath = func(string) (string, error) { return "", fmt.Errorf("not found") }
			cfg := &config.Config{General: config.General{Browser: "edge"}}
			So(shared.ResolveBrowserExecPath(cfg), ShouldEqual, "")
		})

		Convey("When Browser is unknown, returns empty", func() {
			cfg := &config.Config{General: config.General{Browser: "firefox"}}
			So(shared.ResolveBrowserExecPath(cfg), ShouldEqual, "")
		})
	})
}

func TestValidateBrowserConfig(t *testing.T) {
	Convey("Given ValidateBrowserConfig", t, func() {
		origLookPath := shared.LookPath
		defer func() { shared.LookPath = origLookPath }()

		Convey("Returns nil for chrome", func() {
			cfg := &config.Config{General: config.General{Browser: "chrome"}}
			So(ValidateBrowserConfig(cfg), ShouldBeNil)
		})

		Convey("Returns nil for empty browser (chromedp default)", func() {
			cfg := &config.Config{General: config.General{Browser: ""}}
			So(ValidateBrowserConfig(cfg), ShouldBeNil)
		})

		Convey("Returns nil when BrowserPath is set regardless of Browser", func() {
			shared.LookPath = func(string) (string, error) { return "", fmt.Errorf("not found") }
			cfg := &config.Config{General: config.General{Browser: "edge", BrowserPath: "/custom/msedge"}}
			So(ValidateBrowserConfig(cfg), ShouldBeNil)
		})

		Convey("Returns nil when edge binary found on PATH", func() {
			shared.LookPath = func(name string) (string, error) {
				if name == "msedge" {
					return "/usr/bin/msedge", nil
				}
				return "", fmt.Errorf("not found")
			}
			cfg := &config.Config{General: config.General{Browser: "edge"}}
			So(ValidateBrowserConfig(cfg), ShouldBeNil)
		})

		Convey("Returns error when edge requested but no binary found", func() {
			shared.LookPath = func(string) (string, error) { return "", fmt.Errorf("not found") }
			cfg := &config.Config{General: config.General{Browser: "edge"}}
			err := ValidateBrowserConfig(cfg)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "edge")
		})
	})
}
