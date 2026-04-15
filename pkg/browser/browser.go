package browser

import "context"

// Browser abstracts chromedp actions so login providers can be unit tested
// with a mock implementation.
type Browser interface {
	Navigate(ctx context.Context, url string) error
	WaitVisible(ctx context.Context, sel string) error
	Click(ctx context.Context, sel string) error
	SendKeys(ctx context.Context, sel string, value string) error
}
