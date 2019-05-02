package kiosk

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

// GrafanaKioskAnonymous creates a chrome-based kiosk using a local grafana-server account
func GrafanaKioskAnonymous(urlPtr *string, kioskMode int, autoFit *bool, isPlayList *bool) {
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
	chromedp.ListenTarget(taskCtx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *runtime.EventConsoleAPICalled:
			log.Printf("console.%s call:\n", ev.Type)
			for _, arg := range ev.Args {
				log.Printf("%s - %s\n", arg.Type, arg.Value)
			}
		}
	})
	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx); err != nil {
		panic(err)
	}

	// Give browser time to load next page (this can be prone to failure, explore different options vs sleeping)
	time.Sleep(2000 * time.Millisecond)

	var generatedURL = GenerateURL(*urlPtr, kioskMode, autoFit, isPlayList)
	/*
		Launch chrome and look for main-view element
	*/
	log.Println("Navigating to ", generatedURL)
	if err := chromedp.Run(taskCtx,
		chromedp.Navigate(generatedURL),
		chromedp.WaitVisible("//div[@class=\"main-view\"]", chromedp.BySearch),
		// wait forever (for now)
		chromedp.WaitVisible("notnputPassword", chromedp.ByID),
	); err != nil {
		panic(err)
	}
	log.Println("Sleep before exit...")
	// wait here for the process to exit
	time.Sleep(2000 * time.Millisecond)
	log.Println("Exit...")

}
