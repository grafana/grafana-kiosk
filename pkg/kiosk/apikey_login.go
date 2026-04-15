package kiosk

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/chromedp"
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

	waitForBrowserStartup(cfg)

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
				if IsDataSourceQueryRequest(ev.Request.URL, u.Scheme, u.Host) {
					if cfg.General.DebugEnabled {
						log.Println("Appending Content-Type Header for Metric Query")
					}
					fetchReq.Headers = append(
						fetchReq.Headers,
						&fetch.HeaderEntry{Name: "Content-Type", Value: "application/json"},
					)
				}
				// Append Bearer token to all requests matching the target host
				if IsTargetHostRequest(requestURL.Host, u.Host) {
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
		cycleWindowState(cfg),
		fetch.Enable().WithPatterns([]*fetch.RequestPattern{{URLPattern: u.Scheme + "://" + u.Host + "/*"}}),
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
