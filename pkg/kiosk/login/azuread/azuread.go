package azuread

import (
	"context"
	"log"

	"github.com/chromedp/chromedp"

	"github.com/grafana/grafana-kiosk/pkg/browser"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/config"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/login/shared"
)

// Run creates a chrome-based kiosk using an Azure Active Directory authenticated account.
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

	if err := chromedp.Run(taskCtx,
		shared.WaitForPageLoad(cfg),
		shared.CycleWindowState(cfg),
	); err != nil {
		panic(err)
	}

	if err := azureADLoginFlow(taskCtx, cfg, b, shared.GenerateURL(cfg), messages); err != nil {
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
func azureADLoginFlow(ctx context.Context, cfg *config.Config, b browser.Browser, dashboardURL string, messages chan string) error {
	log.Printf("Navigating to %s", dashboardURL)

	log.Println("waiting for azuread login button")
	if err := b.Navigate(ctx, dashboardURL); err != nil {
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
	shared.SleepPageLoad(cfg)

	return shared.RunMessageLoop(ctx, b, dashboardURL, messages)
}
