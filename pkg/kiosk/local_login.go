package kiosk

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/grafana/grafana-kiosk/pkg/browser"
)

// LocalLoginBypassURL extracts the base URL and appends /login/local to bypass
// OAuth auto-login and use a local account instead.
func LocalLoginBypassURL(rawURL string) string {
	startIndex := strings.Index(rawURL, "://") + 3
	slashIndex := strings.Index(rawURL[startIndex:], "/")

	var baseURL string
	if slashIndex == -1 {
		baseURL = rawURL
	} else {
		baseURL = rawURL[:startIndex+slashIndex]
	}

	return baseURL + "/login/local"
}

// loginWithCredentials waits for login fields and submits credentials.
func loginWithCredentials(ctx context.Context, b browser.Browser, username, password string) error {
	if err := b.WaitVisible(ctx, `//input[@name="user"]`); err != nil {
		return err
	}
	if err := b.SendKeys(ctx, `//input[@name="user"]`, username); err != nil {
		return err
	}
	return b.SendKeys(ctx, `//input[@name="password"]`, password+kb.Enter)
}

// GrafanaKioskLocal creates a chrome-based kiosk using a local grafana-server account.
func GrafanaKioskLocal(ctx context.Context, cfg *Config, dir string, b browser.Browser, messages chan string) {
	opts := generateExecutorOptions(dir, cfg)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	// also set up a custom logger
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	listenChromeEvents(taskCtx, cfg, targetCrashed)

	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx); err != nil {
		panic(err)
	}

	waitForBrowserStartup(cfg)

	if err := chromedp.Run(taskCtx,
		waitForPageLoad(cfg),
		cycleWindowState(cfg),
	); err != nil {
		panic(err)
	}

	if err := localLoginFlow(taskCtx, cfg, b, GenerateURL(cfg), messages); err != nil {
		panic(err)
	}
}

// localLoginFlow drives the local-account login sequence and then blocks until
// context is cancelled or a message triggers a reload. Extracted for testability.
func localLoginFlow(ctx context.Context, cfg *Config, b browser.Browser, generatedURL string, messages chan string) error {
	log.Println("Navigating to ", generatedURL)
	delay := time.Duration(cfg.General.PageLoadDelayMS) * time.Millisecond

	/*
		Launch chrome and login with local user account

		name=user, type=text
		id=inputPassword, type=password, name=password
	*/
	if cfg.GoAuth.AutoLogin {
		// if AutoLogin is set, get the base URL and append the local login bypass before navigating to the full url
		bypassURL := LocalLoginBypassURL(cfg.Target.URL)

		log.Println("Bypassing autoLogin using URL ", bypassURL)

		if err := b.Navigate(ctx, bypassURL); err != nil {
			return err
		}

		log.Printf("Sleeping %d MS before checking for login fields", cfg.General.PageLoadDelayMS)
		time.Sleep(delay)

		if err := loginWithCredentials(ctx, b, cfg.Target.Username, cfg.Target.Password); err != nil {
			return err
		}

		log.Printf("Sleeping %d MS before checking for topnav", cfg.General.PageLoadDelayMS)
		time.Sleep(delay)

		if err := b.WaitVisible(ctx, `//img[@alt="User avatar"]`); err != nil {
			return err
		}

		log.Printf("Sleeping %d MS before navigating to final url", cfg.General.PageLoadDelayMS)
		time.Sleep(delay)

		if err := b.Navigate(ctx, generatedURL); err != nil {
			return err
		}
	} else {
		log.Printf("Sleeping %d MS before navigating to final url", cfg.General.PageLoadDelayMS)
		time.Sleep(delay)

		if err := b.Navigate(ctx, generatedURL); err != nil {
			return err
		}

		log.Printf("Sleeping %d MS before checking for login fields", cfg.General.PageLoadDelayMS)
		time.Sleep(delay)

		if err := loginWithCredentials(ctx, b, cfg.Target.Username, cfg.Target.Password); err != nil {
			return err
		}
	}

	// blocking wait until context is cancelled or a message triggers a reload
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-messages:
			if err := b.Navigate(ctx, generatedURL); err != nil {
				return nil
			}
			log.Println("Chromium output:", msg)
		}
	}
}
