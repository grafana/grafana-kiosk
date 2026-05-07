package kiosk

import (
	"fmt"
	"strings"

	"github.com/grafana/grafana-kiosk/pkg/kiosk/config"
	"github.com/grafana/grafana-kiosk/pkg/kiosk/login/shared"
)

// ValidateBrowserConfig returns an error if the browser configuration cannot
// be satisfied: specifically when "edge" is requested but no Edge binary is
// found on PATH, in standard Windows install locations, and no explicit
// BrowserPath is set.
func ValidateBrowserConfig(cfg *config.Config) error {
	if strings.ToLower(cfg.General.Browser) != "edge" || cfg.General.BrowserPath != "" {
		return nil
	}
	if shared.ResolveBrowserExecPath(cfg) != "" {
		return nil
	}
	return fmt.Errorf("browser 'edge' requested but no Edge binary found on PATH (tried: %v) — install Edge or use -browser-path to specify the executable",
		shared.EdgeBinaryCandidates)
}
