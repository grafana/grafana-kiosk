package kiosk

// Config configuration for backend.
type Config struct {
	General struct {
		AutoFit        bool   `yaml:"autofit" env:"KIOSK_AUTOFIT" env-default:"true" env-description:"fit panels to screen"`
		LXDEEnabled    bool   `yaml:"lxde" env:"KIOSK_LXDE_ENABLED" env-default:"false" env-description:"initialize LXDE for kiosk mode"`
		LXDEHome       string `yaml:"lxde-home" env:"KIOSK_LXDE_HOME" env-default:"/home/pi" env-description:"path to home directory of LXDE user running X Server"`
		Mode           string `yaml:"kiosk-mode" env:"KIOSK_MODE" env-default:"full" env-description:"[full|tv|disabled]"`
		WindowPosition string `yaml:"window-position" env:"KIOSK_WINDOW_POSITION" env-default:"0,0" env-description:"Top Left Position of Kiosk"`
	} `yaml:"general"`
	Target struct {
		IgnoreCertificateErrors bool   `yaml:"ignore-certificate-errors" env:"KIOSK_IGNORE_CERTIFICATE_ERRORS" env-description:"ignore SSL/TLS certificate errors" env-default:"false"`
		IsPlayList              bool   `yaml:"playlist" env:"KIOSK_IS_PLAYLIST" env-default:"false" env-description:"URL is a playlist"`
		UseMFA                  bool   `yaml:"use-mfa" env:"KIOSK_USE_MFA" env-default:"false" env-description:"MFA is enabled for given account"`
		LoginMethod             string `yaml:"login-method" env:"KIOSK_LOGIN_METHOD" env-default:"anon" env-description:"[anon|local|gcom|goauth|idtoken]"`
		Password                string `yaml:"password" env:"KIOSK_LOGIN_PASSWORD" env-default:"guest" env-description:"password"`
		URL                     string `yaml:"URL" env:"KIOSK_URL" env-default:"https://play.grafana.org" env-description:"URL to Grafana server"`
		Username                string `yaml:"username" env:"KIOSK_LOGIN_USER" env-default:"guest" env-description:"username"`
	} `yaml:"target"`
	GOAUTH struct {
		AutoLogin     bool   `yaml:"auto-login" env:"KIOSK_GOAUTH_AUTO_LOGIN" env-description:"[false|true]"`
		UsernameField string `yaml:"fieldname-username" env:"KIOSK_GOAUTH_FIELD_USER" env-description:"Username html input name value"`
		PasswordField string `yaml:"fieldname-password" env:"KIOSK_GOAUTH_FIELD_PASSWORD" env-description:"Password html input name value"`
	} `yaml:"goauth"`
	IDTOKEN struct {
		KeyFile  string `yaml:"idtoken-keyfile" env:"KIOSK_IDTOKEN_KEYFILE" env-default:"key.json" env-description:"JSON Credentials for idtoken"`
		Audience string `yaml:"idtoken-audience" env:"KIOSK_IDTOKEN_AUDIENCE" env-description:"Audience for idtoken, tpyically your oauth client id"`
	} `yaml:"idtoken"`
}
