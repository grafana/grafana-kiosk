package kiosk

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/grafana/grafana-kiosk/pkg/browser/browsertest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGenericOauthLoginFlow(t *testing.T) {
	baseCfg := func() *Config {
		return &Config{
			General: General{Mode: "full", AutoFit: true, PageLoadDelayMS: 0},
			Target:  Target{URL: "https://grafana.example.com/d/abc", Username: "admin", Password: "secret"},
			GoAuth:  GoAuth{UsernameField: "username", PasswordField: "password"},
		}
	}

	Convey("Given genericOauthLoginFlow without AutoLogin", t, func() {
		mock := browsertest.NewMock()
		cfg := baseCfg()
		url := GenerateURL(cfg)

		Convey("Returns error if Navigate fails", func() {
			mock.Errors["Navigate"] = errors.New("refused")
			err := genericOauthLoginFlow(context.Background(), cfg, mock, url, make(chan string))
			So(err, ShouldNotBeNil)
		})

		Convey("Full sequence: navigate → click oauth button → credentials", func() {
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() { done <- genericOauthLoginFlow(ctx, cfg, mock, url, make(chan string)) }()
			time.Sleep(10 * time.Millisecond)
			cancel()
			<-done

			So(mock.CallsTo("Navigate")[0].Args[0], ShouldEqual, url)
			So(mock.CallsTo("WaitVisible")[0].Args[0], ShouldContainSubstring, "generic_oauth")
			So(mock.CallsTo("Click")[0].Args[0], ShouldContainSubstring, "generic_oauth")
			So(mock.CallCount("SendKeys"), ShouldEqual, 2)
		})
	})

	Convey("Given genericOauthLoginFlow with AutoLogin", t, func() {
		mock := browsertest.NewMock()
		cfg := baseCfg()
		cfg.GoAuth.AutoLogin = true
		url := GenerateURL(cfg)

		Convey("Skips OAuth button click", func() {
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() { done <- genericOauthLoginFlow(ctx, cfg, mock, url, make(chan string)) }()
			time.Sleep(10 * time.Millisecond)
			cancel()
			<-done

			// No Click on generic_oauth button
			for _, c := range mock.CallsTo("Click") {
				So(c.Args[0], ShouldNotContainSubstring, "generic_oauth")
			}
		})
	})

	Convey("Given genericOauthLoginFlow with WaitForPasswordField", t, func() {
		mock := browsertest.NewMock()
		cfg := baseCfg()
		cfg.GoAuth.WaitForPasswordField = true
		cfg.GoAuth.AutoLogin = true
		url := GenerateURL(cfg)

		Convey("Sends username with Enter then waits for password field", func() {
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() { done <- genericOauthLoginFlow(ctx, cfg, mock, url, make(chan string)) }()
			time.Sleep(10 * time.Millisecond)
			cancel()
			<-done

			sendCalls := mock.CallsTo("SendKeys")
			So(sendCalls, ShouldHaveLength, 2)
			// username sent with Enter appended
			So(sendCalls[0].Args[1], ShouldContainSubstring, cfg.Target.Username)
		})
	})

	Convey("Given genericOauthLoginFlow with WaitForStaySignedInPrompt", t, func() {
		mock := browsertest.NewMock()
		cfg := baseCfg()
		cfg.GoAuth.AutoLogin = true
		cfg.GoAuth.WaitForStaySignedInPrompt = true
		url := GenerateURL(cfg)

		Convey("Clicks Yes button after credentials", func() {
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() { done <- genericOauthLoginFlow(ctx, cfg, mock, url, make(chan string)) }()
			time.Sleep(10 * time.Millisecond)
			cancel()
			<-done

			clickCalls := mock.CallsTo("Click")
			So(clickCalls[len(clickCalls)-1].Args[0], ShouldContainSubstring, `value="Yes"`)
		})
	})
}
