package kiosk

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

// GrafanaKioskGCOM creates a chrome-based kiosk using a grafana.com authenticated account.
func GrafanaKioskGCOM(cfg *Config, messages chan string) {
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

	listenChromeEvents(taskCtx, cfg, targetCrashed)

	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx); err != nil {
		panic(err)
	}
	// Give browser time to load
	log.Printf("Sleeping %d MS before navigating to url", cfg.General.PageLoadDelayMS)
	time.Sleep(time.Duration(cfg.General.PageLoadDelayMS) * time.Millisecond)

	var generatedURL = GenerateURL(cfg.Target.URL, cfg.General.Mode, cfg.General.AutoFit, cfg.Target.IsPlayList)

	log.Println("Navigating to ", generatedURL)
	/*
		Launch chrome, click the grafana.com button, fill out login form and submit
	*/
	// XPATH of grafana.com login button = //*[@href="login/grafana_com"]/i
	// XPATH for grafana.com login (new) = //a[contains(@href,'login/grafana_com')]

	// chromedp.WaitVisible(`//*[@href="login/grafana_com"]/i`, chromedp.BySearch),

	// Click the grafana_com login button
	if err := chromedp.Run(taskCtx,
		chromedp.Navigate(generatedURL),
		chromedp.ActionFunc(func(context.Context) error {
			log.Println("waiting for login dialog")

			return nil
		}),
		chromedp.WaitVisible(`//a[contains(@href,'login/grafana_com')]`, chromedp.BySearch),
		chromedp.ActionFunc(func(context.Context) error {
			log.Println("gcom login dialog detected")

			return nil
		}),
		chromedp.Click(`//a[contains(@href,'login/grafana_com')]/..`, chromedp.BySearch),
		chromedp.ActionFunc(func(context.Context) error {
			log.Println("gcom button clicked")

			return nil
		}),
	); err != nil {
		panic(err)
	}
	// Give browser time to load next page (this can be prone to failure, explore different options vs sleeping)
	time.Sleep(2000 * time.Millisecond)
	// Fill out grafana_com login page
	if err := chromedp.Run(taskCtx,
		chromedp.WaitVisible(`//input[@name="login"]`, chromedp.BySearch),
		chromedp.SendKeys(`//input[@name="login"]`, cfg.Target.Username, chromedp.BySearch),
		chromedp.Click(`#submit`, chromedp.ByID),
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
