package goauth

import (
	"context"
	"log"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"

	"github.com/grafana/grafana-kiosk/pkg/browser"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/config"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/login/shared"
)

// Run creates a chrome-based kiosk using a oauth2 authenticated account.
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

	if err := genericOauthLoginFlow(taskCtx, cfg, b, shared.GenerateURL(cfg), messages); err != nil {
		panic(err)
	}
}

// genericOauthLoginFlow navigates to the Grafana login page, optionally clicks
// the OAuth button, fills in credentials, handles stay-signed-in prompts, then
// blocks until context is cancelled or a message triggers a reload.
func genericOauthLoginFlow(ctx context.Context, cfg *config.Config, b browser.Browser, dashboardURL string, messages chan string) error {
	log.Printf("Navigating to %s", dashboardURL)
	log.Println("Oauth_Auto_Login enabled: ", cfg.GoAuth.AutoLogin)

	if err := b.Navigate(ctx, dashboardURL); err != nil {
		return err
	}

	if !cfg.GoAuth.AutoLogin {
		if err := b.WaitVisible(ctx, `//*[@href="login/generic_oauth"]`); err != nil {
			return err
		}
		if err := b.Click(ctx, `//*[@href="login/generic_oauth"]`); err != nil {
			return err
		}
	}

	shared.WaitForBrowserStartup(cfg)

	if cfg.GoAuth.WaitForPasswordField {
		if err := b.WaitVisible(ctx, `//input[@name="`+cfg.GoAuth.UsernameField+`"]`); err != nil {
			return err
		}
		if err := b.SendKeys(ctx, `//input[@name="`+cfg.GoAuth.UsernameField+`"]`, cfg.Target.Username+kb.Enter); err != nil {
			return err
		}
		if err := b.WaitVisible(ctx, `//input[@name="`+cfg.GoAuth.PasswordField+`" and not(@class="`+cfg.GoAuth.WaitForPasswordFieldIgnoreClass+`")]`); err != nil {
			return err
		}
		if err := b.SendKeys(ctx, `//input[@name="`+cfg.GoAuth.PasswordField+`"]`, cfg.Target.Password+kb.Enter); err != nil {
			return err
		}
	} else {
		if err := b.WaitVisible(ctx, `//input[@name="`+cfg.GoAuth.UsernameField+`"]`); err != nil {
			return err
		}
		if err := b.SendKeys(ctx, `//input[@name="`+cfg.GoAuth.UsernameField+`"]`, cfg.Target.Username); err != nil {
			return err
		}
		if err := b.SendKeys(ctx, `//input[@name="`+cfg.GoAuth.PasswordField+`"]`, cfg.Target.Password+kb.Enter); err != nil {
			return err
		}
	}

	if cfg.GoAuth.WaitForStaySignedInPrompt {
		if err := b.WaitVisible(ctx, `//input[@type="submit" and @value="Yes"]`); err != nil {
			return err
		}
		if err := b.Click(ctx, `//input[@type="submit" and @value="Yes"]`); err != nil {
			return err
		}
	}

	return shared.RunMessageLoop(ctx, b, dashboardURL, messages)
}
