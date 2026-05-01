package gcom

import (
	"context"
	"log"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"

	"github.com/grafana/grafana-kiosk/pkg/browser"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/config"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/login/shared"
)

// Run creates a chrome-based kiosk using a grafana.com authenticated account.
func Run(ctx context.Context, cfg *config.Config, dir string, b browser.Browser, messages chan string) {
	opts := shared.GenerateExecutorOptions(dir, cfg)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	shared.ListenBrowserEvents(taskCtx, cfg, shared.TargetCrashed)

	if err := chromedp.Run(taskCtx); err != nil {
		panic(err)
	}

	shared.WaitForBrowserStartup(cfg)

	if err := chromedp.Run(taskCtx,
		shared.WaitForPageLoad(cfg),
		shared.CycleWindowState(cfg),
	); err != nil {
		panic(err)
	}

	if err := gcomLoginFlow(taskCtx, cfg, b, shared.GenerateURL(cfg), messages); err != nil {
		panic(err)
	}
}

// gcomLoginFlow navigates to the Grafana login page, clicks the grafana.com
// login button, fills in credentials, then blocks until context is cancelled
// or a message triggers a reload.
func gcomLoginFlow(ctx context.Context, cfg *config.Config, b browser.Browser, dashboardURL string, messages chan string) error {
	log.Printf("Navigating to %s", dashboardURL)

	if err := b.Navigate(ctx, dashboardURL); err != nil {
		return err
	}

	log.Println("waiting for login dialog")
	if err := b.WaitVisible(ctx, `//a[contains(@href,'login/grafana_com')]`); err != nil {
		return err
	}
	log.Println("gcom login dialog detected")
	if err := b.Click(ctx, `//a[contains(@href,'login/grafana_com')]`); err != nil {
		return err
	}
	log.Println("gcom button clicked")
	if err := b.WaitVisible(ctx, `//input[@name="login"]`); err != nil {
		return err
	}
	if err := b.SendKeys(ctx, `//input[@name="login"]`, cfg.Target.Username); err != nil {
		return err
	}
	if err := b.Click(ctx, `#submit`); err != nil {
		return err
	}
	if err := b.SendKeys(ctx, `//input[@name="password"]`, cfg.Target.Password+kb.Enter); err != nil {
		return err
	}

	return shared.RunMessageLoop(ctx, b, dashboardURL, messages)
}
