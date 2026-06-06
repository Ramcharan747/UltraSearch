package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// BingEngine implements SearchEngine for Bing Search.
type BingEngine struct{}

func (bi *BingEngine) Name() string { return "bing" }

func (bi *BingEngine) BuildSearchURL(query string, limit int, filters SearchFilters) string {
	base, err := url.Parse("https://www.bing.com/search")
	if err != nil {
		return "https://www.bing.com/search?q=" + url.QueryEscape(query)
	}
	params := url.Values{}
	params.Set("q", query)
	if limit > 0 {
		params.Set("count", fmt.Sprintf("%d", limit))
	}
	// Bing supports language via setlang and market via mkt
	lang := filters.Language
	if lang != "" && lang != "browser" {
		params.Set("setlang", lang)
	}
	country := filters.Country
	if country != "" && country != "browser" {
		// Bing uses mkt format like "en-US"
		mkt := lang + "-" + strings.ToUpper(country)
		params.Set("mkt", mkt)
	}
	safe := filters.SafeSearch
	if safe == "active" {
		params.Set("safeSearch", "Strict")
	} else if safe == "off" {
		params.Set("safeSearch", "Off")
	}
	base.RawQuery = params.Encode()
	return base.String()
}

func (bi *BingEngine) ExtractJS() string {
	return `(maxResults) => {
		const out = [];

		// ── Bing AI / Copilot Answer extraction ──
		const aiSelectors = [
			'#gptans',
			'#b_results .b_ans .b_rich',
			'[class*="copilot"]',
			'.b_ai_container',
			'#sydneyContainer',
			'.rai_message',
			'.b_ans:has(.b_rich)'
		];
		let aiText = '';
		for (const sel of aiSelectors) {
			try {
				const el = document.querySelector(sel);
				if (el) {
					const clone = el.cloneNode(true);
					clone.querySelectorAll('button, svg, style, script, [role="dialog"]').forEach(e => e.remove());
					const text = clone.innerText.replace(/\n{3,}/g, '\n\n').trim();
					if (text.length > aiText.length) aiText = text;
				}
			} catch(e) {}
		}

		if (aiText.length > 30) {
			out.push({
				rank: 0,
				title: "🔷 Bing Copilot Answer",
				url: window.location.href,
				snippet: aiText.substring(0, 5000)
			});
		}

		// ── Bing organic search results ──
		const algos = document.querySelectorAll('#b_results .b_algo');
		for (const el of algos) {
			const linkEl = el.querySelector('h2 a');
			const snippetEl = el.querySelector('.b_caption p, .b_lineclamp2, .b_paractl, .b_dList p');
			
			if (linkEl && linkEl.href) {
				const title = linkEl.innerText || '';
				const snippet = snippetEl ? snippetEl.innerText.replace(/\n/g, ' ').trim() : '';
				out.push({
					rank: out.filter(r => r.rank > 0).length + 1,
					title: title.trim(),
					url: linkEl.href,
					snippet: snippet.substring(0, 1000)
				});
			}
			if (out.length >= maxResults + 1) break;
		}
		return out;
	}`
}

func (bi *BingEngine) PollReadyJS() string {
	return `(() => {
		// Check if Bing Copilot answer has loaded
		const aiSelectors = ['#gptans', '.rai_message', '#sydneyContainer', '[class*="copilot"]'];
		for (const sel of aiSelectors) {
			try {
				const el = document.querySelector(sel);
				if (el && el.innerText.length > 50) {
					const now = Date.now();
					if (window._bingAIPrevLen === undefined) {
						window._bingAIPrevLen = el.innerText.length;
						window._bingAILastChange = now;
						return false;
					}
					if (el.innerText.length !== window._bingAIPrevLen) {
						window._bingAIPrevLen = el.innerText.length;
						window._bingAILastChange = now;
						return false;
					}
					if (now - window._bingAILastChange > 2000) return true;
					return false;
				}
			} catch(e) {}
		}
		// Fallback: wait for organic results
		const results = document.querySelectorAll('#b_results .b_algo');
		if (results.length > 0) {
			if (!window._bingResultTime) window._bingResultTime = Date.now();
			if (Date.now() - window._bingResultTime > 500) return true;
		}
		return false;
	})()`
}

func (bi *BingEngine) DetectChallenge(bodyText string) bool {
	lower := strings.ToLower(bodyText)
	return strings.Contains(lower, "unusual traffic") ||
		strings.Contains(lower, "automated requests") ||
		strings.Contains(lower, "verify you're a human") ||
		strings.Contains(lower, "captcha")
}

func (bi *BingEngine) SolveChallenge(ctx context.Context, currentX, currentY float64) (bool, error) {
	return false, fmt.Errorf("bing captcha solver not implemented yet")
}

func (bi *BingEngine) AIOverviewTitle() string { return "🔷 Bing Copilot Answer" }
