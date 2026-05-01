package kiosk

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/grafana/grafana-kiosk/pkg/kiosk/config"
)

// edgeBinaryCandidates lists executable names to look up on PATH when the
// user requests the Edge browser. Order matters: first match wins.
var edgeBinaryCandidates = []string{
	"msedge",
	"microsoft-edge",
	"microsoft-edge-stable",
}

// lookPath is overridable in tests.
var lookPath = exec.LookPath

// resolveBrowserExecPath returns the explicit browser executable path that
// should be passed to chromedp.ExecPath. An empty string means "let chromedp
// auto-detect" (the default Chrome lookup).
//
// Precedence:
//  1. cfg.General.BrowserPath (verbatim) if set
//  2. cfg.General.Browser == "edge" → PATH lookup
//  3. cfg.General.Browser == "chrome" or empty → "" (chromedp default)
func resolveBrowserExecPath(cfg *config.Config) string {
	if cfg.General.BrowserPath != "" {
		return cfg.General.BrowserPath
	}
	switch strings.ToLower(cfg.General.Browser) {
	case "", "chrome":
		return ""
	case "edge":
		for _, name := range edgeBinaryCandidates {
			if p, err := lookPath(name); err == nil {
				return p
			}
		}
		log.Println("Browser 'edge' requested but no Edge binary found on PATH; set -browser-path or KIOSK_BROWSER_PATH to the Edge executable")
		return ""
	}
	return ""
}

// ValidateBrowserConfig returns an error if the browser configuration cannot
// be satisfied: specifically when "edge" is requested but no Edge binary is
// found on PATH and no explicit BrowserPath is set.
func ValidateBrowserConfig(cfg *config.Config) error {
	if cfg.General.BrowserPath != "" {
		return nil
	}
	if strings.ToLower(cfg.General.Browser) != "edge" {
		return nil
	}
	for _, name := range edgeBinaryCandidates {
		if _, err := lookPath(name); err == nil {
			return nil
		}
	}
	return fmt.Errorf("browser 'edge' requested but no Edge binary found on PATH (tried: %v) — install Edge or use -browser-path to specify the executable",
		edgeBinaryCandidates)
}
