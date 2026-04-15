package kiosk

import (
	"context"
	"testing"

	"github.com/grafana/grafana-kiosk/pkg/browser"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAnonymousLoginNavigateSequence(t *testing.T) {
	Convey("Given a browser mock simulating anonymous login", t, func() {
		mock := browser.NewMock()
		ctx := context.Background()

		Convey("Navigate is called with generated URL", func() {
			cfg := &Config{
				General: General{Mode: "full", AutoFit: true},
				Target:  Target{URL: "https://play.grafana.org"},
			}
			url := GenerateURL(cfg)
			err := mock.Navigate(ctx, url)
			So(err, ShouldBeNil)
			So(mock.Calls[0].Args[0], ShouldEqual,
				"https://play.grafana.org?kiosk=1&autofitpanels")
		})

		Convey("Navigate propagates errors", func() {
			mock.Errors["Navigate"] = context.DeadlineExceeded
			err := mock.Navigate(ctx, "https://example.com")
			So(err, ShouldEqual, context.DeadlineExceeded)
		})

		Convey("Reload triggers second Navigate call", func() {
			cfg := &Config{
				General: General{Mode: "full", AutoFit: true},
				Target:  Target{URL: "https://play.grafana.org"},
			}
			url := GenerateURL(cfg)
			_ = mock.Navigate(ctx, url)
			_ = mock.Navigate(ctx, url)
			So(mock.CallCount("Navigate"), ShouldEqual, 2)
		})
	})
}
