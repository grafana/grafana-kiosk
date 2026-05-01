package anonymous

import (
	"context"
	"log"

	"github.com/chromedp/chromedp"

	"github.com/grafana/grafana-kiosk/pkg/browser"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/config"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/login/shared"
)

// Run starts an anonymous kiosk session without login.
func Run(ctx context.Context, cfg *config.Config, dir string, b browser.Browser, messages chan string) {
	taskCtx, cancel := shared.NewBrowserContext(ctx, cfg, dir, shared.ConsoleAPICall|shared.TargetCrashed)
	defer cancel()

	if err := chromedp.Run(taskCtx, shared.CycleWindowState(cfg)); err != nil {
		panic(err)
	}

	if err := anonymousLoginFlow(taskCtx, cfg, b, shared.GenerateURL(cfg), messages); err != nil {
		panic(err)
	}
}

// anonymousLoginFlow navigates to dashboardURL, waits for page load, then blocks until
// context is cancelled or a message triggers a reload.
func anonymousLoginFlow(ctx context.Context, cfg *config.Config, b browser.Browser, dashboardURL string, messages chan string) error {
	log.Printf("Navigating to %s", dashboardURL)
	if err := b.Navigate(ctx, dashboardURL); err != nil {
		return err
	}
	shared.SleepPageLoad(cfg)
	return shared.RunMessageLoop(ctx, b, dashboardURL, messages)
}
