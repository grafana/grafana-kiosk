package kiosk

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/chromedp"
)

// GenerateURL constructs URL with appropriate parameters for kiosk mode.
func GenerateURL(cfg *Config) string {
	parsedURI, _ := url.ParseRequestURI(cfg.Target.URL)
	parsedQuery, _ := url.ParseQuery(parsedURI.RawQuery)

	switch cfg.General.Mode {
	case "tv": // TV
		parsedQuery.Set("kiosk", "tv") // no sidebar, topnav without buttons
		log.Printf("KioskMode: TV")
	case "full": // FULLSCREEN
		parsedQuery.Set("kiosk", "1") // sidebar and topnav always shown
		log.Printf("KioskMode: Fullscreen")
	case "disabled": // FULLSCREEN
		log.Printf("KioskMode: Disabled")
	default: // disabled
		parsedQuery.Set("kiosk", "1") // sidebar and topnav always shown
		log.Printf("KioskMode: Fullscreen")
	}
	if cfg.General.HideLinks {
		parsedQuery.Set("_dash.hideLinks", "")
	}
	if cfg.General.HideTimePicker {
		parsedQuery.Set("_dash.hideTimePicker", "")
	}
	if cfg.General.HideVariables {
		parsedQuery.Set("_dash.hideVariables", "")
	}
	// a playlist should also go inactive immediately
	if cfg.Target.IsPlayList {
		parsedQuery.Set("inactive", "1")
	}
	parsedURI.RawQuery = parsedQuery.Encode()
	// grafana is not parsing autofitpanels that uses an equals sign, so leave it out
	if cfg.General.AutoFit {
		if len(parsedQuery) > 0 {
			parsedURI.RawQuery += "&autofitpanels"
		} else {
			parsedURI.RawQuery += "autofitpanels"
		}
	}

	return parsedURI.String()
}

func generateExecutorOptions(dir string, cfg *Config) []chromedp.ExecAllocatorOption {
	// agent should not have the v prefix
	buildVersion := strings.TrimPrefix(cfg.BuildInfo.Version, "v")
	kioskVersion := fmt.Sprintf("GrafanaKiosk/%s (%s %s)", buildVersion, runtime.GOOS, runtime.GOARCH)
	userAgent := fmt.Sprintf("Mozilla/5.0 (X11; CrOS armv7l 13597.84.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 %s", kioskVersion)

	// See https://peter.sh/experiments/chromium-command-line-switches/

	// --start-fullscreen
	//    Specifies if the browser should start in fullscreen mode, like if the user had pressed F11 right after startup. ↪
	// --start-maximized
	//    Starts the browser maximized, regardless of any previous settings
	// --disable-gpu
	//    Disables GPU hardware acceleration. If software renderer is not in place, then the GPU process won't launch.--disable-gpu

	// Chromium 130+ enables HTTPS-First Mode by default, which blocks
	// plain HTTP page loads with ERR_BLOCKED_BY_CLIENT. Disable it when
	// the target URL uses HTTP.
	disableFeatures := "Translate"
	if strings.HasPrefix(cfg.Target.URL, "http://") {
		disableFeatures += ",HttpsUpgrades"
	}

	execAllocatorOption := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("autoplay-policy", "no-user-gesture-required"),
		chromedp.Flag("bwsi", true),
		chromedp.Flag("check-for-update-interval", "31536000"),
		chromedp.Flag("password-store", "basic"), // prevent key store popup
		chromedp.Flag("disable-features", disableFeatures),
		chromedp.Flag("disable-notifications", true),
		chromedp.Flag("disable-overlay-scrollbar", true),
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("disable-search-engine-choice-screen", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("ignore-certificate-errors", cfg.Target.IgnoreCertificateErrors),
		chromedp.Flag("incognito", cfg.General.Incognito),
		chromedp.Flag("kiosk", true),
		chromedp.Flag("noerrdialogs", true),
		chromedp.Flag("start-fullscreen", true),
		chromedp.Flag("start-maximized", true),
		chromedp.Flag("user-agent", userAgent),
		chromedp.Flag("window-position", cfg.General.WindowPosition),
		chromedp.UserDataDir(dir),
	}

	if !cfg.General.GPUEnabled {
		execAllocatorOption = append(
			execAllocatorOption,
			chromedp.Flag("disable-gpu", cfg.General.GPUEnabled))
	}
	if cfg.General.OzonePlatform != "" {
		execAllocatorOption = append(
			execAllocatorOption,
			chromedp.Flag("ozone-platform", cfg.General.OzonePlatform))
	}
	if cfg.General.WindowSize != "" {
		execAllocatorOption = append(
			execAllocatorOption,
			chromedp.Flag("kiosk", false),
			chromedp.Flag("start-fullscreen", false),
			// force app mode (no address bar and controls)
			chromedp.Flag("app", "data:text/html,<title>Grafana</title>"),
			chromedp.Flag("window-size", cfg.General.WindowSize))
	}
	if cfg.General.ScaleFactor != "" {
		execAllocatorOption = append(
			execAllocatorOption,
			chromedp.Flag("force-device-scale-factor", cfg.General.ScaleFactor))
	}

	return execAllocatorOption
}

// cycleWindowState cycles the browser window state via CDP before navigation.
// When no custom window size is set, it cycles normal → fullscreen.
// When a custom window size is set, it cycles minimized → normal with the
// specified dimensions. This forces Chrome to properly register viewport
// dimensions so Grafana sees the correct size on initial page load.
func cycleWindowState(cfg *Config) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		log.Println("Resetting window state via CDP")
		windowID, _, err := browser.GetWindowForTarget().Do(ctx)
		if err != nil {
			return fmt.Errorf("get window for target: %w", err)
		}
		if cfg.General.WindowSize != "" {
			return cycleWindowToSize(windowID, cfg.General.WindowSize, ctx)
		}
		err = browser.SetWindowBounds(windowID, &browser.Bounds{
			WindowState: browser.WindowStateNormal,
		}).Do(ctx)
		if err != nil {
			return fmt.Errorf("set window normal: %w", err)
		}
		time.Sleep(100 * time.Millisecond)
		return browser.SetWindowBounds(windowID, &browser.Bounds{
			WindowState: browser.WindowStateFullscreen,
		}).Do(ctx)
	})
}

// cycleWindowToSize cycles the window from minimized to normal with the
// specified dimensions to force Chrome to register the correct viewport.
func cycleWindowToSize(windowID browser.WindowID, windowSize string, ctx context.Context) error {
	parts := strings.SplitN(windowSize, ",", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid window-size format: %q", windowSize)
	}
	width, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
	if err != nil {
		return fmt.Errorf("parse window width: %w", err)
	}
	height, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
	if err != nil {
		return fmt.Errorf("parse window height: %w", err)
	}
	err = browser.SetWindowBounds(windowID, &browser.Bounds{
		WindowState: browser.WindowStateMinimized,
	}).Do(ctx)
	if err != nil {
		return fmt.Errorf("set window minimized: %w", err)
	}
	time.Sleep(100 * time.Millisecond)
	return browser.SetWindowBounds(windowID, &browser.Bounds{
		WindowState: browser.WindowStateNormal,
		Width:       width,
		Height:      height,
	}).Do(ctx)
}

// waitForPageLoad pauses to allow the browser to finish loading.
func waitForPageLoad(cfg *Config) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(_ context.Context) error {
		if cfg.General.PageLoadDelayMS <= 0 {
			return nil
		}
		log.Printf("Sleeping %d MS for page load", cfg.General.PageLoadDelayMS)
		time.Sleep(time.Duration(cfg.General.PageLoadDelayMS) * time.Millisecond)
		return nil
	})
}

// waitForBrowserStartup pauses to allow the browser process to become idle.
func waitForBrowserStartup(cfg *Config) {
	if cfg.General.PageLoadDelayMS <= 0 {
		return
	}
	log.Printf("Sleeping %d MS waiting for browser startup", cfg.General.PageLoadDelayMS)
	time.Sleep(time.Duration(cfg.General.PageLoadDelayMS) * time.Millisecond)
}
