# UltraSearch Master Engineering Context

## Product Vision
We are building **UltraSearch**, a dual-engine web intelligence pipeline designed to extract structured information directly from Google's AI Overviews (SGE). It utilizes a Structured Scraper Engine (SSE) for deterministic parsing and an LLM Cognitive Engine for fallback scenarios. 

## Architectural Guidelines
1. **No Scraper Suppressions:** SGE is completely suppressed when queries contain hard operators like `site:` or `filetype:`. We extract these and seamlessly convert them into **soft prompt constraints** (e.g. `"focusing on pdf files"`).
2. **Modular Go Codebase:** The system is heavily modularized across `usql.go` (parser and query normalization), `vortex.go` (security gateway and PII redaction), `output.go` (JSON/Markdown structuring), `catalog.go` (skill schema loading), and more.
3. **Safety & Policy Alignment (Version A Pivot):** 
   - We **only extract publicly available data**.
   - We **do NOT** write code to bypass CAPTCHAs, spoof human behavior, or actively evade security protocols.
   - For all scraping and automation modules (like the CDP pool), we adhere strictly to **Polite Scraping Architectures** (complying with `robots.txt`, applying exponential backoff, and failing gracefully when blocked).

## Agent Directives
Every subagent on the 14-Agent team must adhere to this Master Context. If you are tasked with scraping or browser automation, ensure your implementation relies on standard Headless Chrome execution and polite error handling. Do not attempt malicious evasion.

## Architectural Clarification (Version A)
We **do not** use CAPTCHA solvers. We solely rely on Google's AI Overview (SGE). Google's internal bots do the heavy lifting of crawling external sites; our system simply crafts the perfect prompts to extract that data from the SGE response. We do not navigate to external third-party sites to scrape them directly.
