package kiosk

import (
	"context"
	"time"

	"fmt"
	"log"
	"os"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"

	"github.com/chromedp/chromedp"

	"google.golang.org/api/idtoken"
)

// GrafanaKioskIDToken creates a chrome-based kiosk using a oauth2 authenticated account.
func GrafanaKioskIDToken(cfg *Config, messages chan string) {
	dir, err := os.MkdirTemp(os.TempDir(), "chromedp-kiosk")
	if err != nil {
		panic(err)
	}

	defer os.RemoveAll(dir)

	opts := generateExecutorOptions(dir, cfg)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// also set up a custom logger
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	listenChromeEvents(taskCtx, cfg, targetCrashed)

	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx); err != nil {
		panic(err)
	}

	// Give browser time to load
	log.Printf("Sleeping %d MS before navigating to url", cfg.General.PageLoadDelayMS)
	time.Sleep(time.Duration(cfg.General.PageLoadDelayMS) * time.Millisecond)

	var generatedURL = GenerateURL(cfg.Target.URL, cfg.General.Mode, cfg.General.AutoFit, cfg.Target.IsPlayList)

	log.Println("Navigating to ", generatedURL)

	log.Printf("Token is using audience %s and reading from %s", cfg.IDToken.Audience, cfg.IDToken.KeyFile)
	tokenSource, err := idtoken.NewTokenSource(context.Background(), cfg.IDToken.Audience, idtoken.WithCredentialsFile(cfg.IDToken.KeyFile))

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

	if err := chromedp.Run(taskCtx, enableFetch(generatedURL)); err != nil {
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

// GetExecutor returns executor for chromedp
func GetExecutor(ctx context.Context) context.Context {
	c := chromedp.FromContext(ctx)

	return cdp.WithExecutor(ctx, c.Target)
}

func enableFetch(url string) chromedp.Tasks {
	return chromedp.Tasks{
		fetch.Enable(),
		chromedp.Navigate(url),
	}
}
