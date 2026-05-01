package aws

import (
	"context"
	"log"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"

	"github.com/grafana/grafana-kiosk/pkg/browser"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/config"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/login/shared"
)

// Run provides login for AWS Managed Grafana instances.
func Run(ctx context.Context, cfg *config.Config, dir string, b browser.Browser, messages chan string) {
	taskCtx, cancel := shared.NewBrowserContext(ctx, cfg, dir, shared.TargetCrashed)
	defer cancel()

	if err := chromedp.Run(taskCtx,
		shared.WaitForPageLoad(cfg),
		shared.CycleWindowState(cfg),
	); err != nil {
		panic(err)
	}

	if err := awsLoginFlow(taskCtx, cfg, b, shared.GenerateURL(cfg), messages); err != nil {
		panic(err)
	}
}

// awsLoginFlow navigates to the AWS Managed Grafana login page, accepts the
// cookie banner, clicks the SSO button, fills in credentials, waits for MFA
// if enabled, then blocks until context is cancelled or a message triggers a reload.
func awsLoginFlow(ctx context.Context, cfg *config.Config, b browser.Browser, dashboardURL string, messages chan string) error {
	log.Printf("Navigating to %s", dashboardURL)
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
	return shared.RunMessageLoop(ctx, b, dashboardURL, messages)
}
