package kiosk

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

	if path := resolveBrowserExecPath(cfg); path != "" {
		log.Printf("Using browser executable: %s", path)
		execAllocatorOption = append(execAllocatorOption, chromedp.ExecPath(path))
	}

	return execAllocatorOption
}

// edgeBinaryCandidates lists executable names to look up on PATH when the
// user requests the Edge browser. Order matters: first match wins.
var edgeBinaryCandidates = []string{
	"msedge",
	"microsoft-edge",
	"microsoft-edge-stable",
}

// lookPath is overridable in tests.
var lookPath = exec.LookPath

// resolveBrowserExecPath returns the explicit browser executable path that
// should be passed to chromedp.ExecPath. An empty string means "let chromedp
// auto-detect" (the default Chrome lookup).
//
// Precedence:
//  1. cfg.General.BrowserPath (verbatim) if set
//  2. cfg.General.Browser == "edge" → PATH lookup
//  3. cfg.General.Browser == "chrome" or empty → "" (chromedp default)
func resolveBrowserExecPath(cfg *Config) string {
	if cfg.General.BrowserPath != "" {
		return cfg.General.BrowserPath
	}
	switch strings.ToLower(cfg.General.Browser) {
	case "", "chrome":
		return ""
	case "edge":
		for _, name := range edgeBinaryCandidates {
			if p, err := lookPath(name); err == nil {
				return p
			}
		}
		log.Println("Browser 'edge' requested but no Edge binary found on PATH; set -browser-path or KIOSK_BROWSER_PATH to the Edge executable")
		return ""
	}
	return ""
}

// isFullscreenMode reports whether the given kiosk mode requires fullscreen.
// "full" and the empty default both mean fullscreen; "tv" and "disabled" do not.
func isFullscreenMode(mode string) bool {
	return mode == "full" || mode == ""
}

// cycleWindowState cycles the browser window state via CDP before navigation.
// When no custom window size is set, it cycles normal → fullscreen.
// When a custom window size is set, it cycles minimized → normal with the
// specified dimensions. This forces Chrome to properly register viewport
// dimensions so Grafana sees the correct size on initial page load.
func cycleWindowState(cfg *Config) chromedp.ActionFunc {
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
// When fullscreen is true, it then cycles to fullscreen to force Chrome to
// register the correct viewport. Otherwise it stays at the requested size.
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
