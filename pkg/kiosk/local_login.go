package kiosk

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

// GrafanaKioskLocal creates a chrome-based kiosk using a local grafana-server account.
func GrafanaKioskLocal(cfg *Config, messages chan string) {
	dir, err := os.MkdirTemp(os.TempDir(), "chromedp-kiosk")
	if err != nil {
		panic(err)
	}

	log.Println("Using temp dir:", dir)
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

	var generatedURL = GenerateURL(cfg.Target.URL, cfg.General.Mode, cfg.General.AutoFit, cfg.Target.IsPlayList)

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
		startIndex := strings.Index(cfg.Target.URL, "://") + 3
		endIndex := strings.Index(cfg.Target.URL[startIndex:], "/") + startIndex
		baseURL := cfg.Target.URL[:endIndex]
		bypassURL := baseURL + "/login/local"

		log.Println("Bypassing Azure AD autoLogin at ", bypassURL)

		if err := chromedp.Run(taskCtx,
			chromedp.Navigate(bypassURL),
			chromedp.WaitVisible(`//input[@name="user"]`, chromedp.BySearch),
			chromedp.SendKeys(`//input[@name="user"]`, cfg.Target.Username, chromedp.BySearch),
			chromedp.SendKeys(`//input[@name="password"]`, cfg.Target.Password+kb.Enter, chromedp.BySearch),
			chromedp.WaitVisible(`//img[@alt="User avatar"]`, chromedp.BySearch),
			chromedp.Navigate(generatedURL),
		); err != nil {
			panic(err)
		}
	} else {
		if err := chromedp.Run(taskCtx,
			chromedp.Navigate(generatedURL),
			chromedp.WaitVisible(`//input[@name="user"]`, chromedp.BySearch),
			chromedp.SendKeys(`//input[@name="user"]`, cfg.Target.Username, chromedp.BySearch),
			chromedp.SendKeys(`//input[@name="password"]`, cfg.Target.Password+kb.Enter, chromedp.BySearch),
		); err != nil {
			panic(err)
		}
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
