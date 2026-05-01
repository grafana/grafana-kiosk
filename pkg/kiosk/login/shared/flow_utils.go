package shared

import (
	"context"
	"log"
	"time"

	"github.com/grafana/grafana-kiosk/pkg/browser"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/config"
)

// SleepPageLoad waits cfg.General.PageLoadDelayMS milliseconds if the delay is
// positive. Used after navigation to allow the page to finish loading.
func SleepPageLoad(cfg *config.Config) {
	if cfg.General.PageLoadDelayMS > 0 {
		log.Printf("Sleeping %d MS for page load", cfg.General.PageLoadDelayMS)
		time.Sleep(time.Duration(cfg.General.PageLoadDelayMS) * time.Millisecond)
	}
}

// RunMessageLoop blocks until ctx is cancelled or a message triggers a reload
// navigation to dashboardURL. Used by all inner login flow functions.
func RunMessageLoop(ctx context.Context, b browser.Browser, dashboardURL string, messages chan string) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case messageFromBrowser := <-messages:
			if err := b.Navigate(ctx, dashboardURL); err != nil {
				return err
			}
			log.Println("Browser output:", messageFromBrowser)
		}
	}
}
