package kiosk

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

// GrafanaKioskLocal creates a chrome-based kiosk using a local grafana-server account.
func GrafanaKioskLocal(cfg *Config, messages chan string) {
	dir, err := os.MkdirTemp(os.TempDir(), "chromedp-kiosk")
	if err != nil {
		panic(err)
	}

	log.Println("Using temp dir:", dir)
	defer os.RemoveAll(dir)

	opts := generateExecutorOptions(dir, cfg)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// also set up a custom logger
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	listenChromeEvents(taskCtx, targetCrashed)

	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx); err != nil {
		panic(err)
	}

	var generatedURL = GenerateURL(cfg.Target.URL, cfg.General.Mode, cfg.General.AutoFit, cfg.Target.IsPlayList)

	log.Println("Navigating to ", generatedURL)
	/*
		Launch chrome and login with local user account

		name=user, type=text
		id=inputPassword, type=password, name=password
	*/
	// Give browser time to load next page (this can be prone to failure, explore different options vs sleeping)
	time.Sleep(2000 * time.Millisecond)

	if err := chromedp.Run(taskCtx,
		chromedp.Navigate(generatedURL),
		chromedp.WaitVisible(`//input[@name="user"]`, chromedp.BySearch),
		chromedp.SendKeys(`//input[@name="user"]`, cfg.Target.Username, chromedp.BySearch),
		chromedp.SendKeys(`//input[@name="password"]`, cfg.Target.Password+kb.Enter, chromedp.BySearch),
	); err != nil {
		panic(err)
	}
	// blocking wait
	for {
		messageFromChrome := <-messages
		if err := chromedp.Run(taskCtx,
			chromedp.Navigate(generatedURL),
		); err != nil {
			panic(err)
		}
		log.Println("Chromium output:", messageFromChrome)
	}
}
