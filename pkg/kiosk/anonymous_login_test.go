package kiosk

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/grafana/grafana-kiosk/pkg/browser"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAnonymousLoginFlow(t *testing.T) {
	Convey("Given anonymousLoginFlow", t, func() {
		mock := browser.NewMock()
		cfg := &Config{General: General{PageLoadDelayMS: 0}}
		url := "https://play.grafana.org?kiosk=1&autofitpanels"

		Convey("Navigates to the provided URL", func() {
			mock.Errors["Navigate"] = errors.New("stop")
			_ = anonymousLoginFlow(context.Background(), cfg, mock, url, make(chan string))
			So(mock.Calls[0].Args[0], ShouldEqual, url)
		})

		Convey("Returns error if initial Navigate fails", func() {
			mock.Errors["Navigate"] = errors.New("connection refused")
			err := anonymousLoginFlow(context.Background(), cfg, mock, url, make(chan string))
			So(err, ShouldNotBeNil)
			So(mock.CallCount("Navigate"), ShouldEqual, 1)
		})

		Convey("Exits cleanly on context cancel", func() {
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() { done <- anonymousLoginFlow(ctx, cfg, mock, url, make(chan string)) }()
			cancel()
			So(<-done, ShouldBeNil)
		})

		Convey("Navigates again on message from browser", func() {
			ctx, cancel := context.WithCancel(context.Background())
			messages := make(chan string, 1)
			done := make(chan error, 1)
			go func() { done <- anonymousLoginFlow(ctx, cfg, mock, url, messages) }()
			messages <- "console event"
			time.Sleep(10 * time.Millisecond)
			cancel()
			<-done
			So(mock.CallCount("Navigate"), ShouldEqual, 2)
			So(mock.CallsTo("Navigate")[1].Args[0], ShouldEqual, url)
		})
	})
}
