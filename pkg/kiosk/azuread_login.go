package kiosk

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/chromedp"

	"github.com/grafana/grafana-kiosk/pkg/browser"
)

// GrafanaKioskAzureAD creates a chrome-based kiosk using an Azure Active Directory authenticated account.
func GrafanaKioskAzureAD(ctx context.Context, cfg *Config, dir string, b browser.Browser, messages chan string) {
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

	if err := chromedp.Run(taskCtx,
		waitForPageLoad(cfg),
		cycleWindowState(cfg),
	); err != nil {
		panic(err)
	}

	if err := azureADLoginFlow(taskCtx, cfg, b, GenerateURL(cfg), messages); err != nil {
		panic(err)
	}
}

// azureADLoginFlow navigates to Grafana, clicks the AzureAD login button, fills
// in Microsoft credentials, then blocks until context is cancelled or a message
// triggers a reload.
//
// Microsoft login page fields:
//
//	email:    input[name="loginfmt"]  (type="email")
//	password: input[name="passwd"]    (type="password")
//	next/submit button: input[id="idSIButton9"]
func azureADLoginFlow(ctx context.Context, cfg *Config, b browser.Browser, url string, messages chan string) error {
	log.Println("Navigating to ", url)

	log.Println("waiting for azuread login button")
	if err := b.Navigate(ctx, url); err != nil {
		return err
	}
	if err := b.WaitVisible(ctx, `//a[contains(@href,'login/azuread')]`); err != nil {
		return err
	}
	log.Println("azuread login button detected")
	if err := b.Click(ctx, `//a[contains(@href,'login/azuread')]`); err != nil {
		return err
	}
	log.Println("azuread button clicked, waiting for Microsoft login page")
	time.Sleep(1 * time.Second)

	if err := b.WaitVisible(ctx, `//input[@name="loginfmt"]`); err != nil {
		return err
	}
	log.Println("Microsoft login page detected, entering username")
	if err := b.SendKeys(ctx, `//input[@name="loginfmt"]`, cfg.Target.Username); err != nil {
		return err
	}
	if err := b.Click(ctx, `//input[@id="idSIButton9"]`); err != nil {
		return err
	}
	log.Println("username submitted, waiting for password field")
	time.Sleep(1 * time.Second)

	if err := b.WaitVisible(ctx, `//input[@name="passwd"]`); err != nil {
		return err
	}
	log.Println("password field detected, entering password")
	if err := b.SendKeys(ctx, `//input[@name="passwd"]`, cfg.Target.Password); err != nil {
		return err
	}
	if err := b.WaitVisible(ctx, `//input[@id="idSIButton9"]`); err != nil {
		return err
	}
	log.Println("clicking sign in button")
	if err := b.Click(ctx, `//input[@id="idSIButton9"]`); err != nil {
		return err
	}
	log.Println("sign in button clicked")
	time.Sleep(1 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil
		case messageFromBrowser := <-messages:
			if err := b.Navigate(ctx, url); err != nil {
				return nil
			}
			log.Println("Browser output:", messageFromBrowser)
		}
	}
}
