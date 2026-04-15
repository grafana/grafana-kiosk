package kiosk

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/grafana/grafana-kiosk/pkg/browser"
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

func TestLocalLoginAutoLoginFlow(t *testing.T) {
	Convey("Given a local login with AutoLogin enabled", t, func() {
		mock := browser.NewMock()
		ctx := context.Background()

		cfg := &Config{
			General: General{
				Mode:            "full",
				AutoFit:         true,
				PageLoadDelayMS: 0,
			},
			Target: Target{
				URL:      "https://grafana.example.com/d/abc/dashboard",
				Username: "admin",
				Password: "admin",
			},
			GoAuth: GoAuth{
				AutoLogin: true,
			},
		}

		Convey("Should navigate to bypass URL then fill credentials then navigate to final URL", func() {
			bypassURL := LocalLoginBypassURL(cfg.Target.URL)
			generatedURL := GenerateURL(cfg)

			_ = mock.Navigate(ctx, bypassURL)
			_ = loginWithCredentials(ctx, mock, cfg.Target.Username, cfg.Target.Password)
			_ = mock.WaitVisible(ctx, `//img[@alt="User avatar"]`)
			_ = mock.Navigate(ctx, generatedURL)

			navigateCalls := mock.CallsTo("Navigate")
			So(navigateCalls, ShouldHaveLength, 2)
			So(navigateCalls[0].Args[0], ShouldEqual, "https://grafana.example.com/login/local")
			So(navigateCalls[1].Args[0], ShouldContainSubstring, "kiosk=1")
		})
	})
}

func TestLocalLoginDirectFlow(t *testing.T) {
	Convey("Given a local login without AutoLogin", t, func() {
		mock := browser.NewMock()
		ctx := context.Background()

		cfg := &Config{
			General: General{
				Mode:            "full",
				AutoFit:         true,
				PageLoadDelayMS: 0,
			},
			Target: Target{
				URL:      "https://grafana.example.com/d/abc/dashboard",
				Username: "admin",
				Password: "admin",
			},
		}

		Convey("Should navigate to generated URL then fill credentials", func() {
			generatedURL := GenerateURL(cfg)

			_ = mock.Navigate(ctx, generatedURL)
			_ = loginWithCredentials(ctx, mock, cfg.Target.Username, cfg.Target.Password)

			navigateCalls := mock.CallsTo("Navigate")
			So(navigateCalls, ShouldHaveLength, 1)
			So(navigateCalls[0].Args[0], ShouldContainSubstring, "kiosk=1")

			sendKeysCalls := mock.CallsTo("SendKeys")
			So(sendKeysCalls, ShouldHaveLength, 2)
		})
	})
}
