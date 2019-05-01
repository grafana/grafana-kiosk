package kiosk

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

// GrafanaKioskGCOM creates a chrome-based kiosk using a grafana.com authenticated account
func GrafanaKioskGCOM(urlPtr *string, usernamePtr *string, passwordPtr *string, autoFit bool) {
	dir, err := ioutil.TempDir("", "chromedp-example")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		// chromedp.DisableGPU, // needed?
		chromedp.Flag("noerrdialogs", true),
		chromedp.Flag("kiosk", true),
		chromedp.Flag("bwsi", true),
		chromedp.Flag("incognito", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-notifications", true),
		chromedp.Flag("disable-overlay-scrollbar", true),
		chromedp.UserDataDir(dir),
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// also set up a custom logger
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx); err != nil {
		panic(err)
	}

	/*
		Launch chrome, click the grafana.com button, fill out login form and submit
	*/
	// XPATH of grafana.com login button = //*[@href="login/grafana_com"]/i

	// Click the grafana_com login button
	if err := chromedp.Run(taskCtx,
		chromedp.Navigate(*urlPtr),
		chromedp.WaitVisible("//*[@href=\"login/grafana_com\"]/i", chromedp.BySearch),
		chromedp.Click("//*[@href=\"login/grafana_com\"]/..", chromedp.BySearch),
	); err != nil {
		panic(err)
	}
	// Give browser time to load next page (this can be prone to failure, explore different options vs sleeping)
	time.Sleep(2000 * time.Millisecond)
	// Fill out grafana_com login page
	if err := chromedp.Run(taskCtx,
		chromedp.WaitVisible("//input[@name=\"login\"]", chromedp.BySearch),
		chromedp.SendKeys("//input[@name=\"login\"]", *usernamePtr, chromedp.BySearch),
		chromedp.SendKeys("//input[@name=\"password\"]", *passwordPtr+kb.Enter, chromedp.BySearch),
		chromedp.WaitVisible("notinputPassword", chromedp.ByID),
	); err != nil {
		panic(err)
	}

}
