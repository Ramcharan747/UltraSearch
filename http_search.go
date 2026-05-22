package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
)

type CookieInfo struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type SessionItem struct {
	ID        string            `json:"id"`
	Headers   map[string]string `json:"headers"`
	Cookies   []CookieInfo      `json:"cookies"`
	UseCount  int               `json:"use_count"`
	Blocked   bool              `json:"blocked"`
	CreatedAt time.Time         `json:"created_at"`
}

type SessionPoolConfig struct {
	Sessions []SessionItem `json:"sessions"`
}

type SessionPoolManager struct {
	mu       sync.Mutex
	sessions []SessionItem
	waiters  []chan struct{}
}

var (
	poolManager       SessionPoolManager
	ReplenishCallback func()
	httpClient        *http.Client
	httpClientOnce    sync.Once
)

func init() {
	poolManager.LoadPool()
	PrewarmConnection()
}

func (m *SessionPoolManager) LoadPool() {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile("solver/session_config.json")
	if err != nil {
		return
	}

	var poolConfig SessionPoolConfig
	if err := json.Unmarshal(data, &poolConfig); err == nil && len(poolConfig.Sessions) > 0 {
		// Filter out blocked, depleted, or stale (older than 1 hour) sessions
		var validSessions []SessionItem
		for _, s := range poolConfig.Sessions {
			if s.Blocked {
				continue
			}
			if s.UseCount >= 5 {
				continue
			}
			if !s.CreatedAt.IsZero() && time.Since(s.CreatedAt) > 1*time.Hour {
				log.Printf("🧹 [Session Pool] Discarding stale session %s (created %v ago)", s.ID, time.Since(s.CreatedAt))
				continue
			}
			if s.CreatedAt.IsZero() {
				s.CreatedAt = time.Now()
			}
			validSessions = append(validSessions, s)
		}
		m.sessions = validSessions
		log.Printf("🔑 [Session Pool] Loaded %d valid sessions successfully.", len(m.sessions))
		return
	}

	// Try fallback to legacy config read
	type LegacyConfig struct {
		Headers map[string]string `json:"headers"`
		Cookies []CookieInfo      `json:"cookies"`
	}
	var legacy LegacyConfig
	if err := json.Unmarshal(data, &legacy); err == nil && len(legacy.Cookies) > 0 {
		m.sessions = []SessionItem{
			{
				ID:        fmt.Sprintf("s_%d", time.Now().UnixNano()),
				Headers:   legacy.Headers,
				Cookies:   legacy.Cookies,
				CreatedAt: time.Now(),
			},
		}
		log.Println("🔑 [Session Pool] Migrated legacy single session config.")
	}
}

func (m *SessionPoolManager) SavePool() {
	m.mu.Lock()
	defer m.mu.Unlock()

	cfg := SessionPoolConfig{
		Sessions: m.sessions,
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Printf("⚠️ [Session Pool] Failed to marshal pool: %v", err)
		return
	}

	_ = os.MkdirAll("solver", 0755)
	err = os.WriteFile("solver/session_config.json", data, 0644)
	if err != nil {
		log.Printf("⚠️ [Session Pool] Failed to write session_config.json: %v", err)
	} else {
		log.Printf("💾 [Session Pool] Successfully saved updated pool of %d sessions to solver/session_config.json", len(m.sessions))
	}
}

func (m *SessionPoolManager) AddSession(headers map[string]string, cookies []*network.Cookie) string {
	m.mu.Lock()
	
	id := fmt.Sprintf("s_%d", time.Now().UnixNano())
	item := SessionItem{
		ID:        id,
		Headers:   headers,
		Cookies:   []CookieInfo{},
		CreatedAt: time.Now(),
	}
	for _, c := range cookies {
		item.Cookies = append(item.Cookies, CookieInfo{
			Name:  c.Name,
			Value: c.Value,
		})
	}
	
	// Add it to our sessions list
	m.sessions = append(m.sessions, item)
	
	// Notify the first waiter if any
	var waiter chan struct{}
	if len(m.waiters) > 0 {
		waiter = m.waiters[0]
		m.waiters = m.waiters[1:]
	}
	m.mu.Unlock()
	
	if waiter != nil {
		select {
		case waiter <- struct{}{}:
		default:
		}
	}
	
	m.SavePool()
	return id
}

func (m *SessionPoolManager) GetSession() (SessionItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Look for a valid session (UseCount < 5, not blocked)
	for _, s := range m.sessions {
		if !s.Blocked && s.UseCount < 5 {
			return s, nil
		}
	}
	return SessionItem{}, fmt.Errorf("no active/valid session found in pool")
}

func (m *SessionPoolManager) GetSessionOrWait(timeout time.Duration) (SessionItem, error) {
	m.mu.Lock()
	
	// Helper function to check out a session from the list if one is available
	checkout := func() (SessionItem, bool) {
		for i, s := range m.sessions {
			if !s.Blocked && s.UseCount < 5 {
				m.sessions[i].UseCount++
				log.Printf("🔄 [Session Pool] Session %s checked out (UseCount=%d/5)", s.ID, m.sessions[i].UseCount)
				
				item := m.sessions[i]
				evicted := false
				if m.sessions[i].UseCount >= 5 {
					log.Printf("♻️ [Session Pool] Session %s reached 5 checkouts. Evicting it.", s.ID)
					m.sessions = append(m.sessions[:i], m.sessions[i+1:]...)
					evicted = true
				}
				m.mu.Unlock()
				
				m.SavePool()
				if evicted {
					go TriggerReplenish()
				}
				return item, true
			}
		}
		return SessionItem{}, false
	}
	
	// Try to get a valid session immediately
	if item, found := checkout(); found {
		return item, nil
	}
	
	// If no session is available and timeout is <= 0, return immediately
	if timeout <= 0 {
		m.mu.Unlock()
		return SessionItem{}, fmt.Errorf("no active/valid session found in pool")
	}
	
	// Create waiter channel and register it
	ch := make(chan struct{}, 1)
	m.waiters = append(m.waiters, ch)
	m.mu.Unlock()
	
	// Wait for notification or timeout
	select {
	case <-ch:
		// Woken up because a session was added!
		// Re-acquire lock and try to fetch the session
		m.mu.Lock()
		if item, found := checkout(); found {
			return item, nil
		}
		m.mu.Unlock()
		return SessionItem{}, fmt.Errorf("woken up but no active session found")
		
	case <-time.After(timeout):
		// Timeout! We must remove our waiter channel from the list
		m.mu.Lock()
		defer m.mu.Unlock()
		for i, w := range m.waiters {
			if w == ch {
				m.waiters = append(m.waiters[:i], m.waiters[i+1:]...)
				break
			}
		}
		return SessionItem{}, fmt.Errorf("timeout waiting for active session")
	}
}

func (m *SessionPoolManager) ReleaseSession(id string, blocked bool) {
	m.mu.Lock()
	evicted := false
	for i, s := range m.sessions {
		if s.ID == id {
			if blocked {
				m.sessions[i].Blocked = true
				log.Printf("🚫 [Session Pool] Session %s marked as BLOCKED.", id)
				m.sessions = append(m.sessions[:i], m.sessions[i+1:]...)
				evicted = true
			}
			break
		}
	}
	m.mu.Unlock()

	if evicted {
		m.SavePool()
		// Trigger background replenishment
		go TriggerReplenish()
	}
}

func (m *SessionPoolManager) ActiveCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, s := range m.sessions {
		if !s.Blocked && s.UseCount < 5 {
			count++
		}
	}
	return count
}

func TriggerReplenish() {
	if ReplenishCallback != nil {
		ReplenishCallback()
	}
}

func loadSessionConfig() {
	poolManager.LoadPool()
}

func saveSessionConfig(headers map[string]string, cookies []*network.Cookie) {
	poolManager.AddSession(headers, cookies)
}


func PrewarmConnection() {
	go func() {
		client := getHTTPClient()
		req, err := http.NewRequest("HEAD", "https://www.google.com/", nil)
		if err != nil {
			return
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36")
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
			log.Println("🔥 [HTTP Search] Connection pre-warmed successfully.")
		} else {
			log.Printf("⚠️ [HTTP Search] Failed to pre-warm connection: %v", err)
		}
	}()
}

func getHTTPClient() *http.Client {
	httpClientOnce.Do(func() {
		transport := &http.Transport{
			MaxIdleConns:        100,
			IdleConnTimeout:     90 * time.Second,
			MaxIdleConnsPerHost: 100,
			DisableKeepAlives:   false,
		}
		httpClient = &http.Client{
			Transport: transport,
			Timeout:   10 * time.Second,
		}
	})
	return httpClient
}

func runHTTPSearch(q string, maxResults int, filters SearchFilters) ([]SearchResult, error) {
	// Wait up to 3 seconds for a session to become available
	session, err := poolManager.GetSessionOrWait(3 * time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to get active session: %w", err)
	}

	client := getHTTPClient()
	searchURL := BuildSearchURL(q, maxResults+10, filters)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	// Apply headers
	for k, v := range session.Headers {
		req.Header.Set(k, v)
	}
	// Fallback headers if not present
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36")
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	}
	if filters.Language != "" && filters.Language != "browser" {
		langs := getLanguagesForCode(filters.Language)
		var parts []string
		qVal := 1.0
		for _, l := range langs {
			if qVal == 1.0 {
				parts = append(parts, l)
			} else {
				parts = append(parts, fmt.Sprintf("%s;q=%.1f", l, qVal))
			}
			qVal -= 0.1
			if qVal < 0.1 {
				qVal = 0.1
			}
		}
		req.Header.Set("Accept-Language", strings.Join(parts, ","))
	} else if req.Header.Get("Accept-Language") == "" {
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	}

	// Apply cookies
	var cookieStrings []string
	for _, c := range session.Cookies {
		cookieStrings = append(cookieStrings, fmt.Sprintf("%s=%s", c.Name, c.Value))
	}
	if len(cookieStrings) > 0 {
		req.Header.Set("Cookie", strings.Join(cookieStrings, "; "))
	}

	t0 := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		poolManager.ReleaseSession(session.ID, true)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		poolManager.ReleaseSession(session.ID, true)
		return nil, fmt.Errorf("blocked by rate limit (status 429)")
	}

	if resp.StatusCode != http.StatusOK {
		poolManager.ReleaseSession(session.ID, false)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if resp.Request != nil && (strings.Contains(resp.Request.URL.Path, "/sorry/") || strings.Contains(resp.Request.URL.RawQuery, "continue=https://www.google.com/search")) {
		poolManager.ReleaseSession(session.ID, true)
		return nil, fmt.Errorf("blocked by captcha (redirected to /sorry/)")
	}

	poolManager.ReleaseSession(session.ID, false)

	log.Printf("⏱️ [HTTP Search] Google HTTP request took %v", time.Since(t0))

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var out []SearchResult

	// 1. Parse SGE (Google AI Overview)
	aiContainer := doc.Find(".s7d4ef")
	if aiContainer.Length() > 0 {
		var aiText string
		paragraphs := aiContainer.Find("div.n6owBd")
		if paragraphs.Length() > 0 {
			paragraphs.Each(func(i int, s *goquery.Selection) {
				aiText += s.Text() + " "
			})
		} else {
			aiText = aiContainer.Text()
		}
		aiText = strings.TrimSpace(strings.ReplaceAll(aiText, "\n", " "))
		
		lowerText := strings.ToLower(aiText)
		isErrorText := strings.Contains(lowerText, "not available for this search") || 
			strings.Contains(lowerText, "can't generate") || 
			strings.Contains(lowerText, "try again later")
			
		if len(aiText) > 30 && !isErrorText {
			if len(aiText) > 3000 {
				aiText = aiText[:3000]
			}
			out = append(out, SearchResult{
				Rank:    0,
				Title:   "✨ Google AI Overview",
				URL:     searchURL,
				Snippet: aiText,
			})
		}
	}

	// 2. Parse organic results
	doc.Find("a h3").Each(func(i int, s *goquery.Selection) {
		if len(out) >= maxResults {
			return
		}
		a := s.Closest("a")
		if a.Length() == 0 {
			return
		}
		href, exists := a.Attr("href")
		if !exists || strings.Contains(href, "google.com") {
			return
		}

		title := s.Text()
		snippet := ""

		// Find closest snippet container
		parent := a.Closest("[data-sokoban-container]")
		if parent.Length() == 0 {
			parent = a.Closest(".g")
		}
		if parent.Length() == 0 {
			parent = a.Parent().Parent().Parent()
		}

		if parent.Length() > 0 {
			var maxLen int
			parent.Find("div, span").Each(func(j int, child *goquery.Selection) {
				text := strings.TrimSpace(strings.ReplaceAll(child.Text(), "\n", " "))
				if len(text) > maxLen && text != title && !strings.Contains(text, "›") && child.Find("h3").Length() == 0 {
					maxLen = len(text)
					snippet = text
				}
			})
		}

		if len(snippet) > 1000 {
			snippet = snippet[:1000]
		}

		out = append(out, SearchResult{
			Rank:    len(out) + 1,
			Title:   title,
			URL:     href,
			Snippet: snippet,
		})
	})

	return out, nil
}
