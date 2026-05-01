package kiosk

import (
	"context"
	"fmt"
	"log"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"

	"github.com/chromedp/chromedp"

	"github.com/grafana/grafana-kiosk/pkg/browser"

	"google.golang.org/api/idtoken"
	"google.golang.org/api/option"
)

// GrafanaKioskIDToken creates a chrome-based kiosk using a oauth2 authenticated account.
func GrafanaKioskIDToken(ctx context.Context, cfg *Config, dir string, b browser.Browser, messages chan string) {
	opts := generateExecutorOptions(dir, cfg)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	// also set up a custom logger
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	listenBrowserEvents(taskCtx, cfg, targetCrashed)

	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx); err != nil {
		panic(err)
	}

	waitForBrowserStartup(cfg)

	log.Printf("Token is using audience %s and reading from %s", cfg.IDToken.Audience, cfg.IDToken.KeyFile)
	tokenSource, err := idtoken.NewTokenSource(context.Background(), cfg.IDToken.Audience, option.WithAuthCredentialsFile(option.ServiceAccount, cfg.IDToken.KeyFile))

	if err != nil {
		panic(err)
	}

	chromedp.ListenTarget(taskCtx, func(ev interface{}) {
		//nolint:gocritic // future events can be handled here
		switch ev := ev.(type) {
		case *fetch.EventRequestPaused:
			go func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("idtoken fetch handler error: %v", r)
					}
				}()
				fetchReq := fetch.ContinueRequest(ev.RequestID)
				for k, v := range ev.Request.Headers {
					fetchReq.Headers = append(fetchReq.Headers, &fetch.HeaderEntry{Name: k, Value: fmt.Sprintf("%v", v)})
				}
				token, err := tokenSource.Token()
				if err != nil {
					log.Printf("idtoken token fetch error: %v", err)
					return
				}
				fetchReq.Headers = append(fetchReq.Headers, &fetch.HeaderEntry{Name: "Authorization", Value: "Bearer " + token.AccessToken})
				if err = fetchReq.Do(GetExecutor(taskCtx)); err != nil {
					log.Printf("idtoken fetchReq error: %v", err)
				}
			}()
		}
	})

	generatedURL := GenerateURL(cfg)
	log.Println("Navigating to ", generatedURL)

	// fetch.Enable and Navigate must be in the same chromedp.Run batch so no
	// unfiltered request can slip through the interception window.
	if err := chromedp.Run(taskCtx,
		waitForPageLoad(cfg),
		cycleWindowState(cfg),
		fetch.Enable(),
		chromedp.Navigate(generatedURL),
	); err != nil {
		panic(err)
	}

	if err := idtokenLoginFlow(taskCtx, b, generatedURL, messages); err != nil {
		panic(err)
	}
}

// idtokenLoginFlow blocks until context is cancelled or a message triggers a
// reload. The initial navigation is handled by the outer function to keep
// fetch interception and navigation atomic. This wrapper exists to maintain
// naming consistency with the other login providers and to give tests a
// stable target.
func idtokenLoginFlow(ctx context.Context, b browser.Browser, dashboardURL string, messages chan string) error {
	return runMessageLoop(ctx, b, dashboardURL, messages)
}

// GetExecutor returns executor for chromedp
func GetExecutor(ctx context.Context) context.Context {
	c := chromedp.FromContext(ctx)

	return cdp.WithExecutor(ctx, c.Target)
}

