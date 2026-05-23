package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	f, err := os.Open("scratch/sge_dom_dump.html")
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer f.Close()

	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		log.Fatalf("Failed to parse HTML: %v", err)
	}

	// Find the selection for "Pros"
	doc.Find("div, span, li, p").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text == "Pros" || text == "Cons" {
			class, _ := s.Attr("class")
			fmt.Printf("Element [%d] <%s class=%q>: %s\n", i, goquery.NodeName(s), class, text)
			// Let's print the parent's HTML structure
			parent := s.Parent()
			parentClass, _ := parent.Attr("class")
			fmt.Printf("  Parent: <%s class=%q>\n", goquery.NodeName(parent), parentClass)
			
			// Print siblings text
			parent.Children().Each(func(j int, sibling *goquery.Selection) {
				sibText := strings.TrimSpace(sibling.Text())
				sibTagName := goquery.NodeName(sibling)
				sibClass, _ := sibling.Attr("class")
				fmt.Printf("    Sibling [%d] <%s class=%q>: %s\n", j, sibTagName, sibClass, strings.ReplaceAll(sibText, "\n", " "))
			})
		}
	})
}
