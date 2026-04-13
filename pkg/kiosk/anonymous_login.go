package kiosk

import (
	"context"
	"log"

	"github.com/chromedp/chromedp"
)

// GrafanaKioskAnonymous creates a chrome-based kiosk using a local grafana-server account.
func GrafanaKioskAnonymous(ctx context.Context, cfg *Config, dir string, messages chan string) {
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
		chromedp.Navigate(generatedURL),
		waitForPageLoad(cfg),
	); err != nil {
		panic(err)
	}
	// blocking wait until context is cancelled or a message triggers a reload
	for {
		select {
		case <-ctx.Done():
			return
		case messageFromChrome := <-messages:
			if err := chromedp.Run(taskCtx,
				chromedp.Navigate(generatedURL),
			); err != nil {
				return
			}
			log.Println("Chromium output:", messageFromChrome)
		}
	}
}
