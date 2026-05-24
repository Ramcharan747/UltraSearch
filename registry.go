package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// SkillBook represents a parsed community or core declarative schema template
type SkillBook struct {
	Name       string   `json:"name"`
	Version    string   `json:"version"`
	Author     string   `json:"author"`
	TrustTier  string   `json:"trust_tier"`
	Domains    []string `json:"domains"`
	FilePath   string   `json:"file_path"`
	RawContent string   `json:"-"`
}

// Global dynamic Skill Book catalog registry
var GlobalRegistry []SkillBook

// LoadSkillBookRegistry reads and catalogs all active Skill Books from the target directory
func LoadSkillBookRegistry(dirPath string) error {
	GlobalRegistry = nil
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return fmt.Errorf("skill book directory '%s' does not exist", dirPath)
	}

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".md" {
			// Skip directories like "unverified" in walk, we load them separately if needed
			if strings.Contains(path, "/unverified/") {
				return nil
			}

			sb, err := ParseSkillBookFile(path)
			if err == nil {
				GlobalRegistry = append(GlobalRegistry, sb)
			} else {
				log.Printf("⚠️ [WI-OS Registry] Skipping invalid skill book %s: %v", filepath.Base(path), err)
			}
		}
		return nil
	})

	if err == nil {
		log.Printf("📚 [WI-OS Registry] Cataloged %d active Skill Books into edge memory.", len(GlobalRegistry))
	}
	return err
}

// ParseSkillBookFile opens a markdown Skill Book, extracts Frontmatter, and returns a SkillBook object
func ParseSkillBookFile(filePath string) (SkillBook, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return SkillBook{}, err
	}

	content := string(bytes)
	if !strings.HasPrefix(content, "---") {
		return SkillBook{}, fmt.Errorf("missing YAML frontmatter block starting prefix '---'")
	}

	// Find the end of frontmatter block
	endIdx := strings.Index(content[3:], "---")
	if endIdx == -1 {
		return SkillBook{}, fmt.Errorf("unterminated frontmatter block matching suffix '---'")
	}
	endIdx += 3 // Adjust for prefix offset

	frontmatter := content[3:endIdx]
	rawBody := content[endIdx+3:]

	sb := SkillBook{
		FilePath:   filePath,
		RawContent: rawBody,
	}

	// Parse custom frontmatter lines (handwritten parser to avoid external dependencies)
	lines := strings.Split(frontmatter, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		colonIdx := strings.Index(line, ":")
		if colonIdx == -1 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(line[:colonIdx]))
		val := strings.TrimSpace(line[colonIdx+1:])

		switch key {
		case "name":
			sb.Name = strings.Trim(val, "\"' ")
		case "version":
			sb.Version = strings.Trim(val, "\"' ")
		case "author":
			sb.Author = strings.Trim(val, "\"' ")
		case "trust_tier":
			sb.TrustTier = strings.Trim(val, "\"' ")
		case "domains":
			// Parse bracketed array format e.g. [venture_capital, startups]
			val = strings.Trim(val, "[] ")
			parts := strings.Split(val, ",")
			for _, part := range parts {
				p := strings.TrimSpace(part)
				if p != "" {
					sb.Domains = append(sb.Domains, strings.Trim(p, "\"' "))
				}
			}
		}
	}

	if sb.Name == "" {
		return SkillBook{}, fmt.Errorf("missing required 'name' field in frontmatter")
	}

	return sb, nil
}

// SemanticRouteQuery maps an incoming natural query to the best-fit Skill Book schema (Resolving Scale)
func SemanticRouteQuery(queryText string) (SkillBook, float64, bool) {
	if len(GlobalRegistry) == 0 {
		return SkillBook{}, 0.0, false
	}

	var bestFit SkillBook
	maxScore := 0.0
	cleanQuery := strings.ReplaceAll(strings.ToLower(queryText), "_", " ")

	for _, sb := range GlobalRegistry {
		// Calculate similarity of the query against:
		// 1. The Skill Book name
		nameClean := strings.ReplaceAll(strings.ToLower(sb.Name), "_", " ")
		nameScore := CosineSimilarity(cleanQuery, nameClean)

		// 2. The domains list
		var domainsClean []string
		for _, d := range sb.Domains {
			domainsClean = append(domainsClean, strings.ReplaceAll(strings.ToLower(d), "_", " "))
		}
		domainsText := strings.Join(domainsClean, " ")
		domainScore := CosineSimilarity(cleanQuery, domainsText)

		score := nameScore
		if domainScore > score {
			score = domainScore
		}

		// Absolute word match gives an immediate boost for matched domain fields
		for _, domain := range domainsClean {
			words := strings.Fields(domain)
			for _, w := range words {
				if len(w) > 3 && strings.Contains(cleanQuery, w) {
					score += 0.20 // Apply boost for matching keyword (e.g. funding, capital)
				}
			}
		}

		if score > maxScore {
			maxScore = score
			bestFit = sb
		}
	}

	// We apply a minimum correlation floor to prevent false positive matches
	if maxScore > 0.15 {
		return bestFit, maxScore, true
	}

	return SkillBook{}, maxScore, false
}

// RunRegistryDiagnostics validates the Skill Book cataloging and Semantic Router
func RunRegistryDiagnostics() {
	fmt.Println("\n[*] Running WI-OS Registry & Semantic Router Diagnostics...")
	err := LoadSkillBookRegistry("ai_skills")
	if err != nil {
		fmt.Printf("❌ Failed to catalog Skill Books: %v\n", err)
		return
	}

	// Diagnostic Test 1: Route venture funding query
	q1 := "Who are the key lead investors and seed funding rounds of Databricks?"
	fmt.Printf("\nQuery 1: '%s'\n", q1)
	book1, score1, found1 := SemanticRouteQuery(q1)
	if found1 {
		fmt.Printf("🎯 Router Match:  %s (Confidence: %.4f)\n", book1.Name, score1)
		fmt.Printf("   Author:        %s | Version: %s\n", book1.Author, book1.Version)
		fmt.Printf("   Domains:       %v\n", book1.Domains)
	} else {
		fmt.Printf("⚠️ Router Match:  None (Highest Score: %.4f)\n", score1)
	}

	// Diagnostic Test 2: Route academic search
	q2 := "Retrieve highly cited research papers about Large Language Model context scaling on arxiv"
	fmt.Printf("\nQuery 2: '%s'\n", q2)
	book2, score2, found2 := SemanticRouteQuery(q2)
	if found2 {
		fmt.Printf("🎯 Router Match:  %s (Confidence: %.4f)\n", book2.Name, score2)
		fmt.Printf("   Domains:       %v\n", book2.Domains)
	} else {
		fmt.Printf("⚠️ Router Match:  None (Highest Score: %.4f)\n", score2)
	}
}
