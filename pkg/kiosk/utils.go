package kiosk

import (
	"fmt"
	"log"
	"net/url"
	"runtime"
	"strings"

	"github.com/chromedp/chromedp"
)

// GenerateURL constructs URL with appropriate parameters for kiosk mode.
func GenerateURL(anURL string, kioskMode string, autoFit bool, isPlayList bool) string {
	parsedURI, _ := url.ParseRequestURI(anURL)
	parsedQuery, _ := url.ParseQuery(parsedURI.RawQuery)

	switch kioskMode {
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
	// a playlist should also go inactive immediately
	if isPlayList {
		parsedQuery.Set("inactive", "1")
	}
	parsedURI.RawQuery = parsedQuery.Encode()
	// grafana is not parsing autofitpanels that uses an equals sign, so leave it out
	if autoFit {
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

	execAllocatorOption := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("autoplay-policy", "no-user-gesture-required"),
		chromedp.Flag("bwsi", true),
		chromedp.Flag("check-for-update-interval", "31536000"),
		chromedp.Flag("disable-features", "Translate"),
		chromedp.Flag("disable-notifications", true),
		chromedp.Flag("disable-overlay-scrollbar", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("ignore-certificate-errors", cfg.Target.IgnoreCertificateErrors),
		chromedp.Flag("incognito", true),
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
			chromedp.Flag("window-size", cfg.General.WindowSize))
	}
	if cfg.General.ScaleFactor != "" {
		execAllocatorOption = append(
			execAllocatorOption,
			chromedp.Flag("force-device-scale-factor", cfg.General.ScaleFactor))
	}

	return execAllocatorOption
}
