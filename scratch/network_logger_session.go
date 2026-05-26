package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type Session struct {
	ID      string            `json:"id"`
	Headers map[string]string `json:"headers"`
	Cookies []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"cookies"`
}

type SessionConfig struct {
	Sessions []Session `json:"sessions"`
}

func main() {
	// Read a session from session_config.json
	data, err := os.ReadFile("solver/session_config.json")
	if err != nil {
		log.Fatalf("Failed to read session config: %v", err)
	}

	var config SessionConfig
	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatalf("Failed to unmarshal session config: %v", err)
	}

	if len(config.Sessions) == 0 {
		log.Fatal("No sessions found in config")
	}

	session := config.Sessions[0]

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

	err = chromedp.Run(ctx,
		network.Enable(),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Inject cookies
			for _, c := range session.Cookies {
				err := network.SetCookie(c.Name, c.Value).
					WithDomain(".google.com").
					WithPath("/").
					WithSecure(true).
					WithHTTPOnly(false).
					Do(ctx)
				if err != nil {
					return err
				}
			}
			return nil
		}),
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
