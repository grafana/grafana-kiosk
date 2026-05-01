//go:build integration

// Package integration contains end-to-end tests that start a real Grafana
// instance via testcontainers and run grafana-kiosk against it in headless
// mode.
//
// Run with: CGO_ENABLED=0 go test -tags integration -v ./pkg/kiosk/integration/...
//
// Requires Docker to be running.
package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/grafana/grafana-kiosk/pkg/browser"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/config"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/login/anonymous"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/login/local"
)

// startGrafana starts a Grafana container and returns its base URL and a
// cleanup function. The container uses the official Grafana image with the
// default admin credentials (admin/admin).
func startGrafana(t *testing.T) (baseURL string, cleanup func()) {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "grafana/grafana:latest",
		ExposedPorts: []string{"3000/tcp"},
		Env: map[string]string{
			"GF_SECURITY_ADMIN_PASSWORD": "admin",
			"GF_AUTH_ANONYMOUS_ENABLED":  "true",
			"GF_AUTH_ANONYMOUS_ORG_ROLE": "Viewer",
		},
		WaitingFor: wait.ForHTTP("/api/health").
			WithPort("3000").
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("start grafana container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("get container host: %v", err)
	}
	port, err := container.MappedPort(ctx, "3000")
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("get container port: %v", err)
	}

	url := fmt.Sprintf("http://%s:%s", host, port.Port())
	return url, func() { _ = container.Terminate(ctx) }
}

// browserPath returns the path to a suitable Chromium-based browser for
// headless testing. Checks common locations and the KIOSK_BROWSER_PATH env var.
func browserPath() string {
	if p := os.Getenv("KIOSK_BROWSER_PATH"); p != "" {
		return p
	}
	candidates := []string{
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
		"/usr/bin/google-chrome",
		"/usr/bin/chromium",
		"/usr/bin/chromium-browser",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

// baseCfg returns a minimal config with headless mode enabled.
func baseCfg(grafanaURL string) *config.Config {
	return &config.Config{
		General: config.General{
			Mode:            "full",
			AutoFit:         true,
			Incognito:       true,
			PageLoadDelayMS: 2000,
			Headless:        true,
			BrowserPath:     browserPath(),
		},
		Target: config.Target{
			URL: grafanaURL,
		},
	}
}

// tempDir creates a throwaway user data dir for Chromium.
func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "grafana-kiosk-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestAnonymousLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires Docker")
	}

	grafanaURL, cleanup := startGrafana(t)
	defer cleanup()

	Convey("Given a headless anonymous kiosk session", t, func() {
		cfg := baseCfg(grafanaURL)
		dir := tempDir(t)
		messages := make(chan string)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		done := make(chan struct{})
		go func() {
			defer close(done)
			defer func() { recover() }() // context cancel may cause panic on clean shutdown
			anonymous.Run(ctx, cfg, dir, &browser.ChromeDP{}, messages)
		}()

		// Give the kiosk time to navigate then cancel cleanly
		time.Sleep(5 * time.Second)
		cancel()
		<-done

		Convey("Kiosk session completed without panic", func() {
			So(true, ShouldBeTrue)
		})
	})
}

func TestLocalLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires Docker")
	}

	grafanaURL, cleanup := startGrafana(t)
	defer cleanup()

	Convey("Given a headless local login kiosk session", t, func() {
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
			defer func() { recover() }() // context cancel may cause panic on clean shutdown
			local.Run(ctx, cfg, dir, &browser.ChromeDP{}, messages)
		}()

		time.Sleep(10 * time.Second)
		cancel()
		<-done

		Convey("Kiosk session completed without panic", func() {
			So(true, ShouldBeTrue)
		})
	})
}
