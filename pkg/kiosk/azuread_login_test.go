package kiosk

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/grafana/grafana-kiosk/pkg/browser/browsertest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAzureADLoginFlow(t *testing.T) {
	baseCfg := func() *Config {
		return &Config{
			General: General{Mode: "full", AutoFit: true, PageLoadDelayMS: 0},
			Target:  Target{URL: "https://grafana.example.com/d/abc", Username: "user@example.com", Password: "secret"},
		}
	}

	Convey("Given azureADLoginFlow", t, func() {
		mock := browsertest.NewMock()
		cfg := baseCfg()
		url := GenerateURL(cfg)

		Convey("Returns error if Navigate fails", func() {
			mock.Errors["Navigate"] = errors.New("refused")
			err := azureADLoginFlow(context.Background(), cfg, mock, url, make(chan string))
			So(err, ShouldNotBeNil)
		})

		Convey("Full sequence: navigate → azuread button → email → password → sign in", func() {
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() { done <- azureADLoginFlow(ctx, cfg, mock, url, make(chan string)) }()
			time.Sleep(10 * time.Millisecond)
			cancel()
			<-done

			So(mock.CallsTo("Navigate")[0].Args[0], ShouldEqual, url)
			waitCalls := mock.CallsTo("WaitVisible")
			So(waitCalls[0].Args[0], ShouldContainSubstring, "login/azuread")
			So(waitCalls[1].Args[0], ShouldContainSubstring, "loginfmt")
			So(waitCalls[2].Args[0], ShouldContainSubstring, "passwd")

			sendCalls := mock.CallsTo("SendKeys")
			So(sendCalls[0].Args[1], ShouldEqual, cfg.Target.Username)
			So(sendCalls[1].Args[1], ShouldEqual, cfg.Target.Password)

			// idSIButton9 clicked twice (after email and after password)
			clickCalls := mock.CallsTo("Click")
			So(clickCalls[0].Args[0], ShouldContainSubstring, "login/azuread")
			So(clickCalls[1].Args[0], ShouldContainSubstring, "idSIButton9")
		})

		Convey("Returns error if WaitVisible fails", func() {
			mock.Errors["WaitVisible"] = errors.New("timeout")
			err := azureADLoginFlow(context.Background(), cfg, mock, url, make(chan string))
			So(err, ShouldNotBeNil)
			So(mock.CallCount("Navigate"), ShouldEqual, 1)
		})

		Convey("Returns error if SendKeys fails", func() {
			mock.Errors["SendKeys"] = errors.New("element not found")
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() { done <- azureADLoginFlow(ctx, cfg, mock, url, make(chan string)) }()
			time.Sleep(10 * time.Millisecond)
			cancel()
			err := <-done
			So(err, ShouldNotBeNil)
		})

		Convey("Exits cleanly on context cancel", func() {
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() { done <- azureADLoginFlow(ctx, cfg, mock, url, make(chan string)) }()
			cancel()
			So(<-done, ShouldBeNil)
		})

		// Reload test waits for the full flow including 3x1s hardcoded sleeps.
		Convey("Reloads on message after full login sequence", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			messages := make(chan string, 1)
			done := make(chan error, 1)
			go func() { done <- azureADLoginFlow(ctx, cfg, mock, url, messages) }()
			time.Sleep(4 * time.Second)
			messages <- "reload"
			time.Sleep(10 * time.Millisecond)
			cancel()
			<-done
			So(mock.CallCount("Navigate"), ShouldEqual, 2)
		})
	})
}
