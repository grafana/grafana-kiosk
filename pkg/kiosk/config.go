package kiosk

// BuildInfo contains the build version
type BuildInfo struct {
	Version string `yaml:"version,omitempty"`
}

// General non-site specific configurations
type General struct {
	LXDEEnabled     bool   `yaml:"lxde" env:"KIOSK_GENERAL_LXDE_ENABLED" env-default:"false" env-description:"initialize LXDE for kiosk mode"`
	LXDEHome        string `yaml:"lxde-home" env:"KIOSK_GENERAL_LXDE_HOME" env-default:"/home/pi" env-description:"path to home directory of LXDE user running X Server"`
	PageLoadDelayMS int64  `yaml:"page-load-delay-ms" env:"KIOSK_GENERAL_PAGE_LOAD_DELAY_MS" env-default:"2000" env-description:"milliseconds to wait before expecting page load"`
}

// GrafanaOptions grafana specific flags
type GrafanaOptions struct {
	AutoFit   bool   `yaml:"autofit" env:"KIOSK_GRAFANA_AUTOFIT" env-default:"true" env-description:"fit panels to screen"`
	KioskMode string `yaml:"kiosk-mode" env:"KIOSK_GRAFANA_MODE" env-default:"full" env-description:"[full|tv|disabled]"`
}

// ChromeDPFlags flags specific to chrome
type ChromeDPFlags struct {
	DebugEnabled     bool   `yaml:"debug" env:"KIOSK_CHROMEDP_DEBUG" env-default:"false" env-description:"enables debug output"`
	GPUEnabled       bool   `yaml:"gpu-enabled" env:"KIOSK_CHROMEDP_GPU_ENABLED" env-default:"true" env-description:"Enable/Disable GPU support"`
	IncognitoEnabled bool   `yaml:"incognito-enabled" env:"KIOSK_CHROMEDP_INCOGNITO_ENABLED" env-default:"true" env-description:"Enable/Disable Incognito Mode"`
	Kiosk            bool   `yaml:"kiosk,omitempty" env:"KIOSK_CHROMEDP_KIOSK" env-default:"true" env-description:"pass kiosk flag to chromedp"`
	OzonePlatform    string `yaml:"ozone-platform" env:"KIOSK_CHROMEDP_OZONE_PLATFORM" env-default:"" env-description:"Set ozone-platform option (wayland|cast|drm|wayland|x11)"`
	ScaleFactor      string `yaml:"scale-factor" env:"KIOSK_CHROMEDP_SCALE_FACTOR" env-default:"1.0" env-description:"Scale factor, like zoom"`
	StartFullscreen  bool   `yaml:"start-fullscreen,omitempty" env:"KIOSK_CHROMEDP_START_FULLSCREEN" env-default:"true" env-description:"Scale factor, like zoom"`
	StartMaximized   bool   `yaml:"start-maximized,omitempty" env:"KIOSK_CHROMEDP_START_MAXIMIZED" env-default:"true" env-description:"Scale factor, like zoom"`
	WindowPosition   string `yaml:"window-position" env:"KIOSK_CHROMEDP_WINDOW_POSITION" env-default:"0,0" env-description:"Top Left Position of Kiosk"`
	WindowSize       string `yaml:"window-size" env:"KIOSK_CHROMEDP_WINDOW_SIZE" env-default:"" env-description:"Size of Kiosk in pixels (width,height)"`
}

// Target the dashboard/playlist details
type Target struct {
	IgnoreCertificateErrors bool   `yaml:"ignore-certificate-errors" env:"KIOSK_TARGET_IGNORE_CERTIFICATE_ERRORS" env-description:"ignore SSL/TLS certificate errors" env-default:"false"`
	IsPlayList              bool   `yaml:"playlist" env:"KIOSK_TARGET_IS_PLAYLIST" env-default:"false" env-description:"URL is a playlist"`
	LoginMethod             string `yaml:"login-method" env:"KIOSK_TARGET_LOGIN_METHOD" env-default:"anon" env-description:"[anon|local|gcom|goauth|idtoken|apikey]"`
	Password                string `yaml:"password" env:"KIOSK_TARGET_LOGIN_PASSWORD" env-default:"guest" env-description:"password"`
	URL                     string `yaml:"URL" env:"KIOSK_TARGET_URL" env-default:"https://play.grafana.org" env-description:"URL to Grafana server"`
	Username                string `yaml:"username" env:"KIOSK_TARGET_LOGIN_USER" env-default:"guest" env-description:"username"`
	UseMFA                  bool   `yaml:"use-mfa" env:"KIOSK_TARGET_USE_MFA" env-default:"false" env-description:"MFA is enabled for given account"`
}

// GoAuth Generic OAuth
type GoAuth struct {
	AutoLogin     bool   `yaml:"auto-login" env:"KIOSK_GOAUTH_AUTO_LOGIN" env-description:"[false|true]"`
	UsernameField string `yaml:"fieldname-username" env:"KIOSK_GOAUTH_FIELD_USER" env-description:"Username html input name value"`
	PasswordField string `yaml:"fieldname-password" env:"KIOSK_GOAUTH_FIELD_PASSWORD" env-description:"Password html input name value"`
}

// IDToken token based login
type IDToken struct {
	KeyFile  string `yaml:"idtoken-keyfile" env:"KIOSK_IDTOKEN_KEYFILE" env-default:"key.json" env-description:"JSON Credentials for idtoken"`
	Audience string `yaml:"idtoken-audience" env:"KIOSK_IDTOKEN_AUDIENCE" env-description:"Audience for idtoken, tpyically your oauth client id"`
}

// Bearer Bearer parameter for login
type Bearer struct {
	APIKey              string `yaml:"api-key,omitempty" env:"KIOSK_BEARER_APIKEY" env-description:"Legacy API Key"`
	ServiceAccountToken string `yaml:"service-account-token,omitempty" env:"KIOSK_BEARER_SERVICE_ACCOUNT_TOKEN" env-description:"Service Account Token"`
}

// Config configuration for backend.
type Config struct {
	BuildInfo      BuildInfo      `yaml:"buildinfo"`
	General        General        `yaml:"general"`
	ChromeDPFlags  ChromeDPFlags  `yaml:"chromedp-options,omitempty"`
	GrafanaOptions GrafanaOptions `yaml:"grafana-options,omitempty"`
	Target         Target         `yaml:"target,omitempty"`
	GoAuth         GoAuth         `yaml:"goauth,omitempty"`
	IDToken        IDToken        `yaml:"idtoken,omitempty"`
	Bearer         Bearer         `yaml:"bearer,omitempty"`
}
