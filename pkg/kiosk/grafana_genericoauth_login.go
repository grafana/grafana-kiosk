package kiosk

import (
	"context"
	"log"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"

	"github.com/grafana/grafana-kiosk/pkg/browser"
)

// GrafanaKioskGenericOauth creates a chrome-based kiosk using a oauth2 authenticated account.
func GrafanaKioskGenericOauth(ctx context.Context, cfg *Config, dir string, b browser.Browser, messages chan string) {
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

	if err := genericOauthLoginFlow(taskCtx, cfg, b, GenerateURL(cfg), messages); err != nil {
		panic(err)
	}
}

// genericOauthLoginFlow navigates to the Grafana login page, optionally clicks
// the OAuth button, fills in credentials, handles stay-signed-in prompts, then
// blocks until context is cancelled or a message triggers a reload.
func genericOauthLoginFlow(ctx context.Context, cfg *Config, b browser.Browser, url string, messages chan string) error {
	log.Println("Navigating to ", url)
	log.Println("Oauth_Auto_Login enabled: ", cfg.GoAuth.AutoLogin)

	if err := b.Navigate(ctx, url); err != nil {
		return err
	}

	if !cfg.GoAuth.AutoLogin {
		// XPATH of Generic OAUTH login button = //*[@href="login/generic_oauth"]
		if err := b.WaitVisible(ctx, `//*[@href="login/generic_oauth"]`); err != nil {
			return err
		}
		if err := b.Click(ctx, `//*[@href="login/generic_oauth"]`); err != nil {
			return err
		}
	}

	waitForBrowserStartup(cfg)

	// Fill out OAuth login page
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
