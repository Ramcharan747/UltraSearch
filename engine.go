package main

import (
	"context"
	"strings"

	"go_search/solver"
)
// SearchEngine defines the interface each search engine must implement.
type SearchEngine interface {
	Name() string
	BuildSearchURL(query string, limit int, filters SearchFilters) string
	ExtractJS() string             // Browser JS to extract AI overview + organic results
	PollReadyJS() string           // Browser JS to poll for page readiness (AI streaming done)
	DetectChallenge(bodyText string) bool // Is this a CAPTCHA page?
	SolveChallenge(ctx context.Context, currentX, currentY float64) (bool, error) // Solve the CAPTCHA
	AIOverviewTitle() string       // e.g. "✨ Google AI Overview", "🦁 Brave AI Answer"
}

var engineRegistry = map[string]SearchEngine{
	"google": &GoogleEngine{},
	"brave":  &BraveEngine{},
	"bing":   &BingEngine{},
}

func GetEngine(name string) SearchEngine {
	if e, ok := engineRegistry[name]; ok {
		return e
	}
	return engineRegistry["google"] // default fallback
}

func ListEngines() []string {
	names := make([]string, 0, len(engineRegistry))
	for k := range engineRegistry {
		names = append(names, k)
	}
	return names
}

// GoogleEngine wraps the existing Google search functionality.
type GoogleEngine struct{}

func (g *GoogleEngine) Name() string { return "google" }

func (g *GoogleEngine) BuildSearchURL(query string, limit int, filters SearchFilters) string {
	return BuildSearchURL(query, limit, filters) // delegate to existing function
}

func (g *GoogleEngine) ExtractJS() string {
	return extractJS // the existing const in main.go
}

func (g *GoogleEngine) PollReadyJS() string {
	return `(() => {
		const aiContainer = document.querySelector('.s7d4ef');
		if (aiContainer) {
			const text = aiContainer.innerText.toLowerCase();
			if (text.includes("not available") || text.includes("can't generate") || text.includes("try again") || text.includes("check your connection")) {
				return true;
			}
			const currentLen = aiContainer.innerText.length;
			const now = Date.now();
			if (currentLen < 150) return false;
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
			if (now - window._sgeLastChangeTime > 1500) return true;
			return false;
		}
		const results = document.querySelectorAll('a h3');
		if (results.length > 0) {
			if (!window._firstResultSeenTime) window._firstResultSeenTime = Date.now();
			if (Date.now() - window._firstResultSeenTime > 400) return true;
		}
		return false;
	})()`
}

func (g *GoogleEngine) DetectChallenge(bodyText string) bool {
	lower := strings.ToLower(bodyText)
	return strings.Contains(lower, "/sorry/") ||
		strings.Contains(lower, "unusual traffic") ||
		strings.Contains(lower, "verify you are human") ||
		strings.Contains(lower, "just a moment") ||
		strings.Contains(lower, "checking your browser")
}

func (g *GoogleEngine) SolveChallenge(ctx context.Context, currentX, currentY float64) (bool, error) {
	return solver.DefeatCaptcha(ctx, currentX, currentY)
}

func (g *GoogleEngine) AIOverviewTitle() string { return "✨ Google AI Overview" }
