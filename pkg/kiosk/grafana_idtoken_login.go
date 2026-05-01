package kiosk

import (
	"context"
	"fmt"
	"log"
	"time"

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
				fetchReq := fetch.ContinueRequest(ev.RequestID)
				for k, v := range ev.Request.Headers {
					fetchReq.Headers = append(fetchReq.Headers, &fetch.HeaderEntry{Name: k, Value: fmt.Sprintf("%v", v)})
				}
				token, err := tokenSource.Token()
				if err != nil {
					panic(fmt.Errorf("idtoken.NewClient: %w", err))
				}
				fetchReq.Headers = append(fetchReq.Headers, &fetch.HeaderEntry{Name: "Authorization", Value: "Bearer " + token.AccessToken})
				err = fetchReq.Do(GetExecutor(taskCtx))
				if err != nil {
					panic(fmt.Errorf("idtoken.NewClient fetchReq error: %w", err))
				}
			}()
		}
	})

	if err := chromedp.Run(taskCtx,
		waitForPageLoad(cfg),
		cycleWindowState(cfg),
		fetch.Enable(),
	); err != nil {
		panic(err)
	}

	if err := idtokenLoginFlow(taskCtx, cfg, b, GenerateURL(cfg), messages); err != nil {
		panic(err)
	}
}

// idtokenLoginFlow navigates to url (with fetch interception already enabled),
// then blocks until context is cancelled or a message triggers a reload.
func idtokenLoginFlow(ctx context.Context, cfg *Config, b browser.Browser, url string, messages chan string) error {
	log.Println("Navigating to ", url)
	if err := b.Navigate(ctx, url); err != nil {
		return err
	}
	if cfg.General.PageLoadDelayMS > 0 {
		log.Printf("Sleeping %d MS for page load", cfg.General.PageLoadDelayMS)
		time.Sleep(time.Duration(cfg.General.PageLoadDelayMS) * time.Millisecond)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-messages:
			if err := b.Navigate(ctx, url); err != nil {
				return nil
			}
			log.Println("Browser output:", msg)
		}
	}
}

// GetExecutor returns executor for chromedp
func GetExecutor(ctx context.Context) context.Context {
	c := chromedp.FromContext(ctx)

	return cdp.WithExecutor(ctx, c.Target)
}

