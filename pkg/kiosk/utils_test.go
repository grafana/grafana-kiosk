package kiosk

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"unsafe"

	"github.com/chromedp/chromedp"
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
		Convey("Playlist mode adds inactive parameter", func() {
			conf := GetConfig()
			conf.General.Mode = "full"
			conf.Target.IsPlayList = true
			anURL := GenerateURL(&conf)
			So(anURL, ShouldEqual, "https://play.grafana/com?inactive=1&kiosk=1&autofitpanels")
		})
		Convey("Playlist mode without autofit", func() {
			conf := GetConfig()
			conf.General.Mode = "tv"
			conf.General.AutoFit = false
			conf.Target.IsPlayList = true
			anURL := GenerateURL(&conf)
			So(anURL, ShouldEqual, "https://play.grafana/com?inactive=1&kiosk=tv")
		})
	})
}

// applyOptions applies ExecAllocatorOption functions to an ExecAllocator and
// returns the resulting initFlags map. This uses reflection to access the
// unexported initFlags field for testing purposes.
func applyOptions(opts []chromedp.ExecAllocatorOption) map[string]interface{} {
	var alloc chromedp.ExecAllocator
	// Initialize the unexported initFlags map via reflection
	v := reflect.ValueOf(&alloc).Elem()
	flagsField := v.FieldByName("initFlags")
	flagsPtr := unsafe.Pointer(flagsField.UnsafeAddr())
	initFlags := (*map[string]interface{})(flagsPtr)
	*initFlags = make(map[string]interface{})
	// Apply all options
	for _, opt := range opts {
		opt(&alloc)
	}
	return *initFlags
}

// TestGenerateExecutorOptions tests the generateExecutorOptions function
func TestGenerateExecutorOptions(t *testing.T) {
	Convey("Given executor option generation", t, func() {
		Convey("When using default config with GPU disabled", func() {
			cfg := &Config{
				BuildInfo: BuildInfo{Version: "v1.0.0"},
				General: General{
					Incognito:      true,
					WindowPosition: "0,0",
				},
				Target: Target{
					IgnoreCertificateErrors: false,
				},
			}
			opts := generateExecutorOptions("/tmp/test", cfg)
			flags := applyOptions(opts)

			Convey("Should set base Chrome flags", func() {
				So(flags["autoplay-policy"], ShouldEqual, "no-user-gesture-required")
				So(flags["bwsi"], ShouldEqual, true)
				So(flags["check-for-update-interval"], ShouldEqual, "31536000")
				So(flags["password-store"], ShouldEqual, "basic")
				So(flags["disable-features"], ShouldEqual, "Translate")
				So(flags["disable-notifications"], ShouldEqual, true)
				So(flags["disable-overlay-scrollbar"], ShouldEqual, true)
				So(flags["hide-scrollbars"], ShouldEqual, true)
				So(flags["disable-search-engine-choice-screen"], ShouldEqual, true)
				So(flags["disable-sync"], ShouldEqual, true)
				So(flags["incognito"], ShouldEqual, true)
				So(flags["kiosk"], ShouldEqual, true)
				So(flags["noerrdialogs"], ShouldEqual, true)
				So(flags["start-fullscreen"], ShouldEqual, true)
				So(flags["start-maximized"], ShouldEqual, true)
				So(flags["window-position"], ShouldEqual, "0,0")
			})

			Convey("Should set user data dir", func() {
				So(flags["user-data-dir"], ShouldEqual, "/tmp/test")
			})

			Convey("Should set ignore-certificate-errors to false", func() {
				So(flags["ignore-certificate-errors"], ShouldEqual, false)
			})

			Convey("Should disable GPU when GPUEnabled is false", func() {
				So(flags["disable-gpu"], ShouldEqual, false)
			})
		})

		Convey("When GPU is enabled", func() {
			cfg := &Config{
				BuildInfo: BuildInfo{Version: "v2.0.0"},
				General: General{
					GPUEnabled:     true,
					WindowPosition: "0,0",
				},
			}
			opts := generateExecutorOptions("/tmp/test", cfg)
			flags := applyOptions(opts)

			Convey("Should not set disable-gpu flag", func() {
				_, exists := flags["disable-gpu"]
				So(exists, ShouldBeFalse)
			})
		})

		Convey("When ozone platform is set", func() {
			cfg := &Config{
				BuildInfo: BuildInfo{Version: "v1.0.0"},
				General: General{
					OzonePlatform:  "wayland",
					WindowPosition: "0,0",
				},
			}
			opts := generateExecutorOptions("/tmp/test", cfg)
			flags := applyOptions(opts)

			Convey("Should set ozone-platform flag", func() {
				So(flags["ozone-platform"], ShouldEqual, "wayland")
			})
		})

		Convey("When ozone platform is empty", func() {
			cfg := &Config{
				BuildInfo: BuildInfo{Version: "v1.0.0"},
				General: General{
					OzonePlatform:  "",
					WindowPosition: "0,0",
				},
			}
			opts := generateExecutorOptions("/tmp/test", cfg)
			flags := applyOptions(opts)

			Convey("Should not set ozone-platform flag", func() {
				_, exists := flags["ozone-platform"]
				So(exists, ShouldBeFalse)
			})
		})

		Convey("When window size is set", func() {
			cfg := &Config{
				BuildInfo: BuildInfo{Version: "v1.0.0"},
				General: General{
					WindowSize:     "1920,1080",
					WindowPosition: "0,0",
				},
			}
			opts := generateExecutorOptions("/tmp/test", cfg)
			flags := applyOptions(opts)

			Convey("Should override kiosk to false", func() {
				So(flags["kiosk"], ShouldEqual, false)
			})

			Convey("Should override start-fullscreen to false", func() {
				So(flags["start-fullscreen"], ShouldEqual, false)
			})

			Convey("Should set app mode", func() {
				So(flags["app"], ShouldEqual, "data:text/html,<title>Grafana</title>")
			})

			Convey("Should set window-size flag", func() {
				So(flags["window-size"], ShouldEqual, "1920,1080")
			})
		})

		Convey("When window size is empty", func() {
			cfg := &Config{
				BuildInfo: BuildInfo{Version: "v1.0.0"},
				General: General{
					WindowSize:     "",
					WindowPosition: "0,0",
				},
			}
			opts := generateExecutorOptions("/tmp/test", cfg)
			flags := applyOptions(opts)

			Convey("Should keep kiosk as true", func() {
				So(flags["kiosk"], ShouldEqual, true)
			})

			Convey("Should keep start-fullscreen as true", func() {
				So(flags["start-fullscreen"], ShouldEqual, true)
			})

			Convey("Should not set app flag", func() {
				_, exists := flags["app"]
				So(exists, ShouldBeFalse)
			})

			Convey("Should not set window-size flag", func() {
				_, exists := flags["window-size"]
				So(exists, ShouldBeFalse)
			})
		})

		Convey("When scale factor is set", func() {
			cfg := &Config{
				BuildInfo: BuildInfo{Version: "v1.0.0"},
				General: General{
					ScaleFactor:    "1.5",
					WindowPosition: "0,0",
				},
			}
			opts := generateExecutorOptions("/tmp/test", cfg)
			flags := applyOptions(opts)

			Convey("Should set force-device-scale-factor flag", func() {
				So(flags["force-device-scale-factor"], ShouldEqual, "1.5")
			})
		})

		Convey("When scale factor is empty", func() {
			cfg := &Config{
				BuildInfo: BuildInfo{Version: "v1.0.0"},
				General: General{
					ScaleFactor:    "",
					WindowPosition: "0,0",
				},
			}
			opts := generateExecutorOptions("/tmp/test", cfg)
			flags := applyOptions(opts)

			Convey("Should not set force-device-scale-factor flag", func() {
				_, exists := flags["force-device-scale-factor"]
				So(exists, ShouldBeFalse)
			})
		})

		Convey("When ignore certificate errors is true", func() {
			cfg := &Config{
				BuildInfo: BuildInfo{Version: "v1.0.0"},
				General: General{
					WindowPosition: "0,0",
				},
				Target: Target{
					IgnoreCertificateErrors: true,
				},
			}
			opts := generateExecutorOptions("/tmp/test", cfg)
			flags := applyOptions(opts)

			Convey("Should set ignore-certificate-errors to true", func() {
				So(flags["ignore-certificate-errors"], ShouldEqual, true)
			})
		})

		Convey("When version has v prefix", func() {
			cfg := &Config{
				BuildInfo: BuildInfo{Version: "v1.2.3"},
				General: General{
					WindowPosition: "0,0",
				},
			}
			opts := generateExecutorOptions("/tmp/test", cfg)
			flags := applyOptions(opts)

			Convey("Should strip v prefix from user agent", func() {
				userAgent := flags["user-agent"].(string)
				expectedKioskVersion := fmt.Sprintf("GrafanaKiosk/1.2.3 (%s %s)", runtime.GOOS, runtime.GOARCH)
				So(userAgent, ShouldContainSubstring, expectedKioskVersion)
				So(userAgent, ShouldNotContainSubstring, "GrafanaKiosk/v")
			})
		})

		Convey("When version has no v prefix", func() {
			cfg := &Config{
				BuildInfo: BuildInfo{Version: "1.2.3"},
				General: General{
					WindowPosition: "0,0",
				},
			}
			opts := generateExecutorOptions("/tmp/test", cfg)
			flags := applyOptions(opts)

			Convey("Should use version as-is in user agent", func() {
				userAgent := flags["user-agent"].(string)
				expectedKioskVersion := fmt.Sprintf("GrafanaKiosk/1.2.3 (%s %s)", runtime.GOOS, runtime.GOARCH)
				So(userAgent, ShouldContainSubstring, expectedKioskVersion)
			})
		})

		Convey("When user agent format is checked", func() {
			cfg := &Config{
				BuildInfo: BuildInfo{Version: "v3.0.0"},
				General: General{
					WindowPosition: "0,0",
				},
			}
			opts := generateExecutorOptions("/tmp/test", cfg)
			flags := applyOptions(opts)

			Convey("Should start with Chrome user agent string", func() {
				userAgent := flags["user-agent"].(string)
				So(userAgent, ShouldStartWith, "Mozilla/5.0")
				So(userAgent, ShouldContainSubstring, "Chrome/")
				So(userAgent, ShouldContainSubstring, "Safari/")
			})
		})

		Convey("When all optional flags are set", func() {
			cfg := &Config{
				BuildInfo: BuildInfo{Version: "v1.0.0"},
				General: General{
					GPUEnabled:     false,
					OzonePlatform:  "x11",
					WindowSize:     "800,600",
					ScaleFactor:    "2.0",
					WindowPosition: "100,200",
				},
				Target: Target{
					IgnoreCertificateErrors: true,
				},
			}
			opts := generateExecutorOptions("/tmp/test", cfg)
			flags := applyOptions(opts)

			Convey("Should set disable-gpu", func() {
				So(flags["disable-gpu"], ShouldEqual, false)
			})

			Convey("Should set ozone-platform", func() {
				So(flags["ozone-platform"], ShouldEqual, "x11")
			})

			Convey("Should override kiosk and fullscreen for window size", func() {
				So(flags["kiosk"], ShouldEqual, false)
				So(flags["start-fullscreen"], ShouldEqual, false)
				So(flags["app"], ShouldEqual, "data:text/html,<title>Grafana</title>")
				So(flags["window-size"], ShouldEqual, "800,600")
			})

			Convey("Should set scale factor", func() {
				So(flags["force-device-scale-factor"], ShouldEqual, "2.0")
			})

			Convey("Should set window position", func() {
				So(flags["window-position"], ShouldEqual, "100,200")
			})

			Convey("Should set ignore-certificate-errors", func() {
				So(flags["ignore-certificate-errors"], ShouldEqual, true)
			})
		})

		Convey("When no-first-run and no-default-browser-check are set", func() {
			cfg := &Config{
				BuildInfo: BuildInfo{Version: "v1.0.0"},
				General: General{
					WindowPosition: "0,0",
				},
			}
			opts := generateExecutorOptions("/tmp/test", cfg)
			flags := applyOptions(opts)

			Convey("Should set no-first-run", func() {
				So(flags["no-first-run"], ShouldEqual, true)
			})

			Convey("Should set no-default-browser-check", func() {
				So(flags["no-default-browser-check"], ShouldEqual, true)
			})
		})

		Convey("When version string contains build metadata", func() {
			cfg := &Config{
				BuildInfo: BuildInfo{Version: "v1.0.0-31-gabcdef"},
				General: General{
					WindowPosition: "0,0",
				},
			}
			opts := generateExecutorOptions("/tmp/test", cfg)
			flags := applyOptions(opts)

			Convey("Should strip v prefix and preserve build metadata", func() {
				userAgent := flags["user-agent"].(string)
				So(userAgent, ShouldContainSubstring, "GrafanaKiosk/1.0.0-31-gabcdef")
				So(strings.Contains(userAgent, "GrafanaKiosk/v"), ShouldBeFalse)
			})
		})

		Convey("When incognito is disabled", func() {
			cfg := &Config{
				BuildInfo: BuildInfo{Version: "v1.0.0"},
				General: General{
					Incognito:      false,
					WindowPosition: "0,0",
				},
			}
			opts := generateExecutorOptions("/tmp/test", cfg)
			flags := applyOptions(opts)

			Convey("Should set incognito to false", func() {
				So(flags["incognito"], ShouldEqual, false)
			})
		})
	})
}
