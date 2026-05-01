//go:build integration

// Functional tests for local login.
// Assert that the browser completes the login form and reaches the Grafana dashboard.
package integration

import (
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/grafana/grafana-kiosk/pkg/kiosk/login/shared"
)

func TestLocalLoginReachesDashboard(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires Docker")
	}

	grafanaURL, cleanup := startGrafana(t)
	defer cleanup()

	Convey("Given local login to Grafana in kiosk mode", t, func() {
		cfg := baseCfg(grafanaURL)
		cfg.Target.Username = "admin"
		cfg.Target.Password = "admin"
		dir := tempDir(t)

		taskCtx, cancel := newHeadlessBrowserContext(t, cfg, dir)
		defer cancel()

		bypassURL := grafanaURL + "/login/local"
		kioskURL := shared.GenerateURL(cfg)

		var title string
		var noLoginForm bool

		// Navigate to local login bypass, fill credentials, then load the kiosk URL.
		err := chromedp.Run(taskCtx,
			chromedp.Navigate(bypassURL),
			chromedp.Sleep(2*time.Second),
			chromedp.WaitVisible(`//input[@name="user"]`, chromedp.BySearch),
			chromedp.SendKeys(`//input[@name="user"]`, cfg.Target.Username, chromedp.BySearch),
			chromedp.SendKeys(`//input[@name="password"]`, cfg.Target.Password+"\r", chromedp.BySearch),
			chromedp.Sleep(3*time.Second),
			chromedp.Navigate(kioskURL),
			chromedp.Sleep(3*time.Second),
			chromedp.Title(&title),
			chromedp.Evaluate(`document.querySelector('input[name="user"]') === null`, &noLoginForm),
		)

		Convey("Login and navigation complete without error", func() {
			So(err, ShouldBeNil)
		})

		Convey("Page title contains Grafana", func() {
			So(strings.ToLower(title), ShouldContainSubstring, "grafana")
		})

		Convey("Login form is absent after authentication", func() {
			So(noLoginForm, ShouldBeTrue)
		})
	})
}
