//go:build ignore

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

	// We want to find the main content body.
	// Let's print all elements that have text inside .s7d4ef, but formatted nicely
	fmt.Println("=== WALKING DOM ===")
	walk(doc.Find(".s7d4ef"), 0)
}

func walk(s *goquery.Selection, depth int) {
	s.Children().Each(func(i int, child *goquery.Selection) {
		tagName := goquery.NodeName(child)
		class, _ := child.Attr("class")
		id, _ := child.Attr("id")
		
		// If it's a leaf node or has direct text, let's print it
		text := strings.TrimSpace(child.Text())
		
		// We only care about printing nodes that contain text
		if len(text) > 0 {
			// Print details
			indent := strings.Repeat("  ", depth)
			
			// Show node type
			tagDesc := tagName
			if class != "" {
				tagDesc += "." + strings.ReplaceAll(class, " ", ".")
			}
			if id != "" {
				tagDesc += "#" + id
			}

			// If it has children, print its tag and recurse
			if child.Children().Length() > 0 {
				// Only print if the text is not identical to the children's concatenated text (to avoid duplication)
				// Actually, let's print the tag name, and if it has direct text nodes, print them.
				directText := ""
				child.Contents().Each(func(j int, node *goquery.Selection) {
					if goquery.NodeName(node) == "#text" {
						directText += node.Text()
					}
				})
				directText = strings.TrimSpace(directText)
				if directText != "" {
					fmt.Printf("%s<%s> (direct: %q)\n", indent, tagDesc, directText)
				} else {
					fmt.Printf("%s<%s>\n", indent, tagDesc)
				}
				walk(child, depth+1)
			} else {
				// Leaf node with text
				fmt.Printf("%s<%s>: %q\n", indent, tagDesc, text)
			}
		}
	})
}
