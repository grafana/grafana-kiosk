package apikey

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/chromedp"

	"github.com/grafana/grafana-kiosk/pkg/browser"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/config"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/login/shared"
)

// IsDataSourceQueryRequest checks if the request URL is a datasource query API
// call to the target host. It matches both the legacy /api/ds/query path and
// the newer /apis/query.grafana.app/.../query path.
func IsDataSourceQueryRequest(requestURL, targetScheme, targetHost string) bool {
	prefix := targetScheme + "://" + targetHost
	if !strings.HasPrefix(requestURL, prefix) {
		return false
	}
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

// Run creates a chrome-based kiosk using a grafana api key.
func Run(ctx context.Context, cfg *config.Config, dir string, b browser.Browser, messages chan string) {
	opts := shared.GenerateExecutorOptions(dir, cfg)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	shared.ListenBrowserEvents(taskCtx, cfg, shared.ConsoleAPICall|shared.TargetCrashed)

	if err := chromedp.Run(taskCtx); err != nil {
		panic(err)
	}

	shared.WaitForBrowserStartup(cfg)

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
				if IsDataSourceQueryRequest(ev.Request.URL, targetURL.Scheme, targetURL.Host) {
					if cfg.General.DebugEnabled {
						log.Println("Appending Content-Type Header for Metric Query")
					}
					fetchReq.Headers = append(
						fetchReq.Headers,
						&fetch.HeaderEntry{Name: "Content-Type", Value: "application/json"},
					)
				}
				if IsTargetHostRequest(requestURL.Host, targetURL.Host) {
					if cfg.General.DebugEnabled {
						log.Println("Appending Header Authorization: Bearer REDACTED")
					}
					fetchReq.Headers = append(
						fetchReq.Headers,
						&fetch.HeaderEntry{Name: "Authorization", Value: "Bearer " + cfg.APIKey.APIKey},
					)
				}
				if err = fetchReq.Do(shared.GetExecutor(taskCtx)); err != nil {
					log.Printf("apikey fetchReq error: %v", err)
				}
			}()
		}
	})

	if err := chromedp.Run(taskCtx,
		shared.CycleWindowState(cfg),
		fetch.Enable().WithPatterns([]*fetch.RequestPattern{{URLPattern: targetURL.Scheme + "://" + targetURL.Host + "/*"}}),
	); err != nil {
		panic(err)
	}

	if err := apikeyLoginFlow(taskCtx, cfg, b, shared.GenerateURL(cfg), messages); err != nil {
		panic(err)
	}
}

// apikeyLoginFlow navigates to url (with fetch interception already enabled),
// waits for page load, then blocks until context is cancelled or a message
// triggers a reload.
func apikeyLoginFlow(ctx context.Context, cfg *config.Config, b browser.Browser, dashboardURL string, messages chan string) error {
	log.Printf("Navigating to %s", dashboardURL)
	if err := b.Navigate(ctx, dashboardURL); err != nil {
		return err
	}
	shared.SleepPageLoad(cfg)
	return shared.RunMessageLoop(ctx, b, dashboardURL, messages)
}
