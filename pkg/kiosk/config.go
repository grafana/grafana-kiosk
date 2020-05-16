package kiosk

// Config configuration for backend
type Config struct {
	General struct {
		AutoFit     bool   `yaml:"autofit" env:"KIOSK_AUTOFIT" env-default:"true" env-description:"fit panels to screen"`
		LXDEEnabled bool   `yaml:"lxde" env:"KIOSK_LXDE_ENABLED" env-default:"false" env-description:"initialize LXDE for kiosk mode"`
		LXDEHome    string `yaml:"lxde-home" env:"KIOSK_MODE" env-default:"/home/pi" env-description:"path to home directory of LXDE user running X Server"`
		Mode        string `yaml:"mode" env:"KIOSK_MODE" env-default:"full" env-description:"[full|tv|disabled]"`
	} `yaml:"general"`
	Target struct {
		IgnoreCertificateErrors bool   `yaml:"ignore-certificate-errors" env:"KIOSK_IGNORE_CERTIFICATE_ERRORS" env-description:"ignore SSL/TLS certificate errors" env-default:"false"`
		IsPlayList              bool   `yaml:"playlist" env:"KIOSK_IS_PLAYLIST" env-default:"false" env-description:"URL is a playlist"`
		LoginMethod             string `yaml:"login-method" env:"KIOSK_LOGIN_METHOD" env-default:"anon" env-description:"[anon|local|gcom]"`
		Password                string `yaml:"password" env:"KIOSK_LOGIN_PASSWORD" env-default:"guest" env-description:"password"`
		URL                     string `yaml:"url" env:"KIOSK_URL" env-default:"https://play.grafana.org" env-description:"URL to Grafana server"`
		Username                string `yaml:"username" env:"KIOSK_LOGIN_USER" env-default:"guest" env-description:"username"`
	} `yaml:"target"`
}
