package kiosk

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/chromedp"
)

// GrafanaKioskAPIKey creates a chrome-based kiosk using a grafana api key.
func GrafanaKioskAPIKey(cfg *Config, messages chan string) {
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

	listenChromeEvents(taskCtx, cfg, consoleAPICall|targetCrashed)

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
		Launch chrome and look for main-view element
	*/
	u, err := url.Parse(cfg.Target.URL)
	if err != nil {
		panic(fmt.Errorf("url.Parse: %w", err))
	}
	chromedp.ListenTarget(taskCtx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *fetch.EventRequestPaused:
			go func() {
				fetchReq := fetch.ContinueRequest(ev.RequestID)
				fetchReq.Headers = append(
					fetchReq.Headers,
					&fetch.HeaderEntry{Name: "Authorization", Value: "Bearer " + cfg.APIKey.APIKey},
				)
				err = fetchReq.Do(GetExecutor(taskCtx))
				if err != nil {
					panic(fmt.Errorf("apikey fetchReq error: %w", err))
				}
			}()
		}
	})
	if err := chromedp.Run(
		taskCtx,
		fetch.Enable().WithPatterns([]*fetch.RequestPattern{{URLPattern: u.Scheme + "://" + u.Host + "/*"}}),
		chromedp.Navigate(generatedURL),
		chromedp.WaitVisible(`//div[@class="main-view"]`, chromedp.BySearch),
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
