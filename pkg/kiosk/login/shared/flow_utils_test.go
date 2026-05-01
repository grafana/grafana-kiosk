package shared

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/grafana/grafana-kiosk/pkg/browser/browsertest"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSleepPageLoad(t *testing.T) {
	Convey("Given SleepPageLoad", t, func() {
		Convey("Does not sleep when PageLoadDelayMS is zero", func() {
			cfg := &config.Config{General: config.General{PageLoadDelayMS: 0}}
			start := time.Now()
			SleepPageLoad(cfg)
			So(time.Since(start), ShouldBeLessThan, 50*time.Millisecond)
		})

		Convey("Does not sleep when PageLoadDelayMS is negative", func() {
			cfg := &config.Config{General: config.General{PageLoadDelayMS: -1}}
			start := time.Now()
			SleepPageLoad(cfg)
			So(time.Since(start), ShouldBeLessThan, 50*time.Millisecond)
		})

		Convey("Sleeps when PageLoadDelayMS is positive", func() {
			cfg := &config.Config{General: config.General{PageLoadDelayMS: 1}}
			start := time.Now()
			SleepPageLoad(cfg)
			So(time.Since(start), ShouldBeGreaterThanOrEqualTo, 1*time.Millisecond)
		})
	})
}

func TestRunMessageLoopInShared(t *testing.T) {
	Convey("Given RunMessageLoop", t, func() {
		mock := browsertest.NewMock()
		dashboardURL := "https://grafana.example.com/d/abc?kiosk=1"

		Convey("Exits cleanly on context cancel", func() {
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() { done <- RunMessageLoop(ctx, mock, dashboardURL, make(chan string)) }()
			cancel()
			So(<-done, ShouldBeNil)
			So(mock.CallCount("Navigate"), ShouldEqual, 0)
		})

		Convey("Navigates to dashboardURL on message", func() {
			ctx, cancel := context.WithCancel(context.Background())
			messages := make(chan string, 1)
			done := make(chan error, 1)
			go func() { done <- RunMessageLoop(ctx, mock, dashboardURL, messages) }()
			messages <- "reload"
			time.Sleep(10 * time.Millisecond)
			cancel()
			<-done
			So(mock.CallCount("Navigate"), ShouldEqual, 1)
			So(mock.CallsTo("Navigate")[0].Args[0], ShouldEqual, dashboardURL)
		})

		Convey("Returns error if Navigate fails on reload", func() {
			mock.Errors["Navigate"] = errors.New("browser crashed")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			messages := make(chan string, 1)
			done := make(chan error, 1)
			go func() { done <- RunMessageLoop(ctx, mock, dashboardURL, messages) }()
			messages <- "reload"
			err := <-done
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "browser crashed")
		})
	})
}
