package kiosk

import (
	"log"
	"net/url"

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
	if autoFit {
		parsedURI.RawQuery += "&autofitpanels"
	}

	return parsedURI.String()
}

func generateExecutorOptions(dir string, windowPosition string, ignoreCertificateErrors bool) []chromedp.ExecAllocatorOption {
	return []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("noerrdialogs", true),
		chromedp.Flag("kiosk", true),
		chromedp.Flag("bwsi", true),
		chromedp.Flag("incognito", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-notifications", true),
		chromedp.Flag("disable-overlay-scrollbar", true),
		chromedp.Flag("window-position", windowPosition),
		chromedp.Flag("check-for-update-interval", "31536000"),
		chromedp.Flag("ignore-certificate-errors", ignoreCertificateErrors),
		chromedp.Flag("test-type", ignoreCertificateErrors),
		chromedp.Flag("autoplay-policy", "no-user-gesture-required"),
		chromedp.UserDataDir(dir),
	}
}
