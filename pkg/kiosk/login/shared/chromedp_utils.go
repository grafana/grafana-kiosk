package shared

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/chromedp"

	"github.com/grafana/grafana-kiosk/pkg/kiosk/config"
)

// EdgeBinaryCandidates lists executable names to look up on PATH when the
// user requests the Edge browser. Order matters: first match wins.
var EdgeBinaryCandidates = []string{
	"msedge",
	"microsoft-edge",
	"microsoft-edge-stable",
}

// LookPath is overridable in tests.
var LookPath = exec.LookPath

// GenerateURL constructs URL with appropriate parameters for kiosk mode.
func GenerateURL(cfg *config.Config) string {
	parsedURI, _ := url.ParseRequestURI(cfg.Target.URL)
	parsedQuery, _ := url.ParseQuery(parsedURI.RawQuery)

	switch cfg.General.Mode {
	case "tv": // TV
		parsedQuery.Set("kiosk", "tv") // no sidebar, topnav without buttons
		log.Printf("KioskMode: TV")
	case "full": // FULLSCREEN
		parsedQuery.Set("kiosk", "1") // sidebar and topnav always shown
		log.Printf("KioskMode: Fullscreen")
	case "disabled":
		log.Printf("KioskMode: Disabled")
	default: // disabled
		parsedQuery.Set("kiosk", "1") // sidebar and topnav always shown
		log.Printf("KioskMode: Fullscreen")
	}
	if cfg.General.HideLinks {
		parsedQuery.Set("_dash.hideLinks", "true")
	}
	if cfg.General.HideLogo {
		parsedQuery.Set("hideLogo", "1")
	}
	if cfg.General.HidePlaylistNav {
		parsedQuery.Set("_dash.hidePlaylistNav", "true")
	}
	if cfg.General.HideTimePicker {
		parsedQuery.Set("_dash.hideTimePicker", "true")
	}
	if cfg.General.HideVariables {
		parsedQuery.Set("_dash.hideVariables", "true")
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

// GenerateExecutorOptions builds the chromedp ExecAllocator options from cfg.
func GenerateExecutorOptions(dir string, cfg *config.Config) []chromedp.ExecAllocatorOption {
	// agent should not have the v prefix
	buildVersion := strings.TrimPrefix(cfg.BuildInfo.Version, "v")
	versionTag := fmt.Sprintf("GrafanaKiosk/%s (%s %s)", buildVersion, runtime.GOOS, runtime.GOARCH)
	userAgent := fmt.Sprintf("Mozilla/5.0 (X11; CrOS armv7l 13597.84.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 %s", versionTag)

	// Chromium 130+ enables HTTPS-First Mode by default, which blocks
	// plain HTTP page loads with ERR_BLOCKED_BY_CLIENT. Disable it when
	// the target URL uses HTTP.
	disableFeatures := "Translate"
	if strings.HasPrefix(cfg.Target.URL, "http://") {
		disableFeatures += ",HttpsUpgrades"
	}

	execAllocatorOption := []chromedp.ExecAllocatorOption{
		// Skip first-run wizard and default browser prompt on startup.
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		// Allow media (video, audio) to play without a user gesture — needed for
		// dashboard panels that embed auto-playing content.
		chromedp.Flag("autoplay-policy", "no-user-gesture-required"),
		// Browse Without Sign In — suppress Google account sign-in prompts.
		chromedp.Flag("bwsi", true),
		// Set update check interval to 1 year — suppresses update-available prompts.
		chromedp.Flag("check-for-update-interval", "31536000"),
		// Use the basic (non-OS) password store to prevent OS key-ring popups.
		chromedp.Flag("password-store", "basic"),
		// Disable translation bar and (for HTTP URLs) HTTPS upgrades.
		chromedp.Flag("disable-features", disableFeatures),
		// Suppress browser notification permission prompts.
		chromedp.Flag("disable-notifications", true),
		// Hide overlay scrollbars for a cleaner kiosk appearance.
		chromedp.Flag("disable-overlay-scrollbar", true),
		chromedp.Flag("hide-scrollbars", true),
		// Suppress the one-time search engine choice dialog (Chrome 121+).
		chromedp.Flag("disable-search-engine-choice-screen", true),
		// No Google account sync — kiosk has no user account.
		chromedp.Flag("disable-sync", true),
		// Prevent JS timers from being throttled when the window loses focus —
		// critical for Grafana auto-refresh on playlist or multi-monitor setups.
		chromedp.Flag("disable-background-timer-throttling", true),
		// Prevent the renderer from being given lower priority when unfocused.
		chromedp.Flag("disable-renderer-backgrounding", true),
		// Prevent backgrounding when another window covers this one.
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		// Suppress "Page Unresponsive" dialogs that would disrupt the kiosk
		// display when Grafana takes time to load a heavy dashboard.
		chromedp.Flag("disable-hang-monitor", true),
		// Disable crash reporting — no crash dialogs or background reporter process.
		chromedp.Flag("disable-breakpad", true),
		// Prevent background component update checks from interfering with rendering.
		chromedp.Flag("disable-component-update", true),
		// Prevent Safe Browsing database updates from running in the background.
		chromedp.Flag("safebrowsing-disable-auto-update", true),
		// Record metrics locally only — suppresses telemetry uploads to Google.
		chromedp.Flag("metrics-recording-only", true),
		// Allow popups/new windows from Grafana drill-down links without prompting.
		chromedp.Flag("disable-popup-blocking", true),
		// Configurable: ignore TLS certificate errors (e.g., self-signed certs).
		chromedp.Flag("ignore-certificate-errors", cfg.Target.IgnoreCertificateErrors),
		// Configurable: run in incognito mode (no persistent profile state).
		chromedp.Flag("incognito", cfg.General.Incognito),
		// Enable Chromium kiosk mode — hides address bar and browser chrome.
		chromedp.Flag("kiosk", true),
		// Suppress error dialogs (e.g., renderer crash dialogs).
		chromedp.Flag("noerrdialogs", true),
		chromedp.Flag("start-fullscreen", true),
		chromedp.Flag("start-maximized", true),
		// Custom user-agent identifies the kiosk version to Grafana logs.
		chromedp.Flag("user-agent", userAgent),
		chromedp.Flag("window-position", cfg.General.WindowPosition),
		chromedp.UserDataDir(dir),
	}

	if cfg.General.Headless {
		execAllocatorOption = append(execAllocatorOption,
			chromedp.Flag("headless", "new"),
		)
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
		fullscreen := isFullscreenMode(cfg.General.Mode)
		if fullscreen {
			log.Printf("window-size %s with kiosk mode %q: window will cycle to fullscreen via CDP", cfg.General.WindowSize, cfg.General.Mode)
		}
		execAllocatorOption = append(
			execAllocatorOption,
			chromedp.Flag("kiosk", fullscreen),
			chromedp.Flag("start-fullscreen", fullscreen),
			// force app mode (no address bar and controls)
			chromedp.Flag("app", "data:text/html,<title>Grafana</title>"),
			chromedp.Flag("window-size", cfg.General.WindowSize))
	}
	if cfg.General.ScaleFactor != "" {
		execAllocatorOption = append(
			execAllocatorOption,
			chromedp.Flag("force-device-scale-factor", cfg.General.ScaleFactor))
	}

	if path := ResolveBrowserExecPath(cfg); path != "" {
		log.Printf("Using browser executable: %s", path)
		execAllocatorOption = append(execAllocatorOption, chromedp.ExecPath(path))
	}

	return execAllocatorOption
}

// ResolveBrowserExecPath returns the explicit browser executable path to pass to
// chromedp.ExecPath. An empty string means "let chromedp auto-detect".
// ResolveBrowserExecPath returns the explicit browser executable path to pass to
// chromedp.ExecPath. An empty string means "let chromedp auto-detect".
func ResolveBrowserExecPath(cfg *config.Config) string {
	if cfg.General.BrowserPath != "" {
		return cfg.General.BrowserPath
	}
	switch strings.ToLower(cfg.General.Browser) {
	case "", "chrome":
		return ""
	case "edge":
		for _, name := range EdgeBinaryCandidates {
			if p, err := LookPath(name); err == nil {
				return p
			}
		}
		log.Println("Browser 'edge' requested but no Edge binary found on PATH; set -browser-path or KIOSK_BROWSER_PATH to the Edge executable")
		return ""
	}
	return ""
}

// isFullscreenMode reports whether the given kiosk mode requires fullscreen.
func isFullscreenMode(mode string) bool {
	return mode == "full" || mode == ""
}

// CycleWindowState cycles the browser window state via CDP before navigation.
func CycleWindowState(cfg *config.Config) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		log.Println("Cycling window state via CDP")
		windowID, _, err := browser.GetWindowForTarget().Do(ctx)
		if err != nil {
			return fmt.Errorf("get window for target: %w", err)
		}
		if cfg.General.WindowSize != "" {
			return cycleWindowToSize(ctx, windowID, cfg.General.WindowSize, isFullscreenMode(cfg.General.Mode))
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

// cycleWindowToSize sets the window to the specified dimensions.
func cycleWindowToSize(ctx context.Context, windowID browser.WindowID, windowSize string, fullscreen bool) error {
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
		Width:  width,
		Height: height,
	}).Do(ctx)
	if err != nil {
		return fmt.Errorf("set window size: %w", err)
	}

	if !fullscreen {
		return nil
	}

	time.Sleep(100 * time.Millisecond)

	return browser.SetWindowBounds(windowID, &browser.Bounds{
		WindowState: browser.WindowStateFullscreen,
	}).Do(ctx)
}

// WaitForPageLoad pauses to allow the browser to finish loading.
func WaitForPageLoad(cfg *config.Config) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(_ context.Context) error {
		if cfg.General.PageLoadDelayMS <= 0 {
			return nil
		}
		log.Printf("Sleeping %d MS for page load", cfg.General.PageLoadDelayMS)
		time.Sleep(time.Duration(cfg.General.PageLoadDelayMS) * time.Millisecond)
		return nil
	})
}

// WaitForBrowserStartup pauses to allow the browser process to become idle.
func WaitForBrowserStartup(cfg *config.Config) {
	if cfg.General.PageLoadDelayMS <= 0 {
		return
	}
	log.Printf("Sleeping %d MS waiting for browser startup", cfg.General.PageLoadDelayMS)
	time.Sleep(time.Duration(cfg.General.PageLoadDelayMS) * time.Millisecond)
}
