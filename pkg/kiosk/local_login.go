package kiosk

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

// GrafanaKioskLocal creates a chrome-based kiosk using a local grafana-server account
func GrafanaKioskLocal(cfg *Config) {
	dir, err := ioutil.TempDir("", "chromedp-example")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("noerrdialogs", true),
		chromedp.Flag("kiosk", true),
		chromedp.Flag("bwsi", true),
		chromedp.Flag("incognito", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-notifications", true),
		chromedp.Flag("disable-overlay-scrollbar", true),
		chromedp.Flag("ignore-certificate-errors", cfg.Target.IgnoreCertificateErrors),
		chromedp.Flag("test-type", cfg.Target.IgnoreCertificateErrors),
		chromedp.Flag("window-position", cfg.General.WindowPosition),
		chromedp.Flag("check-for-update-interval", "31536000"),
		chromedp.UserDataDir(dir),
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// also set up a custom logger
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	listenChromeEvents(taskCtx, targetCrashed)

	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx); err != nil {
		panic(err)
	}

	anURL := cfg.Target.URL
	if cfg.Target.IsPlayList {
		anURL, err = gatherPlayListUID(anURL, cfg.Target.Username, cfg.Target.Password, cfg.Target.IgnoreCertificateErrors)
		if err != nil {
			panic(err)
		}
	}
	var generatedURL = GenerateURL(anURL, cfg.General.Mode, cfg.General.AutoFit, cfg.Target.IsPlayList)
	log.Println("Navigating to ", generatedURL)
	/*
		Launch chrome and login with local user account

		name=user, type=text
		id=inputPassword, type=password, name=password
	*/
	// Give browser time to load next page (this can be prone to failure, explore different options vs sleeping)
	time.Sleep(2000 * time.Millisecond)

	if err := chromedp.Run(taskCtx,
		chromedp.Navigate(generatedURL),
		chromedp.WaitVisible(`//input[@name="user"]`, chromedp.BySearch),
		chromedp.SendKeys(`//input[@name="user"]`, cfg.Target.Username, chromedp.BySearch),
		chromedp.SendKeys(`//input[@name="password"]`, cfg.Target.Password+kb.Enter, chromedp.BySearch),
		chromedp.WaitVisible(`notinputPassword`, chromedp.ByID),
	); err != nil {
		panic(err)
	}
}

func gatherPlayListUID(anURL, username, password string, ignoreCertErrors bool) (string, error) {
	client, err := NewGrafanaClient(anURL, username, password, ignoreCertErrors)
	if err != nil {
		log.Println("unable to create grafana Client")
		return "", err
	}
	uid, err := GetPlayListUID(anURL, client)
	if err != nil {
		log.Println("")
		return "", fmt.Errorf("unable to get the uid from the id defined %e", err)
	}

	log.Printf("Playlist uid: %s", uid)

	// replace the id with uid
	urlOfAnURL, err := url.Parse(anURL)
	if err != nil {
		return "", err
	}
	uidURL := UrlChangeIDtoUID(urlOfAnURL, uid)

	/*
		err = nil
		anURL, err = ChangeIDtoUID(anURL, uid)
		if err != nil {
			return "", err
		}
	*/
	log.Println("URL using uid:", uidURL.String())
	return uidURL.String(), nil
}
