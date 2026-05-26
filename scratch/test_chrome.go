package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/chromedp/chromedp"
)

func main() {
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", "new"),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-infobars", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("disable-features", "DownloadFonts,FontAccess"),
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.Flag("mute-audio", true),
		chromedp.WindowSize(1440, 900),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"),
	}

	tempDir, err := os.MkdirTemp("", "ultrasearch-chrome-*")
	if err != nil {
		log.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	fmt.Printf("Temp user data dir: %s\n", tempDir)

	opts = append(opts, chromedp.UserDataDir(tempDir))

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancelBrowser := chromedp.NewContext(allocCtx)
	defer cancelBrowser()

	fmt.Println("Starting Chrome...")
	err = chromedp.Run(ctx, chromedp.Navigate("https://www.google.com"))
	if err != nil {
		log.Fatalf("failed to run Chrome: %v", err)
	}
	fmt.Println("Success!")
}
