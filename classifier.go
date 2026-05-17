package main

import (
	"io"
	"net/http"
	"strings"
	"time"
)

// --- Tier Constants ---
const (
	TierStatic      = 1 // Pure HTTP GET, curl-speed
	TierJSRender    = 2 // Needs headless Chrome for JS rendering
	TierBotProtect  = 3 // Needs stealth browser + ML solver
	TierLoginWall   = 4 // Needs stealth + cookie injection
	TierUnreachable = 5 // Dead, skip
)

// --- Domain Knowledge Cache ---
// Pre-classified domains to skip the HTTP probe entirely
var domainTierCache = map[string]int{
	// T2: JS-rendered SPAs
	"reddit.com":       TierJSRender,
	"www.reddit.com":   TierJSRender,
	"old.reddit.com":   TierJSRender,
	"medium.com":       TierJSRender,
	"quora.com":        TierJSRender,
	"www.quora.com":    TierJSRender,
	"twitter.com":      TierJSRender,
	"x.com":            TierJSRender,
	"facebook.com":     TierJSRender,

	// T3: Bot-protected
	"pitchbook.com":          TierBotProtect,
	"www.pitchbook.com":      TierBotProtect,
	"crunchbase.com":         TierBotProtect,
	"www.crunchbase.com":     TierBotProtect,
	"g2.com":                 TierBotProtect,
	"www.g2.com":             TierBotProtect,
	"glassdoor.com":          TierBotProtect,
	"www.glassdoor.com":      TierBotProtect,
	"zillow.com":             TierBotProtect,
	"www.zillow.com":         TierBotProtect,
	"indeed.com":             TierBotProtect,
	"www.indeed.com":         TierBotProtect,

	// T4: Login-walled
	"linkedin.com":           TierJSRender, // Public articles work with JS render
	"www.linkedin.com":       TierJSRender,
}

// extractDomain pulls the hostname from a URL string
func extractDomain(rawURL string) string {
	// Quick parse without importing net/url to keep it fast
	s := rawURL
	if idx := strings.Index(s, "://"); idx >= 0 {
		s = s[idx+3:]
	}
	if idx := strings.Index(s, "/"); idx >= 0 {
		s = s[:idx]
	}
	if idx := strings.Index(s, "?"); idx >= 0 {
		s = s[:idx]
	}
	// Also check the bare domain (strip www. for matching)
	return strings.ToLower(s)
}

// baseDomain strips "www." for matching subdomains
func baseDomain(domain string) string {
	return strings.TrimPrefix(domain, "www.")
}

// lookupDomainTier checks if a URL's domain has a known tier.
// Returns 0 if not in cache (needs probing).
func lookupDomainTier(rawURL string) int {
	domain := extractDomain(rawURL)

	// Direct match
	if tier, ok := domainTierCache[domain]; ok {
		return tier
	}

	// Check for subdomain match (e.g., "blossomstreetventures.medium.com" → medium.com)
	parts := strings.Split(domain, ".")
	if len(parts) > 2 {
		parentDomain := strings.Join(parts[len(parts)-2:], ".")
		if tier, ok := domainTierCache[parentDomain]; ok {
			return tier
		}
	}

	return 0 // Unknown, needs HTTP probe
}

// --- HTTP Probe Classification ---

// ProbeResult holds the outcome of an HTTP probe
type ProbeResult struct {
	Tier     int
	HTML     string // Only populated if Tier == TierStatic
	StatusCode int
}

// probeURL does a fast HTTP GET and classifies the response
func probeURL(rawURL string, client *http.Client) ProbeResult {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return ProbeResult{Tier: TierUnreachable}
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		// Timeout or DNS failure
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline") {
			return ProbeResult{Tier: TierJSRender} // Could be a slow SPA
		}
		return ProbeResult{Tier: TierUnreachable}
	}
	defer resp.Body.Close()

	// --- Status code classification ---
	if resp.StatusCode == 403 || resp.StatusCode == 503 {
		return ProbeResult{Tier: TierBotProtect, StatusCode: resp.StatusCode}
	}
	if resp.StatusCode == 429 {
		return ProbeResult{Tier: TierBotProtect, StatusCode: 429}
	}
	if resp.StatusCode == 401 {
		return ProbeResult{Tier: TierLoginWall, StatusCode: 401}
	}
	if resp.StatusCode >= 400 {
		return ProbeResult{Tier: TierUnreachable, StatusCode: resp.StatusCode}
	}

	// --- Body analysis ---
	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)
	bodyLower := strings.ToLower(body)

	// Too short = probably a redirect stub or challenge
	if len(body) < 500 {
		return ProbeResult{Tier: TierJSRender, StatusCode: resp.StatusCode}
	}

	// Bot protection signatures
	if containsAny(bodyLower, []string{
		"cf-turnstile",
		"challenges.cloudflare.com",
		"just a moment",
		"enable javascript and cookies",
		"datadome",
		"verify you are human",
		"_cf_chl_opt",
		"cf-browser-verification",
	}) {
		return ProbeResult{Tier: TierBotProtect, StatusCode: resp.StatusCode}
	}

	// SPA / JS-required signatures
	if containsAny(bodyLower, []string{
		"<noscript>",
		"<div id=\"root\"></div>",
		"<div id=\"app\"></div>",
		"<div id=\"__next\"></div>",
		"window.__initial_state__",
		"you need to enable javascript",
	}) && len(body) < 5000 {
		return ProbeResult{Tier: TierJSRender, StatusCode: resp.StatusCode}
	}

	// Login wall signatures
	if containsAny(bodyLower, []string{
		"sign in to view",
		"log in to continue",
		"create an account",
		"join to view",
	}) {
		return ProbeResult{Tier: TierLoginWall, StatusCode: resp.StatusCode}
	}

	// If we got here, it's real static HTML
	return ProbeResult{Tier: TierStatic, HTML: body, StatusCode: resp.StatusCode}
}

// --- Content Quality ---

// ContentQuality scores extracted text. Returns true if quality is sufficient.
func ContentQuality(text string) bool {
	if len(text) < 80 {
		return false
	}

	// Check it's not just navigation/footer garbage
	words := strings.Fields(text)
	if len(words) < 15 {
		return false
	}

	// Check for common garbage patterns
	lower := strings.ToLower(text)
	garbagePatterns := []string{
		"enable javascript",
		"cookies are required",
		"your browser does not support",
		"please enable cookies",
		"just a moment",
		"checking your browser",
		"performing security verification",
	}
	for _, p := range garbagePatterns {
		if strings.Contains(lower, p) && len(text) < 300 {
			return false
		}
	}

	return true
}

// --- Helpers ---

func containsAny(s string, patterns []string) bool {
	for _, p := range patterns {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}

// SharedHTTPClient returns a reusable client with connection pooling
func SharedHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 4 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        30,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     30 * time.Second,
			DisableKeepAlives:   false,
		},
		// Don't follow redirects that go to login pages
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
}
