package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/ilyakaznacheev/cleanenv"

	"github.com/grafana/grafana-kiosk/pkg/initialize"
	"github.com/grafana/grafana-kiosk/pkg/kiosk"
)

var (
	// Version this is set during build time using git tags
	Version string
)

// Args command-line parameters.
type Args struct {
	AutoFit                 bool
	IgnoreCertificateErrors bool
	IsPlayList              bool
	OauthAutoLogin          bool
	LXDEEnabled             bool
	UseMFA                  bool
	Audience                string
	KeyFile                 string
	APIKey                  string
	LXDEHome                string
	ConfigPath              string
	Mode                    string
	LoginMethod             string
	URL                     string
	Username                string
	Password                string
	UsernameField           string
	PasswordField           string
	WindowPosition          string
	WindowSize              string
	ScaleFactor             string
}

// ProcessArgs processes and handles CLI arguments.
func ProcessArgs(cfg interface{}) Args {
	var processedArgs Args

	flagSettings := flag.NewFlagSet("grafana-kiosk", flag.ContinueOnError)
	flagSettings.StringVar(&processedArgs.ConfigPath, "c", "", "Path to configuration file (config.yaml)")
	flagSettings.StringVar(&processedArgs.LoginMethod, "login-method", "anon", "[anon|local|gcom|goauth|idtoken|apikey|aws]")
	flagSettings.StringVar(&processedArgs.Username, "username", "guest", "username")
	flagSettings.StringVar(&processedArgs.Password, "password", "guest", "password")
	flagSettings.BoolVar(&processedArgs.UseMFA, "use-mfa", false, "password")
	flagSettings.StringVar(&processedArgs.Mode, "kiosk-mode", "full", "Kiosk Display Mode [full|tv|disabled]\nfull = No TOPNAV and No SIDEBAR\ntv = No SIDEBAR\ndisabled = omit option\n")
	flagSettings.StringVar(&processedArgs.URL, "URL", "https://play.grafana.org", "URL to Grafana server")
	flagSettings.StringVar(&processedArgs.WindowPosition, "window-position", "0,0", "Top Left Position of Kiosk")
	flagSettings.StringVar(&processedArgs.WindowSize, "window-size", "", "Size of Kiosk in pixels (width,height)")
	flagSettings.StringVar(&processedArgs.ScaleFactor, "scale-factor", "1.0", "Scale factor, sort of zoom")
	flagSettings.BoolVar(&processedArgs.IsPlayList, "playlists", false, "URL is a playlist")
	flagSettings.BoolVar(&processedArgs.AutoFit, "autofit", true, "Fit panels to screen")
	flagSettings.BoolVar(&processedArgs.LXDEEnabled, "lxde", false, "Initialize LXDE for kiosk mode")
	flagSettings.StringVar(&processedArgs.LXDEHome, "lxde-home", "/home/pi", "Path to home directory of LXDE user running X Server")
	flagSettings.BoolVar(&processedArgs.IgnoreCertificateErrors, "ignore-certificate-errors", false, "Ignore SSL/TLS certificate error")
	flagSettings.BoolVar(&processedArgs.OauthAutoLogin, "auto-login", false, "oauth_auto_login is enabled in grafana config")
	flagSettings.StringVar(&processedArgs.UsernameField, "field-username", "username", "Fieldname for the username")
	flagSettings.StringVar(&processedArgs.PasswordField, "field-password", "password", "Fieldname for the password")
	flagSettings.StringVar(&processedArgs.Audience, "audience", "", "idtoken audience")
	flagSettings.StringVar(&processedArgs.KeyFile, "keyfile", "key.json", "idtoken json credentials")
	flagSettings.StringVar(&processedArgs.APIKey, "apikey", "", "apikey")

	fu := flagSettings.Usage
	flagSettings.Usage = func() {
		fu()

		envHelp, _ := cleanenv.GetDescription(cfg, nil)

		fmt.Fprintln(flagSettings.Output())
		fmt.Fprintln(flagSettings.Output(), envHelp)
	}

	err := flagSettings.Parse(os.Args[1:])
	if err != nil {
		os.Exit(-1)
	}

	return processedArgs
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

	log.Println("DISPLAY=", displayEnv)

	var xAuthorityEnv = os.Getenv("XAUTHORITY")
	if xAuthorityEnv == "" {
		log.Println("XAUTHORITY not set, autosetting")
		// use HOME of current user
		var homeEnv = os.Getenv("HOME")

		if err := os.Setenv("XAUTHORITY", homeEnv+"/.Xauthority"); err != nil {
			log.Println("Error setting XAUTHORITY", err.Error())
		}
		xAuthorityEnv = os.Getenv("XAUTHORITY")
	}

	log.Println("XAUTHORITY=", xAuthorityEnv)
}

func summary(cfg *kiosk.Config) {
	// general
	log.Println("AutoFit:", cfg.General.AutoFit)
	log.Println("LXDEEnabled:", cfg.General.LXDEEnabled)
	log.Println("LXDEHome:", cfg.General.LXDEHome)
	log.Println("Mode:", cfg.General.Mode)
	log.Println("WindowPosition:", cfg.General.WindowPosition)
	log.Println("WindowSize:", cfg.General.WindowSize)
	log.Println("ScaleFactor:", cfg.General.ScaleFactor)
	// target
	log.Println("URL:", cfg.Target.URL)
	log.Println("LoginMethod:", cfg.Target.LoginMethod)
	log.Println("Username:", cfg.Target.Username)
	log.Println("Password:", "*redacted*")
	log.Println("IgnoreCertificateErrors:", cfg.Target.IgnoreCertificateErrors)
	log.Println("IsPlayList:", cfg.Target.IsPlayList)
	log.Println("UseMFA:", cfg.Target.UseMFA)
	// goauth
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
	args := ProcessArgs(&cfg)

	// validate auth methods
	switch args.LoginMethod {
	case "goauth", "anon", "local", "gcom", "idtoken", "apikey", "aws":
	default:
		log.Println("Invalid auth method", args.LoginMethod)
		os.Exit(-1)
	}

	// check if config specified
	if args.ConfigPath != "" {
		// read configuration from the file and then override with environment variables
		if err := cleanenv.ReadConfig(args.ConfigPath, &cfg); err != nil {
			log.Println("Error reading config file", err)
			os.Exit(-1)
		} else {
			log.Println("Using config from", args.ConfigPath)
		}
	} else {
		log.Println("No config specified, using environment and args")
		// no config, use environment and args
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			log.Println("Error reading config from environment", err)
		}
		cfg.Target.URL = args.URL
		cfg.Target.LoginMethod = args.LoginMethod
		cfg.Target.Username = args.Username
		cfg.Target.Password = args.Password
		cfg.Target.IgnoreCertificateErrors = args.IgnoreCertificateErrors
		cfg.Target.IsPlayList = args.IsPlayList
		cfg.Target.UseMFA = args.UseMFA
		//
		cfg.General.AutoFit = args.AutoFit
		cfg.General.LXDEEnabled = args.LXDEEnabled
		cfg.General.LXDEHome = args.LXDEHome
		cfg.General.Mode = args.Mode
		cfg.General.WindowPosition = args.WindowPosition
		cfg.General.WindowSize = args.WindowSize
		//
		cfg.GoAuth.AutoLogin = args.OauthAutoLogin
		cfg.GoAuth.UsernameField = args.UsernameField
		cfg.GoAuth.PasswordField = args.PasswordField

		cfg.IDToken.Audience = args.Audience
		cfg.IDToken.KeyFile = args.KeyFile

		cfg.APIKey.APIKey = args.APIKey
	}

	// make sure the url has content
	if cfg.Target.URL == "" {
		os.Exit(1)
	}
	// validate url
	_, err := url.ParseRequestURI(cfg.Target.URL)
	if err != nil {
		panic(err)
	}

	summary(&cfg)

	if cfg.General.LXDEEnabled {
		initialize.LXDE(cfg.General.LXDEHome)
	}

	// for linux/X display must be set
	setEnvironment()
	log.Println("method ", cfg.Target.LoginMethod)

	messages := make(chan string)
	switch cfg.Target.LoginMethod {
	case "local":
		log.Printf("Launching local login kiosk")
		kiosk.GrafanaKioskLocal(&cfg, messages)
	case "gcom":
		log.Printf("Launching GCOM login kiosk")
		kiosk.GrafanaKioskGCOM(&cfg, messages)
	case "goauth":
		log.Printf("Launching Generic Oauth login kiosk")
		kiosk.GrafanaKioskGenericOauth(&cfg, messages)
	case "idtoken":
		log.Printf("Launching idtoken oauth kiosk")
		kiosk.GrafanaKioskIDToken(&cfg, messages)
	case "apikey":
		log.Printf("Launching apikey kiosk")
		kiosk.GrafanaKioskAPIKey(&cfg, messages)
	case "aws":
		log.Printf("Launcing AWS SSO kiosk")
		kiosk.GrafanaKioskAWSLogin(&cfg, messages)
	default:
		log.Printf("Launching ANON login kiosk")
		kiosk.GrafanaKioskAnonymous(&cfg, messages)
	}
}
