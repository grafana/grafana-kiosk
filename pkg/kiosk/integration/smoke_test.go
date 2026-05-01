//go:build integration

// Smoke tests verify that kiosk sessions start and complete without panicking.
// They do not assert on browser state — use the functional tests for that.
package integration

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/grafana/grafana-kiosk/pkg/browser"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/login/anonymous"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/login/local"
)

func TestSmokeAnonymousLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires Docker")
	}

	grafanaURL, cleanup := startGrafana(t)
	defer cleanup()

	Convey("Smoke: anonymous kiosk session does not panic", t, func() {
		cfg := baseCfg(grafanaURL)
		dir := tempDir(t)
		messages := make(chan string)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		done := make(chan struct{})
		go func() {
			defer close(done)
			defer func() { recover() }()
			anonymous.Run(ctx, cfg, dir, &browser.ChromeDP{}, messages)
		}()

		time.Sleep(5 * time.Second)
		cancel()
		<-done

		So(true, ShouldBeTrue)
	})
}

func TestSmokeLocalLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires Docker")
	}

	grafanaURL, cleanup := startGrafana(t)
	defer cleanup()

	Convey("Smoke: local login kiosk session does not panic", t, func() {
		cfg := baseCfg(grafanaURL)
		cfg.Target.Username = "admin"
		cfg.Target.Password = "admin"
		cfg.Target.LoginMethod = "local"
		dir := tempDir(t)
		messages := make(chan string)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		done := make(chan struct{})
		go func() {
			defer close(done)
			defer func() { recover() }()
			local.Run(ctx, cfg, dir, &browser.ChromeDP{}, messages)
		}()

		time.Sleep(10 * time.Second)
		cancel()
		<-done

		So(true, ShouldBeTrue)
	})
}
