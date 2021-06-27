package kiosk

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

// GrafanaKioskGenericOauth creates a chrome-based kiosk using a oauth2 authenticated account
func GrafanaKioskGenericOauth(cfg *Config) {
	dir, err := ioutil.TempDir("", "chromedp-example")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		// chromedp.DisableGPU, // needed?
		chromedp.Flag("noerrdialogs", true),
		chromedp.Flag("kiosk", true),
		chromedp.Flag("bwsi", true),
		chromedp.Flag("incognito", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-notifications", true),
		chromedp.Flag("disable-overlay-scrollbar", true),
		chromedp.UserDataDir(dir),
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// also set up a custom logger
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	listenChromeEvents(taskCtx, targetCrashed)

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
	log.Println("Oauth_Auto_Login enabeld: ", cfg.GOAUTH.AutoLogin)
	if cfg.GOAUTH.AutoLogin {
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

	// Give browser time to load next page (this can be prone to failure, explore different options vs sleeping)
	time.Sleep(2000 * time.Millisecond)
	// Fill out OAUTH login page
	if err := chromedp.Run(taskCtx,
		chromedp.WaitVisible(`//input[@name="`+cfg.GOAUTH.UsernameField+`"]`, chromedp.BySearch),
		chromedp.SendKeys(`//input[@name="`+cfg.GOAUTH.UsernameField+`"]`, cfg.Target.Username, chromedp.BySearch),
		chromedp.SendKeys(`//input[@name="`+cfg.GOAUTH.PasswordField+`"]`, cfg.Target.Password+kb.Enter, chromedp.BySearch),
		chromedp.WaitVisible(`notinputPassword`, chromedp.ByID),
	); err != nil {
		panic(err)
	}
}
