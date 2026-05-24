//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	// Find absolute path of sge_dom_dump.html
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Getwd failed: %v", err)
	}
	htmlPath := filepath.Join(wd, "scratch", "sge_dom_dump.html")
	fileURL := "file://" + filepath.ToSlash(htmlPath)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()

	ctx, cancelTimeout := context.WithTimeout(ctx, 15*time.Second)
	defer cancelTimeout()

	log.Printf("Loading local SGE dump: %s", fileURL)

	err = chromedp.Run(ctx,
		chromedp.Navigate(fileURL),
	)
	if err != nil {
		log.Fatalf("Navigation failed: %v", err)
	}

	extractorJS := `(() => {
		const aiContainer = document.querySelector('.s7d4ef');
		if (!aiContainer) return "❌ SGE container not found";
		
		// Clone container
		const clone = aiContainer.cloneNode(true);
		
		// Remove UI elements
		const toRemove = clone.querySelectorAll('button, svg, style, script, [role="dialog"]');
		toRemove.forEach(el => el.remove());
		
		// Format code blocks
		const preBlocks = clone.querySelectorAll('pre');
		preBlocks.forEach(pre => {
			const codeText = pre.innerText;
			let lang = '';
			if (codeText.includes('package main') || codeText.includes('func main()') || codeText.includes('go ')) {
				lang = 'go';
			} else if (codeText.includes('fn main()') || codeText.includes('let mut') || codeText.includes('impl ')) {
				lang = 'rust';
			} else if (codeText.includes('def ') || (codeText.includes('import ') && codeText.includes(':\n'))) {
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
			marker.innerText = '\n' + String.fromCharCode(96) + String.fromCharCode(96) + String.fromCharCode(96) + lang + '\n' + codeText + '\n' + String.fromCharCode(96) + String.fromCharCode(96) + String.fromCharCode(96) + '\n';
			pre.parentNode.replaceChild(marker, pre);
		});
		
		// Format tables
		const tables = clone.querySelectorAll('table');
		tables.forEach(table => {
			let mdTable = '\n';
			const rows = table.querySelectorAll('tr');
			rows.forEach((row, rowIndex) => {
				const cols = row.querySelectorAll('th, td');
				let mdRow = '|';
				cols.forEach(col => {
					mdRow += ' ' + col.innerText.replace(/\n/g, ' ').trim() + ' |';
				});
				mdTable += mdRow + '\n';
				if (rowIndex === 0) {
					let mdSep = '|';
					cols.forEach(() => {
						mdSep += ' --- |';
					});
					mdTable += mdSep + '\n';
				}
			});
			mdTable += '\n';
			const marker = document.createElement('div');
			marker.innerText = mdTable;
			table.parentNode.replaceChild(marker, table);
		});

		// Append to body for layout calculation (innerText requires layout)
		clone.style.position = 'absolute';
		clone.style.left = '-9999px';
		clone.style.top = '-9999px';
		document.body.appendChild(clone);

		let text = clone.innerText;
		clone.remove();
		
		// Clean up multiple consecutive newlines
		text = text.replace(/\n{3,}/g, '\n\n');
		return text.trim();
	})()`

	var result string
	err = chromedp.Run(ctx,
		chromedp.Evaluate(extractorJS, &result),
	)
	if err != nil {
		log.Fatalf("Evaluation failed: %v", err)
	}

	fmt.Println("\n=== EXTRACTED TEXT CONTENT ===")
	fmt.Println(result)
	fmt.Println("==============================")
}
