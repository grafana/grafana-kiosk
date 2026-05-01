package shared

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/grafana/grafana-kiosk/pkg/kiosk/config"
)

func TestGenerateURL(t *testing.T) {
	Convey("Given GenerateURL", t, func() {
		base := "https://play.grafana.org"

		Convey("Full kiosk mode adds kiosk=1", func() {
			cfg := &config.Config{General: config.General{Mode: "full"}, Target: config.Target{URL: base}}
			So(GenerateURL(cfg), ShouldContainSubstring, "kiosk=1")
		})

		Convey("TV kiosk mode adds kiosk=tv", func() {
			cfg := &config.Config{General: config.General{Mode: "tv"}, Target: config.Target{URL: base}}
			So(GenerateURL(cfg), ShouldContainSubstring, "kiosk=tv")
		})

		Convey("Disabled kiosk mode adds no kiosk param", func() {
			cfg := &config.Config{General: config.General{Mode: "disabled"}, Target: config.Target{URL: base}}
			result := GenerateURL(cfg)
			So(result, ShouldNotContainSubstring, "kiosk")
		})

		Convey("AutoFit appends autofitpanels", func() {
			cfg := &config.Config{General: config.General{Mode: "full", AutoFit: true}, Target: config.Target{URL: base}}
			So(GenerateURL(cfg), ShouldContainSubstring, "autofitpanels")
		})

		Convey("HideLinks adds _dash.hideLinks", func() {
			cfg := &config.Config{General: config.General{Mode: "full", HideLinks: true}, Target: config.Target{URL: base}}
			So(GenerateURL(cfg), ShouldContainSubstring, "_dash.hideLinks")
		})

		Convey("HideLogo adds hideLogo", func() {
			cfg := &config.Config{General: config.General{Mode: "full", HideLogo: true}, Target: config.Target{URL: base}}
			So(GenerateURL(cfg), ShouldContainSubstring, "hideLogo")
		})

		Convey("IsPlayList adds inactive=1", func() {
			cfg := &config.Config{General: config.General{Mode: "full"}, Target: config.Target{URL: base, IsPlayList: true}}
			So(GenerateURL(cfg), ShouldContainSubstring, "inactive=1")
		})

		Convey("Preserves base URL host and path", func() {
			cfg := &config.Config{General: config.General{Mode: "full"}, Target: config.Target{URL: "https://example.com/d/abc"}}
			result := GenerateURL(cfg)
			So(result, ShouldStartWith, "https://example.com/d/abc")
		})
	})
}

func TestResolveBrowserExecPathInShared(t *testing.T) {
	Convey("Given ResolveBrowserExecPath", t, func() {
		origLookPath := LookPath
		defer func() { LookPath = origLookPath }()

		Convey("BrowserPath set wins over Browser", func() {
			LookPath = func(string) (string, error) { return "/unused", nil }
			cfg := &config.Config{General: config.General{Browser: "edge", BrowserPath: "/custom/msedge"}}
			So(ResolveBrowserExecPath(cfg), ShouldEqual, "/custom/msedge")
		})

		Convey("Empty Browser returns empty (chromedp default)", func() {
			cfg := &config.Config{General: config.General{Browser: ""}}
			So(ResolveBrowserExecPath(cfg), ShouldEqual, "")
		})

		Convey("Browser=chrome returns empty", func() {
			cfg := &config.Config{General: config.General{Browser: "chrome"}}
			So(ResolveBrowserExecPath(cfg), ShouldEqual, "")
		})

		Convey("Browser=edge with msedge on PATH returns path", func() {
			LookPath = func(name string) (string, error) {
				if name == "msedge" {
					return "/usr/bin/msedge", nil
				}
				return "", fmt.Errorf("not found")
			}
			cfg := &config.Config{General: config.General{Browser: "edge"}}
			So(ResolveBrowserExecPath(cfg), ShouldEqual, "/usr/bin/msedge")
		})

		Convey("Browser=edge with no binary returns empty", func() {
			LookPath = func(string) (string, error) { return "", fmt.Errorf("not found") }
			cfg := &config.Config{General: config.General{Browser: "edge"}}
			So(ResolveBrowserExecPath(cfg), ShouldEqual, "")
		})
	})
}
