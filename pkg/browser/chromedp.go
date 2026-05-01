package browser

import (
	"context"

	"github.com/chromedp/chromedp"
)

// ChromeDP implements Browser by delegating to chromedp.
type ChromeDP struct{}

func (c *ChromeDP) Navigate(ctx context.Context, url string) error {
	return chromedp.Run(ctx, chromedp.Navigate(url))
}

func (c *ChromeDP) WaitVisible(ctx context.Context, sel string) error {
	return chromedp.Run(ctx, chromedp.WaitVisible(sel, chromedp.BySearch))
}

func (c *ChromeDP) WaitNotVisible(ctx context.Context, sel string) error {
	return chromedp.Run(ctx, chromedp.WaitNotVisible(sel, chromedp.BySearch))
}

func (c *ChromeDP) Click(ctx context.Context, sel string) error {
	return chromedp.Run(ctx, chromedp.Click(sel, chromedp.BySearch))
}

func (c *ChromeDP) SendKeys(ctx context.Context, sel string, value string) error {
	return chromedp.Run(ctx, chromedp.SendKeys(sel, value, chromedp.BySearch))
}
