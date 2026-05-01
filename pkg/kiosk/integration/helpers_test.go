//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/grafana/grafana-kiosk/pkg/kiosk/config"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/login/shared"
)

// startGrafana starts a Grafana container and returns its base URL and a
// cleanup function. The container has anonymous access and admin/admin credentials.
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

// browserPath returns the path to a Chrome/Chromium binary for headless testing.
// Override with KIOSK_BROWSER_PATH env var.
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

// baseCfg returns a minimal headless kiosk config.
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

// tempDir creates a throwaway Chromium user data directory.
func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "grafana-kiosk-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

// newHeadlessBrowserContext creates a chromedp task context using the same
// options as the kiosk (headless, same browser path).
func newHeadlessBrowserContext(t *testing.T, cfg *config.Config, dir string) (context.Context, context.CancelFunc) {
	t.Helper()
	taskCtx, cancel := shared.NewBrowserContext(context.Background(), cfg, dir, 0)
	return taskCtx, cancel
}
