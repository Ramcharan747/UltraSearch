# UltraSearch: Master Engineering & Architecture Guide v2.0

> **This document is the single source of truth for ALL subagents working on UltraSearch.**
> Read it completely before starting any work. If you're unclear on any constraint, ASK — don't assume.

---

## 1. PRODUCT VISION

**UltraSearch** is a dual-engine query pipeline that extracts structured, verified data from **Google's AI Overview (SGE)**. We don't scrape third-party websites. We prompt Google, and Google's LLM (which already has access to its entire index) does the heavy lifting. We just:

1. **Craft the optimal query** (USQL compiler + prompt optimizer)
2. **Capture the SGE response** (headless Chrome CDP pool)
3. **Parse and structure the output** (DOM optimizer + CSS selector engine + LLM fallback)
4. **Validate and secure the data** (Vortex security gateway + PII scrubber + consensus engine)

### The Big Insight
Google's AI Overview is essentially a free, publicly available API to Google's entire knowledge graph. The challenge is **prompt engineering** — knowing how to ask to get the right response. This is what our 10K Query Boundary Research Study is mapping.

---

## 2. NON-NEGOTIABLE COMPLIANCE RULES (Version A)

> **CRITICAL: Every subagent MUST follow these. Violating any of these is grounds for immediate task rejection.**

| Rule | Description |
|------|-------------|
| **Public Data Only** | We ONLY parse data returned via standard Google search queries. Zero access to private APIs, internal systems, or gated content. |
| **No CAPTCHA Solvers** | We do NOT use any CAPTCHA solving services (2Captcha, AntiCaptcha, etc.). If Google blocks us, we back off. |
| **No Proxy Rotation** | We do NOT use residential proxies, datacenter proxy pools, or IP rotation to evade detection. |
| **No Human Mimicry** | We do NOT simulate human mouse movements, scroll patterns, or typing to bypass bot detection. |
| **No Direct Scraping** | We do NOT navigate to third-party websites (Crunchbase, LinkedIn, etc.) to scrape their HTML. Google's own crawlers already have this data indexed. |
| **Polite Client Protocol** | Rate limiting (3-10s between queries), exponential backoff on errors, and graceful degradation if blocked. |
| **No Login/Auth Bypass** | We do NOT attempt to log into any service or bypass authentication walls. |

**What we DO**: We send search queries to Google, capture the public AI Overview response, and parse it. That's it. Google's LLM does the work, not us.

---

## 3. ARCHITECTURE OVERVIEW

```
┌─────────────────────────────────────────────────────────────────┐
│                        USER INPUT                                │
│    USQL Query: SEARCH "NVIDIA" RETURN {revenue, market_cap}      │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────────┐
│                   USQL COMPILER (usql.go, parser.go)              │
│  • Parse AST (brackets, schemas, union types)                     │
│  • Strip hard operators (site:, filetype:) → soft constraints     │
│  • Compile final optimized search prompt                          │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────────┐
│                   VORTEX SECURITY ROUTER (vortex.go)              │
│  • PII scrubbing (emails, SSNs, API keys)                         │
│  • Adversarial injection detection                                │
│  • Source reputation scoring                                      │
│  • Cosine similarity correlation check                            │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────────┐
│              QUERY EXECUTION LAYER (browser.go, dag_executor.go)  │
│  • Headless Chrome CDP context pool                               │
│  • Concurrent DAG-based query scheduler                           │
│  • Rate limiting & exponential backoff                            │
│  • Google AI Overview (SGE) capture                               │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────────┐
│              DUAL ENGINE SELECTION (gateway logic in main.go)     │
│                                                                   │
│  ┌─────────────────────┐  ┌────────────────────────────────┐     │
│  │  STRUCTURED SCRAPER │  │  COGNITIVE LLM FALLBACK        │     │
│  │  ENGINE (SSE)       │  │  ENGINE (fallback.go)          │     │
│  │  • DOM optimization │  │  • LLM-based text parsing      │     │
│  │  • CSS selectors    │  │  • Unstructured synthesis       │     │
│  │  • Canary validation│  │  • Auto-activates on SSE drift │     │
│  └─────────┬───────────┘  └──────────┬─────────────────────┘     │
│            │                          │                           │
│            └──────────┬───────────────┘                           │
│                       │                                           │
└───────────────────────┼──────────────────────────────────────────┘
                        │
                        ▼
┌──────────────────────────────────────────────────────────────────┐
│              POST-PROCESSING PIPELINE                             │
│  • Bayesian Consensus Engine (consensus.go)                       │
│  • Naive Bayes Pre-Screener (ml_classifier.go)                    │
│  • Evolutionary Prompt Optimizer (optimizer_loop.go)               │
│  • Output Formatter (output.go) → JSON/Table/Markdown             │
│  • Telemetry Logger (telemetry.go) → Prometheus metrics           │
└──────────────────────────────────────────────────────────────────┘
```

---

## 4. FILE MAP & COMPONENT DETAILS

All Go code is in `/Users/ramcharan/Desktop/UltraSearch/` under `package main`.

### 4.1 Query Translation & Parser
| File | Purpose | Lines |
|------|---------|-------|
| `usql.go` | USQL AST parser, soft constraint compiler, query normalizer | ~39K bytes |
| `parser.go` | Core parsing utilities | ~1K bytes |
| `usql_test.go` | USQL unit tests | ~2K bytes |

**Key Design Decision**: Google SGE *suppresses* AI Overviews when strict search operators (`site:`, `filetype:`, `inurl:`) are present. Our normalizer converts these to natural language "soft constraints":
- `site:arxiv.org` → `"focusing on arxiv.org publications"`
- `filetype:pdf` → `"from PDF documents"`
- `intitle:revenue` → `"with revenue in the title"`

### 4.2 Security & Gateway
| File | Purpose | Lines |
|------|---------|-------|
| `vortex.go` | PII scrubber, injection detector, reputation registry | ~14K bytes |
| `quarantine.go` | CSS selector drift detection + canary validation | ~5K bytes |
| `classifier.go` | Query classification and routing | ~7K bytes |

### 4.3 Execution Layer
| File | Purpose | Lines |
|------|---------|-------|
| `browser.go` | Headless Chrome CDP context pool | ~3K bytes |
| `dag_executor.go` | Concurrent DAG-based query scheduler | ~19K bytes |
| `http_search.go` | HTTP-based search execution | ~14K bytes |
| `main.go` | CLI entry point, server mode, all integration | ~93K bytes |

### 4.4 Intelligence Layer
| File | Purpose | Lines |
|------|---------|-------|
| `ml_classifier.go` | Local Naive Bayes SGE pre-screener | ~8K bytes |
| `consensus.go` | Bayesian fact consensus + graph reconciliation | ~27K bytes |
| `optimizer_loop.go` | 50-generation evolutionary prompt optimizer | ~19K bytes |

### 4.5 Output & Telemetry
| File | Purpose | Lines |
|------|---------|-------|
| `output.go` | JSON/Table/Markdown formatters | ~4K bytes |
| `registry.go` | Standard library functions (CLEAN_PII, CONVERT_CURRENCY) | ~8K bytes |
| `function_registry.go` | Function registry for USQL | ~5K bytes |
| `telemetry.go` | Prometheus metrics + zerolog | ~2K bytes |
| `catalog.go` | Skill book catalog | ~5K bytes |
| `contribution.go` | Community contribution gateway | ~10K bytes |
| `fallback.go` | Cognitive LLM fallback engine | ~3K bytes |
| `dom.go` | DOM optimization (strip scripts/styles/SVGs) | ~3K bytes |

### 4.6 Tests
| File | Purpose |
|------|---------|
| `ml_classifier_test.go` | ML pre-screener unit tests |
| `dag_executor_test.go` | DAG scheduler unit tests |
| `consensus_test.go` | Consensus engine unit tests |
| `optimizer_loop_test.go` | Evolutionary optimizer tests |

---

## 5. 10K QUERY BOUNDARY RESEARCH STUDY

### 5.1 Purpose
Map exactly how Google's AI Overview responds to different query formulations. We want to know:
- Which queries trigger SGE vs standard search results?
- How does query size affect response quality?
- Which "magic words" alter responses?
- How do persona/role framing affect output?
- Which jailbreak patterns produce the best results?

### 5.2 Query Generation (COMPLETE ✅)
Generated **10,000 queries** across:

| Dimension | Count | Details |
|-----------|-------|---------|
| Domains | 15 | business, tech, medicine, legal, science, education, engineering, arts, social sciences, agriculture, environment, sports, travel, consumer, government |
| Personas | 12 | neutral, researcher, analyst, journalist, student, professional, curious, expert, fact-checker, decision-maker, developer, medical |
| Output Styles | 8 | default, table, JSON, bullets, comparison, step-by-step, sources, pros/cons |
| Jailbreak Categories | 11 | baseline, persona, output-control, authority-anchoring, multi-perspective, constraint-stacking, decomposition, adversarial, temporal, geographic, meta |
| Sizes | 3 | small (3-8 words), medium (15-40 words), large (60-200 words) |
| Magic Word Categories | 8 | precision, comprehensiveness, recency, authority, exclusion, trigger, quantity, temporal |

### 5.3 Query Files
```
/Users/ramcharan/Desktop/UltraSearch/query_research/
├── taxonomy.json                    # Full taxonomy config
├── jailbreak_templates.json         # 88 prompt templates
├── generator.py                     # Generator script
├── collector.py                     # SGE response collector
├── analyzer.py                      # Statistical analyzer
├── queries/
│   ├── all_queries.jsonl            # Master file (10K)
│   ├── batch_0001.jsonl             # Queries 1-500
│   ├── ...
│   └── batch_0020.jsonl             # Queries 9501-10000
│   └── generation_stats.json        # Distribution stats
├── results/                         # SGE responses (pending)
└── analysis/                        # Analysis reports (in progress)
```

### 5.4 Next Steps
1. **Quality Audit** — 5 analysis agents are reviewing query quality
2. **Canary Test** — Run first 100 queries against real SGE
3. **Full Execution** — Process all 10K queries with rate limiting
4. **Statistical Analysis** — Identify trigger patterns and magic words
5. **Report Generation** — Final comprehensive research report

---

## 6. CLI REFERENCE

```bash
# Basic query
./ultrasearch --query "NVIDIA revenue 2024"

# AI Overview only (SGE capture mode)
./ultrasearch --only-ai --query "NVIDIA revenue 2024"

# Fast AI mode (AI Overview + URLs, no deep scrape)
./ultrasearch --fast-ai --query "NVIDIA revenue 2024"

# USQL structured query
./ultrasearch --usql 'SEARCH "NVIDIA" RETURN {revenue: number, market_cap: number}'

# Evolutionary optimizer
./ultrasearch --optimize --optimize-generations 50

# Diagnostics
./ultrasearch --vortex-diag

# HTTP server mode
./ultrasearch --serve --port 8080

# Stress test
./ultrasearch --stress --stress-count 30 --stress-concurrency 2
```

---

## 7. SUBAGENT COORDINATION (Nexus)

Subagents persist state via the nexus directory:
```
/Users/ramcharan/Desktop/UltraSearch/nexus/
├── master_engineering_guide.md      # THIS FILE
├── <agent_name>/
│   ├── memory.md                    # Short-term context
│   ├── soul.md                      # Behavioral persona
│   └── logs.md                      # Activity trail
└── global_streamline.md             # Team billboard
```

### Rules for Subagents
1. **Always read this guide first** before starting work
2. **Write modular code** — no God functions, no 500-line blocks
3. **Test everything** — every new function needs a test
4. **Stay in `package main`** — all Go files in root directory
5. **Name functions uniquely** — prefix with module name (e.g., `MLClassifierTokenize`)
6. **Never violate Version A compliance** — if in doubt, ask
7. **Log your work** — update your nexus memory and logs after each task
8. **Report blockers immediately** — don't spin on problems

---

## 8. GLOSSARY

| Term | Meaning |
|------|---------|
| **SGE** | Search Generative Experience — Google's AI Overview in search results |
| **USQL** | UltraSearch Query Language — our custom query syntax |
| **SSE** | Structured Scraper Engine — deterministic CSS selector-based parsing |
| **Cognitive Mode** | LLM-based fallback for unstructured parsing |
| **Vortex** | Security gateway (PII scrub + injection detection + reputation) |
| **Canary** | Staging pipeline for CSS selector validation |
| **Soft Constraint** | Natural language replacement for suppressed search operators |
| **CDP** | Chrome DevTools Protocol — browser automation interface |
| **DAG** | Directed Acyclic Graph — query decomposition scheduler |
| **Magic Words** | Specific terms that alter SGE response characteristics |
| **Nexus** | Inter-agent coordination directory |

---

*Last Updated: 2026-05-24 | Version 2.0 | Maintainer: CEO Agent*
