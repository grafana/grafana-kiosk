package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/grafana/grafana-kiosk/pkg/initialize"
	"github.com/grafana/grafana-kiosk/pkg/kiosk"
)

// LoginMethod specifies the type of login to be used by the kiosk
type LoginMethod int

// Login Methods
const (
	ANONYMOUS LoginMethod = 0
	LOCAL     LoginMethod = 1
	GCOM      LoginMethod = 2
)

// Kiosk Modes
const (
	// TV will hide the sidebar but allow usage of menu
	TV int = 0
	// NORMAL will disable sidebar and top navigation bar
	NORMAL int = 1
	// DISABLED will omit kiosk option
	DISABLED int = 2
)

var (
	loginMethod = LOCAL
	kioskMode   = NORMAL
)

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
	var Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %v\n", os.Args[0])
		flag.PrintDefaults()
	}
	urlPtr := flag.String("URL", "https://play.grafana.org", "URL to Grafana server (Required)")
	methodPtr := flag.String("login-method", "anon", "login method: [anon|local|gcom]")
	usernamePtr := flag.String("username", "guest", "username (Required)")
	passwordPtr := flag.String("password", "guest", "password (Required)")
	ignoreCertificateErrors := flag.Bool("ignore-certificate-errors", false, "ignore SSL/TLS certificate errors")
	// kiosk=tv includes sidebar menu
	// kiosk no sidebar ever
	kioskModePtr := flag.String("kiosk-mode", "full", "kiosk mode [full|tv|disabled]")
	autoFit := flag.Bool("autofit", true, "autofit panels in kiosk mode")
	// when the URL is a playlist, append "inactive" to the URL
	isPlayList := flag.Bool("playlist", false, "URL is a playlist: [true|false]")
	LXDEEnabled := flag.Bool("lxde", false, "initialize LXDE for kiosk mode")
	LXDEHomePtr := flag.String("lxde-home", "/home/pi", "path to home directory of LXDE user running X Server")
	flag.Parse()

	// make sure the url has content
	if *urlPtr == "" {
		Usage()
		os.Exit(1)
	}
	// validate url
	_, err := url.ParseRequestURI(*urlPtr)
	if err != nil {
		Usage()
		panic(err)
	}

	if *isPlayList {
		log.Printf("playlist")
	}

	if *LXDEEnabled {
		initialize.LXDE(*LXDEHomePtr)
	}
	switch *kioskModePtr {
	case "tv": // NO SIDEBAR ACCESS
		kioskMode = TV
	case "full": // NO TOPNAV or SIDEBAR
		kioskMode = NORMAL
	case "disabled": // NO TOPNAV or SIDEBAR
		kioskMode = DISABLED
	default:
		kioskMode = NORMAL
	}

	switch *methodPtr {
	case "anon":
		loginMethod = ANONYMOUS
	case "local":
		loginMethod = LOCAL
	case "gcom":
		loginMethod = GCOM
	default:
		loginMethod = ANONYMOUS
	}

	// for linux/X display must be set
	setEnvironment()

	switch loginMethod {
	case LOCAL:
		log.Printf("Launching local login kiosk")
		kiosk.GrafanaKioskLocal(urlPtr, usernamePtr, passwordPtr, kioskMode, autoFit, isPlayList, ignoreCertificateErrors)
	case GCOM:
		log.Printf("Launching GCOM login kiosk")
		kiosk.GrafanaKioskGCOM(urlPtr, usernamePtr, passwordPtr, kioskMode, autoFit, isPlayList)
	case ANONYMOUS:
		log.Printf("Launching ANON login kiosk")
		kiosk.GrafanaKioskAnonymous(urlPtr, kioskMode, autoFit, isPlayList, ignoreCertificateErrors)
	}
}
