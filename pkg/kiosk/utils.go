package kiosk

import (
	"log"
	"net/url"
)

// GenerateURL constructs URL with appropriate parameters for kiosk mode
func GenerateURL(anURL string, kioskMode string, autoFit bool, isPlayList bool) string {
	u, _ := url.ParseRequestURI(anURL)
	q, _ := url.ParseQuery(u.RawQuery)

	switch kioskMode {
	case "tv": // TV
		q.Set("kiosk", "tv") // no sidebar, topnav without buttons
		log.Printf("KioskMode: TV")
	case "full": // FULLSCREEN
		q.Set("kiosk", "1") // sidebar and topnav always shown
		log.Printf("KioskMode: Fullscreen")
	case "disabled": // FULLSCREEN
		log.Printf("KioskMode: Disabled")
	default: // disabled
		q.Set("kiosk", "1") // sidebar and topnav always shown
		log.Printf("KioskMode: Fullscreen")
	}
	// a playlist should also go inactive immediately
	if isPlayList {
		q.Set("inactive", "1")
	}
	u.RawQuery = q.Encode()
	if autoFit {
		u.RawQuery += "&autofitpanels"
	}
	return u.String()
}
