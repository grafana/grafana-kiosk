//go:build integration

// Functional tests for anonymous kiosk login.
// Assert that the browser loads Grafana in kiosk mode and no login form appears.
package integration

import (
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/grafana/grafana-kiosk/pkg/kiosk/login/shared"
)

func TestAnonymousPageLoadsInKioskMode(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires Docker")
	}

	grafanaURL, cleanup := startGrafana(t)
	defer cleanup()

	Convey("Given anonymous access to Grafana in kiosk mode", t, func() {
		cfg := baseCfg(grafanaURL)
		dir := tempDir(t)

		taskCtx, cancel := newHeadlessBrowserContext(t, cfg, dir)
		defer cancel()

		kioskURL := shared.GenerateURL(cfg)

		var title, currentURL string
		var noLoginForm bool

		err := chromedp.Run(taskCtx,
			chromedp.Navigate(kioskURL),
			chromedp.Sleep(3*time.Second),
			chromedp.Title(&title),
			chromedp.Location(&currentURL),
			chromedp.Evaluate(`document.querySelector('input[name="user"]') === null`, &noLoginForm),
		)

		Convey("Page loads without error", func() {
			So(err, ShouldBeNil)
		})

		Convey("Page title contains Grafana", func() {
			So(strings.ToLower(title), ShouldContainSubstring, "grafana")
		})

		Convey("URL contains kiosk=1 param", func() {
			So(currentURL, ShouldContainSubstring, "kiosk=1")
		})

		Convey("Login form is absent — anonymous access granted", func() {
			So(noLoginForm, ShouldBeTrue)
		})
	})
}
