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
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/target"
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
	Content string `json:"content"`
	Tier    int    `json:"tier"` // 1=Static, 2=JSRender, 3=Stealth, 4=Login/Persistence
}

// TelemetryLog captures automated failure tracking during the testing week
type TelemetryLog struct {
	Timestamp   string `json:"timestamp"`
	Query       string `json:"query"`
	TargetURL   string `json:"target_url"`
	Tier        int    `json:"tier"`
	Status      string `json:"status"`
	ContentLen  int    `json:"content_length"`
}

// SearchResponse represents the results for a single query
type SearchResponse struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
	Error   string         `json:"error,omitempty"`
}

type SearchFilters struct {
	Language   string `json:"language"`    // hl, e.g. "en", "hi", "browser"
	Country    string `json:"country"`     // gl, e.g. "us", "in", "browser"
	Uule       string `json:"uule"`        // Location encoded parameter, e.g. "browser"
	SafeSearch string `json:"safe_search"` // safe, e.g. "active", "off", "browser"
	Tbs        string `json:"tbs"`         // Search tools, e.g. "qdr:d", "qdr:w", "browser"
}

type FilterProfileManager struct {
	filePath string
	profiles map[string]SearchFilters
	mu       sync.RWMutex
}

func NewFilterProfileManager(filePath string) *FilterProfileManager {
	mgr := &FilterProfileManager{
		filePath: filePath,
		profiles: make(map[string]SearchFilters),
	}
	// Initial presets
	mgr.profiles["browser"] = SearchFilters{
		Language:   "browser",
		Country:    "browser",
		Uule:       "browser",
		SafeSearch: "browser",
		Tbs:        "browser",
	}
	mgr.profiles["us_english"] = SearchFilters{
		Language:   "en",
		Country:    "us",
		SafeSearch: "off",
	}
	mgr.profiles["india_hindi"] = SearchFilters{
		Language:   "hi",
		Country:    "in",
		SafeSearch: "off",
	}
	mgr.profiles["uk_english"] = SearchFilters{
		Language:   "en",
		Country:    "gb",
		SafeSearch: "off",
	}
	mgr.profiles["safe_active"] = SearchFilters{
		SafeSearch: "active",
	}
	mgr.profiles["fresh_day"] = SearchFilters{
		Tbs: "qdr:d",
	}

	mgr.Load()
	return mgr
}

func (m *FilterProfileManager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var loaded map[string]SearchFilters
	if err := json.Unmarshal(data, &loaded); err != nil {
		return err
	}

	for name, filters := range loaded {
		m.profiles[name] = filters
	}
	return nil
}

func (m *FilterProfileManager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := json.MarshalIndent(m.profiles, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.filePath, data, 0644)
}

func (m *FilterProfileManager) Get(name string) (SearchFilters, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	f, ok := m.profiles[name]
	return f, ok
}

func (m *FilterProfileManager) Set(name string, filters SearchFilters) error {
	m.mu.Lock()
	m.profiles[name] = filters
	m.mu.Unlock()
	return m.Save()
}

func (m *FilterProfileManager) List() map[string]SearchFilters {
	m.mu.RLock()
	defer m.mu.RUnlock()

	copyMap := make(map[string]SearchFilters)
	for k, v := range m.profiles {
		copyMap[k] = v
	}
	return copyMap
}

var filterManager *FilterProfileManager

func detectSystemLanguage() string {
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("defaults", "read", "-g", "AppleLocale")
		if out, err := cmd.Output(); err == nil {
			loc := strings.TrimSpace(string(out))
			loc = strings.Split(loc, "@")[0]
			parts := strings.Split(loc, "_")
			if len(parts) > 0 && len(parts[0]) > 0 {
				return strings.ToLower(parts[0])
			}
		}
	}
	langEnv := os.Getenv("LANG")
	if langEnv == "" {
		langEnv = os.Getenv("LC_ALL")
	}
	if langEnv == "" {
		langEnv = os.Getenv("LC_MESSAGES")
	}
	if langEnv != "" && langEnv != "C" && langEnv != "POSIX" && langEnv != "C.UTF-8" {
		parts := strings.Split(langEnv, ".")
		if len(parts) > 0 {
			langCountry := parts[0]
			langParts := strings.Split(langCountry, "_")
			if len(langParts) > 0 && len(langParts[0]) > 0 {
				return strings.ToLower(langParts[0])
			}
		}
	}
	return "en"
}

func detectSystemCountry() string {
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("defaults", "read", "-g", "AppleLocale")
		if out, err := cmd.Output(); err == nil {
			loc := strings.TrimSpace(string(out))
			loc = strings.Split(loc, "@")[0]
			parts := strings.Split(loc, "_")
			if len(parts) > 1 && len(parts[1]) > 0 {
				return strings.ToLower(parts[1])
			}
		}
	}
	langEnv := os.Getenv("LANG")
	if langEnv == "" {
		langEnv = os.Getenv("LC_ALL")
	}
	if langEnv != "" && langEnv != "C" && langEnv != "POSIX" && langEnv != "C.UTF-8" {
		parts := strings.Split(langEnv, ".")
		if len(parts) > 0 {
			langCountry := parts[0]
			langParts := strings.Split(langCountry, "_")
			if len(langParts) > 1 && len(langParts[1]) > 0 {
				return strings.ToLower(langParts[1])
			}
		}
	}
	return "us"
}

func (m *FilterProfileManager) Resolve(filters SearchFilters) SearchFilters {
	res := filters
	if res.Language == "browser" || res.Language == "" {
		res.Language = detectSystemLanguage()
	}
	if res.Country == "browser" || res.Country == "" {
		res.Country = detectSystemCountry()
	}
	if res.Uule == "browser" {
		res.Uule = ""
	}
	if res.SafeSearch == "browser" {
		res.SafeSearch = "off"
	}
	if res.Tbs == "browser" {
		res.Tbs = ""
	}
	return res
}

func BuildSearchURL(q string, limit int, filters SearchFilters) string {
	base, err := url.Parse("https://www.google.com/search")
	if err != nil {
		return "https://www.google.com/search?q=" + url.QueryEscape(q)
	}
	params := url.Values{}
	params.Set("q", q)
	if limit > 0 {
		params.Set("num", fmt.Sprintf("%d", limit))
	}

	lang := filters.Language
	if lang == "browser" {
		lang = detectSystemLanguage()
	}
	if lang != "" {
		params.Set("hl", lang)
	}

	country := filters.Country
	if country == "browser" {
		country = detectSystemCountry()
	}
	if country != "" {
		params.Set("gl", country)
	}

	uule := filters.Uule
	if uule == "browser" {
		uule = ""
	}
	if uule != "" {
		params.Set("uule", uule)
	}

	safe := filters.SafeSearch
	if safe == "browser" {
		safe = "off"
	}
	if safe != "" {
		params.Set("safe", safe)
	}

	tbs := filters.Tbs
	if tbs == "browser" {
		tbs = ""
	}
	if tbs != "" {
		params.Set("tbs", tbs)
	}

	base.RawQuery = params.Encode()
	return base.String()
}

func getLanguagesForCode(lang string) []string {
	if lang == "" || lang == "browser" {
		lang = detectSystemLanguage()
	}
	switch lang {
	case "en":
		return []string{"en-US", "en"}
	case "hi":
		return []string{"hi-IN", "hi", "en-US", "en"}
	case "es":
		return []string{"es-ES", "es", "en-US", "en"}
	case "fr":
		return []string{"fr-FR", "fr", "en-US", "en"}
	case "de":
		return []string{"de-DE", "de", "en-US", "en"}
	case "ja":
		return []string{"ja-JP", "ja", "en-US", "en"}
	case "zh":
		return []string{"zh-CN", "zh", "en-US", "en"}
	default:
		return []string{lang, "en-US", "en"}
	}
}

func GetStealthScript(lang string) string {
	langs := getLanguagesForCode(lang)
	langStr := "['en-US', 'en']"
	if len(langs) > 0 {
		var parts []string
		for _, l := range langs {
			parts = append(parts, fmt.Sprintf("'%s'", l))
		}
		langStr = "[" + strings.Join(parts, ", ") + "]"
	}
	return strings.Replace(solver.StealthScript, "['en-US', 'en']", langStr, 1)
}

func GetAcceptLangOption(lang string) chromedp.ExecAllocatorOption {
	langs := getLanguagesForCode(lang)
	return chromedp.Flag("accept-lang", strings.Join(langs, ","))
}

func GetReplenishFilters() SearchFilters {
	if filterManager != nil {
		if f, ok := filterManager.Get("default"); ok {
			return filterManager.Resolve(f)
		}
		if f, ok := filterManager.Get("browser"); ok {
			return filterManager.Resolve(f)
		}
	}
	return filterManager.Resolve(SearchFilters{
		Language:   "browser",
		Country:    "browser",
		Uule:       "browser",
		SafeSearch: "browser",
		Tbs:        "browser",
	})
}

const extractJS = `(maxResults) => {
    const out = [];

    // Attempt to extract Google AI Overview (SGE) if present
    const aiContainer = document.querySelector('.s7d4ef');
    if (aiContainer) {
        // Clone container
        const clone = aiContainer.cloneNode(true);
        
        // Remove UI elements (buttons, svgs, styles, scripts, dialogs)
        // Keep hidden/carousel elements since SGE uses them for structured content
        const toRemove = clone.querySelectorAll('button, svg, style, script, [role="dialog"]');
        toRemove.forEach(el => el.remove());
        
        // Format code blocks dynamically
        const preBlocks = clone.querySelectorAll('pre');
        preBlocks.forEach(pre => {
            const codeText = pre.innerText;
            let lang = '';
            if (codeText.includes('package main') || codeText.includes('func main()') || codeText.includes('go ')) {
                lang = 'go';
            } else if (codeText.includes('fn main()') || codeText.includes('let mut') || codeText.includes('impl ')) {
                lang = 'rust';
            } else if (codeText.includes('def ') || (codeText.includes('import ') && codeText.includes(':\\n'))) {
                lang = 'python';
            } else if (codeText.includes('const ') || codeText.includes('let ') || codeText.includes('function ')) {
                lang = 'javascript';
            } else if (codeText.includes('<html>') || codeText.includes('class=') || codeText.includes('</div>')) {
                lang = 'html';
            } else if (codeText.includes('public class ') || codeText.includes('public static void main')) {
                lang = 'java';
            } else if (codeText.includes('#include <')) {
                lang = 'cpp';
            }
            
            const marker = document.createElement('div');
            marker.innerText = '\\n' + String.fromCharCode(96) + String.fromCharCode(96) + String.fromCharCode(96) + lang + '\\n' + codeText + '\\n' + String.fromCharCode(96) + String.fromCharCode(96) + String.fromCharCode(96) + '\\n';
            pre.parentNode.replaceChild(marker, pre);
        });
        
        // Format tables
        const tables = clone.querySelectorAll('table');
        tables.forEach(table => {
            let mdTable = '\\n';
            const rows = table.querySelectorAll('tr');
            rows.forEach((row, rowIndex) => {
                const cols = row.querySelectorAll('th, td');
                let mdRow = '|';
                cols.forEach(col => {
                    mdRow += ' ' + col.innerText.replace(/\\n/g, ' ').trim() + ' |';
                });
                mdTable += mdRow + '\\n';
                if (rowIndex === 0) {
                    let mdSep = '|';
                    cols.forEach(() => {
                        mdSep += ' --- |';
                    });
                    mdTable += mdSep + '\\n';
                }
            });
            mdTable += '\\n';
            const marker = document.createElement('div');
            marker.innerText = mdTable;
            table.parentNode.replaceChild(marker, table);
        });

        // Append to body for layout calculation (innerText requires layout)
        clone.style.position = 'absolute';
        clone.style.left = '-9999px';
        clone.style.top = '-9999px';
        document.body.appendChild(clone);

        let aiText = clone.innerText;
        clone.remove();
        
        aiText = aiText.replace(/\\n{3,}/g, '\\n\\n').trim();
        const lowerText = aiText.toLowerCase();
        const isErrorText = lowerText.includes("not available for this search") || 
                            lowerText.includes("can't generate") || 
                            lowerText.includes("try again later");
                            
        if (aiText.length > 30 && !isErrorText) {
            out.push({ 
                rank: 0, 
                title: "✨ Google AI Overview", 
                url: window.location.href, 
                snippet: aiText.substring(0, 5000)
            });
        }
    }

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
            
            snippet = snippet.replace(/\\n/g, ' ').trim();
            out.push({ rank: out.length + 1, title: h3.innerText, url: a.href, snippet: snippet.substring(0, 1000) });
        }
        if (out.length >= maxResults) break;
    }
    return out;
}`

var globalImmunizer *VortexImmunizer

func init() {
	err := solver.LoadTrajectories("solver/trajectories.json")
	if err != nil {
		log.Printf("⚠️ Could not load trajectories: %v", err)
	}
	filterManager = NewFilterProfileManager("filters.json")
	ReplenishCallback = func() {
		ReplenishSessionPool(5)
	}
	globalImmunizer = NewVortexImmunizer()
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

func worker(id int, queries <-chan string, results chan<- SearchResponse, searchBrowserCtx context.Context, maxResults int, fetchContent bool, aiMode string, showBrowser bool, headless bool, filters SearchFilters, wg *sync.WaitGroup) {
	defer wg.Done()
	httpClient := SharedHTTPClient()

	for q := range queries {
		start := time.Now()
		
		var res []SearchResult
		var err error

		// Attempt direct HTTP Search first ONLY if aiMode is "none" (since HTTP search cannot render client-side SGE)
		var runHTTP = (aiMode == "none")
		if runHTTP {
			res, err = runHTTPSearch(q, maxResults, filters)
		}
		
		if runHTTP && err == nil && len(res) > 0 {
			log.Printf("   🚀 W%d: '%s' -> Direct HTTP Search SUCCESS (Total = %v)", id, q, time.Since(start))
			
			// Filter out AI Overview (rank 0) since aiMode is "none"
			var filteredRes []SearchResult
			for _, r := range res {
				if r.Rank != 0 {
					filteredRes = append(filteredRes, r)
				}
			}
			res = filteredRes

			if !fetchContent {
				results <- SearchResponse{Query: q, Results: res}
				continue
			}
		} else {
			if aiMode != "none" {
				log.Printf("   🔍 W%d: '%s' -> Forcing browser search to fetch AI Overview (AI Mode: %s)", id, q, aiMode)
			} else {
				log.Printf("   ⚠️ W%d: Direct HTTP Search failed for '%s', falling back to browser search... (Error: %v)", id, q, err)
			}
			
			// --- PHASE 0: Google Search (Browser Fallback / Sniffer) ---
			tSetupStart := time.Now()
			// Use separate browser context for isolation
			ctx, cancel, tabErr := createIsolatedTab(searchBrowserCtx)
			if tabErr != nil {
				log.Printf("   ❌ W%d: Failed to spawn isolated browser tab: %v", id, tabErr)
				results <- SearchResponse{Query: q, Error: tabErr.Error()}
				continue
			}
			ctx, cancelTimeout := context.WithTimeout(ctx, 25*time.Second)
			
			var pageURL string
			searchURL := BuildSearchURL(q, maxResults+10, filters)

			var tSetup, tNavigate, tPoll, tEvaluate time.Duration

			var capturedHeaders map[string]string
			var captureMu sync.Mutex
			chromedp.ListenTarget(ctx, func(ev interface{}) {
				if ev, ok := ev.(*network.EventRequestWillBeSent); ok {
					if strings.Contains(ev.Request.URL, "google.com/search") && ev.Request.Method == "GET" {
						captureMu.Lock()
						if capturedHeaders == nil {
							capturedHeaders = make(map[string]string)
							for k, v := range ev.Request.Headers {
								if strVal, ok := v.(string); ok {
									if strings.ToLower(k) != "cookie" {
										capturedHeaders[k] = strVal
									}
								}
							}
						}
						captureMu.Unlock()
					}
				}
			})

			var cookies []*network.Cookie
			// 1. Initial setup
			err = chromedp.Run(ctx,
				chromedp.ActionFunc(func(ctx context.Context) error {
					// Setup network block and stealth script
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
					langs := getLanguagesForCode(filters.Language)
					err = network.SetExtraHTTPHeaders(network.Headers{
						"Accept-Language": strings.Join(langs, ","),
					}).Do(ctx)
					if err != nil {
						return err
					}
					_, err = page.AddScriptToEvaluateOnNewDocument(GetStealthScript(filters.Language)).Do(ctx)
					tSetup = time.Since(tSetupStart)
					return err
				}),
			)

			if err == nil {
				// 2. Navigate
				tNavigateStart := time.Now()
				err = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
					_, _, _, _, err := page.Navigate(searchURL).Do(ctx)
					return err
				}))
				tNavigate = time.Since(tNavigateStart)
			}

			if err == nil {
				// Check if browser got redirected to sorry page and solve CAPTCHA
				var currentLoc string
				_ = chromedp.Run(ctx, chromedp.Location(&currentLoc))
				if strings.Contains(strings.ToLower(currentLoc), "sorry") {
					log.Printf("   ⚠️ W%d: Browser fallback hit CAPTCHA, attempting to solve...", id)
					solved, solveErr := solver.DefeatCaptcha(ctx, 200, 400)
					if solveErr != nil {
						log.Printf("   ❌ W%d: CAPTCHA solver error: %v", id, solveErr)
					} else if solved {
						log.Printf("   ✅ W%d: CAPTCHA solved, waiting for Google redirect...", id)
						time.Sleep(2 * time.Second)
					}
				}
			}

			if err == nil {
				// 3. Poll with retry
				tPollStart := time.Now()
				var pollErr error
				for attempt := 1; attempt <= 5; attempt++ {
					pollErr = chromedp.Run(ctx, chromedp.Poll(`(() => {
						const aiContainer = document.querySelector('.s7d4ef');
						if (aiContainer) {
							const text = aiContainer.innerText.toLowerCase();
							if (text.includes("not available") || text.includes("can't generate") || text.includes("try again") || text.includes("check your connection")) {
								return true;
							}
							
							const currentLen = aiContainer.innerText.length;
							const now = Date.now();
							
							// If it hasn't started streaming actual content yet, keep waiting
							if (currentLen < 150) {
								return false;
							}
							
							if (window._sgePrevLen === undefined) {
								window._sgePrevLen = currentLen;
								window._sgeLastChangeTime = now;
								return false;
							}
							
							if (currentLen !== window._sgePrevLen) {
								window._sgePrevLen = currentLen;
								window._sgeLastChangeTime = now;
								return false;
							}
							
							if (now - window._sgeLastChangeTime > 1500) {
								return true;
							}
							return false;
						}
						const results = document.querySelectorAll('a h3');
						if (results.length > 0) {
							if (!window._firstResultSeenTime) {
								window._firstResultSeenTime = Date.now();
							}
							if (Date.now() - window._firstResultSeenTime > 400) {
								return true;
							}
						}
						return false;
					})()`, nil, chromedp.WithPollingInterval(150*time.Millisecond)))
					
					if pollErr == nil {
						break
					}
					if strings.Contains(pollErr.Error(), "navigated or closed") {
						log.Printf("   ⚠️ W%d: Poll attempt %d failed due to navigation/closed context, retrying...", id, attempt)
						time.Sleep(100 * time.Millisecond)
						continue
					}
					break
				}
				tPoll = time.Since(tPollStart)
				if pollErr != nil {
					err = pollErr
				}
			}

			if err == nil {
				// 4. Extract and retrieve cookies with retry
				var evalErr error
				for attempt := 1; attempt <= 3; attempt++ {
					tEvaluateStart := time.Now()
					evalErr = chromedp.Run(ctx,
						chromedp.ActionFunc(func(ctx context.Context) error {
							err := chromedp.Location(&pageURL).Do(ctx)
							if err != nil {
								return err
							}
							err = chromedp.Evaluate(fmt.Sprintf("(%s)(%d)", extractJS, maxResults), &res).Do(ctx)
							return err
						}),
						chromedp.ActionFunc(func(ctx context.Context) error {
							var err error
							cookies, err = network.GetCookies().WithURLs([]string{"https://www.google.com"}).Do(ctx)
							return err
						}),
					)
					tEvaluate = time.Since(tEvaluateStart)
					if evalErr == nil {
						break
					}
					if strings.Contains(evalErr.Error(), "navigated or closed") {
						log.Printf("   ⚠️ W%d: Evaluate attempt %d failed due to navigation/closed context, retrying...", id, attempt)
						time.Sleep(100 * time.Millisecond)
						continue
					}
					break
				}
				if evalErr != nil {
					err = evalErr
				}
			}

			log.Printf("⏱️ W%d Timings: Setup/Init Context = %v | page.Navigate = %v | DOM Polling = %v | JS Evaluate = %v",
				id, tSetup, tNavigate, tPoll, tEvaluate)

			cancelTimeout()
			cancel()

			if err == nil && len(capturedHeaders) > 0 && len(cookies) > 0 {
				saveSessionConfig(capturedHeaders, cookies)
			}

			if err != nil {
				log.Printf("   ❌ W%d: '%s' -> Error: %v", id, q, err)
				results <- SearchResponse{Query: q, Error: err.Error()}
				continue
			} else if strings.Contains(strings.ToLower(pageURL), "sorry") {
				log.Printf("   ⚠️ W%d: '%s' -> BLOCKED", id, q)
				results <- SearchResponse{Query: q, Error: "blocked_by_captcha"}
				continue
			}

			// Run Vortex Output Immunizer on any Google AI Overview (Rank == 0) SGE results
			for i, r := range res {
				if r.Rank == 0 {
					var sgeSources []string
					for _, organicRes := range res {
						if organicRes.Rank > 0 {
							sgeSources = append(sgeSources, organicRes.URL)
						}
					}
					
					log.Printf("🛡️ [Vortex] Sanitizing Google AI Overview output via Go Security Gateway...")
					startTime := time.Now()
					_, verdict := globalImmunizer.ProcessSGEResponse(q, r.Snippet, sgeSources, int(time.Since(startTime).Milliseconds()))
					log.Printf("🛡️ [Vortex] Sanitization complete. Verdict: %s", verdict)
					
					if verdict != "SAFE" && verdict != "BYPASSED_TRUSTED" && verdict != "NO_JSON_FOUND" && verdict != "PARSING_ERROR" {
						res[i].Snippet = fmt.Sprintf("⚠️ [Vortex Security Alert] Indirect Prompt Injection Attack Neutralized.\nVerdict: %s", verdict)
					}
				}
			}

			// Filter results based on aiMode
			var filteredRes []SearchResult
			for _, r := range res {
				if aiMode == "only" {
					if r.Rank == 0 {
						filteredRes = append(filteredRes, r)
					}
				} else if aiMode == "none" {
					if r.Rank != 0 {
						filteredRes = append(filteredRes, r)
					}
				} else { // "both"
					filteredRes = append(filteredRes, r)
				}
			}
			res = filteredRes

			if aiMode == "only" || !fetchContent {
				log.Printf("   ✅ W%d: '%s' -> %d results (content skipped/only-ai)", id, q, len(res))
				results <- SearchResponse{Query: q, Results: res}
				continue
			}
		}

		// Collect organic URLs for deep content extraction (maximum 5)
		var organicIdxs []int
		for i := 0; i < len(res); i++ {
			if res[i].Rank > 0 {
				organicIdxs = append(organicIdxs, i)
				if len(organicIdxs) >= 5 {
					break
				}
			}
		}

		if len(organicIdxs) == 0 {
			log.Printf("   ✅ W%d: '%s' -> %d results (no organic results for content extraction)", id, q, len(res))
			results <- SearchResponse{Query: q, Results: res}
			continue
		}

		// --- PHASE 1: CLASSIFY all URLs ---
		type classifiedURL struct {
			idx   int
			tier  int
			html  string // Populated for T1 (static)
		}

		classifyCh := make(chan classifiedURL, len(organicIdxs))
		
		for _, idx := range organicIdxs {
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
			}(idx)
		}

		// Collect classifications
		var t1Results []classifiedURL // Static HTML - already have content
		var t2Idxs []int             // JS-render needed
		var t3Idxs []int             // Bot-protected
		var t4Idxs []int             // Login-walled

		for i := 0; i < len(organicIdxs); i++ {
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
			id, len(organicIdxs), len(t1Results), len(t2Idxs), len(t3Idxs), len(t4Idxs))


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
				chromedp.Flag("mute-audio", true),
				chromedp.WindowSize(1440, 900),
				chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"),
			}
			jsOpts = append(jsOpts, GetAcceptLangOption(filters.Language))
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

					tabCtx, tabCancel, tabErr := createIsolatedTab(jsParent)
					if tabErr != nil {
						log.Printf("   ⚠️ W%d: Failed to spawn isolated tab for %s: %v", id, res[idx].URL, tabErr)
						t2Mu.Lock()
						t2Escalate = append(t2Escalate, idx)
						t2Mu.Unlock()
						return
					}
					tabCtx, tabTimeout := context.WithTimeout(tabCtx, 10*time.Second)

					var htmlDump string
					err := chromedp.Run(tabCtx,
						chromedp.ActionFunc(func(c context.Context) error {
							if err := network.Enable().Do(c); err != nil {
								return err
							}
							langs := getLanguagesForCode(filters.Language)
							if err := network.SetExtraHTTPHeaders(network.Headers{
								"Accept-Language": strings.Join(langs, ","),
							}).Do(c); err != nil {
								return err
							}
							_, err := page.AddScriptToEvaluateOnNewDocument(GetStealthScript(filters.Language)).Do(c)
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
				chromedp.Flag("headless", headless),
				chromedp.Flag("enable-automation", false),
				chromedp.Flag("disable-blink-features", "AutomationControlled"),
				chromedp.Flag("disable-infobars", true),
				chromedp.Flag("disable-gpu", true),
				chromedp.Flag("no-sandbox", true),
				chromedp.Flag("disable-extensions", false),
				chromedp.Flag("disable-features", "DownloadFonts,FontAccess"),
				chromedp.Flag("mute-audio", true),
				chromedp.WindowSize(1440, 900),
				chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"),
			}
			if !showBrowser && !headless {
				stealthOpts = append(stealthOpts, chromedp.Flag("window-position", "-2400,-2400"))
			}
			stealthOpts = append(stealthOpts, GetAcceptLangOption(filters.Language))
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
				
				parkCtx, parkCancel, tabErr := createIsolatedTab(stealthParent)
				if tabErr != nil {
					log.Printf("   ❌ W%d: Failed to spawn isolated park tab: %v", id, tabErr)
					continue
				}
				parkCtx, parkTimeout := context.WithTimeout(parkCtx, 20*time.Second)

				// Step 1: Park on root domain
				err := chromedp.Run(parkCtx,
					chromedp.ActionFunc(func(c context.Context) error {
						if err := network.Enable().Do(c); err != nil {
							return err
						}
						langs := getLanguagesForCode(filters.Language)
						if err := network.SetExtraHTTPHeaders(network.Headers{
							"Accept-Language": strings.Join(langs, ","),
						}).Do(c); err != nil {
							return err
						}
						_, err := page.AddScriptToEvaluateOnNewDocument(GetStealthScript(filters.Language)).Do(c)
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
						
						fallbackCtx, fallbackCancel, tabErr := createIsolatedTab(stealthParent)
						if tabErr != nil {
							log.Printf("   ❌ W%d: Failed to spawn isolated fallback tab: %v", id, tabErr)
							continue
						}
						fallbackCtx, fallbackTimeout := context.WithTimeout(fallbackCtx, 15*time.Second)
						
						err = chromedp.Run(fallbackCtx,
							chromedp.ActionFunc(func(c context.Context) error {
								if err := network.Enable().Do(c); err != nil {
									return err
								}
								langs := getLanguagesForCode(filters.Language)
								if err := network.SetExtraHTTPHeaders(network.Headers{
									"Accept-Language": strings.Join(langs, ","),
								}).Do(c); err != nil {
									return err
								}
								_, err = page.AddScriptToEvaluateOnNewDocument(GetStealthScript(filters.Language)).Do(c)
								return err
							}),
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
		for _, idx := range organicIdxs {
			if res[idx].Content != "" {
				contentCount++
			} else {
				failedURLs = append(failedURLs, fmt.Sprintf("T%d:%s", res[idx].Tier, res[idx].URL))
			}
		}

		if len(failedURLs) > 0 {
			log.Printf("   ⚠️ W%d: Failed extractions: %v", id, failedURLs)
		}

		log.Printf("   ✅ W%d: '%s' -> %d results, %d/%d content (%.1fs)",
			id, q, len(res), contentCount, len(organicIdxs), time.Since(start).Seconds())
		
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
	fastAIFlag := flag.Bool("fast-ai", false, "Fast AI Mode: Skips deep scraping and instantly returns the AI Overview and URLs")
	onlyAIFlag := flag.Bool("only-ai", false, "Only AI Overview Mode: Skips deep scraping and only returns the Google AI Overview if it exists")
	noAIFlag := flag.Bool("no-ai", false, "No AI Mode: Skips AI Overview and instantly returns the 10 URLs (Snippets only or deep content)")
	showBrowserFlag := flag.Bool("show-browser", false, "Show browser GUI visually on-screen during stealth operations (default: false)")
	headlessFlag := flag.Bool("headless", true, "Run stealth browsers in real headless mode (default: true)")
	serveFlag := flag.Bool("serve", false, "Start an HTTP API server for AI Agents")
	portFlag := flag.String("port", "8080", "Port for the HTTP server")
	formatFlag := flag.String("output-format", "json", "Output format (json, llm-dense)")
	outputFlag := flag.String("output", "ultra_results.json", "Output JSON file path")
	vortexDiagFlag := flag.Bool("vortex-diag", false, "Run Vortex Go Security and Telemetry Gateway diagnostics")
	
	stressFlag := flag.Bool("stress", false, "Run stress test suite")
	stressCountFlag := flag.Int("stress-count", 30, "Total number of queries to run in stress test")
	stressConcurrencyFlag := flag.Int("stress-concurrency", 2, "Concurrency level for stress test")
	stressDelayFlag := flag.Int("stress-delay", 500, "Delay in milliseconds between queries in each worker")
	stressSelfHealFlag := flag.Bool("stress-self-heal", true, "Trigger self-healing browser fallback when captcha blocked")

	profileNameFlag := flag.String("filter-profile", "browser", "Name of the profile to use")
	hlFlag := flag.String("hl", "", "Language filter override")
	glFlag := flag.String("gl", "", "Country/region override")
	uuleFlag := flag.String("uule", "", "Geolocation override")
	safeFlag := flag.String("safe", "", "SafeSearch override")
	tbsFlag := flag.String("tbs", "", "Search tools override")
	saveProfileFlag := flag.String("save-profile", "", "Save the resolved filters under this profile name")
	listProfilesFlag := flag.Bool("list-profiles", false, "List all saved profiles and exit")
	
	usqlFlag := flag.String("usql", "", "Execute a structured USQL statement")
	semanticRouteFlag := flag.String("semantic-route", "", "Resolve a query to matching Skill Books")
	integrateSkillFlag := flag.String("integrate-skill", "", "Integrate a community Skill Book template (Sandbox intake)")
	promoteSkillFlag := flag.String("promote-skill", "", "Promote a staged Skill Book to the active engine (Human-in-the-loop verification)")
	listStagedFlag := flag.Bool("list-staged", false, "List all staged Skill Books awaiting human promotion")
	installFlag := flag.String("install", "", "Install a community Skill Book template from a GitHub URL")
	
	flag.Parse()

	if *vortexDiagFlag {
		RunVortexDiagnostics()
		RunUSQLDiagnostics()
		RunRegistryDiagnostics()
		RunContributionDiagnostics()
		return
	}

	if *integrateSkillFlag != "" {
		cg := NewContributionGateway("ai_skills", filepath.Join("ai_skills", "unverified"))
		status, err := cg.IntakeSkillBook(*integrateSkillFlag)
		if err != nil {
			log.Fatalf("❌ Intake failed: %v", err)
		}
		fmt.Printf("🎉 Staged cleanly! Contribution status: %s. Staged in ai_skills/unverified/ awaiting human review.\n", status)
		return
	}

	if *promoteSkillFlag != "" {
		cg := NewContributionGateway("ai_skills", filepath.Join("ai_skills", "unverified"))
		err := cg.PromoteSkillBook(*promoteSkillFlag)
		if err != nil {
			log.Fatalf("❌ Promotion failed: %v", err)
		}
		fmt.Println("🎉 Human review verified successfully. Skill Book promoted to Active catalog!")
		return
	}

	if *listStagedFlag {
		cg := NewContributionGateway("ai_skills", filepath.Join("ai_skills", "unverified"))
		list, err := cg.ListStagedSkillBooks()
		if err != nil {
			log.Fatalf("❌ Failed to list staged Skill Books: %v", err)
		}
		fmt.Println("📬 Staged Skill Books awaiting human review:")
		for _, f := range list {
			fmt.Printf("  - %s\n", f)
		}
		return
	}

	if *installFlag != "" {
		cg := NewContributionGateway("ai_skills", filepath.Join("ai_skills", "unverified"))
		staged, err := cg.DownloadCommunitySkill(*installFlag)
		if err != nil {
			log.Fatalf("❌ Installation failed: %v", err)
		}
		fmt.Printf("🎉 Successfully installed community templates into staging area:\n")
		for _, f := range staged {
			fmt.Printf("  - Staged: %s (ai_skills/unverified/%s)\n", f, f)
		}
		fmt.Println("📬 Awaiting human review. Promote them to the active catalog via: ./ultrasearch -promote-skill <filename>")
		return
	}

	if *semanticRouteFlag != "" {
		_ = LoadSkillBookRegistry("ai_skills")
		book, score, found := SemanticRouteQuery(*semanticRouteFlag)
		if found {
			fmt.Printf("🎯 Best-fit Skill Book: %s (Cosine Correlation: %.4f)\n", book.Name, score)
			fmt.Printf("   Author:             %s\n", book.Author)
			fmt.Printf("   Version:            %s\n", book.Version)
			fmt.Printf("   Active Domains:     %v\n", book.Domains)
		} else {
			fmt.Printf("⚠️ No matching Skill Book found. Highest correlation: %.4f (below threshold 0.15)\n", score)
		}
		return
	}

	if *usqlFlag != "" {
		query, err := ParseHybridQuery(*usqlFlag)
		if err != nil {
			LogQueryFailure(*usqlFlag, "", "HYBRID_PARSE_ERROR", err.Error(), SearchFilters{})
			log.Fatalf("❌ Hybrid Parse Error: %v", err)
		}

		// Initialize global registry and attempt semantic routing if FROM is missing in query
		if len(query.Sources) == 0 {
			_ = LoadSkillBookRegistry("ai_skills")
			if book, _, found := SemanticRouteQuery(query.SearchEntity); found {
				log.Printf("🎯 [USQL Engine] Auto-routed query to Skill Book: %s", book.Name)
				// Bind Skill Book sources dynamically
				query.Sources = book.Domains
			}
		}

		dorkQuery := query.CompileToDorkQuery()
		log.Printf("🤖 [USQL Compiler] AST compiled into search dork: %q", dorkQuery)

		// Resolve search filters
		resolvedFilters, found := filterManager.Get(*profileNameFlag)
		if !found {
			resolvedFilters = SearchFilters{
				Language:   "browser",
				Country:    "browser",
				Uule:       "browser",
				SafeSearch: "browser",
				Tbs:        "browser",
			}
		}
		if *hlFlag != "" {
			resolvedFilters.Language = *hlFlag
		}
		if *glFlag != "" {
			resolvedFilters.Country = *glFlag
		}
		if *uuleFlag != "" {
			resolvedFilters.Uule = *uuleFlag
		}
		if *safeFlag != "" {
			resolvedFilters.SafeSearch = *safeFlag
		}
		if *tbsFlag != "" {
			resolvedFilters.Tbs = *tbsFlag
		}

		ctx := GetGlobalBrowserCtx()
		results := runSearchPipeline(ctx, []string{dorkQuery}, *limitFlag, *workersFlag, *contentFlag, "only", *showBrowserFlag, *headlessFlag, resolvedFilters)

		log.Printf("📊 [USQL Engine] Scraped %d raw search answers. Compiling schemas...", len(results))

		type USQLResponse struct {
			Query        string                 `json:"usql_query"`
			Entity       string                 `json:"search_entity"`
			TargetSchema map[string]interface{} `json:"target_schema"`
			Data         map[string]interface{} `json:"data"`
			Error        string                 `json:"error,omitempty"`
		}

		var finalPayload USQLResponse
		finalPayload.Query = *usqlFlag
		finalPayload.Entity = query.SearchEntity
		finalPayload.TargetSchema = query.ReturnFields

		foundData := false

		for _, r := range results {
			for _, resItem := range r.Results {
				if resItem.Rank == 0 {
					var rawMap map[string]interface{}
					snippet := resItem.Snippet
					
					// Scan snippet for refusal templates
					lowerSnippet := strings.ToLower(snippet)
					if strings.Contains(lowerSnippet, "not available for this search") ||
						strings.Contains(lowerSnippet, "can't generate") ||
						strings.Contains(lowerSnippet, "try again later") ||
						strings.Contains(lowerSnippet, "i cannot fulfill") ||
						strings.Contains(lowerSnippet, "i cannot provide") {
						LogQueryFailure(*usqlFlag, dorkQuery, "SGE_REFUSAL", snippet, resolvedFilters)
						finalPayload.Error = "Google AI Overview refused to generate: " + snippet
						foundData = true
						break
					}

					jsonStart := strings.Index(snippet, "{")
					jsonEnd := strings.LastIndex(snippet, "}") + 1
					if jsonStart != -1 && jsonEnd > jsonStart {
						_ = json.Unmarshal([]byte(snippet[jsonStart:jsonEnd]), &rawMap)
					}

					if rawMap != nil {
						filteredData := make(map[string]interface{})
						for key := range query.ReturnFields {
							foundVal := false
							for k, v := range rawMap {
								if strings.EqualFold(k, key) {
									filteredData[key] = v
									foundVal = true
									break
								}
							}
							if !foundVal {
								filteredData[key] = nil
							}
						}
						finalPayload.Data = EvaluateUSQLFunctions(query.ReturnFields, filteredData)
						foundData = true
					} else {
						LogQueryFailure(*usqlFlag, dorkQuery, "SGE_SCHEMALESS_RESPONSE", snippet, resolvedFilters)
						finalPayload.Data = ExtractSchemaFromText(snippet, query.ReturnFields)
						foundData = true
					}
				}
			}
		}

		if !foundData {
			LogQueryFailure(*usqlFlag, dorkQuery, "NO_SEARCH_RESULTS", "Zero search responses returned from browser context", resolvedFilters)
			finalPayload.Error = "no structured SGE overview resolved"
		}

		outBytes, _ := json.MarshalIndent(finalPayload, "", "  ")
		fmt.Println("\n=== USQL QUERY RESULT ===")
		fmt.Println(string(outBytes))
		fmt.Println("=========================")
		return
	}

	if *listProfilesFlag {
		profiles := filterManager.List()
		data, _ := json.MarshalIndent(profiles, "", "  ")
		fmt.Println(string(data))
		return
	}

	resolvedFilters, found := filterManager.Get(*profileNameFlag)
	if !found {
		resolvedFilters = SearchFilters{
			Language:   "browser",
			Country:    "browser",
			Uule:       "browser",
			SafeSearch: "browser",
			Tbs:        "browser",
		}
	}
	if *hlFlag != "" {
		resolvedFilters.Language = *hlFlag
	}
	if *glFlag != "" {
		resolvedFilters.Country = *glFlag
	}
	if *uuleFlag != "" {
		resolvedFilters.Uule = *uuleFlag
	}
	if *safeFlag != "" {
		resolvedFilters.SafeSearch = *safeFlag
	}
	if *tbsFlag != "" {
		resolvedFilters.Tbs = *tbsFlag
	}

	if *saveProfileFlag != "" {
		if err := filterManager.Set(*saveProfileFlag, resolvedFilters); err != nil {
			log.Fatalf("Failed to save profile %q: %v", *saveProfileFlag, err)
		}
		log.Printf("💾 Saved profile %q to filters.json", *saveProfileFlag)
		if *queryFlag == "" && *bundleFlag == "" && !*serveFlag && !*stressFlag {
			return
		}
	}

	resolvedFilters = filterManager.Resolve(resolvedFilters)

	// Proactively pre-fill the session pool at startup if we have fewer than 5 active sessions
	if poolManager.ActiveCount() < 5 {
		log.Println("🔑 [Session Pool] Startup pool count low. Pre-filling session pool...")
		ReplenishSessionPool(5)
	}

	if *stressFlag {
		runStressTest(*stressCountFlag, *stressConcurrencyFlag, *stressDelayFlag, *stressSelfHealFlag, *limitFlag, *showBrowserFlag, *headlessFlag, resolvedFilters)
		return
	}

	if *serveFlag {
		_ = LoadSkillBookRegistry("ai_skills")
		StartRegistryWatcher("ai_skills", 2*time.Second)
		log.Printf("🚀 Starting UltraSearch API Server on :%s", *portFlag)
		
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
			chromedp.Flag("blink-settings", "imagesEnabled=false"),
			chromedp.Flag("mute-audio", true),
			chromedp.WindowSize(1440, 900),
			chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"),
		}
		opts = append(opts, GetAcceptLangOption(resolvedFilters.Language))

		allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
		defer cancelAlloc()
		
		browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
		defer cancelBrowser()
		
		if err := chromedp.Run(browserCtx); err != nil {
			log.Fatalf("Failed to start persistent browser: %v", err)
		}

		http.HandleFunc("/profiles", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if r.Method == http.MethodGet {
				json.NewEncoder(w).Encode(filterManager.List())
				return
			}
			if r.Method == http.MethodPost || r.Method == http.MethodPut {
				name := r.URL.Query().Get("name")
				if name == "" {
					http.Error(w, "Missing 'name' parameter", http.StatusBadRequest)
					return
				}

				var f SearchFilters
				if r.Header.Get("Content-Type") == "application/json" {
					if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
						http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
						return
					}
				} else {
					f = SearchFilters{
						Language:   r.URL.Query().Get("hl"),
						Country:    r.URL.Query().Get("gl"),
						Uule:       r.URL.Query().Get("uule"),
						SafeSearch: r.URL.Query().Get("safe"),
						Tbs:        r.URL.Query().Get("tbs"),
					}
					if f.Language == "" && f.Country == "" && f.Uule == "" && f.SafeSearch == "" && f.Tbs == "" {
						f = SearchFilters{
							Language:   "browser",
							Country:    "browser",
							Uule:       "browser",
							SafeSearch: "browser",
							Tbs:        "browser",
						}
					}
				}

				if err := filterManager.Set(name, f); err != nil {
					http.Error(w, fmt.Sprintf("Failed to save profile: %v", err), http.StatusInternalServerError)
					return
				}
				log.Printf("💾 Saved profile %q via /profiles API", name)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status":  "success",
					"profile": name,
					"filters": f,
				})
				return
			}
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		})

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
			
			aiMode := "none"
			if *fastAIFlag {
				aiMode = "both"
			}
			if *onlyAIFlag {
				aiMode = "only"
			}
			if *noAIFlag {
				aiMode = "none"
			}

			// Override by query params if provided
			if m := r.URL.Query().Get("ai_mode"); m != "" {
				if m == "only" || m == "none" || m == "both" {
					aiMode = m
				}
			} else {
				// Legacy / alternate parameters
				if r.URL.Query().Get("only_ai") == "true" || r.URL.Query().Get("only-ai") == "true" {
					aiMode = "only"
				} else if r.URL.Query().Get("fast_ai") == "true" || r.URL.Query().Get("fast-ai") == "true" ||
					r.URL.Query().Get("ai_overview") == "true" || r.URL.Query().Get("ai-overview") == "true" {
					aiMode = "both"
				} else if r.URL.Query().Get("no_ai") == "true" || r.URL.Query().Get("no-ai") == "true" {
					aiMode = "none"
				}
			}

			content := *contentFlag
			// Default content to false if AI mode is only or both, unless explicitly overridden
			if aiMode == "only" || aiMode == "both" {
				content = false
			}
			if c := r.URL.Query().Get("content"); c != "" {
				content = (c == "true")
			}

			// Resolve filters for this request
			reqProfile := r.URL.Query().Get("profile")
			if reqProfile == "" {
				reqProfile = "browser"
			}
			reqFilters, found := filterManager.Get(reqProfile)
			if !found {
				reqFilters = SearchFilters{
					Language:   "browser",
					Country:    "browser",
					Uule:       "browser",
					SafeSearch: "browser",
					Tbs:        "browser",
				}
			}

			// Individual overrides
			if hl := r.URL.Query().Get("hl"); hl != "" {
				reqFilters.Language = hl
			}
			if gl := r.URL.Query().Get("gl"); gl != "" {
				reqFilters.Country = gl
			}
			if uule := r.URL.Query().Get("uule"); uule != "" {
				reqFilters.Uule = uule
			}
			if safe := r.URL.Query().Get("safe"); safe != "" {
				reqFilters.SafeSearch = safe
			}
			if tbs := r.URL.Query().Get("tbs"); tbs != "" {
				reqFilters.Tbs = tbs
			}

			// Save profile via API if requested
			if saveProfName := r.URL.Query().Get("save_profile"); saveProfName != "" {
				if err := filterManager.Set(saveProfName, reqFilters); err != nil {
					log.Printf("⚠️ Failed to save profile %q via API: %v", saveProfName, err)
				} else {
					log.Printf("💾 Saved profile %q via API request", saveProfName)
				}
			}

			log.Printf("📡 API Request: q='%s' limit=%d content=%v aiMode=%s showBrowser=%v headless=%v filters=%+v", query, limit, content, aiMode, *showBrowserFlag, *headlessFlag, reqFilters)
			responses := runSearchPipeline(browserCtx, []string{query}, limit, *workersFlag, content, aiMode, *showBrowserFlag, *headlessFlag, reqFilters)
			
			w.Header().Set("Content-Type", "application/json")
			if len(responses) > 0 {
				json.NewEncoder(w).Encode(responses[0])
			} else {
				json.NewEncoder(w).Encode(SearchResponse{Query: query, Error: "no results"})
			}
		})

		http.HandleFunc("/usql", func(w http.ResponseWriter, r *http.Request) {
			queryStr := r.URL.Query().Get("q")
			if queryStr == "" {
				http.Error(w, "Missing 'q' parameter containing USQL query", http.StatusBadRequest)
				return
			}

			query, err := ParseHybridQuery(queryStr)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				LogQueryFailure(queryStr, "", "HYBRID_PARSE_ERROR", err.Error(), SearchFilters{})
				json.NewEncoder(w).Encode(map[string]string{"error": "USQL Hybrid Parse Error: " + err.Error()})
				return
			}

			if len(query.Sources) == 0 {
				if book, _, found := SemanticRouteQuery(query.SearchEntity); found {
					query.Sources = book.Domains
				}
			}

			dorkQuery := query.CompileToDorkQuery()
			reqFilters := SearchFilters{
				Language:   "browser",
				Country:    "browser",
				Uule:       "browser",
				SafeSearch: "browser",
				Tbs:        "browser",
			}
			if query.Language != "" {
				reqFilters.Language = query.Language
			}
			if query.Country != "" {
				reqFilters.Country = query.Country
			}
			if query.SafeSearch != "" {
				reqFilters.SafeSearch = query.SafeSearch
			}

			responses := runSearchPipeline(browserCtx, []string{dorkQuery}, 5, *workersFlag, false, "only", *showBrowserFlag, *headlessFlag, reqFilters)

			type USQLResponse struct {
				Query        string                 `json:"usql_query"`
				Entity       string                 `json:"search_entity"`
				TargetSchema map[string]interface{} `json:"target_schema"`
				Data         map[string]interface{} `json:"data"`
				Error        string                 `json:"error,omitempty"`
			}

			var finalPayload USQLResponse
			finalPayload.Query = queryStr
			finalPayload.Entity = query.SearchEntity
			finalPayload.TargetSchema = query.ReturnFields

			foundData := false
			for _, r := range responses {
				for _, resItem := range r.Results {
					if resItem.Rank == 0 {
						var rawMap map[string]interface{}
						snippet := resItem.Snippet

						// Scan for SGE Refusals
						lowerSnippet := strings.ToLower(snippet)
						if strings.Contains(lowerSnippet, "not available for this search") ||
							strings.Contains(lowerSnippet, "can't generate") ||
							strings.Contains(lowerSnippet, "try again later") ||
							strings.Contains(lowerSnippet, "i cannot fulfill") ||
							strings.Contains(lowerSnippet, "i cannot provide") {
							LogQueryFailure(queryStr, dorkQuery, "SGE_REFUSAL", snippet, reqFilters)
							finalPayload.Error = "Google AI Overview refused to generate: " + snippet
							foundData = true
							break
						}

						jsonStart := strings.Index(snippet, "{")
						jsonEnd := strings.LastIndex(snippet, "}") + 1
						if jsonStart != -1 && jsonEnd > jsonStart {
							_ = json.Unmarshal([]byte(snippet[jsonStart:jsonEnd]), &rawMap)
						}

						if rawMap != nil {
							filteredData := make(map[string]interface{})
							for key := range query.ReturnFields {
								foundVal := false
								for k, v := range rawMap {
									if strings.EqualFold(k, key) {
										filteredData[key] = v
										foundVal = true
										break
									}
								}
								if !foundVal {
									filteredData[key] = nil
								}
							}
							finalPayload.Data = EvaluateUSQLFunctions(query.ReturnFields, filteredData)
							foundData = true
						} else {
							LogQueryFailure(queryStr, dorkQuery, "SGE_SCHEMALESS_RESPONSE", snippet, reqFilters)
							finalPayload.Data = ExtractSchemaFromText(snippet, query.ReturnFields)
							foundData = true
						}
					}
				}
			}

			if !foundData {
				LogQueryFailure(queryStr, dorkQuery, "NO_SEARCH_RESULTS", "Zero search responses returned from browser context", reqFilters)
				finalPayload.Error = "no structured SGE overview resolved"
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(finalPayload)
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

	aiMode := "none"
	if *fastAIFlag {
		aiMode = "both"
	}
	if *onlyAIFlag {
		aiMode = "only"
	}
	if *noAIFlag {
		aiMode = "none"
	}

	fetchContent := *contentFlag
	if *fastAIFlag || aiMode == "only" {
		fetchContent = false
	}

	log.Printf("🚀 Starting UltraSearch CLI with %d workers. Content: %v (FastAI: %v, AI Mode: %s)", *workersFlag, fetchContent, *fastAIFlag, aiMode)
	responses := runSearchPipeline(nil, queries, *limitFlag, *workersFlag, fetchContent, aiMode, *showBrowserFlag, *headlessFlag, resolvedFilters)

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
func runSearchPipeline(sharedBrowserCtx context.Context, queries []string, maxResults int, numWorkers int, fetchContent bool, aiMode string, showBrowser bool, headless bool, filters SearchFilters) []SearchResponse {
	filters = filterManager.Resolve(filters)
	startTotal := time.Now()

	var browserCtx context.Context
	var cancelBrowser context.CancelFunc
	var allocCtx context.Context
	var cancelAlloc context.CancelFunc

	if sharedBrowserCtx != nil {
		browserCtx = sharedBrowserCtx
	} else {
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
			chromedp.Flag("blink-settings", "imagesEnabled=false"), // natively block images
			chromedp.Flag("mute-audio", true),
			chromedp.WindowSize(1440, 900),
			chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"),
		}
		opts = append(opts, GetAcceptLangOption(filters.Language))

		allocCtx, cancelAlloc = chromedp.NewExecAllocator(context.Background(), opts...)
		browserCtx, cancelBrowser = chromedp.NewContext(allocCtx)
		if err := chromedp.Run(browserCtx); err != nil {
			log.Fatalf("Failed to start browser: %v", err)
		}
	}
	if cancelBrowser != nil {
		defer cancelBrowser()
	}
	if cancelAlloc != nil {
		defer cancelAlloc()
	}

	queriesChan := make(chan string, len(queries))
	resultsChan := make(chan SearchResponse, len(queries))

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(i, queriesChan, resultsChan, browserCtx, maxResults, fetchContent, aiMode, showBrowser, headless, filters, &wg)
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
			organicCount := 0
			for ri, r := range resp.Results {
				if r.Rank == 0 {
					continue
				}
				organicCount++
				if organicCount > 5 {
					break
				}
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
				chromedp.Flag("headless", headless),
				chromedp.Flag("enable-automation", false),
				chromedp.Flag("disable-blink-features", "AutomationControlled"),
				chromedp.Flag("disable-infobars", true),
				chromedp.Flag("disable-gpu", true),
				chromedp.Flag("no-sandbox", true),
				chromedp.Flag("disable-extensions", false),
				chromedp.Flag("disable-features", "DownloadFonts,FontAccess"),
				chromedp.Flag("mute-audio", true),
				chromedp.WindowSize(1440, 900),
				chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"),
			}
			if !showBrowser && !headless {
				retryOpts = append(retryOpts, chromedp.Flag("window-position", "-2400,-2400"))
			}
			retryOpts = append(retryOpts, GetAcceptLangOption(filters.Language))
			retryAlloc, retryAllocCancel := chromedp.NewExecAllocator(context.Background(), retryOpts...)
			retryParent, retryParentCancel := chromedp.NewContext(retryAlloc)
			chromedp.Run(retryParent)

			recovered := 0
			for _, rt := range retryList {
				tabCtx, tabCancel, tabErr := createIsolatedTab(retryParent)
				if tabErr != nil {
					log.Printf("   ❌ Failed to spawn isolated retry tab: %v", tabErr)
					continue
				}
				tabCtx, tabTimeout := context.WithTimeout(tabCtx, 15*time.Second)

				var htmlDump string
				err := chromedp.Run(tabCtx,
					chromedp.ActionFunc(func(c context.Context) error {
						if err := network.Enable().Do(c); err != nil {
							return err
						}
						langs := getLanguagesForCode(filters.Language)
						if err := network.SetExtraHTTPHeaders(network.Headers{
							"Accept-Language": strings.Join(langs, ","),
						}).Do(c); err != nil {
							return err
						}
						_, err := page.AddScriptToEvaluateOnNewDocument(GetStealthScript(filters.Language)).Do(c)
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
		organicCount := 0
		for _, r := range resp.Results {
			if r.Rank == 0 {
				continue
			}
			organicCount++
			if organicCount > 5 {
				break
			}
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

	// Automatically record telemetry for usage week analysis
	writeTelemetryLog(responses)

	return responses
}

func writeTelemetryLog(responses []SearchResponse) {
	file, err := os.OpenFile("usage_telemetry.jsonl", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	
	for _, resp := range responses {
		for i, r := range resp.Results {
			if i >= 5 { break }
			
			status := "SUCCESS"
			if r.Content == "" && r.Tier >= 2 {
				status = "FAILED"
			}
			
			logEntry := TelemetryLog{
				Timestamp:  time.Now().Format(time.RFC3339),
				Query:      resp.Query,
				TargetURL:  r.URL,
				Tier:       r.Tier,
				Status:     status,
				ContentLen: len(r.Content),
			}
			json.NewEncoder(file).Encode(logEntry)
		}
	}
}

func runStressTest(count int, concurrency int, delayMs int, selfHeal bool, limit int, showBrowser bool, headless bool, filters SearchFilters) {
	log.Printf("🔥 STARTING STRESS TEST: count=%d, concurrency=%d, delay=%dms, selfHeal=%v", count, concurrency, delayMs, selfHeal)

	// Build a pool of different queries to rotate through
	baseQueries := []string{
		"what is quantum computing",
		"how does photosynthesis work",
		"history of the internet",
		"why is the sky blue",
		"largest ocean in the world",
		"how to learn go programming",
		"speed of light in mph",
		"capital of australia",
		"what is the theory of relativity",
		"how do airplanes fly",
		"definition of artificial intelligence",
		"who wrote hamlet",
		"origin of the word algorithm",
		"distance from earth to moon",
		"average height of mount everest",
		"benefits of regular exercise",
		"symptoms of common cold",
		"how to cook pasta",
		"what is the water cycle",
		"invention of the telephone",
		"how many countries in europe",
		"what is photosynthesis",
		"structure of an atom",
		"how is cheese made",
		"deepest lake in the world",
		"who discovered gravity",
		"what is the speed of sound",
		"tallest building in the world",
		"how does a battery work",
		"what is black hole",
	}

	queries := make([]string, count)
	for i := 0; i < count; i++ {
		queries[i] = baseQueries[i%len(baseQueries)]
	}

	// Stats tracking
	var successCount int64
	var captchaCount int64
	var errorCount int64
	var healTriggerCount int64
	var totalDuration int64 // in ms
	var statsMu sync.Mutex

	type RequestResult struct {
		Index    int
		Query    string
		Status   string // "SUCCESS", "CAPTCHA_BLOCKED", "ERROR"
		Duration time.Duration
		Healed   bool
	}
	var results []RequestResult
	var resultsMu sync.Mutex

	// We need a browser context for the fallback/self-healing.
	var browserCtx context.Context
	var cancelBrowser context.CancelFunc
	var allocCtx context.Context
	var cancelAlloc context.CancelFunc

	if selfHeal {
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
			chromedp.Flag("blink-settings", "imagesEnabled=false"),
			chromedp.Flag("mute-audio", true),
			chromedp.WindowSize(1440, 900),
			chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"),
		}
		opts = append(opts, GetAcceptLangOption(filters.Language))
		allocCtx, cancelAlloc = chromedp.NewExecAllocator(context.Background(), opts...)
		browserCtx, cancelBrowser = chromedp.NewContext(allocCtx)
		// Start browser
		if err := chromedp.Run(browserCtx); err != nil {
			log.Fatalf("Failed to start browser for self-healing: %v", err)
		}
		defer cancelBrowser()
		defer cancelAlloc()
	}

	queriesChan := make(chan int, count)
	for i := 0; i < count; i++ {
		queriesChan <- i
	}
	close(queriesChan)

	var wg sync.WaitGroup
	var healMu sync.Mutex
	var lastHealTime time.Time

	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for idx := range queriesChan {
				q := queries[idx]
				start := time.Now()
				
				log.Printf("🔄 [W%d] Req %d/%d: Querying '%s' via direct HTTP...", workerID, idx+1, count, q)
				
				// Try direct HTTP Search
				httpRes, err := runHTTPSearch(q, limit, filters)
				
				status := "SUCCESS"
				healed := false
				
				if err != nil {
					log.Printf("⚠️ [W%d] Req %d/%d: HTTP Search failed: %v", workerID, idx+1, count, err)
					
					isCaptcha := strings.Contains(err.Error(), "blocked by captcha") || strings.Contains(err.Error(), "status 429")
					
					if isCaptcha && selfHeal {
						status = "CAPTCHA_BLOCKED"
						
						// Trigger self-healing: fallback to browser search
						healMu.Lock()
						// Check if another worker just healed it recently (within 5 seconds) to avoid redundant browser launches
						if time.Since(lastHealTime) > 5*time.Second {
							log.Printf("🛡️ [W%d] CAPTCHA detected. Triggering browser fallback to self-heal...", workerID)
							statsMu.Lock()
							healTriggerCount++
							statsMu.Unlock()
							
							// Run search through browser fallback (Phase 0) to capture fresh headers/cookies
							ctx, cancel, tabErr := createIsolatedTab(browserCtx)
							if tabErr != nil {
								log.Printf("🛡️ [W%d] Self-heal: Failed to spawn isolated tab: %v", workerID, tabErr)
								healMu.Unlock()
								continue
							}
							ctx, cancelTimeout := context.WithTimeout(ctx, 25*time.Second)
							
							searchURL := BuildSearchURL(q, limit+10, filters)
							
							var pageURL string
							var res []SearchResult
							var capturedHeaders map[string]string
							var captureMu sync.Mutex
							var cookies []*network.Cookie
							
							chromedp.ListenTarget(ctx, func(ev interface{}) {
								if ev, ok := ev.(*network.EventRequestWillBeSent); ok {
									if strings.Contains(ev.Request.URL, "google.com/search") && ev.Request.Method == "GET" {
										captureMu.Lock()
										if capturedHeaders == nil {
											capturedHeaders = make(map[string]string)
											for k, v := range ev.Request.Headers {
												if strVal, ok := v.(string); ok {
													if strings.ToLower(k) != "cookie" {
														capturedHeaders[k] = strVal
													}
												}
											}
										}
										captureMu.Unlock()
									}
								}
							})

							runErr := chromedp.Run(ctx,
								chromedp.ActionFunc(func(ctx context.Context) error {
									err := network.Enable().Do(ctx)
									if err != nil {
										return err
									}
									_, err = page.AddScriptToEvaluateOnNewDocument(GetStealthScript(filters.Language)).Do(ctx)
									return err
								}),
							)

							if runErr == nil {
								runErr = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
									_, _, _, _, err := page.Navigate(searchURL).Do(ctx)
									return err
								}))
							}

							if runErr == nil {
								// Check for sorry page redirect
								chromedp.Run(ctx, chromedp.Location(&pageURL))
								if strings.Contains(strings.ToLower(pageURL), "sorry") {
									log.Printf("   ⚠️ [W%d] Browser Fallback hit CAPTCHA, attempting to solve...", workerID)
									solved, solveErr := solver.DefeatCaptcha(ctx, 200, 400)
									if solveErr != nil {
										log.Printf("   ❌ CAPTCHA solver error: %v", solveErr)
									} else if solved {
										log.Printf("   ✅ CAPTCHA solved, waiting for Google redirect...")
										time.Sleep(2 * time.Second)
									}
								}
							}

							if runErr == nil {
								// Poll for results
								runErr = chromedp.Run(ctx, chromedp.Poll(`(() => {
									const results = document.querySelectorAll('a h3');
									return results.length > 0;
								})()`, nil, chromedp.WithPollingInterval(100*time.Millisecond)))
							}

							if runErr == nil {
								// Extract details and cookies
								runErr = chromedp.Run(ctx,
									chromedp.ActionFunc(func(ctx context.Context) error {
										err := chromedp.Location(&pageURL).Do(ctx)
										if err != nil {
											return err
										}
										err = chromedp.Evaluate(fmt.Sprintf("(%s)(%d)", extractJS, limit), &res).Do(ctx)
										return err
									}),
									chromedp.ActionFunc(func(ctx context.Context) error {
										var err error
										cookies, err = network.GetCookies().WithURLs([]string{"https://www.google.com"}).Do(ctx)
										return err
									}),
								)
							}

							cancelTimeout()
							cancel()

							if runErr == nil && len(capturedHeaders) > 0 && len(cookies) > 0 {
								saveSessionConfig(capturedHeaders, cookies)
								lastHealTime = time.Now()
								healed = true
								log.Printf("✅ [W%d] Self-healing SUCCESS: Fresh cookies loaded and saved. Retrying HTTP search...", workerID)
								
								// Retry the direct HTTP search once with fresh config
								retryRes, retryErr := runHTTPSearch(q, limit, filters)
								if retryErr == nil {
									status = "SUCCESS"
									httpRes = retryRes
									log.Printf("🎉 [W%d] Retry SUCCESS: HTTP request recovered after healing!", workerID)
								} else {
									log.Printf("❌ [W%d] Retry FAILED after healing: %v", workerID, retryErr)
									status = "ERROR"
								}
							} else {
								log.Printf("❌ [W%d] Self-healing FAILED: Browser search error: %v", workerID, runErr)
								status = "ERROR"
							}
						} else {
							// Another worker recently healed, so we reload config and retry direct HTTP search
							log.Printf("🛡️ [W%d] CAPTCHA detected. Another worker recently self-healed, reloading session config...", workerID)
							loadSessionConfig()
							retryRes, retryErr := runHTTPSearch(q, limit, filters)
							if retryErr == nil {
								status = "SUCCESS"
								httpRes = retryRes
								healed = true
								log.Printf("🎉 [W%d] Retry SUCCESS: Recovered using session from other worker!", workerID)
							} else {
								log.Printf("❌ [W%d] Retry FAILED after session reload: %v", workerID, retryErr)
								status = "ERROR"
							}
						}
						healMu.Unlock()
					} else {
						status = "ERROR"
					}
				}

				dur := time.Since(start)
				
				statsMu.Lock()
				if status == "SUCCESS" {
					successCount++
					totalDuration += dur.Milliseconds()
				} else if status == "CAPTCHA_BLOCKED" {
					captchaCount++
				} else {
					errorCount++
				}
				statsMu.Unlock()

				resultsMu.Lock()
				results = append(results, RequestResult{
					Index:    idx,
					Query:    q,
					Status:   status,
					Duration: dur,
					Healed:   healed,
				})
				resultsMu.Unlock()

				log.Printf("📊 [W%d] Req %d/%d Done. Status: %s | Dur: %v | SGE: %v", workerID, idx+1, count, status, dur, len(httpRes) > 0 && httpRes[0].Rank == 0)

				if delayMs > 0 {
					time.Sleep(time.Duration(delayMs) * time.Millisecond)
				}
			}
		}(w)
	}

	wg.Wait()

	// Print summary stats
	log.Println(strings.Repeat("=", 60))
	log.Println("🏁 STRESS TEST RESULTS:")
	log.Println(strings.Repeat("=", 60))
	log.Printf("Total Queries Executed:   %d", count)
	log.Printf("Successful Requests:      %d (%.1f%%)", successCount, float64(successCount)/float64(count)*100)
	log.Printf("Captcha Blocks (Total):   %d (%.1f%%)", captchaCount, float64(captchaCount)/float64(count)*100)
	log.Printf("Connection/Other Errors:  %d (%.1f%%)", errorCount, float64(errorCount)/float64(count)*100)
	log.Printf("Self-Healing Triggers:    %d", healTriggerCount)
	if successCount > 0 {
		log.Printf("Average Latency (Success): %v", time.Duration(totalDuration/successCount)*time.Millisecond)
	}
	log.Println(strings.Repeat("-", 60))
	log.Println("Detailed Timeline:")
	for _, r := range results {
		healStr := ""
		if r.Healed {
			healStr = " (Healed)"
		}
		log.Printf("  #%-3d: %-35q | Status: %-15s | Latency: %-8v%s", r.Index+1, r.Query, r.Status, r.Duration, healStr)
	}
	log.Println(strings.Repeat("=", 60))
}

var (
	globalBrowserCtx    context.Context
	globalBrowserCancel context.CancelFunc
	globalAllocCtx      context.Context
	globalAllocCancel   context.CancelFunc
	globalBrowserOnce   sync.Once
	replenishMu         sync.Mutex
	isReplenishing      bool
)

func GetGlobalBrowserCtx() context.Context {
	globalBrowserOnce.Do(func() {
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
			chromedp.Flag("blink-settings", "imagesEnabled=false"),
			chromedp.Flag("mute-audio", true),
			chromedp.WindowSize(1440, 900),
			chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"),
		}
		opts = append(opts, GetAcceptLangOption(GetReplenishFilters().Language))
		globalAllocCtx, globalAllocCancel = chromedp.NewExecAllocator(context.Background(), opts...)
		globalBrowserCtx, globalBrowserCancel = chromedp.NewContext(globalAllocCtx)
		if err := chromedp.Run(globalBrowserCtx); err != nil {
			log.Printf("⚠️ Failed to start global browser context: %v", err)
		} else {
			log.Println("🌐 [Global Browser] Started global browser context successfully.")
		}
	})
	return globalBrowserCtx
}

func ReplenishSessionPool(targetCount int) {
	replenishMu.Lock()
	if isReplenishing {
		replenishMu.Unlock()
		return
	}
	isReplenishing = true
	replenishMu.Unlock()

	defer func() {
		replenishMu.Lock()
		isReplenishing = false
		replenishMu.Unlock()
	}()

	active := poolManager.ActiveCount()
	if active >= targetCount {
		return
	}

	needed := targetCount - active
	if needed > 5 {
		needed = 5 // limit concurrency to 5 max
	}

	log.Printf("🔄 [Session Pool] Replenishing session pool: active=%d, target=%d, launching %d parallel browsers...", active, targetCount, needed)

	parentCtx := GetGlobalBrowserCtx()
	if parentCtx == nil {
		log.Println("❌ [Session Pool] Cannot replenish, global browser context is nil.")
		return
	}

	var wg sync.WaitGroup
	for i := 0; i < needed; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			log.Printf("🔑 [Replenish W%d] Spawning isolated browser tab...", workerID)
			ctx, cancel, tabErr := createIsolatedTab(parentCtx)
			if tabErr != nil {
				log.Printf("❌ [Replenish W%d] Failed to spawn isolated browser tab: %v", workerID, tabErr)
				return
			}
			ctx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
			defer cancelTimeout()
			defer cancel()

			replenishQueries := []string{
				"weather today",
				"local news",
				"what time is it",
				"google translate",
				"calculator",
				"map of new york",
				"stock market today",
				"world news",
				"dictionary",
				"internet speed test",
			}
			qIdx := (int(time.Now().UnixNano()) + workerID) % len(replenishQueries)
			if qIdx < 0 {
				qIdx = -qIdx
			}
			q := replenishQueries[qIdx]
			searchURL := BuildSearchURL(q, 10, GetReplenishFilters())
			var pageURL string
			var capturedHeaders map[string]string
			var captureMu sync.Mutex
			var cookies []*network.Cookie

			chromedp.ListenTarget(ctx, func(ev interface{}) {
				if ev, ok := ev.(*network.EventRequestWillBeSent); ok {
					if strings.Contains(ev.Request.URL, "google.com/search") && ev.Request.Method == "GET" {
						captureMu.Lock()
						if capturedHeaders == nil {
							capturedHeaders = make(map[string]string)
							for k, v := range ev.Request.Headers {
								if strVal, ok := v.(string); ok {
									if strings.ToLower(k) != "cookie" {
										capturedHeaders[k] = strVal
									}
								}
							}
						}
						captureMu.Unlock()
					}
				}
			})

			// Run the pre-fetch sequence
			err := chromedp.Run(ctx,
				chromedp.ActionFunc(func(c context.Context) error {
					err := network.Enable().Do(c)
					if err != nil {
						return err
					}
					langs := getLanguagesForCode(GetReplenishFilters().Language)
					err = network.SetExtraHTTPHeaders(network.Headers{
						"Accept-Language": strings.Join(langs, ","),
					}).Do(c)
					if err != nil {
						return err
					}
					_, err = page.AddScriptToEvaluateOnNewDocument(GetStealthScript(GetReplenishFilters().Language)).Do(c)
					return err
				}),
				chromedp.Navigate(searchURL),
				chromedp.Location(&pageURL),
			)

			if err == nil && strings.Contains(strings.ToLower(pageURL), "sorry") {
				log.Printf("⚠️ [Replenish W%d] Encountered CAPTCHA, attempting to solve...", workerID)
				solved, solveErr := solver.DefeatCaptcha(ctx, 200, 400)
				if solveErr != nil {
					log.Printf("❌ [Replenish W%d] CAPTCHA solve error: %v", workerID, solveErr)
				} else if solved {
					log.Printf("✅ [Replenish W%d] CAPTCHA solved, waiting...", workerID)
					time.Sleep(2 * time.Second)
					chromedp.Run(ctx, chromedp.Location(&pageURL))
				}
			}

			if err == nil {
				// Poll to verify search results exist
				err = chromedp.Run(ctx, chromedp.Poll(`(() => {
					return document.querySelectorAll('a h3').length > 0;
				})()`, nil, chromedp.WithPollingInterval(100*time.Millisecond)))
			}

			if err == nil {
				// Extract cookies
				err = chromedp.Run(ctx, chromedp.ActionFunc(func(c context.Context) error {
					var err error
					cookies, err = network.GetCookies().WithURLs([]string{"https://www.google.com"}).Do(c)
					return err
				}))
			}

			if err == nil && len(capturedHeaders) > 0 && len(cookies) > 0 {
				sessionID := poolManager.AddSession(capturedHeaders, cookies)
				log.Printf("🎉 [Replenish W%d] Successfully captured and added session %s", workerID, sessionID)
			} else {
				log.Printf("❌ [Replenish W%d] Failed to capture session: err=%v, headers=%d, cookies=%d", 
					workerID, err, len(capturedHeaders), len(cookies))
			}
		}(i)
	}

	wg.Wait()
	log.Printf("🔄 [Session Pool] Replenishment complete. Active session count: %d", poolManager.ActiveCount())
}

// createIsolatedTab spawns a completely isolated browser context (incognito profile)
// and creates a new window in it, returning a chromedp context attached to the new target.
// This bypasses the CDP -32000 "Failed to open new tab - no browser is open" error in headless mode.
func createIsolatedTab(parentCtx context.Context) (context.Context, context.CancelFunc, error) {
	var browserContextID cdp.BrowserContextID
	err := chromedp.Run(parentCtx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		browserContextID, err = target.CreateBrowserContext().Do(cdp.WithExecutor(ctx, chromedp.FromContext(parentCtx).Browser))
		return err
	}))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create browser context: %w", err)
	}

	var targetID target.ID
	err = chromedp.Run(parentCtx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		targetID, err = target.CreateTarget("about:blank").
			WithBrowserContextID(browserContextID).
			WithNewWindow(true).
			Do(cdp.WithExecutor(ctx, chromedp.FromContext(parentCtx).Browser))
		return err
	}))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create target: %w", err)
	}

	ctx, cancel := chromedp.NewContext(parentCtx, chromedp.WithTargetID(targetID))
	return ctx, cancel, nil
}

// LogQueryFailure records parser, SGE, and organic search breakdowns inside a query failure forensic log.
func LogQueryFailure(queryText, dorkQuery, failureType, rawSGEResponse string, resolvedFilters SearchFilters) {
	file, err := os.OpenFile("query_failures.jsonl", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	failureEntry := map[string]interface{}{
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
		"raw_query":        queryText,
		"compiled_dork":    dorkQuery,
		"failure_type":     failureType, // e.g. "HYBRID_PARSE_ERROR", "NO_SEARCH_RESULTS", "SGE_REFUSAL", "SGE_SCHEMALESS_RESPONSE"
		"raw_sge_response": rawSGEResponse,
		"filters":          resolvedFilters,
	}

	data, err := json.Marshal(failureEntry)
	if err == nil {
		_, _ = file.Write(append(data, '\n'))
	}
}

