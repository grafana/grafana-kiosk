package shared

import (
	"context"
	"log"
	"strings"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/inspector"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"

	"github.com/grafana/grafana-kiosk/pkg/kiosk/config"
)

// BrowserEvents is a bitmask of browser event categories to log.
type BrowserEvents int

const (
	// ConsoleAPICall enables logging of console.* calls from the page.
	ConsoleAPICall BrowserEvents = 1 << iota
	// TargetCrashed enables reload-on-crash handling.
	TargetCrashed
)

// ListenBrowserEvents subscribes to chromedp target events for the given context.
func ListenBrowserEvents(taskCtx context.Context, cfg *config.Config, events BrowserEvents) {
	chromedp.ListenTarget(taskCtx, func(ev any) {
		switch ev := ev.(type) {
		case *runtime.EventConsoleAPICalled:
			if events&ConsoleAPICall != 0 {
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
			if events&TargetCrashed != 0 {
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

// NewBrowserContext creates a chromedp allocator and task context, starts the
// browser, registers event listeners, and waits for browser startup. The
// returned cancel must be deferred by the caller.
func NewBrowserContext(ctx context.Context, cfg *config.Config, dir string, events BrowserEvents) (taskCtx context.Context, cancel func()) {
	opts := GenerateExecutorOptions(dir, cfg)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, opts...)
	taskCtx, cancelTask := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	ListenBrowserEvents(taskCtx, cfg, events)
	if err := chromedp.Run(taskCtx); err != nil {
		cancelTask()
		cancelAlloc()
		panic(err)
	}
	WaitForBrowserStartup(cfg)
	return taskCtx, func() {
		cancelTask()
		cancelAlloc()
	}
}

// GetExecutor returns executor for chromedp.
func GetExecutor(ctx context.Context) context.Context {
	c := chromedp.FromContext(ctx)

	return cdp.WithExecutor(ctx, c.Target)
}
