package kiosk

import (
	"log"
	"net/url"
)

// GenerateURL constructs URL with appropriate parameters for kiosk mode
func GenerateURL(anURL string, kioskMode int, autoFit *bool, isPlayList *bool) string {
	u, _ := url.ParseRequestURI(anURL)
	q, _ := url.ParseQuery(u.RawQuery)
	switch kioskMode {
	case 0: // TV
		q.Set("kiosk", "tv") // no sidebar, topnav without buttons
		log.Printf("KioskMode: TV")
	case 1: // FULLSCREEN
		q.Set("kiosk", "1") // sidebar and topnav always shown
		log.Printf("KioskMode: Fullscreen")
	default: // disabled
		log.Printf("KioskMode: Disabled")
	}
	// a playlist should also go inactive immediately
	if *isPlayList == true {
		q.Set("inactive", "1")
	}
	u.RawQuery = q.Encode()
	if *autoFit == true {
		u.RawQuery = u.RawQuery + "&autofitpanels"
	}
	return u.String()
}
