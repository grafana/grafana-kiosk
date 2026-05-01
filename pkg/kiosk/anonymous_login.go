package kiosk

import (
	"context"
	"log"

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

	var generatedURL = GenerateURL(cfg)

	log.Println("Navigating to ", generatedURL)

	if err := chromedp.Run(taskCtx,
		cycleWindowState(cfg),
	); err != nil {
		panic(err)
	}

	if err := b.Navigate(taskCtx, generatedURL); err != nil {
		panic(err)
	}

	if err := chromedp.Run(taskCtx, waitForPageLoad(cfg)); err != nil {
		panic(err)
	}
	// blocking wait until context is cancelled or a message triggers a reload
	for {
		select {
		case <-ctx.Done():
			return
		case messageFromChrome := <-messages:
			if err := b.Navigate(taskCtx, generatedURL); err != nil {
				return
			}
			log.Println("Chromium output:", messageFromChrome)
		}
	}
}
