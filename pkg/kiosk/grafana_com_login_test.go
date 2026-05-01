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
		url := "https://grafana.example.com/d/abc?kiosk=1&autofitpanels"

		Convey("Returns error if initial Navigate fails", func() {
			mock.Errors["Navigate"] = errors.New("refused")
			err := gcomLoginFlow(context.Background(), cfg, mock, url, make(chan string))
			So(err, ShouldNotBeNil)
			So(mock.CallCount("Navigate"), ShouldEqual, 1)
		})

		Convey("Returns error if WaitVisible fails", func() {
			mock.Errors["WaitVisible"] = errors.New("timeout")
			err := gcomLoginFlow(context.Background(), cfg, mock, url, make(chan string))
			So(err, ShouldNotBeNil)
		})

		Convey("Full sequence: navigate → click gcom button → credentials", func() {
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() { done <- gcomLoginFlow(ctx, cfg, mock, url, make(chan string)) }()
			time.Sleep(10 * time.Millisecond)
			cancel()
			<-done

			So(mock.CallsTo("Navigate")[0].Args[0], ShouldEqual, url)
			So(mock.CallsTo("WaitVisible")[0].Args[0], ShouldContainSubstring, "login/grafana_com")
			So(mock.CallsTo("Click")[0].Args[0], ShouldContainSubstring, "login/grafana_com")
			So(mock.CallsTo("WaitVisible")[1].Args[0], ShouldContainSubstring, `name="login"`)
		})

		Convey("Reloads on message", func() {
			ctx, cancel := context.WithCancel(context.Background())
			messages := make(chan string, 1)
			done := make(chan error, 1)
			go func() { done <- gcomLoginFlow(ctx, cfg, mock, url, messages) }()
			time.Sleep(10 * time.Millisecond)
			messages <- "reload"
			time.Sleep(10 * time.Millisecond)
			cancel()
			<-done
			So(mock.CallCount("Navigate"), ShouldBeGreaterThanOrEqualTo, 2)
		})
	})
}
