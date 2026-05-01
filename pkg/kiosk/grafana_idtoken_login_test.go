package kiosk

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/grafana/grafana-kiosk/pkg/browser/browsertest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestIdtokenLoginFlow(t *testing.T) {
	Convey("Given idtokenLoginFlow", t, func() {
		mock := browsertest.NewMock()
		cfg := &Config{General: General{PageLoadDelayMS: 0}}
		url := "https://grafana.example.com/d/abc?kiosk=1"

		Convey("Navigates to provided URL", func() {
			mock.Errors["Navigate"] = errors.New("stop")
			_ = idtokenLoginFlow(context.Background(), cfg, mock, url, make(chan string))
			So(mock.Calls[0].Args[0], ShouldEqual, url)
		})

		Convey("Returns error if Navigate fails", func() {
			mock.Errors["Navigate"] = errors.New("refused")
			err := idtokenLoginFlow(context.Background(), cfg, mock, url, make(chan string))
			So(err, ShouldNotBeNil)
		})

		Convey("Exits cleanly on context cancel", func() {
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() { done <- idtokenLoginFlow(ctx, cfg, mock, url, make(chan string)) }()
			cancel()
			So(<-done, ShouldBeNil)
		})

		Convey("Reloads on message", func() {
			ctx, cancel := context.WithCancel(context.Background())
			messages := make(chan string, 1)
			done := make(chan error, 1)
			go func() { done <- idtokenLoginFlow(ctx, cfg, mock, url, messages) }()
			messages <- "reload"
			time.Sleep(10 * time.Millisecond)
			cancel()
			<-done
			So(mock.CallCount("Navigate"), ShouldEqual, 2)
		})
	})
}
