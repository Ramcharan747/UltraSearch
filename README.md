<div align="center">
  <img src="https://upload.wikimedia.org/wikipedia/commons/4/4e/Go_Logo_Blue.svg" alt="Golang" width="80"/>
  <h1>UltraSearch</h1>
  <p><b>The Unrestricted Tavily Alternative for Local AI Agents</b></p>
  
  [![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org)
  [![License](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)
  [![Stealth](https://img.shields.io/badge/Bypass-Cloudflare%20%7C%20DataDome-red.svg)](#stealth)
  [![Agents](https://img.shields.io/badge/AI%20Agents-OpenClaw%20%7C%20AutoGPT-purple.svg)](#api)
</div>

<br/>

**UltraSearch** is a self-hosted, unrestricted web search and extraction engine designed specifically for **Agentic Workflows** (OpenClaw, AutoGPT, LangChain, Cursor). 

Tired of API rate limits, expensive credits, and restrictive scraping policies from commercial tools like Tavily? UltraSearch runs entirely on your local machine, effortlessly bypassing enterprise-grade bot protections (Cloudflare Turnstile, DataDome) to feed raw, pristine, token-optimized data directly into your LLM's context window.

## 🔥 Why UltraSearch?

- **Local API Server (`--serve`)**: Instantly drop UltraSearch into any agent framework as a 1:1 replacement for external search APIs. 
- **LLM-Dense Output (`--output-format=llm-dense`)**: Generates hyper-compressed, whitespace-stripped HTML/Text chunks specifically engineered to maximize your LLM's context window.
- **VS Code Extension Integration**: Includes a native VS Code wrapper (`vscode-ultrasearch/`) so AI assistants like Cursor or GitHub Copilot can directly command the engine to research the web.
- **Automated Defensive Classification:** Intelligently probes targets and classifies them into 4 tiers (Static HTML, JS-Rendered, Bot-Protected, Login-Walled).
- **Human-Mimicry Solvers:** Features a robust ML-trained trajectory generator that simulates human cursor movements and input latency to flawlessly bypass zero-click CAPTCHAs.

---

## 🚀 Installation

Ensure you have [Go 1.21+](https://go.dev/) installed.

```bash
# Clone the repository
git clone https://github.com/Ramcharan747/UltraSearch.git
cd UltraSearch

# Install dependencies
go mod tidy

# Build the CLI tool
go build -o ultrasearch main.go classifier.go
```

---

## 💻 Usage

UltraSearch is designed to be fully controllable via CLI flags.

### Single Query
Run a stealth search on a single query and extract deep content from the top 5 results.
```bash
./ultrasearch -query "best python stealth scraping tools 2025" -limit 5
```

### Bundle Search (Bulk Queries)
Pass a text file (`queries.txt`) containing thousands of queries (one per line) to be processed in parallel.
```bash
./ultrasearch -bundle queries.txt -workers 10 -limit 10
```

### Lightweight Mode
Only extract search snippets and URLs, skipping the deep content extraction (lightning fast).
```bash
./ultrasearch -query "private equity SaaS acquisitions" -content=false
```

### CLI Arguments Reference

| Flag | Default | Description |
| :--- | :--- | :--- |
| `-query` | `""` | A single search query string to execute. |
| `-bundle` | `""` | Path to a text file containing queries (one per line). |
| `-limit` | `10` | Maximum number of search results to process per query. |
| `-workers` | `5` | Number of concurrent processing workers to spawn. |
| `-content` | `true` | Extract full page content (T1-T4). Set to `false` for URL/Snippet only. |
| `-output` | `"ultra_results.json"` | Path to save the extracted JSON data. |

---

## 🧠 Architecture: The 4-Tier Escalation Model

UltraSearch doesn't waste heavy browser resources on simple pages. It routes traffic intelligently:

1. **Tier 1 (Static):** Fast `net/http` extraction. Pure curl speed.
2. **Tier 2 (JS-Rendered):** Headless Chrome tabs for SPAs (React, Next.js).
3. **Tier 3 (Bot-Protected):** Non-headless Chrome with deep CDP stealth flags, fingerprint spoofing, and the trajectory solver engine.
4. **Tier 4 (Domain Persistence):** For aggressive Managed Challenges. Parks on the root domain, clears the firewall, and extracts target sub-pages invisibly via background JS `fetch()` to evade session resets.

---

## 🤝 Contributing
Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📝 License
Distributed under the MIT License. See `LICENSE` for more information.
