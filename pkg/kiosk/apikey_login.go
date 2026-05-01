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

	"github.com/grafana/grafana-kiosk/pkg/browser"
)

// IsDataSourceQueryRequest checks if the request URL is a datasource query API
// call to the target host. It matches both the legacy /api/ds/query path and
// the newer /apis/query.grafana.app/.../query path.
func IsDataSourceQueryRequest(requestURL, targetScheme, targetHost string) bool {
	prefix := targetScheme + "://" + targetHost
	if !strings.HasPrefix(requestURL, prefix) {
		return false
	}
	// Ensure the prefix is followed by "/" or end of string to prevent
	// matching against hosts that share a prefix (e.g., example.com.evil.com)
	rest := requestURL[len(prefix):]
	if len(rest) > 0 && rest[0] != '/' {
		return false
	}

	if strings.Contains(requestURL, "/api/ds/query?") {
		return true
	}

	path := strings.SplitN(rest, "?", 2)[0]
	return strings.Contains(rest, "/apis/query.grafana.app/") && strings.HasSuffix(path, "/query")
}

// IsTargetHostRequest checks if the request URL host matches the target host.
func IsTargetHostRequest(requestHost, targetHost string) bool {
	return requestHost == targetHost
}

// GrafanaKioskAPIKey creates a chrome-based kiosk using a grafana api key.
func GrafanaKioskAPIKey(ctx context.Context, cfg *Config, dir string, b browser.Browser, messages chan string) {
	opts := generateExecutorOptions(dir, cfg)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	// also set up a custom logger
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	listenBrowserEvents(taskCtx, cfg, consoleAPICall|targetCrashed)

	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx); err != nil {
		panic(err)
	}

	waitForBrowserStartup(cfg)

	targetURL, err := url.Parse(cfg.Target.URL)
	if err != nil {
		panic(fmt.Errorf("url.Parse: %w", err))
	}

	chromedp.ListenTarget(taskCtx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *fetch.EventRequestPaused:
			go func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("apikey fetch handler error: %v", r)
					}
				}()
				fetchReq := fetch.ContinueRequest(ev.RequestID)
				requestURL, err := url.Parse(ev.Request.URL)
				if err != nil {
					log.Printf("apikey url.Parse error: %v", err)
					return
				}
				// Add Content-Type header only for datasource query API calls
				if IsDataSourceQueryRequest(ev.Request.URL, targetURL.Scheme, targetURL.Host) {
					if cfg.General.DebugEnabled {
						log.Println("Appending Content-Type Header for Metric Query")
					}
					fetchReq.Headers = append(
						fetchReq.Headers,
						&fetch.HeaderEntry{Name: "Content-Type", Value: "application/json"},
					)
				}
				// Append Bearer token to all requests matching the target host
				if IsTargetHostRequest(requestURL.Host, targetURL.Host) {
					if cfg.General.DebugEnabled {
						log.Println("Appending Header Authorization: Bearer REDACTED")
					}
					fetchReq.Headers = append(
						fetchReq.Headers,
						&fetch.HeaderEntry{Name: "Authorization", Value: "Bearer " + cfg.APIKey.APIKey},
					)
				}
				if err = fetchReq.Do(GetExecutor(taskCtx)); err != nil {
					log.Printf("apikey fetchReq error: %v", err)
				}
			}()
		}
	})

	if err := chromedp.Run(taskCtx,
		cycleWindowState(cfg),
		fetch.Enable().WithPatterns([]*fetch.RequestPattern{{URLPattern: targetURL.Scheme + "://" + targetURL.Host + "/*"}}),
	); err != nil {
		panic(err)
	}

	if err := apikeyLoginFlow(taskCtx, cfg, b, GenerateURL(cfg), messages); err != nil {
		panic(err)
	}
}

// apikeyLoginFlow navigates to url (with fetch interception already enabled),
// waits for page load, then blocks until context is cancelled or a message
// triggers a reload.
func apikeyLoginFlow(ctx context.Context, cfg *Config, b browser.Browser, dashboardURL string, messages chan string) error {
	log.Println("Navigating to ", dashboardURL)
	if err := b.Navigate(ctx, dashboardURL); err != nil {
		return err
	}
	if cfg.General.PageLoadDelayMS > 0 {
		log.Printf("Sleeping %d MS for page load", cfg.General.PageLoadDelayMS)
		time.Sleep(time.Duration(cfg.General.PageLoadDelayMS) * time.Millisecond)
	}
	return runMessageLoop(ctx, b, dashboardURL, messages)
}
