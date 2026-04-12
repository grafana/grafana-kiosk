package kiosk

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

// GrafanaKioskAzureAD creates a chrome-based kiosk using an Azure Active Directory authenticated account.
func GrafanaKioskAzureAD(ctx context.Context, cfg *Config, dir string, messages chan string) {
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

	// Give browser time to load
	log.Printf("Sleeping %d MS before navigating to url", cfg.General.PageLoadDelayMS)
	time.Sleep(time.Duration(cfg.General.PageLoadDelayMS) * time.Millisecond)

	var generatedURL = GenerateURL(cfg)

	log.Println("Navigating to ", generatedURL)
	/*
		Launch chrome, click the azuread button on the Grafana login page,
		which redirects to the Microsoft login page. Fill out the email and
		password fields and submit.

		Microsoft login page fields:
		  email:    input[name="loginfmt"]  (type="email")
		  password: input[name="passwd"]    (type="password")
		  next/submit button: input[id="idSIButton9"]
	*/

	// Click the AzureAD login button on the Grafana login page
	if err := chromedp.Run(taskCtx,
		chromedp.Navigate(generatedURL),
		chromedp.ActionFunc(func(context.Context) error {
			log.Println("waiting for azuread login button")
			return nil
		}),
		chromedp.WaitVisible(`//a[contains(@href,'login/azuread')]`, chromedp.BySearch),
		chromedp.ActionFunc(func(context.Context) error {
			log.Println("azuread login button detected")
			return nil
		}),
		chromedp.Click(`//a[contains(@href,'login/azuread')]`, chromedp.BySearch),
		chromedp.ActionFunc(func(context.Context) error {
			log.Println("azuread button clicked, waiting for Microsoft login page")
			time.Sleep(1 * time.Second)
			return nil
		}),
	); err != nil {
		panic(err)
	}

	// Fill out the Microsoft login email field
	if err := chromedp.Run(taskCtx,
		chromedp.WaitVisible(`//input[@name="loginfmt"]`, chromedp.BySearch),
		chromedp.ActionFunc(func(context.Context) error {
			log.Println("Microsoft login page detected, entering username")
			return nil
		}),
		chromedp.SendKeys(`//input[@name="loginfmt"]`, cfg.Target.Username, chromedp.BySearch),
		chromedp.Click(`//input[@id="idSIButton9"]`, chromedp.BySearch),
		chromedp.ActionFunc(func(context.Context) error {
			log.Println("username submitted, waiting for password field")
			time.Sleep(1 * time.Second)
			return nil
		}),
	); err != nil {
		panic(err)
	}

	// Fill out the Microsoft login password field and click Sign in
	if err := chromedp.Run(taskCtx,
		chromedp.WaitVisible(`//input[@name="passwd"]`, chromedp.BySearch),
		chromedp.ActionFunc(func(context.Context) error {
			log.Println("password field detected, entering password")
			return nil
		}),
		chromedp.SendKeys(`//input[@name="passwd"]`, cfg.Target.Password, chromedp.BySearch),
		chromedp.WaitVisible(`//input[@id="idSIButton9"]`, chromedp.BySearch),
		chromedp.ActionFunc(func(context.Context) error {
			log.Println("clicking sign in button")
			return nil
		}),
		chromedp.Click(`//input[@id="idSIButton9"]`, chromedp.BySearch),
		chromedp.ActionFunc(func(context.Context) error {
			log.Println("sign in button clicked")
			time.Sleep(1 * time.Second)
			return nil
		}),
	); err != nil {
		panic(err)
	}

	if err := chromedp.Run(taskCtx, triggerAutofit(cfg)); err != nil {
		panic(err)
	}

	// blocking wait for reload messages
	for {
		messageFromChrome := <-messages
		if err := chromedp.Run(taskCtx,
			chromedp.Navigate(generatedURL),
			triggerAutofit(cfg),
		); err != nil {
			panic(err)
		}
		log.Println("Chromium output:", messageFromChrome)
	}
}
