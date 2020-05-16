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

// Args command-line parameters
type Args struct {
	AutoFit                 bool
	IgnoreCertificateErrors bool
	IsPlayList              bool
	LXDEEnabled             bool
	LXDEHome                string
	ConfigPath              string
	Mode                    string
	LoginMethod             string
	URL                     string
	Username                string
	Password                string
}

// ProcessArgs processes and handles CLI arguments
func ProcessArgs(cfg interface{}) Args {
	var a Args

	f := flag.NewFlagSet("grafana-kiosk", flag.ContinueOnError)
	f.StringVar(&a.ConfigPath, "c", "config.yml", "Path to configuration file")
	f.BoolVar(&a.AutoFit, "autofit", true, "Fit panels to screen")
	f.BoolVar(&a.LXDEEnabled, "lxde", false, "Initialize LXDE for kiosk mode")
	f.StringVar(&a.LXDEHome, "lxde-home", "/home/pi", "Path to home directory of LXDE user running X Server")
	f.StringVar(&a.Mode, "mode", "full", "Kiosk Display Mode [full|tv|disabled]\nfull = No TOPNAV and No SIDEBAR\ntv = No SIDEBAR\ndisabled = omit option\n")
	f.StringVar(&a.URL, "url", "https://play.grafana.org", "URL to Grafana server")
	f.BoolVar(&a.IgnoreCertificateErrors, "ignore-certificate-errors", false, "Ignore SSL/TLS certificate error")
	f.BoolVar(&a.IsPlayList, "playlists", false, "URL is a playlist")
	f.StringVar(&a.LoginMethod, "login-method", "anon", "[anon|local|gcom]")
	f.StringVar(&a.Username, "username", "guest", "username")
	f.StringVar(&a.Password, "password", "guest", "password")

	// get config usage with wrapped flag usage
	//f.Usage = cleanenv.FUsage(f.Output(), &cfg, nil, f.Usage)

	fu := f.Usage
	f.Usage = func() {
		fu()
		envHelp, _ := cleanenv.GetDescription(cfg, nil)
		fmt.Fprintln(f.Output())
		fmt.Fprintln(f.Output(), envHelp)
	}

	err := f.Parse(os.Args[1:])
	if err != nil {
		f.Usage()
		os.Exit(-1)
	}
	return a
}

func setEnvironment() {
	// for linux/X display must be set
	var displayEnv = os.Getenv("DISPLAY")
	if displayEnv == "" {
		log.Println("DISPLAY not set, autosetting to :0.0")
		os.Setenv("DISPLAY", ":0.0")
		displayEnv = os.Getenv("DISPLAY")
	}
	log.Println("DISPLAY=", displayEnv)

	var xAuthorityEnv = os.Getenv("XAUTHORITY")
	if xAuthorityEnv == "" {
		log.Println("XAUTHORITY not set, autosetting")
		// use HOME of current user
		var homeEnv = os.Getenv("HOME")
		os.Setenv("XAUTHORITY", homeEnv+"/.Xauthority")
		xAuthorityEnv = os.Getenv("XAUTHORITY")
	}
	log.Println("XAUTHORITY=", xAuthorityEnv)
}

func main() {
	var cfg kiosk.Config
	// override
	args := ProcessArgs(&cfg)
	cfg.Target.URL = args.URL
	// read configuration from the file and environment variables
	if err := cleanenv.ReadConfig(args.ConfigPath, &cfg); err != nil {
		log.Printf("No config file specified, continuing")
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

	if cfg.General.LXDEEnabled {
		initialize.LXDE(cfg.General.LXDEHome)
	}

	// for linux/X display must be set
	setEnvironment()

	switch cfg.Target.LoginMethod {
	case "local":
		log.Printf("Launching local login kiosk")
		kiosk.GrafanaKioskLocal(&cfg)
	case "gcom":
		log.Printf("Launching GCOM login kiosk")
		kiosk.GrafanaKioskGCOM(&cfg)
	default:
		log.Printf("Launching ANON login kiosk")
		kiosk.GrafanaKioskAnonymous(&cfg)
	}
}
