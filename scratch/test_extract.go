package main

import (
	"fmt"
	"regexp"
	"strings"
)

func extractNameWithRegex(text string, templates []string) string {
	namePattern := `(\p{Lu}\p{L}*(?:[-'.]\p{L}+)*\.?(?:[ \t]+\p{Lu}\p{L}*(?:[-'.]\p{L}+)*\.?)*)`
	
	bestIndex := -1
	bestName := ""
	
	for _, temp := range templates {
		pattern := strings.Replace(temp, "%s", namePattern, 1)
		re := regexp.MustCompile(pattern)
		
		matches := re.FindAllStringSubmatchIndex(text, -1)
		for _, match := range matches {
			// match is [start, end, group_start, group_end]
			if len(match) >= 4 {
				start := match[0]
				gStart := match[2]
				gEnd := match[3]
				if gStart != -1 && gEnd != -1 {
					name := strings.TrimSpace(text[gStart:gEnd])
					// Avoid common stop words/noise
					nameLower := strings.ToLower(name)
					stopWords := map[string]bool{
						"the": true, "a": true, "an": true, "she": true, "he": true, "they": true,
						"it": true, "this": true, "that": true, "these": true, "those": true,
						"who": true, "which": true, "what": true, "we": true, "you": true,
						"our": true, "his": true, "her": true, "their": true, "its": true,
					}
					if stopWords[nameLower] {
						continue
					}
					// We want the match that starts earliest in the text
					if bestIndex == -1 || start < bestIndex {
						bestIndex = start
						bestName = name
					}
				}
			}
		}
	}
	
	return bestName
}

func main() {
	ceoTemplates := []string{
		`(?i:ceo\s+of\s+[\w\-']+\s+is\s+)%s`,
		`(?i:ceo\s+is\s+)%s`,
		`(?i:ceo\s*:\s*)%s`,
		`(?i:chief\s+executive\s+officer\s+is\s+)%s`,
		`(?i:chief\s+executive\s+is\s+)%s`,
		`(?i:leader\s+is\s+)%s`,
		`%s\s+(?i:is\s+the\s+ceo\s+of\s+[\w\-']+)`,
		`%s\s+(?i:is\s+the\s+co-founder\s+and\s+ceo\s+of\s+[\w\-']+)`,
		`%s\s+(?i:is\s+the\s+co-founder\s+and\s+ceo)`,
		`%s\s+(?i:is\s+co-founder\s+and\s+ceo)`,
		`%s\s+(?i:is\s+the\s+ceo)`,
		`%s\s+(?i:is\s+ceo)`,
		`%s\s+(?i:leads\s+the\s+company)`,
	}
	
	text := `AI Overview
ALI GHODSI is the Co-founder and CEO of Databricks. 
As of early 2026, he leads the data and AI company, which was valued at over $130 billion. Ghodsi took over as CEO in January 2016 and is known for his focus on "data intelligence" and the shift towards AI agents in enterprise. 
Key 2026 Insights from Ali Ghodsi:
AI Agent Shift: He reports that 80% of databases on the Databricks platform are now built by AI agents, not humans.
Market Outlook: Ghodsi is preparing for a potential tech downturn similar to 2022, leading to significant capital raising for the company.
Talent Strategy: He emphasizes paying "top of the market" for AI talent.
Data Focus: He believes the AI race will be won by whoever best manages enterprise data, rather than just having the best model. 
He is a former academic who co-founded Databricks with engineers from UC Berkeley. 
Watch Databricks CEO: We Pay Top of the Market for Talent`

	extracted := extractNameWithRegex(text, ceoTemplates)
	fmt.Printf("Extracted: %q\n", extracted)
}
