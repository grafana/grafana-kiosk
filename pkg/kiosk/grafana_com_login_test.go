package kiosk

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/grafana/grafana-kiosk/pkg/browser/browsertest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGcomLoginFlow(t *testing.T) {
	baseCfg := func() *Config {
		return &Config{
			General: General{Mode: "full", AutoFit: true, PageLoadDelayMS: 0},
			Target:  Target{URL: "https://grafana.example.com/d/abc", Username: "user@example.com", Password: "secret"},
		}
	}

	Convey("Given gcomLoginFlow", t, func() {
		mock := browsertest.NewMock()
		cfg := baseCfg()
		dashboardURL := "https://grafana.example.com/d/abc?kiosk=1&autofitpanels"

		Convey("Returns error if initial Navigate fails", func() {
			mock.Errors["Navigate"] = errors.New("refused")
			err := gcomLoginFlow(context.Background(), cfg, mock, dashboardURL, make(chan string))
			So(err, ShouldNotBeNil)
			So(mock.CallCount("Navigate"), ShouldEqual, 1)
		})

		Convey("Returns error if WaitVisible fails", func() {
			mock.Errors["WaitVisible"] = errors.New("timeout")
			err := gcomLoginFlow(context.Background(), cfg, mock, dashboardURL, make(chan string))
			So(err, ShouldNotBeNil)
		})

		Convey("Returns error if Click fails", func() {
			mock.Errors["Click"] = errors.New("not found")
			err := gcomLoginFlow(context.Background(), cfg, mock, dashboardURL, make(chan string))
			So(err, ShouldNotBeNil)
			So(mock.CallCount("Navigate"), ShouldEqual, 1)
			So(mock.CallCount("WaitVisible"), ShouldEqual, 1)
		})

		Convey("Full sequence: navigate → click gcom button → credentials → message loop", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			messages := make(chan string, 1)
			done := make(chan error, 1)
			go func() { done <- gcomLoginFlow(ctx, cfg, mock, dashboardURL, messages) }()
			time.Sleep(10 * time.Millisecond)
			messages <- "reload"
			time.Sleep(10 * time.Millisecond)
			cancel()
			<-done

			So(mock.CallsTo("Navigate")[0].Args[0], ShouldEqual, dashboardURL)
			So(mock.CallsTo("WaitVisible")[0].Args[0], ShouldContainSubstring, "login/grafana_com")
			So(mock.CallsTo("Click")[0].Args[0], ShouldContainSubstring, "login/grafana_com")
			So(mock.CallsTo("WaitVisible")[1].Args[0], ShouldContainSubstring, `name="login"`)
			So(mock.CallsTo("SendKeys")[0].Args[1], ShouldEqual, cfg.Target.Username)
			So(mock.CallsTo("Click")[1].Args[0], ShouldEqual, `#submit`)
			// Reload Navigate is the second Navigate call
			So(mock.CallCount("Navigate"), ShouldEqual, 2)
		})
	})
}
