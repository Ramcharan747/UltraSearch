package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/go-shiori/go-readability"
	"go_search/solver"
)

// SearchResult represents a single search result
type SearchResult struct {
	Rank    int    `json:"rank"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
	Content string `json:"content,omitempty"`
	Tier    int    `json:"tier,omitempty"`
}

// SearchResponse represents the results for a single query
type SearchResponse struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
	Error   string         `json:"error,omitempty"`
}

const extractJS = `(maxResults) => {
    const out = [];
    const links = document.querySelectorAll('a h3');
    for (const h3 of links) {
        const a = h3.closest('a');
        if (a && a.href && !a.href.includes('google.com')) {
            let snippet = '';
            const parent = a.closest('[data-sokoban-container]') || a.closest('.g') || a.parentElement?.parentElement?.parentElement;
            
            if (parent) {
                const elements = parent.querySelectorAll('div, span');
                let maxLen = 0;
                for (const el of elements) {
                    const text = el.innerText || '';
                    if (text.length > maxLen && text !== h3.innerText && !text.includes('›') && !el.querySelector('h3')) {
                        maxLen = text.length;
                        snippet = text;
                    }
                }
            }
            
            snippet = snippet.replace(/\n/g, ' ').trim();
            out.push({ rank: out.length + 1, title: h3.innerText, url: a.href, snippet: snippet.substring(0, 1000) });
        }
        if (out.length >= maxResults) break;
    }
    return out;
}`

func init() {
	err := solver.LoadTrajectories("solver/trajectories.json")
	if err != nil {
		log.Printf("⚠️ Could not load trajectories: %v", err)
	}
}

// ==================== EXTRACTION FUNCTIONS ====================

// extractText parses HTML into clean text via readability, with goquery fallback
func extractText(html string) string {
	parsed, err := readability.FromReader(strings.NewReader(html), nil)
	if err == nil && len(parsed.TextContent) > 50 {
		text := strings.Join(strings.Fields(parsed.TextContent), " ")
		if len(text) > 2000 {
			text = text[:2000]
		}
		return text
	}
	// Fallback: goquery strip
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err == nil {
		doc.Find("script, style, noscript, nav, footer, header, aside").Remove()
		text := strings.Join(strings.Fields(doc.Find("body").Text()), " ")
		if len(text) > 2000 {
			text = text[:2000]
		}
		if len(text) > 50 {
			return text
		}
	}
	return ""
}

// ==================== WORKER ====================

func worker(id int, queries <-chan string, results chan<- SearchResponse, searchAllocCtx context.Context, maxResults int, fetchContent bool, wg *sync.WaitGroup) {
	defer wg.Done()
	httpClient := SharedHTTPClient()

	for q := range queries {
		start := time.Now()
		
		// --- PHASE 0: Google Search ---
		ctx, cancel := chromedp.NewContext(searchAllocCtx)
		ctx, cancelTimeout := context.WithTimeout(ctx, 25*time.Second)
		
		chromedp.ListenTarget(ctx, func(ev interface{}) {
			switch ev := ev.(type) {
			case *fetch.EventRequestPaused:
				go func() {
					c := chromedp.FromContext(ctx)
					resType := ev.ResourceType
					if resType == network.ResourceTypeImage || resType == network.ResourceTypeMedia || resType == network.ResourceTypeStylesheet || resType == network.ResourceTypeFont {
						fetch.FailRequest(ev.RequestID, network.ErrorReasonFailed).Do(cdp.WithExecutor(ctx, c.Target))
					} else {
						fetch.ContinueRequest(ev.RequestID).Do(cdp.WithExecutor(ctx, c.Target))
					}
				}()
			}
		})

		var res []SearchResult
		var pageURL string
		searchURL := fmt.Sprintf("https://www.google.com/search?q=%s&hl=en&num=%d", url.QueryEscape(q), maxResults+10)

		err := chromedp.Run(ctx,
			fetch.Enable(),
			chromedp.ActionFunc(func(ctx context.Context) error {
				_, err := page.AddScriptToEvaluateOnNewDocument(solver.StealthScript).Do(ctx)
				return err
			}),
			chromedp.Navigate(searchURL),
			chromedp.Sleep(600*time.Millisecond), 
			chromedp.Location(&pageURL),
			chromedp.Evaluate(fmt.Sprintf("(%s)(%d)", extractJS, maxResults), &res),
		)

		cancelTimeout()
		cancel()

		if err != nil {
			log.Printf("   ❌ W%d: '%s' -> Error: %v", id, q, err)
			results <- SearchResponse{Query: q, Error: err.Error()}
			continue
		} else if strings.Contains(strings.ToLower(pageURL), "sorry") {
			log.Printf("   ⚠️ W%d: '%s' -> BLOCKED", id, q)
			results <- SearchResponse{Query: q, Error: "blocked_by_captcha"}
			continue
		}

		// --- PHASE 1: CLASSIFY all URLs ---
		deepLimit := 5
		if deepLimit > len(res) {
			deepLimit = len(res)
		}

		type classifiedURL struct {
			idx   int
			tier  int
			html  string // Populated for T1 (static)
		}

		classifyCh := make(chan classifiedURL, deepLimit)
		
		for i := 0; i < deepLimit; i++ {
			go func(idx int) {
				cu := classifiedURL{idx: idx}

				// Step 1: Check domain cache first (instant)
				cachedTier := lookupDomainTier(res[idx].URL)
				if cachedTier > 0 {
					cu.tier = cachedTier
					classifyCh <- cu
					return
				}

				// Step 2: HTTP probe (curl-speed)
				probe := probeURL(res[idx].URL, httpClient)
				cu.tier = probe.Tier
				cu.html = probe.HTML
				classifyCh <- cu
			}(i)
		}

		// Collect classifications
		var t1Results []classifiedURL // Static HTML - already have content
		var t2Idxs []int             // JS-render needed
		var t3Idxs []int             // Bot-protected
		var t4Idxs []int             // Login-walled

		for i := 0; i < deepLimit; i++ {
			cu := <-classifyCh
			res[cu.idx].Tier = cu.tier
			switch cu.tier {
			case TierStatic:
				t1Results = append(t1Results, cu)
			case TierJSRender:
				t2Idxs = append(t2Idxs, cu.idx)
			case TierBotProtect:
				t3Idxs = append(t3Idxs, cu.idx)
			case TierLoginWall:
				t4Idxs = append(t4Idxs, cu.idx)
			// TierUnreachable: skip
			}
		}

		log.Printf("   📊 W%d: Classified %d URLs → T1:%d T2:%d T3:%d T4:%d",
			id, deepLimit, len(t1Results), len(t2Idxs), len(t3Idxs), len(t4Idxs))

		if !fetchContent {
			log.Printf("   ✅ W%d: '%s' -> %d results (content skipped)", id, q, len(res))
			results <- SearchResponse{Query: q, Results: res}
			continue
		}

		// --- PHASE 2: EXTRACT T1 (instant, already have HTML) ---
		for _, cu := range t1Results {
			text := extractText(cu.html)
			if ContentQuality(text) {
				res[cu.idx].Content = text
			} else {
				// Quality too low → escalate to T2
				t2Idxs = append(t2Idxs, cu.idx)
			}
		}

		// --- PHASE 3: EXTRACT T2 (JS-render via shared headless browser) ---
		if len(t2Idxs) > 0 {
			jsOpts := []chromedp.ExecAllocatorOption{
				chromedp.NoFirstRun,
				chromedp.NoDefaultBrowserCheck,
				chromedp.Flag("headless", true),
				chromedp.Flag("enable-automation", false),
				chromedp.Flag("disable-blink-features", "AutomationControlled"),
				chromedp.Flag("disable-gpu", true),
				chromedp.Flag("no-sandbox", true),
				chromedp.Flag("disable-features", "DownloadFonts,FontAccess"),
				chromedp.WindowSize(1440, 900),
				chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"),
			}

			jsAlloc, jsAllocCancel := chromedp.NewExecAllocator(context.Background(), jsOpts...)
			jsParent, jsParentCancel := chromedp.NewContext(jsAlloc)
			chromedp.Run(jsParent) // Start browser once

			// Process T2 URLs as tabs, 3 at a time
			sem := make(chan struct{}, 3) // Concurrency limiter
			var t2Wg sync.WaitGroup
			var t2Mu sync.Mutex
			var t2Escalate []int // URLs that fail T2 → escalate to T3

			for _, idx := range t2Idxs {
				t2Wg.Add(1)
				sem <- struct{}{} // Acquire slot
				go func(idx int) {
					defer t2Wg.Done()
					defer func() { <-sem }() // Release slot

					tabCtx, tabCancel := chromedp.NewContext(jsParent)
					tabCtx, tabTimeout := context.WithTimeout(tabCtx, 10*time.Second)

					var htmlDump string
					err := chromedp.Run(tabCtx,
						chromedp.ActionFunc(func(c context.Context) error {
							_, err := page.AddScriptToEvaluateOnNewDocument(solver.StealthScript).Do(c)
							return err
						}),
						chromedp.Navigate(res[idx].URL),
						chromedp.Sleep(2*time.Second),
						chromedp.OuterHTML("html", &htmlDump),
					)
					tabTimeout()
					tabCancel()

					if err != nil || len(htmlDump) < 500 {
						t2Mu.Lock()
						t2Escalate = append(t2Escalate, idx)
						t2Mu.Unlock()
						return
					}

					text := extractText(htmlDump)
					if ContentQuality(text) {
						t2Mu.Lock()
						res[idx].Content = text
						t2Mu.Unlock()
					} else {
						t2Mu.Lock()
						t2Escalate = append(t2Escalate, idx)
						t2Mu.Unlock()
					}
				}(idx)
			}
			t2Wg.Wait()
			jsParentCancel()
			jsAllocCancel()

			// Escalate failed T2 → T3
			t3Idxs = append(t3Idxs, t2Escalate...)
		}

		// --- PHASE 4: EXTRACT T3 + T4 (stealth browser with solver) ---
		allStealthIdxs := append(t3Idxs, t4Idxs...)
		if len(allStealthIdxs) > 0 {
			log.Printf("   🛡️ W%d: Stealth browser for %d URLs (T3:%d T4:%d)", id, len(allStealthIdxs), len(t3Idxs), len(t4Idxs))

			stealthOpts := []chromedp.ExecAllocatorOption{
				chromedp.NoFirstRun,
				chromedp.NoDefaultBrowserCheck,
				chromedp.Flag("headless", false),
				chromedp.Flag("enable-automation", false),
				chromedp.Flag("disable-blink-features", "AutomationControlled"),
				chromedp.Flag("disable-infobars", true),
				chromedp.Flag("disable-gpu", true),
				chromedp.Flag("no-sandbox", true),
				chromedp.Flag("disable-extensions", false),
				chromedp.Flag("disable-features", "DownloadFonts,FontAccess"),
				chromedp.WindowSize(1440, 900),
				chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"),
			}

			stealthAlloc, stealthAllocCancel := chromedp.NewExecAllocator(context.Background(), stealthOpts...)
			stealthParent, stealthParentCancel := chromedp.NewContext(stealthAlloc)
			chromedp.Run(stealthParent) // Start browser once

			// Group URLs by root domain
			domainGroups := make(map[string][]int)
			for _, idx := range allStealthIdxs {
				parsedUrl, err := url.Parse(res[idx].URL)
				if err != nil {
					continue
				}
				domain := parsedUrl.Scheme + "://" + parsedUrl.Host + "/"
				domainGroups[domain] = append(domainGroups[domain], idx)
			}

			// Process each domain group sequentially to avoid IP limits
			for domain, idxs := range domainGroups {
				log.Printf("   🚗 W%d: Parking on domain %s for %d targets", id, domain, len(idxs))
				
				parkCtx, parkCancel := chromedp.NewContext(stealthParent)
				parkCtx, parkTimeout := context.WithTimeout(parkCtx, 20*time.Second)

				// Step 1: Park on root domain
				err := chromedp.Run(parkCtx,
					chromedp.ActionFunc(func(c context.Context) error {
						_, err := page.AddScriptToEvaluateOnNewDocument(solver.StealthScript).Do(c)
						return err
					}),
					chromedp.Navigate(domain),
					chromedp.Sleep(3*time.Second),
				)

				// Check for CAPTCHA on the parked page
				if err == nil {
					var bodySnippet string
					chromedp.Run(parkCtx, chromedp.Evaluate(`document.body ? document.body.innerText.substring(0, 300).toLowerCase() : ''`, &bodySnippet))

					needsSolver := strings.Contains(bodySnippet, "verify you are human") ||
						strings.Contains(bodySnippet, "just a moment") ||
						strings.Contains(bodySnippet, "checking your browser") ||
						strings.Contains(bodySnippet, "performing security verification") ||
						strings.Contains(bodySnippet, "enable javascript and cookies") ||
						len(bodySnippet) < 30

					if needsSolver {
						log.Printf("   🛡️ W%d: Challenge on root %s, solving...", id, domain)
						solved, _ := solver.DefeatCaptcha(parkCtx, 200, 400)
						if solved {
							// Wait for clearance
							for j := 0; j < 10; j++ {
								time.Sleep(1 * time.Second)
								var title string
								chromedp.Run(parkCtx, chromedp.Title(&title))
								if title != "Just a moment..." && title != "" {
									log.Printf("   ✅ W%d: Clearance acquired for %s", id, domain)
									break
								}
							}
						}
					}
				}

				// Step 2: Silent Fetch all targets for this domain using the trusted tab
				for _, idx := range idxs {
					targetURL := res[idx].URL
					var htmlDump string
					
					js := fmt.Sprintf(`
						window.fetchResult_%d = null;
						(async () => {
							try {
								const response = await fetch('%s');
								const text = await response.text();
								window.fetchResult_%d = text;
							} catch (e) {
								window.fetchResult_%d = "Fetch failed: " + e.message;
							}
						})();
					`, idx, targetURL, idx, idx)
					
					chromedp.Run(parkCtx, chromedp.Evaluate(js, nil))
					
					// Poll for fetch result
					fetchSuccess := false
					for j := 0; j < 15; j++ {
						time.Sleep(500 * time.Millisecond)
						var fetchRes interface{}
						chromedp.Run(parkCtx, chromedp.Evaluate(fmt.Sprintf("window.fetchResult_%d", idx), &fetchRes))
						if fetchRes != nil {
							if s, ok := fetchRes.(string); ok && !strings.HasPrefix(s, "Fetch failed:") {
								htmlDump = s
								fetchSuccess = true
								log.Printf("   👻 W%d: Silent Fetch success on %s", id, targetURL)
							}
							break
						}
					}

					// Step 3: Fallback to normal navigation if fetch fails or returns challenge
					if !fetchSuccess || len(htmlDump) < 500 || strings.Contains(strings.ToLower(htmlDump[:min(500, len(htmlDump))]), "verify you are human") {
						log.Printf("   ⚠️ W%d: Fetch failed/blocked on %s, falling back to tab navigation", id, targetURL)
						
						fallbackCtx, fallbackCancel := chromedp.NewContext(stealthParent)
						fallbackCtx, fallbackTimeout := context.WithTimeout(fallbackCtx, 15*time.Second)
						
						err = chromedp.Run(fallbackCtx,
							chromedp.Navigate(targetURL),
							chromedp.Sleep(2500*time.Millisecond),
						)
						
						if err == nil {
							var fbSnippet string
							chromedp.Run(fallbackCtx, chromedp.Evaluate(`document.body ? document.body.innerText.substring(0, 300).toLowerCase() : ''`, &fbSnippet))
							if strings.Contains(fbSnippet, "just a moment") || len(fbSnippet) < 30 {
								solver.DefeatCaptcha(fallbackCtx, 200, 400)
								chromedp.Run(fallbackCtx, chromedp.Sleep(2*time.Second))
							}
							chromedp.Run(fallbackCtx, chromedp.OuterHTML("html", &htmlDump))
						}
						fallbackTimeout()
						fallbackCancel()
					}

					// Extract text
					if len(htmlDump) > 500 {
						text := extractText(htmlDump)
						if ContentQuality(text) {
							res[idx].Content = text
						}
					}
				}
				
				parkTimeout()
				parkCancel()
			}

			stealthParentCancel()
			stealthAllocCancel()
		}

		// --- STATS ---
		contentCount := 0
		var failedURLs []string
		for i := 0; i < deepLimit; i++ {
			if res[i].Content != "" {
				contentCount++
			} else {
				failedURLs = append(failedURLs, fmt.Sprintf("T%d:%s", res[i].Tier, res[i].URL))
			}
		}

		if len(failedURLs) > 0 {
			log.Printf("   ⚠️ W%d: Failed extractions: %v", id, failedURLs)
		}

		log.Printf("   ✅ W%d: '%s' -> %d results, %d/%d content (%.1fs)",
			id, q, len(res), contentCount, deepLimit, time.Since(start).Seconds())
		
		results <- SearchResponse{Query: q, Results: res}
	}
}

// formatForAI formats the responses into a clean text block for AI consumption.
func formatForAI(responses []SearchResponse) string {
	var builder strings.Builder
	for _, resp := range responses {
		builder.WriteString(fmt.Sprintf("Search Query: %s\n", resp.Query))
		builder.WriteString(strings.Repeat("-", 50) + "\n")
		
		if resp.Error != "" {
			builder.WriteString(fmt.Sprintf("Error: %s\n\n", resp.Error))
			builder.WriteString(strings.Repeat("=", 50) + "\n")
			continue
		}
		
		for _, item := range resp.Results {
			tierLabel := ""
			switch item.Tier {
			case TierStatic:
				tierLabel = "[HTTP]"
			case TierJSRender:
				tierLabel = "[JS]"
			case TierBotProtect:
				tierLabel = "[STEALTH]"
			case TierLoginWall:
				tierLabel = "[LOGIN]"
			case TierUnreachable:
				tierLabel = "[SKIP]"
			}
			builder.WriteString(fmt.Sprintf("[%d] %s %s\n", item.Rank, tierLabel, item.Title))
			builder.WriteString(fmt.Sprintf("URL: %s\n", item.URL))
			if item.Snippet != "" {
				builder.WriteString(fmt.Sprintf("Snippet: %s\n", item.Snippet))
			}
			if item.Content != "" {
				builder.WriteString(fmt.Sprintf("Content (%d chars): %s...\n", len(item.Content), item.Content))
			}
			builder.WriteString("\n")
		}
		builder.WriteString(strings.Repeat("=", 50) + "\n")
	}
	return builder.String()
}

func main() {
	queryFlag := flag.String("query", "", "Single search query to run")
	bundleFlag := flag.String("bundle", "", "Path to a text file containing queries (one per line)")
	limitFlag := flag.Int("limit", 10, "Maximum search results to process per query")
	workersFlag := flag.Int("workers", 5, "Number of concurrent workers")
	contentFlag := flag.Bool("content", true, "Extract deep content from pages (if false, only returns snippets)")
	serveFlag := flag.Bool("serve", false, "Start an HTTP API server for AI Agents")
	portFlag := flag.String("port", "8080", "Port for the HTTP server")
	formatFlag := flag.String("output-format", "json", "Output format (json, llm-dense)")
	outputFlag := flag.String("output", "ultra_results.json", "Output JSON file path")
	flag.Parse()

	if *serveFlag {
		log.Printf("🚀 Starting UltraSearch API Server on :%s", *portFlag)
		http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query().Get("q")
			if query == "" {
				http.Error(w, "Missing 'q' parameter", http.StatusBadRequest)
				return
			}
			
			limit := 5
			if l := r.URL.Query().Get("limit"); l != "" {
				if parsed, err := strconv.Atoi(l); err == nil {
					limit = parsed
				}
			}
			
			content := true
			if c := r.URL.Query().Get("content"); c == "false" {
				content = false
			}

			log.Printf("📡 API Request: q='%s' limit=%d content=%v", query, limit, content)
			responses := runSearchPipeline([]string{query}, limit, *workersFlag, content)
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(responses[0])
		})
		log.Fatal(http.ListenAndServe(":"+*portFlag, nil))
	}

	var queries []string
	if *queryFlag != "" {
		queries = append(queries, *queryFlag)
	}

	if *bundleFlag != "" {
		file, err := os.Open(*bundleFlag)
		if err != nil {
			log.Fatalf("Could not open bundle file: %v", err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			q := strings.TrimSpace(scanner.Text())
			if q != "" {
				queries = append(queries, q)
			}
		}
	}

	if len(queries) == 0 {
		log.Println("⚠️ No queries provided. Use --query, --bundle, or --serve.")
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("🚀 Starting UltraSearch CLI with %d workers. Content: %v", *workersFlag, *contentFlag)
	
	responses := runSearchPipeline(queries, *limitFlag, *workersFlag, *contentFlag)

	// Save Output
	if *formatFlag == "llm-dense" {
		denseOutput := formatLLMDense(responses)
		_ = os.WriteFile(*outputFlag, []byte(denseOutput), 0644)
		log.Printf("💾 Saved LLM-dense results to %s", *outputFlag)
	} else {
		file, _ := json.MarshalIndent(responses, "", "  ")
		_ = os.WriteFile(*outputFlag, file, 0644)
		log.Printf("💾 Saved JSON results to %s", *outputFlag)
	}
}

func formatLLMDense(responses []SearchResponse) string {
	var sb strings.Builder
	for _, resp := range responses {
		sb.WriteString("<SEARCH q=\"" + resp.Query + "\">\n")
		if resp.Error != "" {
			sb.WriteString("<ERR>" + resp.Error + "</ERR>\n")
			continue
		}
		for _, r := range resp.Results {
			sb.WriteString(fmt.Sprintf("<RES rank=\"%d\" url=\"%s\">\n", r.Rank, r.URL))
			content := r.Content
			if content == "" {
				content = r.Snippet
			}
			// aggressively strip whitespace for tokens
			content = strings.Join(strings.Fields(content), " ")
			sb.WriteString(content + "\n</RES>\n")
		}
		sb.WriteString("</SEARCH>\n")
	}
	return sb.String()
}

func runSearchPipeline(queries []string, maxResults int, numWorkers int, fetchContent bool) []SearchResponse {
	startTotal := time.Now()
	
	// Search allocator: headless for Google only
	opts := []chromedp.ExecAllocatorOption{
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
	}

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	defer cancelBrowser()
	
	if err := chromedp.Run(browserCtx); err != nil {
		log.Fatalf("Failed to start browser: %v", err)
	}

	queriesChan := make(chan string, len(queries))
	resultsChan := make(chan SearchResponse, len(queries))

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(i, queriesChan, resultsChan, allocCtx, maxResults, fetchContent, &wg)
	}

	for _, q := range queries {
		queriesChan <- q
	}
	close(queriesChan)

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var responses []SearchResponse
	successCount := 0
	for resp := range resultsChan {
		responses = append(responses, resp)
		if resp.Error == "" && len(resp.Results) > 0 {
			successCount++
		}
	}
	
	elapsedSearch := time.Since(startTotal).Seconds()
	log.Printf("\n⚡ %d/%d queries in %.1fs", successCount, len(queries), elapsedSearch)

	// === FINAL RETRY PASS: Collect all failed URLs and retry with fresh stealth ===
	if fetchContent {
		type retryTarget struct {
			queryIdx  int
			resultIdx int
			url       string
		}
		var retryList []retryTarget
		for qi, resp := range responses {
			for ri, r := range resp.Results {
				if ri >= 5 { break } // Only retry top 5
				if r.Content == "" && r.Tier >= TierJSRender && r.URL != "" {
					retryList = append(retryList, retryTarget{qi, ri, r.URL})
				}
			}
		}

		if len(retryList) > 0 {
			log.Printf("\n🔄 RETRY PASS: %d failed URLs with fresh stealth session...", len(retryList))

			retryOpts := []chromedp.ExecAllocatorOption{
				chromedp.NoFirstRun,
				chromedp.NoDefaultBrowserCheck,
				chromedp.Flag("headless", false),
				chromedp.Flag("enable-automation", false),
				chromedp.Flag("disable-blink-features", "AutomationControlled"),
				chromedp.Flag("disable-infobars", true),
				chromedp.Flag("disable-gpu", true),
				chromedp.Flag("no-sandbox", true),
				chromedp.Flag("disable-extensions", false),
				chromedp.Flag("disable-features", "DownloadFonts,FontAccess"),
				chromedp.WindowSize(1440, 900),
				chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"),
			}

			retryAlloc, retryAllocCancel := chromedp.NewExecAllocator(context.Background(), retryOpts...)
			retryParent, retryParentCancel := chromedp.NewContext(retryAlloc)
			chromedp.Run(retryParent)

			recovered := 0
			for _, rt := range retryList {
				tabCtx, tabCancel := chromedp.NewContext(retryParent)
				tabCtx, tabTimeout := context.WithTimeout(tabCtx, 15*time.Second)

				var htmlDump string
				err := chromedp.Run(tabCtx,
					chromedp.ActionFunc(func(c context.Context) error {
						_, err := page.AddScriptToEvaluateOnNewDocument(solver.StealthScript).Do(c)
						return err
					}),
					chromedp.Navigate(rt.url),
					chromedp.Sleep(3*time.Second),
				)
				if err == nil {
					// Check for challenge
					var bodySnippet string
					chromedp.Run(tabCtx, chromedp.Evaluate(`document.body ? document.body.innerText.substring(0, 300).toLowerCase() : ''`, &bodySnippet))

					needsSolver := strings.Contains(bodySnippet, "verify you are human") ||
						strings.Contains(bodySnippet, "just a moment") ||
						strings.Contains(bodySnippet, "checking your browser") ||
						strings.Contains(bodySnippet, "performing security verification") ||
						strings.Contains(bodySnippet, "enable javascript and cookies") ||
						len(bodySnippet) < 30

					if needsSolver {
						solved, _ := solver.DefeatCaptcha(tabCtx, 200, 400)
						if solved {
							chromedp.Run(tabCtx, chromedp.Sleep(2*time.Second))
						}
					}

					chromedp.Run(tabCtx, chromedp.OuterHTML("html", &htmlDump))
					if len(htmlDump) > 500 {
						text := extractText(htmlDump)
						if ContentQuality(text) {
							responses[rt.queryIdx].Results[rt.resultIdx].Content = text
							recovered++
							urlPreview := rt.url; if len(urlPreview) > 60 { urlPreview = urlPreview[:60] }
							log.Printf("   🔄 Recovered: %s (%d chars)", urlPreview, len(text))
						}
					}
				}
				tabTimeout()
				tabCancel()
			}

			retryParentCancel()
			retryAllocCancel()
			log.Printf("🔄 Retry recovered %d/%d URLs", recovered, len(retryList))
		}
	}

	// === TIER STATISTICS ===
	tierNames := map[int]string{1: "HTTP", 2: "JS", 3: "STEALTH", 4: "LOGIN", 5: "SKIP"}
	tierTotal := map[int]int{}
	tierOK := map[int]int{}
	totalContent := 0
	for _, resp := range responses {
		for i, r := range resp.Results {
			if i >= 5 { break }
			tierTotal[r.Tier]++
			if r.Content != "" {
				tierOK[r.Tier]++
				totalContent++
			}
		}
	}
	log.Println("\n📊 TIER STATS:")
	for _, t := range []int{1, 2, 3, 4, 5} {
		if tierTotal[t] > 0 {
			pct := 100 * tierOK[t] / tierTotal[t]
			log.Printf("   T%d %-8s: %d/%d extracted (%d%%)", t, tierNames[t], tierOK[t], tierTotal[t], pct)
		}
	}
	
	elapsedTotal := time.Since(startTotal).Seconds()
	log.Printf("\n⚡ FINAL: %d/%d queries, %d URLs with content in %.1fs (%.1fs/query)",
		successCount, len(queries), totalContent, elapsedTotal, elapsedTotal/float64(len(queries)))

	return responses
}
