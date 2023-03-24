package kiosk

import (
	"fmt"
	"log"
	"net/url"
	"runtime"

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
	if autoFit {
		parsedQuery.Set("autofitpanels", "")
	}
	parsedURI.RawQuery = parsedQuery.Encode()

	return parsedURI.String()
}

func generateExecutorOptions(dir string, cfg *Config) []chromedp.ExecAllocatorOption {
	kioskVersion := fmt.Sprintf("GrafanaKiosk/%s (%s %s)", cfg.BuildInfo.Version, runtime.GOOS, runtime.GOARCH)
	userAgent := fmt.Sprintf("Mozilla/5.0 (X11; CrOS armv7l 13597.84.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 %s", kioskVersion)
	return []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("user-agent", userAgent),
		chromedp.Flag("noerrdialogs", true),
		chromedp.Flag("kiosk", true),
		chromedp.Flag("bwsi", true),
		chromedp.Flag("incognito", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-notifications", true),
		chromedp.Flag("disable-overlay-scrollbar", true),
		chromedp.Flag("window-position", cfg.General.WindowPosition),
		chromedp.Flag("check-for-update-interval", "31536000"),
		chromedp.Flag("ignore-certificate-errors", cfg.Target.IgnoreCertificateErrors),
		chromedp.Flag("test-type", cfg.Target.IgnoreCertificateErrors),
		chromedp.UserDataDir(dir),
	}
}
