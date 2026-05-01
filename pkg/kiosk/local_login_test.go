package kiosk

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/grafana/grafana-kiosk/pkg/browser"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLocalLoginBypassURL(t *testing.T) {
	Convey("Given a URL for local login bypass", t, func() {
		Convey("When URL has a path", func() {
			result := LocalLoginBypassURL("https://grafana.example.com/dashboard/db/test")
			So(result, ShouldEqual, "https://grafana.example.com/login/local")
		})

		Convey("When URL has no path", func() {
			result := LocalLoginBypassURL("https://grafana.example.com")
			So(result, ShouldEqual, "https://grafana.example.com/login/local")
		})

		Convey("When URL has a port", func() {
			result := LocalLoginBypassURL("https://localhost:3000/d/abc123")
			So(result, ShouldEqual, "https://localhost:3000/login/local")
		})

		Convey("When URL uses http", func() {
			result := LocalLoginBypassURL("http://grafana.local/playlists/play/1")
			So(result, ShouldEqual, "http://grafana.local/login/local")
		})

		Convey("When URL has a deep path", func() {
			result := LocalLoginBypassURL("https://bkgann3.grafana.net/dashboard/db/sensu-summary")
			So(result, ShouldEqual, "https://bkgann3.grafana.net/login/local")
		})
	})
}

func TestLoginWithCredentials(t *testing.T) {
	Convey("Given loginWithCredentials", t, func() {
		mock := browser.NewMock()
		ctx := context.Background()

		Convey("Should wait for user field then send credentials", func() {
			err := loginWithCredentials(ctx, mock, "admin", "secret")
			So(err, ShouldBeNil)
			So(mock.Calls, ShouldHaveLength, 3)
			So(mock.Calls[0].Method, ShouldEqual, "WaitVisible")
			So(mock.Calls[0].Args[0], ShouldEqual, `//input[@name="user"]`)
			So(mock.Calls[1].Method, ShouldEqual, "SendKeys")
			So(mock.Calls[1].Args[0], ShouldEqual, `//input[@name="user"]`)
			So(mock.Calls[1].Args[1], ShouldEqual, "admin")
			So(mock.Calls[2].Method, ShouldEqual, "SendKeys")
			So(mock.Calls[2].Args[0], ShouldEqual, `//input[@name="password"]`)
			So(mock.Calls[2].Args[1], ShouldContainSubstring, "secret")
		})

		Convey("Should propagate WaitVisible error", func() {
			mock.Errors["WaitVisible"] = context.DeadlineExceeded
			err := loginWithCredentials(ctx, mock, "admin", "secret")
			So(err, ShouldEqual, context.DeadlineExceeded)
			So(mock.Calls, ShouldHaveLength, 1)
		})

		Convey("Should propagate SendKeys error", func() {
			mock.Errors["SendKeys"] = context.DeadlineExceeded
			err := loginWithCredentials(ctx, mock, "admin", "secret")
			So(err, ShouldEqual, context.DeadlineExceeded)
		})
	})
}

func TestLocalLoginFlow(t *testing.T) {
	baseCfg := func() *Config {
		return &Config{
			General: General{Mode: "full", AutoFit: true, PageLoadDelayMS: 0},
			Target:  Target{URL: "https://grafana.example.com/d/abc/dashboard", Username: "admin", Password: "secret"},
		}
	}

	Convey("Given localLoginFlow with AutoLogin", t, func() {
		mock := browser.NewMock()
		cfg := baseCfg()
		cfg.GoAuth = GoAuth{AutoLogin: true}
		generatedURL := GenerateURL(cfg)

		Convey("Full sequence: bypass URL → credentials → avatar wait → dashboard", func() {
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() { done <- localLoginFlow(ctx, cfg, mock, generatedURL, make(chan string)) }()
			time.Sleep(10 * time.Millisecond)
			cancel()
			<-done

			nav := mock.CallsTo("Navigate")
			So(nav, ShouldHaveLength, 2)
			So(nav[0].Args[0], ShouldEqual, "https://grafana.example.com/login/local")
			So(nav[1].Args[0], ShouldContainSubstring, "kiosk=1")
			So(mock.CallsTo("WaitVisible"), ShouldHaveLength, 2) // user field + avatar
			So(mock.CallsTo("SendKeys"), ShouldHaveLength, 2)    // username + password
		})

		Convey("Returns error if Navigate to bypass URL fails", func() {
			mock.Errors["Navigate"] = errors.New("refused")
			err := localLoginFlow(context.Background(), cfg, mock, generatedURL, make(chan string))
			So(err, ShouldNotBeNil)
			So(mock.CallsTo("Navigate")[0].Args[0], ShouldEqual, "https://grafana.example.com/login/local")
		})
	})

	Convey("Given localLoginFlow without AutoLogin", t, func() {
		mock := browser.NewMock()
		cfg := baseCfg()
		generatedURL := GenerateURL(cfg)

		Convey("Full sequence: dashboard URL → credentials", func() {
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() { done <- localLoginFlow(ctx, cfg, mock, generatedURL, make(chan string)) }()
			time.Sleep(10 * time.Millisecond)
			cancel()
			<-done

			nav := mock.CallsTo("Navigate")
			So(nav, ShouldHaveLength, 1)
			So(nav[0].Args[0], ShouldContainSubstring, "kiosk=1")
			So(mock.CallsTo("WaitVisible"), ShouldHaveLength, 1) // user field only
			So(mock.CallsTo("SendKeys"), ShouldHaveLength, 2)
		})

		Convey("Returns error if Navigate fails", func() {
			mock.Errors["Navigate"] = errors.New("refused")
			err := localLoginFlow(context.Background(), cfg, mock, generatedURL, make(chan string))
			So(err, ShouldNotBeNil)
		})
	})
}
