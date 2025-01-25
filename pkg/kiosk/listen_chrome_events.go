package kiosk

import (
	"context"
	"log"
	"strings"

	"github.com/chromedp/cdproto/inspector"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

type chromeEvents int

const (
	consoleAPICall chromeEvents = 1 << iota
	targetCrashed
)

func listenChromeEvents(taskCtx context.Context, cfg *Config, events chromeEvents) {
	chromedp.ListenTarget(taskCtx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *runtime.EventConsoleAPICalled:
			if events&consoleAPICall != 0 {
				log.Printf("console.%s call:", ev.Type)
				for _, arg := range ev.Args {
					log.Printf("	%s - %s", arg.Type, arg.Value)
					if strings.Contains(string(arg.Value), "not correct url correcting") {
						log.Printf("playlist may be broken, restart!")
					}
				}
			}
			if ev.StackTrace != nil {
				log.Printf("console.%s stacktrace:", ev.Type)
				for _, arg := range ev.StackTrace.CallFrames {
					log.Printf("(%s:%d): %s", arg.URL, arg.LineNumber, arg.FunctionName)
				}
			}
		case *inspector.EventTargetCrashed:
			if events&targetCrashed != 0 {
				log.Printf("target crashed, reload...")
				go func() {
					_ = chromedp.Run(taskCtx, chromedp.Reload())
				}()
			}
		default:
			if cfg.General.DebugEnabled {
				log.Printf("Unknown Event: %+v", ev)
			}
		}
	})
}
