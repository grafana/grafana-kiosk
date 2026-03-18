package kiosk

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/chromedp"
)

// GrafanaKioskAPIKey creates a chrome-based kiosk using a grafana api key.
func GrafanaKioskAPIKey(ctx context.Context, cfg *Config, dir string, messages chan string) {
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

	u, err := url.Parse(cfg.Target.URL)
	if err != nil {
		panic(fmt.Errorf("url.Parse: %w", err))
	}
	chromedp.ListenTarget(taskCtx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *fetch.EventRequestPaused:
			go func() {
				fetchReq := fetch.ContinueRequest(ev.RequestID)
				requestURL, err := url.Parse(ev.Request.URL)
				if err != nil {
					panic(fmt.Errorf("url.Parse: %w", err))
				}
				// Add Content-Type header only for datasource query API calls
				if strings.HasPrefix(ev.Request.URL, u.Scheme+"://"+u.Host) &&
					strings.Contains(ev.Request.URL, "/api/ds/query?") {
					if cfg.General.DebugEnabled {
						log.Println("Appending Content-Type Header for Metric Query")
					}
					fetchReq.Headers = append(
						fetchReq.Headers,
						&fetch.HeaderEntry{Name: "Content-Type", Value: "application/json"},
					)
				}
				// Append Bearer token to all requests matching the target host
				if requestURL.Host == u.Host {
					if cfg.General.DebugEnabled {
						log.Println("Appending Header Authorization: Bearer REDACTED")
					}
					fetchReq.Headers = append(
						fetchReq.Headers,
						&fetch.HeaderEntry{Name: "Authorization", Value: "Bearer " + cfg.APIKey.APIKey},
					)
				}
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
	); err != nil {
		panic(err)
	}
	// Give browser time to fully render the dashboard
	log.Printf("Sleeping %d MS before continuing", cfg.General.PageLoadDelayMS)
	time.Sleep(time.Duration(cfg.General.PageLoadDelayMS) * time.Millisecond)
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
