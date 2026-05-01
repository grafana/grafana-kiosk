package kiosk

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"

	"github.com/grafana/grafana-kiosk/pkg/browser"
)

// GrafanaKioskGCOM creates a chrome-based kiosk using a grafana.com authenticated account.
func GrafanaKioskGCOM(ctx context.Context, cfg *Config, dir string, b browser.Browser, messages chan string) {
	opts := generateExecutorOptions(dir, cfg)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	// also set up a custom logger
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	listenBrowserEvents(taskCtx, cfg, targetCrashed)

	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx); err != nil {
		panic(err)
	}

	waitForBrowserStartup(cfg)

	if err := chromedp.Run(taskCtx,
		waitForPageLoad(cfg),
		cycleWindowState(cfg),
	); err != nil {
		panic(err)
	}

	if err := gcomLoginFlow(taskCtx, cfg, b, GenerateURL(cfg), messages); err != nil {
		panic(err)
	}
}

// gcomLoginFlow navigates to the Grafana login page, clicks the grafana.com
// login button, fills in credentials, then blocks until context is cancelled
// or a message triggers a reload.
func gcomLoginFlow(ctx context.Context, cfg *Config, b browser.Browser, url string, messages chan string) error {
	log.Println("Navigating to ", url)

	// XPATH for grafana.com login button = //a[contains(@href,'login/grafana_com')]
	if err := b.Navigate(ctx, url); err != nil {
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

	// Give browser time to load next page
	time.Sleep(3 * time.Second)

	if err := b.WaitVisible(ctx, `//input[@name="login"]`); err != nil {
		return err
	}
	if err := b.SendKeys(ctx, `//input[@name="login"]`, cfg.Target.Username); err != nil {
		return err
	}
	if err := b.Click(ctx, `//*[@id="submit"]`); err != nil {
		return err
	}
	if err := b.SendKeys(ctx, `//input[@name="password"]`, cfg.Target.Password+kb.Enter); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case messageFromBrowser := <-messages:
			if err := b.Navigate(ctx, url); err != nil {
				return nil
			}
			log.Println("Browser output:", messageFromBrowser)
		}
	}
}
