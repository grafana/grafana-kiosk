package kiosk

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

// GrafanaKioskAWSLogin Provides login for AWS Managed Grafana instances
func GrafanaKioskAWSLogin(cfg *Config, messages chan string) {
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
	// Give browser time to load
	log.Printf("Sleeping %d MS before navigating to url", cfg.General.PageLoadDelayMS)
	time.Sleep(time.Duration(cfg.General.PageLoadDelayMS) * time.Millisecond)

	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx); err != nil {
		panic(err)
	}

	var generatedURL = GenerateURL(cfg.Target.URL, cfg.General.Mode, cfg.General.AutoFit, cfg.Target.IsPlayList)

	log.Println("Navigating to ", generatedURL)

	// Give browser time to load next page (this can be prone to failure, explore different options vs sleeping)
	time.Sleep(2000 * time.Millisecond)

	if err := chromedp.Run(taskCtx,
		chromedp.Navigate(generatedURL),
		chromedp.WaitVisible(`//a[contains(@href,'login/sso')]`, chromedp.BySearch),
		chromedp.Click(`//a[contains(@href,'login/sso')]`, chromedp.BySearch),
		chromedp.WaitVisible(`input#awsui-input-0`, chromedp.BySearch),
		chromedp.SendKeys(`input#awsui-input-0`, cfg.Target.Username+kb.Enter, chromedp.BySearch),
		chromedp.WaitVisible(`input#awsui-input-1`, chromedp.BySearch),
		chromedp.SendKeys(`input#awsui-input-1`, cfg.Target.Password+kb.Enter, chromedp.BySearch),
	); err != nil {
		panic(err)
	}

	if cfg.Target.UseMFA {
		if err := chromedp.Run(taskCtx,
			chromedp.WaitNotVisible(`input#awsui-input-2`, chromedp.BySearch),
		); err != nil {
			panic(err)
		}
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
