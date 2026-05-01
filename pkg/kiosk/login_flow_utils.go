package kiosk

import (
	"context"
	"log"

	"github.com/grafana/grafana-kiosk/pkg/browser"
)

// runMessageLoop blocks until ctx is cancelled or a message triggers a reload
// navigation to dashboardURL. Used by all inner login flow functions.
func runMessageLoop(ctx context.Context, b browser.Browser, dashboardURL string, messages chan string) error {
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
