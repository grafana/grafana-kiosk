package kiosk

import (
	"context"
	"testing"
	"time"

	"github.com/grafana/grafana-kiosk/pkg/browser/browsertest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestIdtokenLoginFlow(t *testing.T) {
	Convey("Given idtokenLoginFlow", t, func() {
		mock := browsertest.NewMock()
		dashboardURL := "https://grafana.example.com/d/abc?kiosk=1"

		Convey("Exits cleanly on context cancel", func() {
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() { done <- idtokenLoginFlow(ctx, mock, dashboardURL, make(chan string)) }()
			cancel()
			So(<-done, ShouldBeNil)
		})

		Convey("Reloads on message", func() {
			ctx, cancel := context.WithCancel(context.Background())
			messages := make(chan string, 1)
			done := make(chan error, 1)
			go func() { done <- idtokenLoginFlow(ctx, mock, dashboardURL, messages) }()
			messages <- "reload"
			time.Sleep(10 * time.Millisecond)
			cancel()
			<-done
			So(mock.CallCount("Navigate"), ShouldEqual, 1)
			So(mock.CallsTo("Navigate")[0].Args[0], ShouldEqual, dashboardURL)
		})
	})
}
