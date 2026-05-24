//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"go_search/solver"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run inspect_sge.go <query>")
	}
	query := strings.Join(os.Args[1:], " ")

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", true),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-infobars", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("disable-features", "DownloadFonts,FontAccess"),
		chromedp.WindowSize(1440, 900),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()

	// Set timeout
	ctx, cancelTimeout := context.WithTimeout(ctx, 60*time.Second)
	defer cancelTimeout()

	searchURL := fmt.Sprintf("https://www.google.com/search?q=%s&hl=en", url.QueryEscape(query))
	log.Printf("Navigating to: %s", searchURL)

	var currentLoc string

	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			err := network.Enable().Do(ctx)
			if err != nil {
				return err
			}
			err = network.SetBlockedURLs().WithURLPatterns([]*network.BlockPattern{
				{URLPattern: "*://*:*/*.css", Block: true},
				{URLPattern: "*://*:*/*.woff", Block: true},
				{URLPattern: "*://*:*/*.woff2", Block: true},
				{URLPattern: "*://*:*/*.ttf", Block: true},
				{URLPattern: "*://*:*/*.png", Block: true},
				{URLPattern: "*://*:*/*.jpg", Block: true},
				{URLPattern: "*://*:*/*.jpeg", Block: true},
				{URLPattern: "*://*:*/*.gif", Block: true},
				{URLPattern: "*://*:*/*.svg", Block: true},
				{URLPattern: "*://*:*/*.mp4", Block: true},
				{URLPattern: "*://*:*/*.webm", Block: true},
				{URLPattern: "*://*/*analytics*", Block: true},
				{URLPattern: "*://*/*doubleclick*", Block: true},
			}).Do(ctx)
			if err != nil {
				return err
			}
			_, err = page.AddScriptToEvaluateOnNewDocument(solver.StealthScript).Do(ctx)
			return err
		}),
		chromedp.Navigate(searchURL),
		chromedp.Location(&currentLoc),
	)
	if err != nil {
		log.Fatalf("Navigation failed: %v", err)
	}
	log.Printf("Current location: %s", currentLoc)

	if strings.Contains(strings.ToLower(currentLoc), "sorry") {
		log.Printf("⚠️ Browser hit CAPTCHA, attempting to solve...")
		solved, solveErr := solver.DefeatCaptcha(ctx, 200, 400)
		if solveErr != nil {
			log.Fatalf("❌ CAPTCHA solver error: %v", solveErr)
		} else if solved {
			log.Printf("✅ CAPTCHA solved, waiting for Google redirect...")
			time.Sleep(2 * time.Second)
			_ = chromedp.Run(ctx, chromedp.Location(&currentLoc))
			log.Printf("New location: %s", currentLoc)
		} else {
			log.Fatalf("❌ CAPTCHA could not be solved")
		}
	}

	// Wait up to 15 seconds for the SGE container
	log.Println("Waiting for AI Overview container (.s7d4ef)...")
	err = chromedp.Run(ctx,
		chromedp.Poll(`(() => {
			const el = document.querySelector('.s7d4ef');
			if (el) {
				// Wait for it to change from 'Thinking'
				const text = el.innerText.toLowerCase();
				if (text.includes("thinking") || text.includes("searching")) {
					return false;
				}
				return true;
			}
			// Fallback to organic
			if (document.querySelectorAll('a h3').length > 0) {
				return true;
			}
			return false;
		})()`, nil, chromedp.WithPollingInterval(100*time.Millisecond)),
	)
	if err != nil {
		log.Fatalf("Poll failed: %v", err)
	}

	var exists bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`!!document.querySelector('.s7d4ef')`, &exists),
	)
	if err != nil {
		log.Fatalf("Check exists failed: %v", err)
	}

	if !exists {
		log.Println("❌ SGE container not found (AI Overview suppressed/not generated).")
		return
	}

	log.Println("Found SGE, sleeping 5 seconds to let streaming finish...")
	time.Sleep(5 * time.Second)

	var outerHTML string
	err = chromedp.Run(ctx,
		chromedp.OuterHTML(`.s7d4ef`, &outerHTML),
	)
	if err != nil {
		log.Fatalf("Failed to get SGE HTML: %v", err)
	}

	// Save to a file in scratch directory
	outFile := "scratch/sge_dom_dump.html"
	err = os.WriteFile(outFile, []byte(outerHTML), 0644)
	if err != nil {
		log.Fatalf("Failed to write DOM dump to file: %v", err)
	}

	log.Printf("✅ Saved RAW SGE HTML to %s", outFile)
}
