package kiosk

import (
	"log"
	"net/url"
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
