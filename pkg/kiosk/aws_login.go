package kiosk

import (
	"context"
	"log"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"

	"github.com/grafana/grafana-kiosk/pkg/browser"
)

// GrafanaKioskAWSLogin Provides login for AWS Managed Grafana instances
func GrafanaKioskAWSLogin(ctx context.Context, cfg *Config, dir string, b browser.Browser, messages chan string) {
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

	// cycleWindowState runs before awsLoginFlow's Navigate — no fetch
	// interception is involved so the two-step ordering is safe.
	if err := chromedp.Run(taskCtx,
		waitForPageLoad(cfg),
		cycleWindowState(cfg),
	); err != nil {
		panic(err)
	}

	if err := awsLoginFlow(taskCtx, cfg, b, GenerateURL(cfg), messages); err != nil {
		panic(err)
	}
}

// awsLoginFlow navigates to the AWS Managed Grafana login page, accepts the
// cookie banner, clicks the SSO button, fills in credentials, waits for MFA
// if enabled, then blocks until context is cancelled or a message triggers a reload.
func awsLoginFlow(ctx context.Context, cfg *Config, b browser.Browser, dashboardURL string, messages chan string) error {
	log.Println("Navigating to ", dashboardURL)
	if err := b.Navigate(ctx, dashboardURL); err != nil {
		return err
	}
	if err := b.WaitVisible(ctx, `//a[contains(@href,'login/sso')]`); err != nil {
		return err
	}
	if err := b.WaitVisible(ctx, `div#awsccc-cb-buttons`); err != nil {
		return err
	}
	if err := b.Click(ctx, `//button[contains(@data-id,'awsccc-cb-btn-accept')]`); err != nil {
		return err
	}
	if err := b.Click(ctx, `//a[contains(@href,'login/sso')]`); err != nil {
		return err
	}
	if err := b.WaitVisible(ctx, `input#awsui-input-0`); err != nil {
		return err
	}
	if err := b.SendKeys(ctx, `input#awsui-input-0`, cfg.Target.Username+kb.Enter); err != nil {
		return err
	}
	if err := b.WaitVisible(ctx, `input#awsui-input-1`); err != nil {
		return err
	}
	if err := b.SendKeys(ctx, `input#awsui-input-1`, cfg.Target.Password+kb.Enter); err != nil {
		return err
	}
	if cfg.Target.UseMFA {
		if err := b.WaitNotVisible(ctx, `input#awsui-input-2`); err != nil {
			return err
		}
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case messageFromBrowser := <-messages:
			if err := b.Navigate(ctx, dashboardURL); err != nil {
				return nil
			}
			log.Println("Browser output:", messageFromBrowser)
		}
	}
}
