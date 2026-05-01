package local

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"

	"github.com/grafana/grafana-kiosk/pkg/browser"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/config"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/login/shared"
)

// BypassURL extracts the base URL and appends /login/local to bypass
// OAuth auto-login and use a local account instead.
func BypassURL(rawURL string) string {
	startIndex := strings.Index(rawURL, "://") + 3
	slashIndex := strings.Index(rawURL[startIndex:], "/")

	var baseURL string
	if slashIndex == -1 {
		baseURL = rawURL
	} else {
		baseURL = rawURL[:startIndex+slashIndex]
	}

	return baseURL + "/login/local"
}

// loginWithCredentials waits for login fields and submits credentials.
func loginWithCredentials(ctx context.Context, b browser.Browser, username, password string) error {
	if err := b.WaitVisible(ctx, `//input[@name="user"]`); err != nil {
		return err
	}
	if err := b.SendKeys(ctx, `//input[@name="user"]`, username); err != nil {
		return err
	}
	return b.SendKeys(ctx, `//input[@name="password"]`, password+kb.Enter)
}

// Run creates a chrome-based kiosk using a local grafana-server account.
func Run(ctx context.Context, cfg *config.Config, dir string, b browser.Browser, messages chan string) {
	taskCtx, cancel := shared.NewBrowserContext(ctx, cfg, dir, shared.TargetCrashed)
	defer cancel()

	if err := chromedp.Run(taskCtx,
		shared.WaitForPageLoad(cfg),
		shared.CycleWindowState(cfg),
	); err != nil {
		panic(err)
	}

	if err := localLoginFlow(taskCtx, cfg, b, shared.GenerateURL(cfg), messages); err != nil {
		panic(err)
	}
}

// localLoginFlow drives the local-account login sequence and then blocks until
// context is cancelled or a message triggers a reload.
func localLoginFlow(ctx context.Context, cfg *config.Config, b browser.Browser, dashboardURL string, messages chan string) error {
	log.Printf("Navigating to %s", dashboardURL)
	delay := time.Duration(cfg.General.PageLoadDelayMS) * time.Millisecond

	if cfg.GoAuth.AutoLogin {
		bypassURL := BypassURL(cfg.Target.URL)

		log.Printf("Bypassing autoLogin using URL %s", bypassURL)

		if err := b.Navigate(ctx, bypassURL); err != nil {
			return err
		}

		if delay > 0 {
			log.Printf("Sleeping %d MS before checking for login fields", cfg.General.PageLoadDelayMS)
			time.Sleep(delay)
		}

		if err := loginWithCredentials(ctx, b, cfg.Target.Username, cfg.Target.Password); err != nil {
			return err
		}

		if delay > 0 {
			log.Printf("Sleeping %d MS before checking for topnav", cfg.General.PageLoadDelayMS)
			time.Sleep(delay)
		}

		if err := b.WaitVisible(ctx, `//img[@alt="User avatar"]`); err != nil {
			return err
		}

		if delay > 0 {
			log.Printf("Sleeping %d MS before navigating to final url", cfg.General.PageLoadDelayMS)
			time.Sleep(delay)
		}

		if err := b.Navigate(ctx, dashboardURL); err != nil {
			return err
		}
	} else {
		if delay > 0 {
			log.Printf("Sleeping %d MS before navigating to final url", cfg.General.PageLoadDelayMS)
			time.Sleep(delay)
		}

		if err := b.Navigate(ctx, dashboardURL); err != nil {
			return err
		}

		if delay > 0 {
			log.Printf("Sleeping %d MS before checking for login fields", cfg.General.PageLoadDelayMS)
			time.Sleep(delay)
		}

		if err := loginWithCredentials(ctx, b, cfg.Target.Username, cfg.Target.Password); err != nil {
			return err
		}
	}

	return shared.RunMessageLoop(ctx, b, dashboardURL, messages)
}
