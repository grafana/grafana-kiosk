package kiosk

import "github.com/grafana/grafana-kiosk/pkg/kiosk/config"

// Type aliases — keep existing code compiling during migration.
// Remove this file once all consumers import pkg/kiosk/config directly.
type BuildInfo = config.BuildInfo
type General = config.General
type Target = config.Target
type GoAuth = config.GoAuth
type IDToken = config.IDToken
type APIKey = config.APIKey
type Config = config.Config
