package idtoken

import (
	"context"
	"fmt"
	"log"

	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/chromedp"

	"github.com/grafana/grafana-kiosk/pkg/browser"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/config"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/login/shared"

	gidtoken "google.golang.org/api/idtoken"
	"google.golang.org/api/option"
)

// Run creates a chrome-based kiosk using a oauth2 authenticated account.
func Run(ctx context.Context, cfg *config.Config, dir string, b browser.Browser, messages chan string) {
	opts := shared.GenerateExecutorOptions(dir, cfg)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	shared.ListenBrowserEvents(taskCtx, cfg, shared.TargetCrashed)

	if err := chromedp.Run(taskCtx); err != nil {
		panic(err)
	}

	shared.WaitForBrowserStartup(cfg)

	log.Printf("Token is using audience %s and reading from %s", cfg.IDToken.Audience, cfg.IDToken.KeyFile)
	tokenSource, err := gidtoken.NewTokenSource(context.Background(), cfg.IDToken.Audience, option.WithAuthCredentialsFile(option.ServiceAccount, cfg.IDToken.KeyFile))

	if err != nil {
		panic(err)
	}

	chromedp.ListenTarget(taskCtx, func(ev any) {
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
				if err = fetchReq.Do(shared.GetExecutor(taskCtx)); err != nil {
					log.Printf("idtoken fetchReq error: %v", err)
				}
			}()
		}
	})

	generatedURL := shared.GenerateURL(cfg)
	log.Printf("Navigating to %s", generatedURL)

	// fetch.Enable and Navigate must be in the same chromedp.Run batch so no
	// unfiltered request can slip through the interception window.
	if err := chromedp.Run(taskCtx,
		shared.WaitForPageLoad(cfg),
		shared.CycleWindowState(cfg),
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
// fetch interception and navigation atomic.
func idtokenLoginFlow(ctx context.Context, b browser.Browser, dashboardURL string, messages chan string) error {
	return shared.RunMessageLoop(ctx, b, dashboardURL, messages)
}
