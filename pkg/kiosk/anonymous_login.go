package kiosk

import (
	"context"
	"log"
	"time"

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

	// Give browser time to load
	log.Printf("Sleeping %d MS before navigating to url", cfg.General.PageLoadDelayMS)
	time.Sleep(time.Duration(cfg.General.PageLoadDelayMS) * time.Millisecond)

	var generatedURL = GenerateURL(cfg)

	log.Println("Navigating to ", generatedURL)
	/*
		Launch chrome and look for main-view element
	*/
	if err := chromedp.Run(taskCtx,
		chromedp.Navigate(generatedURL),
		chromedp.WaitVisible(`//div[@class="main-view"]`, chromedp.BySearch),
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
