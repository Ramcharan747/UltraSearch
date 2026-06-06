package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"go_search/solver"
)

// BraveEngine implements SearchEngine for Brave Search.
type BraveEngine struct{}

func (b *BraveEngine) Name() string { return "brave" }

func (b *BraveEngine) BuildSearchURL(query string, limit int, filters SearchFilters) string {
	base, err := url.Parse("https://search.brave.com/search")
	if err != nil {
		return "https://search.brave.com/search?q=" + url.QueryEscape(query)
	}
	params := url.Values{}
	params.Set("q", query)
	if limit > 0 {
		params.Set("count", fmt.Sprintf("%d", limit))
	}
	// Brave supports country via 'country' param
	country := filters.Country
	if country != "" && country != "browser" {
		params.Set("country", country)
	}
	base.RawQuery = params.Encode()
	return base.String()
}

func (b *BraveEngine) ExtractJS() string {
	return `(maxResults) => {
		const out = [];

		// ── Brave AI Answer / AI Overview extraction ──
		const aiSelectors = [
			'.ai-overview',
			'.ai-answer', 
			'#ai-answer',
			'[class*="ai-overview"]',
			'[class*="ai-answer"]',
			'[data-type="ai"]',
			'.answer-module',
			'.infobox',
			'.chatllm-answer-list',
			'[class*="chatllm"]'
		];
		let aiText = '';
		for (const sel of aiSelectors) {
			const el = document.querySelector(sel);
			if (el) {
				const clone = el.cloneNode(true);
				clone.querySelectorAll('button, svg, style, script').forEach(e => e.remove());
				const text = clone.innerText.replace(/\n{3,}/g, '\n\n').trim();
				if (text.length > aiText.length) aiText = text;
			}
		}
		
		// Also check for summarizer panel
		const summarizer = document.querySelector('#summarizer, .summarizer, [class*="summarizer"]');
		if (summarizer) {
			const clone = summarizer.cloneNode(true);
			clone.querySelectorAll('button, svg, style, script').forEach(e => e.remove());
			const text = clone.innerText.replace(/\n{3,}/g, '\n\n').trim();
			if (text.length > aiText.length) aiText = text;
		}

		if (aiText.length > 30) {
			out.push({
				rank: 0,
				title: "🦁 Brave AI Answer",
				url: window.location.href,
				snippet: aiText.substring(0, 5000)
			});
		}

		// ── Brave organic search results ──
		const resultSelectors = [
			'#results .snippet',
			'#results .card', 
			'.web-results .snippet',
			'[data-type="web"]',
			'.fdb'
		];
		
		let resultEls = [];
		for (const sel of resultSelectors) {
			const items = document.querySelectorAll(sel);
			if (items.length > 0) {
				resultEls = Array.from(items);
				break;
			}
		}

		for (const el of resultEls) {
			const linkEl = el.querySelector('a[href]');
			const titleEl = el.querySelector('.title, .search-snippet-title, h2');
			const snippetEl = el.querySelector('.generic-snippet .content, .snippet-description, .snippet-content, .snippet-text, p');
			
			if (linkEl && linkEl.href && !linkEl.href.includes('brave.com')) {
				const title = titleEl ? titleEl.innerText : linkEl.innerText;
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

func (b *BraveEngine) PollReadyJS() string {
	return `(() => {
		// 1. Check for AI answer button and click it
		const llmBtn = document.querySelector('#submit-llm-button');
		if (llmBtn && !window._braveAiClicked) {
			llmBtn.click();
			window._braveAiClicked = true;
			window._braveAIPollStart = Date.now();
			return false; // Give it time to start
		}

		// Check if Brave AI answer has loaded
		const aiSelectors = ['.ai-overview', '.ai-answer', '#ai-answer', '#summarizer', '.summarizer', '.chatllm-answer-list'];
		let aiFound = false;
		for (const sel of aiSelectors) {
			const el = document.querySelector(sel);
			if (el && el.innerText.length > 50) {
				aiFound = true;
				const now = Date.now();
				if (window._braveAIPrevLen === undefined) {
					window._braveAIPrevLen = el.innerText.length;
					window._braveAILastChange = now;
					return false;
				}
				if (el.innerText.length !== window._braveAIPrevLen) {
					window._braveAIPrevLen = el.innerText.length;
					window._braveAILastChange = now;
					return false;
				}
				if (now - window._braveAILastChange > 2000) return true;
				return false;
			}
		}

		// Wait at least 3 seconds if we clicked the AI button and it hasn't shown up
		if (window._braveAIPollStart && (Date.now() - window._braveAIPollStart < 3000)) {
			return false;
		}

		// No AI panel (or finished waiting) — just wait for organic results
		const results = document.querySelectorAll('#results .snippet, #results .card, .web-results .snippet');
		if (results.length > 0) {
			if (!window._braveResultTime) window._braveResultTime = Date.now();
			if (Date.now() - window._braveResultTime > 500) return true;
		}
		return false;
	})()`
}

func (b *BraveEngine) DetectChallenge(bodyText string) bool {
	lower := strings.ToLower(bodyText)
	return strings.Contains(lower, "verifying you're not a bot") ||
		strings.Contains(lower, "quick check before you continue searching") ||
		strings.Contains(lower, "verify you are human")
}

func (b *BraveEngine) SolveChallenge(ctx context.Context, currentX, currentY float64) (bool, error) {
	return solver.DefeatBraveChallenge(ctx, currentX, currentY)
}

func (b *BraveEngine) AIOverviewTitle() string { return "🦁 Brave AI Answer" }
