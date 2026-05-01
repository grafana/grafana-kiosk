package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ilyakaznacheev/cleanenv"

	"github.com/grafana/grafana-kiosk/pkg/browser"
	"github.com/grafana/grafana-kiosk/pkg/initialize"
	"github.com/grafana/grafana-kiosk/pkg/kiosk"
)

var (
	// Version this is set during build time using git tags
	Version string
)

// sanitize removes newlines and carriage returns to prevent log injection
func sanitize(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

// Args command-line parameters.
type Args struct {
	AutoFit                              bool
	IgnoreCertificateErrors              bool
	IsPlayList                           bool
	OauthAutoLogin                       bool
	OauthWaitForPasswordField            bool
	OauthWaitForPasswordFieldIgnoreClass string
	OauthWaitForStaySignedInPrompt       bool
	LXDEEnabled                          bool
	UseMFA                               bool
	Audience                             string
	KeyFile                              string
	APIKey                               string
	LXDEHome                             string
	ConfigPath                           string
	Mode                                 string
	LoginMethod                          string
	URL                                  string
	Username                             string
	PageLoadDelayMS                      int64
	RestartDelayMS                       int64
	Password                             string
	UsernameField                        string
	PasswordField                        string
	WindowPosition                       string
	WindowSize                           string
	ScaleFactor                          string
	Browser                              string
	BrowserPath                          string
	HideLinks                            bool
	HideLogo                             bool
	HidePlaylistNav                      bool
	HideTimePicker                       bool
	HideVariables                        bool
	Incognito                            bool
}

// ProcessArgs processes and handles CLI arguments.
func ProcessArgs(cfg interface{}) (Args, *flag.FlagSet) {
	var processedArgs Args

	flagSettings := flag.NewFlagSet("grafana-kiosk", flag.ContinueOnError)
	flagSettings.StringVar(&processedArgs.ConfigPath, "c", "", "Path to configuration file (config.yaml)")
	flagSettings.StringVar(&processedArgs.LoginMethod, "login-method", "anon", "[anon|local|gcom|goauth|idtoken|apikey|aws|azuread]")
	flagSettings.StringVar(&processedArgs.Username, "username", "guest", "username")
	flagSettings.StringVar(&processedArgs.Password, "password", "guest", "password")
	flagSettings.BoolVar(&processedArgs.UseMFA, "use-mfa", false, "password")
	flagSettings.StringVar(&processedArgs.Mode, "kiosk-mode", "full", "Kiosk Display Mode [full|tv|disabled]\nfull = No TOPNAV and No SIDEBAR\ntv = No SIDEBAR\ndisabled = omit option\n")
	flagSettings.StringVar(&processedArgs.URL, "URL", "https://play.grafana.org", "URL to Grafana server")
	flagSettings.StringVar(&processedArgs.WindowPosition, "window-position", "0,0", "Top Left Position of Kiosk")
	flagSettings.StringVar(&processedArgs.WindowSize, "window-size", "", "Size of Kiosk in pixels (width,height)")
	flagSettings.StringVar(&processedArgs.ScaleFactor, "scale-factor", "1.0", "Scale factor, sort of zoom")
	flagSettings.StringVar(&processedArgs.Browser, "browser", "chrome", "Browser to launch [chrome|edge]")
	flagSettings.StringVar(&processedArgs.BrowserPath, "browser-path", "", "Explicit path to a Chromium-based browser executable; overrides -browser")
	flagSettings.Int64Var(&processedArgs.PageLoadDelayMS, "page-load-delay-ms", 2000, "Delay in milliseconds before navigating to URL")
	flagSettings.Int64Var(&processedArgs.RestartDelayMS, "restart-delay-ms", 5000, "Delay in milliseconds before restarting after a session error")
	flagSettings.BoolVar(&processedArgs.IsPlayList, "playlists", false, "URL is a playlist")
	flagSettings.BoolVar(&processedArgs.AutoFit, "autofit", true, "Fit panels to screen")
	flagSettings.BoolVar(&processedArgs.HideLinks, "hide-links", false, "Hide links in the top nav bar")
	flagSettings.BoolVar(&processedArgs.HideLogo, "hide-logo", false, "Hide Powered by Grafana logo")
	flagSettings.BoolVar(&processedArgs.HidePlaylistNav, "hide-playlist-nav", false, "Hide playlist navigation controls")
	flagSettings.BoolVar(&processedArgs.HideTimePicker, "hide-time-picker", false, "Hide time picker in the top nav bar")
	flagSettings.BoolVar(&processedArgs.HideVariables, "hide-variables", false, "Hide variables in the top nav bar")
	flagSettings.BoolVar(&processedArgs.LXDEEnabled, "lxde", false, "Initialize LXDE for kiosk mode")
	flagSettings.StringVar(&processedArgs.LXDEHome, "lxde-home", "/home/pi", "Path to home directory of LXDE user running X Server")
	flagSettings.BoolVar(&processedArgs.IgnoreCertificateErrors, "ignore-certificate-errors", false, "Ignore SSL/TLS certificate error")
	flagSettings.BoolVar(&processedArgs.Incognito, "incognito", true, "Use incognito mode")
	flagSettings.BoolVar(&processedArgs.OauthAutoLogin, "auto-login", false, "oauth_auto_login is enabled in grafana config")
	flagSettings.BoolVar(&processedArgs.OauthWaitForPasswordField, "wait-for-password-field", false, "oauth_auto_login is enabled in grafana config")
	flagSettings.StringVar(&processedArgs.OauthWaitForPasswordFieldIgnoreClass, "wait-for-password-field-class", "", "oauth_auto_login is enabled in grafana config")
	flagSettings.BoolVar(&processedArgs.OauthWaitForStaySignedInPrompt, "wait-for-stay-signed-in-prompt", false, "oauth_auto_login is enabled in grafana config")
	flagSettings.StringVar(&processedArgs.UsernameField, "field-username", "username", "Fieldname for the username")
	flagSettings.StringVar(&processedArgs.PasswordField, "field-password", "password", "Fieldname for the password")
	flagSettings.StringVar(&processedArgs.Audience, "audience", "", "idtoken audience")
	flagSettings.StringVar(&processedArgs.KeyFile, "keyfile", "key.json", "idtoken json credentials")
	flagSettings.StringVar(&processedArgs.APIKey, "apikey", "", "apikey")

	fu := flagSettings.Usage
	flagSettings.Usage = func() {
		fu()

		envHelp, _ := cleanenv.GetDescription(cfg, nil)

		_, _ = fmt.Fprintln(flagSettings.Output())
		_, _ = fmt.Fprintln(flagSettings.Output(), envHelp)
	}

	err := flagSettings.Parse(os.Args[1:])
	if err != nil {
		log.Printf("Failed to parse flags: %v — run with -help to see available options", err)
		os.Exit(-1)
	}

	return processedArgs, flagSettings
}

// loadConfig reads configuration from a file or environment, then applies
// CLI flag overrides. Flags always take precedence.
func loadConfig(args Args, fs *flag.FlagSet, cfg *kiosk.Config) error {
	if args.ConfigPath != "" {
		// read configuration from the file and then override with environment variables
		if err := cleanenv.ReadConfig(args.ConfigPath, cfg); err != nil {
			return fmt.Errorf("error reading config file: %w", err)
		}
		log.Println("Using config from", args.ConfigPath)
	} else {
		log.Println("No config specified, using environment and args")
		if err := cleanenv.ReadEnv(cfg); err != nil {
			log.Println("Error reading config from environment", err)
		}
	}

	// apply CLI flag overrides (flags take precedence over config file and environment)
	update := map[string]func(){
		"URL":                       func() { cfg.Target.URL = args.URL },
		"login-method":              func() { cfg.Target.LoginMethod = args.LoginMethod },
		"username":                  func() { cfg.Target.Username = args.Username },
		"password":                  func() { cfg.Target.Password = args.Password },
		"ignore-certificate-errors": func() { cfg.Target.IgnoreCertificateErrors = args.IgnoreCertificateErrors },
		"playlists":                 func() { cfg.Target.IsPlayList = args.IsPlayList },
		"use-mfa":                   func() { cfg.Target.UseMFA = args.UseMFA },
		//
		"autofit":            func() { cfg.General.AutoFit = args.AutoFit },
		"lxde":               func() { cfg.General.LXDEEnabled = args.LXDEEnabled },
		"lxde-home":          func() { cfg.General.LXDEHome = args.LXDEHome },
		"kiosk-mode":         func() { cfg.General.Mode = args.Mode },
		"window-position":    func() { cfg.General.WindowPosition = args.WindowPosition },
		"window-size":        func() { cfg.General.WindowSize = args.WindowSize },
		"scale-factor":       func() { cfg.General.ScaleFactor = args.ScaleFactor },
		"browser":            func() { cfg.General.Browser = args.Browser },
		"browser-path":       func() { cfg.General.BrowserPath = args.BrowserPath },
		"page-load-delay-ms":  func() { cfg.General.PageLoadDelayMS = args.PageLoadDelayMS },
		"restart-delay-ms":    func() { cfg.General.RestartDelayMS = args.RestartDelayMS },
		"hide-links":         func() { cfg.General.HideLinks = args.HideLinks },
		"hide-logo":          func() { cfg.General.HideLogo = args.HideLogo },
		"hide-playlist-nav":  func() { cfg.General.HidePlaylistNav = args.HidePlaylistNav },
		"hide-time-picker":   func() { cfg.General.HideTimePicker = args.HideTimePicker },
		"hide-variables":     func() { cfg.General.HideVariables = args.HideVariables },
		"incognito":          func() { cfg.General.Incognito = args.Incognito },
		//
		"auto-login":                     func() { cfg.GoAuth.AutoLogin = args.OauthAutoLogin },
		"field-username":                 func() { cfg.GoAuth.UsernameField = args.UsernameField },
		"field-password":                 func() { cfg.GoAuth.PasswordField = args.PasswordField },
		"wait-for-password-field":        func() { cfg.GoAuth.WaitForPasswordField = args.OauthWaitForPasswordField },
		"wait-for-password-field-class":  func() { cfg.GoAuth.WaitForPasswordFieldIgnoreClass = args.OauthWaitForPasswordFieldIgnoreClass },
		"wait-for-stay-signed-in-prompt": func() { cfg.GoAuth.WaitForStaySignedInPrompt = args.OauthWaitForStaySignedInPrompt },

		"audience": func() { cfg.IDToken.Audience = args.Audience },
		"keyfile":  func() { cfg.IDToken.KeyFile = args.KeyFile },

		"apikey": func() { cfg.APIKey.APIKey = args.APIKey },
	}

	fs.Visit(func(f *flag.Flag) {
		if do, ok := update[f.Name]; ok {
			do()
		}
	})

	return nil
}

func setEnvironment() {
	// for linux/X display must be set
	var displayEnv = os.Getenv("DISPLAY")
	if displayEnv == "" {
		log.Println("DISPLAY not set, autosetting to :0.0")
		if err := os.Setenv("DISPLAY", ":0.0"); err != nil {
			log.Println("Error setting DISPLAY", err.Error())
		}
		displayEnv = os.Getenv("DISPLAY")
	}

	log.Println("DISPLAY=", sanitize(displayEnv)) // #nosec G706 -- sanitized before logging

	var xAuthorityEnv = os.Getenv("XAUTHORITY")
	if xAuthorityEnv == "" {
		log.Println("XAUTHORITY not set, autosetting")
		// use HOME of current user
		var homeEnv = os.Getenv("HOME")

		if err := os.Setenv("XAUTHORITY", homeEnv+"/.Xauthority"); err != nil {
			log.Println("Error setting XAUTHORITY", sanitize(err.Error())) // #nosec G706 -- sanitized before logging
		}
		xAuthorityEnv = os.Getenv("XAUTHORITY")
	}

	log.Println("XAUTHORITY=", sanitize(xAuthorityEnv)) // #nosec G706 -- sanitized before logging
}

func summary(cfg *kiosk.Config) {
	log.Println("*************************************************************")
	logGeneralSettings(cfg)
	logTargetSettings(cfg)
	logGoAuthSettings(cfg)
	log.Println("*************************************************************")
}

func logGeneralSettings(cfg *kiosk.Config) {
	log.Println("--- General -------------------------------------------------")
	log.Println("AutoFit:", cfg.General.AutoFit)
	log.Println("LXDEEnabled:", cfg.General.LXDEEnabled)
	log.Println("LXDEHome:", cfg.General.LXDEHome)
	log.Println("Mode:", cfg.General.Mode)
	log.Println("Incognito:", cfg.General.Incognito)
	log.Println("WindowPosition:", cfg.General.WindowPosition)
	log.Println("WindowSize:", cfg.General.WindowSize)
	log.Println("ScaleFactor:", cfg.General.ScaleFactor)
	log.Println("Browser:", cfg.General.Browser)
	log.Println("BrowserPath:", cfg.General.BrowserPath)
	log.Println("PageLoadDelayMS:", cfg.General.PageLoadDelayMS)
	log.Println("RestartDelayMS:", cfg.General.RestartDelayMS)
	log.Println("HideLinks:", cfg.General.HideLinks)
	log.Println("HideLogo:", cfg.General.HideLogo)
	log.Println("HidePlaylistNav:", cfg.General.HidePlaylistNav)
	log.Println("HideTimePicker:", cfg.General.HideTimePicker)
	log.Println("HideVariables:", cfg.General.HideVariables)
}

func logTargetSettings(cfg *kiosk.Config) {
	log.Println("--- Target --------------------------------------------------")
	log.Println("URL:", cfg.Target.URL)
	log.Println("LoginMethod:", cfg.Target.LoginMethod)
	log.Println("Username:", cfg.Target.Username)
	log.Println("Password:", "*redacted*")
	log.Println("IgnoreCertificateErrors:", cfg.Target.IgnoreCertificateErrors)
	log.Println("IsPlayList:", cfg.Target.IsPlayList)
	log.Println("UseMFA:", cfg.Target.UseMFA)
}

func logGoAuthSettings(cfg *kiosk.Config) {
	log.Println("--- GoAuth --------------------------------------------------")
	log.Println("Fieldname AutoLogin:", cfg.GoAuth.AutoLogin)
	log.Println("Fieldname Username:", cfg.GoAuth.UsernameField)
	log.Println("Fieldname Password:", cfg.GoAuth.PasswordField)
}

func main() {
	var cfg kiosk.Config
	fmt.Println("GrafanaKiosk Version:", Version)
	// set the version
	cfg.BuildInfo.Version = Version

	// override
	args, fs := ProcessArgs(&cfg)

	// validate auth methods
	switch args.LoginMethod {
	case "goauth", "anon", "local", "gcom", "idtoken", "apikey", "aws", "azuread":
	default:
		log.Printf("Invalid login method %q — supported values: anon, local, gcom, goauth, idtoken, apikey, aws, azuread", args.LoginMethod)
		os.Exit(-1)
	}

	if err := loadConfig(args, fs, &cfg); err != nil {
		log.Printf("Failed to load configuration: %v — check your config file and environment variables", err)
		os.Exit(-1)
	}

	// validate browser selection
	switch strings.ToLower(cfg.General.Browser) {
	case "", "chrome", "edge":
	default:
		log.Printf("Invalid browser %q — supported values: chrome, edge (or use -browser-path for a custom binary)", cfg.General.Browser)
		os.Exit(-1)
	}

	// make sure the url has content
	if cfg.Target.URL == "" {
		log.Println("No target URL specified — set -URL or KIOSK_URL")
		os.Exit(-1)
	}
	// validate url
	_, err := url.ParseRequestURI(cfg.Target.URL)
	if err != nil {
		log.Printf("Invalid target URL %q: %v — set a valid URL with -URL or KIOSK_URL", cfg.Target.URL, err)
		os.Exit(-1)
	}

	summary(&cfg)

	if cfg.General.LXDEEnabled {
		initialize.LXDE(cfg.General.LXDEHome)
	}

	// for linux/X display must be set
	setEnvironment()
	log.Println("method ", cfg.Target.LoginMethod)

	dir, err := os.MkdirTemp(os.TempDir(), "chromedp-kiosk")
	if err != nil {
		log.Fatal("Error creating temp dir:", err)
	}
	log.Println("Using temp dir:", dir)
	defer func() {
		log.Println("Cleaning up temp dir:", dir)
		if err := os.RemoveAll(dir); err != nil {
			log.Printf("Error cleaning temporary directory: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		<-sigs
		log.Println("Received signal, shutting down...")
		cancel()
	}()

	messages := make(chan string)
	restartDelay := time.Duration(cfg.General.RestartDelayMS) * time.Millisecond

	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Kiosk session error: %v", r)
				}
			}()
			switch cfg.Target.LoginMethod {
			case "local":
				log.Printf("Launching local login kiosk")
				kiosk.GrafanaKioskLocal(ctx, &cfg, dir, &browser.ChromeDP{}, messages)
			case "gcom":
				log.Printf("Launching GCOM login kiosk")
				kiosk.GrafanaKioskGCOM(ctx, &cfg, dir, &browser.ChromeDP{}, messages)
			case "goauth":
				log.Printf("Launching Generic Oauth login kiosk")
				kiosk.GrafanaKioskGenericOauth(ctx, &cfg, dir, &browser.ChromeDP{}, messages)
			case "idtoken":
				log.Printf("Launching idtoken oauth kiosk")
				kiosk.GrafanaKioskIDToken(ctx, &cfg, dir, &browser.ChromeDP{}, messages)
			case "apikey":
				log.Printf("Launching apikey kiosk")
				kiosk.GrafanaKioskAPIKey(ctx, &cfg, dir, &browser.ChromeDP{}, messages)
			case "aws":
				log.Printf("Launching AWS SSO kiosk")
				kiosk.GrafanaKioskAWSLogin(ctx, &cfg, dir, &browser.ChromeDP{}, messages)
			case "azuread":
				log.Printf("Launching AzureAD login kiosk")
				kiosk.GrafanaKioskAzureAD(ctx, &cfg, dir, &browser.ChromeDP{}, messages)
			default:
				log.Printf("Launching ANON login kiosk")
				kiosk.GrafanaKioskAnonymous(ctx, &cfg, dir, &browser.ChromeDP{}, messages)
			}
		}()

		// Exit on clean shutdown; otherwise wait before restarting.
		select {
		case <-ctx.Done():
			return
		default:
			log.Printf("Kiosk session ended — restarting in %.0f seconds", restartDelay.Seconds())
			select {
			case <-time.After(restartDelay):
			case <-ctx.Done():
				return
			}
		}
	}
}
