package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"

	"github.com/grafana/grafana-kiosk/pkg/kiosk"
)

var (
	// Version this is set during build time using git tags
	Version string
)

// Args command-line parameters.
type Args struct {
	ConfigPath string
}

// ProcessArgs processes and handles CLI arguments.
func ProcessArgs(cfg interface{}) Args {
	var processedArgs Args

	flagSettings := flag.NewFlagSet("migrate-config", flag.ContinueOnError)
	flagSettings.StringVar(&processedArgs.ConfigPath, "c", "", "Path to configuration file (config.yaml)")

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

func summary(cfg *kiosk.Config) {
	// general
	log.Println("AutoFit:", cfg.GrafanaOptions.AutoFit)
	log.Println("LXDEEnabled:", cfg.General.LXDEEnabled)
	log.Println("LXDEHome:", cfg.General.LXDEHome)
	log.Println("GrafanaOptions - Kiosk Mode:", cfg.GrafanaOptions.KioskMode)
	log.Println("WindowPosition:", cfg.ChromeDPFlags.WindowPosition)
	log.Println("WindowSize:", cfg.ChromeDPFlags.WindowSize)
	log.Println("ScaleFactor:", cfg.ChromeDPFlags.ScaleFactor)
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
	fmt.Println("Migrate Config Version:", Version)
	// set the version
	cfg.BuildInfo.Version = Version

	// override
	args := ProcessArgs(&cfg)

	// check if config specified
	if args.ConfigPath != "" {
		// read configuration from the file and then override with environment variables
		if err := cleanenv.ReadConfig(args.ConfigPath, &cfg); err != nil {
			log.Println("Error reading config file", err)
			os.Exit(-1)
		} else {
			log.Println("Using config from", args.ConfigPath)
		}
	}
	summary(&cfg)

}
