package apikey

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"sort"
	"strings"

	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
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

// RequestHeaders returns the headers to continue an intercepted request with.
// The original headers are kept, Content-Type is defaulted to JSON for
// datasource queries, and Authorization is set for the target host.
func RequestHeaders(orig network.Headers, isQuery, isTarget bool, apiKey string) []*fetch.HeaderEntry {
	keys := make([]string, 0, len(orig))
	for k := range orig {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	headers := make([]*fetch.HeaderEntry, 0, len(keys)+2)
	hasContentType := false
	for _, k := range keys {
		if isTarget && strings.EqualFold(k, "Authorization") {
			continue
		}
		if strings.EqualFold(k, "Content-Type") {
			hasContentType = true
		}
		headers = append(headers, &fetch.HeaderEntry{Name: k, Value: fmt.Sprintf("%v", orig[k])})
	}
	if isQuery && !hasContentType {
		headers = append(headers, &fetch.HeaderEntry{Name: "Content-Type", Value: "application/json"})
	}
	if isTarget {
		headers = append(headers, &fetch.HeaderEntry{Name: "Authorization", Value: "Bearer " + apiKey})
	}
	return headers
}

// IsTargetHostRequest checks if the request URL host matches the target host.
func IsTargetHostRequest(requestHost, targetHost string) bool {
	return requestHost == targetHost
}

// Run creates a chrome-based kiosk using a grafana api key.
func Run(ctx context.Context, cfg *config.Config, dir string, b browser.Browser, messages chan string) {
	taskCtx, cancel := shared.NewBrowserContext(ctx, cfg, dir, shared.ConsoleAPICall|shared.TargetCrashed)
	defer cancel()

	targetURL, err := url.Parse(cfg.Target.URL)
	if err != nil {
		panic(fmt.Errorf("url.Parse: %w", err))
	}

	chromedp.ListenTarget(taskCtx, func(ev any) {
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
				isQuery := IsDataSourceQueryRequest(ev.Request.URL, targetURL.Scheme, targetURL.Host)
				isTarget := IsTargetHostRequest(requestURL.Host, targetURL.Host)
				if isQuery || isTarget {
					if isTarget && cfg.General.DebugEnabled {
						log.Println("Appending Header Authorization: Bearer REDACTED")
					}
					// ContinueRequest replaces the whole header set, so start from
					// the original request headers instead of dropping them.
					fetchReq.Headers = RequestHeaders(ev.Request.Headers, isQuery, isTarget, cfg.APIKey.APIKey)
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
