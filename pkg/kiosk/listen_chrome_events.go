package kiosk

import (
	"context"
	"log"

	"github.com/chromedp/cdproto/inspector"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

type chromeEvents int

const (
	consoleAPICall chromeEvents = 1 << iota
	targetCrashed
)

func listenChromeEvents(taskCtx context.Context, cfg *Config, events chromeEvents) {
	headers := make(map[string]interface{})
	if len(cfg.BasicAuth.Username) != 0 && len(cfg.BasicAuth.Password) != 0 {
		headers["Authorization"] = GenerateHTTPBasicAuthHeader(cfg.BasicAuth.Username, cfg.BasicAuth.Password)
	}

	chromedp.ListenTarget(taskCtx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *runtime.EventConsoleAPICalled:
			if events&consoleAPICall != 0 {
				log.Printf("console.%s call:", ev.Type)
				for _, arg := range ev.Args {
					log.Printf("	%s - %s", arg.Type, arg.Value)
				}
			}
		case *inspector.EventTargetCrashed:
			if events&targetCrashed != 0 {
				log.Printf("target crashed, reload...")
				go func() {
					_ = chromedp.Run(taskCtx,
						network.Enable(),
						network.SetExtraHTTPHeaders(network.Headers(headers)),
						network.Enable(),
						network.SetExtraHTTPHeaders(network.Headers(headers)),
						chromedp.Reload())
				}()
			}
		default:
			if cfg.General.DebugEnabled {
				log.Printf("Unknown Event: %+v", ev)
			}
		}
	})
}
