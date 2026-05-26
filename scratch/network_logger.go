package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func main() {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.Flag("headless", "new"),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Enable network events
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			if strings.Contains(ev.Request.URL, "google.com") {
				fmt.Printf("[REQ] ID: %s | Method: %s | Type: %s | URL: %s\n", 
					ev.RequestID, ev.Request.Method, ev.Type, ev.Request.URL[:min(120, len(ev.Request.URL))])
			}
		case *network.EventLoadingFinished:
			fmt.Printf("[FINISH] ID: %s | Time: %v\n", ev.RequestID, time.Now().Format("15:04:05.000"))
		case *network.EventLoadingFailed:
			fmt.Printf("[FAIL] ID: %s | Error: %s\n", ev.RequestID, ev.ErrorText)
		}
	})

	err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate("https://www.google.com/search?q=what+is+the+capital+of+france"),
		chromedp.Sleep(6*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
