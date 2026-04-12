package kiosk

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
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

// GrafanaKioskLocal creates a chrome-based kiosk using a local grafana-server account.
func GrafanaKioskLocal(ctx context.Context, cfg *Config, dir string, messages chan string) {
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

	var generatedURL = GenerateURL(cfg)

	log.Println("Navigating to ", generatedURL)
	/*
		Launch chrome and login with local user account

		name=user, type=text
		id=inputPassword, type=password, name=password
	*/
	// Give browser time to load
	log.Printf("Sleeping %d MS before navigating to url", cfg.General.PageLoadDelayMS)
	time.Sleep(time.Duration(cfg.General.PageLoadDelayMS) * time.Millisecond)

	if cfg.GoAuth.AutoLogin {
		// if AutoLogin is set, get the base URL and append the local login bypass before navigating to the full url
		bypassURL := LocalLoginBypassURL(cfg.Target.URL)

		log.Println("Bypassing autoLogin using URL ", bypassURL)

		if err := chromedp.Run(taskCtx,
			chromedp.Navigate(bypassURL),
			chromedp.ActionFunc(func(context.Context) error {
				log.Printf("Sleeping %d MS before checking for login fields", cfg.General.PageLoadDelayMS)
				time.Sleep(time.Duration(cfg.General.PageLoadDelayMS) * time.Millisecond)
				return nil
			}),
			chromedp.WaitVisible(`//input[@name="user"]`, chromedp.BySearch),
			chromedp.SendKeys(`//input[@name="user"]`, cfg.Target.Username, chromedp.BySearch),
			chromedp.SendKeys(`//input[@name="password"]`, cfg.Target.Password+kb.Enter, chromedp.BySearch),
			chromedp.ActionFunc(func(context.Context) error {
				log.Printf("Sleeping %d MS before checking for topnav", cfg.General.PageLoadDelayMS)
				time.Sleep(time.Duration(cfg.General.PageLoadDelayMS) * time.Millisecond)
				return nil
			}),
			chromedp.WaitVisible(`//img[@alt="User avatar"]`, chromedp.BySearch),
			chromedp.ActionFunc(func(context.Context) error {
				log.Printf("Sleeping %d MS before navigating to final url", cfg.General.PageLoadDelayMS)
				time.Sleep(time.Duration(cfg.General.PageLoadDelayMS) * time.Millisecond)
				return nil
			}),
			chromedp.Navigate(generatedURL),
			postNavigate(cfg),
		); err != nil {
			panic(err)
		}
	} else {
		if err := chromedp.Run(taskCtx,
			chromedp.ActionFunc(func(context.Context) error {
				log.Printf("Sleeping %d MS before navigating to final url", cfg.General.PageLoadDelayMS)
				time.Sleep(time.Duration(cfg.General.PageLoadDelayMS) * time.Millisecond)
				return nil
			}),
			chromedp.Navigate(generatedURL),
			chromedp.ActionFunc(func(context.Context) error {
				log.Printf("Sleeping %d MS before checking for login fields", cfg.General.PageLoadDelayMS)
				time.Sleep(time.Duration(cfg.General.PageLoadDelayMS) * time.Millisecond)
				return nil
			}),
			chromedp.WaitVisible(`//input[@name="user"]`, chromedp.BySearch),
			chromedp.SendKeys(`//input[@name="user"]`, cfg.Target.Username, chromedp.BySearch),
			chromedp.SendKeys(`//input[@name="password"]`, cfg.Target.Password+kb.Enter, chromedp.BySearch),
			postNavigate(cfg),
		); err != nil {
			panic(err)
		}
	}

	// blocking wait until context is cancelled or a message triggers a reload
	for {
		select {
		case <-ctx.Done():
			return
		case messageFromChrome := <-messages:
			if err := chromedp.Run(taskCtx,
				chromedp.Navigate(generatedURL),
				postNavigate(cfg),
			); err != nil {
				return
			}
			log.Println("Chromium output:", messageFromChrome)
		}
	}
}
