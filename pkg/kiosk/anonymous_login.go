package kiosk

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/grafana/grafana-kiosk/pkg/browser"
)

// GrafanaKioskAnonymous creates a chrome-based kiosk using a local grafana-server account.
func GrafanaKioskAnonymous(ctx context.Context, cfg *Config, dir string, b browser.Browser, messages chan string) {
	opts := generateExecutorOptions(dir, cfg)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	// also set up a custom logger
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	listenChromeEvents(taskCtx, cfg, consoleAPICall|targetCrashed)

	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx); err != nil {
		panic(err)
	}

	waitForBrowserStartup(cfg)

	if err := chromedp.Run(taskCtx, cycleWindowState(cfg)); err != nil {
		panic(err)
	}

	if err := anonymousLoginFlow(taskCtx, cfg, b, GenerateURL(cfg), messages); err != nil {
		panic(err)
	}
}

// anonymousLoginFlow navigates to url, waits for page load, then blocks until
// context is cancelled or a message triggers a reload. Extracted for testability.
func anonymousLoginFlow(ctx context.Context, cfg *Config, b browser.Browser, url string, messages chan string) error {
	log.Println("Navigating to ", url)
	if err := b.Navigate(ctx, url); err != nil {
		return err
	}
	if cfg.General.PageLoadDelayMS > 0 {
		log.Printf("Sleeping %d MS for page load", cfg.General.PageLoadDelayMS)
		time.Sleep(time.Duration(cfg.General.PageLoadDelayMS) * time.Millisecond)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-messages:
			if err := b.Navigate(ctx, url); err != nil {
				return nil
			}
			log.Println("Browser output:", msg)
		}
	}
}
