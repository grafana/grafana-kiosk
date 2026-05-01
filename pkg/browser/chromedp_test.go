package browser

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// Compile-time check: ChromeDP implements Browser.
var _ Browser = (*ChromeDP)(nil)

// These tests use a plain context.Background() which has no chromedp executor,
// causing chromedp.Run to return an error immediately. This verifies that each
// method propagates the error rather than silently returning nil.
func TestChromeDP(t *testing.T) {
	Convey("Given a ChromeDP browser", t, func() {
		b := &ChromeDP{}
		ctx := context.Background()

		Convey("Navigate propagates chromedp errors", func() {
			err := b.Navigate(ctx, "https://example.com")
			So(err, ShouldNotBeNil)
		})

		Convey("WaitVisible propagates chromedp errors", func() {
			err := b.WaitVisible(ctx, `//input[@name="user"]`)
			So(err, ShouldNotBeNil)
		})

		Convey("Click propagates chromedp errors", func() {
			err := b.Click(ctx, `//button[@type="submit"]`)
			So(err, ShouldNotBeNil)
		})

		Convey("SendKeys propagates chromedp errors", func() {
			err := b.SendKeys(ctx, `//input[@name="user"]`, "admin")
			So(err, ShouldNotBeNil)
		})
	})
}
