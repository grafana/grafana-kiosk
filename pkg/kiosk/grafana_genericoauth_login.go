package kiosk

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

// GrafanaKioskGenericOauth creates a chrome-based kiosk using a oauth2 authenticated account.
func GrafanaKioskGenericOauth(cfg *Config, messages chan string) {
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
		Launch chrome, click the GENERIC OAUTH button, fill out login form and submit
	*/
	// XPATH of grafana.com for Generic OAUTH login button = //*[@href="login/grafana_com"]/i

	// Click the OAUTH login button
	log.Println("Oauth_Auto_Login enabled: ", cfg.GoAuth.AutoLogin)

	if cfg.GoAuth.AutoLogin {
		if err := chromedp.Run(taskCtx,
			chromedp.Navigate(generatedURL),
		); err != nil {
			panic(err)
		}
	} else {
		if err := chromedp.Run(taskCtx,
			chromedp.Navigate(generatedURL),
			chromedp.WaitVisible(`//*[@href="login/generic_oauth"]`, chromedp.BySearch),
			chromedp.Click(`//*[@href="login/generic_oauth"]`, chromedp.BySearch),
		); err != nil {
			panic(err)
		}
	}

	// Give browser time to load
	log.Printf("Sleeping %d MS before navigating to url", cfg.General.PageLoadDelayMS)
	time.Sleep(time.Duration(cfg.General.PageLoadDelayMS) * time.Millisecond)

	// Fill out OAUTH login page
	if err := chromedp.Run(taskCtx,
		chromedp.WaitVisible(`//input[@name="`+cfg.GoAuth.UsernameField+`"]`, chromedp.BySearch),
		chromedp.SendKeys(`//input[@name="`+cfg.GoAuth.UsernameField+`"]`, cfg.Target.Username, chromedp.BySearch),
		chromedp.SendKeys(`//input[@name="`+cfg.GoAuth.PasswordField+`"]`, cfg.Target.Password+kb.Enter, chromedp.BySearch),
	); err != nil {
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
