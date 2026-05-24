# Implementation Plan: USQL Language Engine & WI-OS Semantic Kernel

This plan details the design and native Go implementation of the **UltraSearch Query Language (USQL)** compiler and the **Web-Intelligence Operating System (WI-OS) Kernel**. It establishes a dynamic, zero-trust registry that resolves query templates at scale using a semantic router, and introduces a safe AI-Human validation pipeline for community-contributed templates.

## User Review Required

> [!IMPORTANT]
> **No Executable Code in Contributions:** To prevent malicious or careless community modifications, all user/community templates are strictly restricted to **declarative markdown Skill Books**. There is absolute separation: users define schemas and search drivers in markdown, and our Go core OS compiles and executes them. No arbitrary code execution is mathematically possible.
> 
> **AI-Human Validation Pipeline:** All newly contributed Skill Books must pass through a two-stage check:
> 1. *AI Sandbox Audit:* Checks dorks and schemas against known cyber threat vectors (creds mining, payload exfiltration, Pastebin breaches).
> 2. *Human-in-the-Loop Approval:* Staged in an `unverified/` buffer, requiring manual human verification before being loaded into the active OS kernel registry.
> 
> **Semantic Router (Auto-Selection at Scale):** Bypasses the tedious manual file navigation. The WI-OS kernel uses our pure Go TF-IDF similarity engine to map incoming natural language search queries directly to the best-fit Skill Book dynamically, resolving schemas and dorking rules in $O(1)$ lookup time.

---

## Proposed Subsystems Architecture

We will implement three Go-native files in the root of the repository:
1. **`usql.go` (The USQL Compiler & AST Engine):**
   - **Lexer/Scanner:** Tokenizes inputs into typed tokens: `SEARCH`, `FROM`, `RETURN`, `WITH`, `CACHE`, `FORMAT`, keys, strings, and types.
   - **Parser:** Validates syntax and generates an Abstract Syntax Tree (AST).
   - **Execution Engine:** Evaluates AST properties and binds them to our Go Scraping backend loops.
2. **`registry.go` (WI-OS Dynamic Skill Book Registry & Semantic Router):**
   - **Frontmatter Parser:** Parses Skill Book markdown frontmatter (`name`, `version`, `domains`, `trust_tier`).
   - **Semantic Router:** Employs our pure Go `CosineSimilarity` engine to match user query terms against Skill Book `domains` and metadata to auto-select schemas at scale.
3. **`contribution.go` (AI-Human Validation Pipeline):**
   - **Intake Gateway:** Command-line hook (`./ultrasearch -integrate-skill <path>`) that parses new books.
   - **Dry-run Security Check:** Evaluates dork templates against injection and cyber-attack rules.
   - **Verification Buffer:** Saves unverified books to `/ai_skills/unverified/` until explicitly promoted.

---

## Technical Specifications

### 1. Abstract Syntax Tree (AST) Structure in Go
```go
type USQLQuery struct {
    SearchEntity string                 // e.g. "company:Databricks"
    Sources      []string               // e.g. ["crunchbase", "pitchbook"]
    ReturnFields map[string]string      // e.g. {"ceo": "string", "latest_valuation": "number"}
    Confidence   float64                // e.g. 0.8
    CacheTTL     time.Duration          // e.g. 24h
    Format       string                 // e.g. "json"
}
```

### 2. Lexical Grammar
- **Keywords:** `SEARCH`, `FROM`, `RETURN`, `WITH`, `CACHE`, `FORMAT`
- **Symbols:** `:`, `,`, `{`, `}`, `>`, `;`
- **Values:** Identifiers, quoted strings (`"Databricks"`), numbers, durations (`24h`)

---

## Proposed Changes

### [NEW] [usql.go](file:///Users/ramcharan/Desktop/UltraSearch/usql.go)
Contains the lexical scanner, parser, AST generation, and execution compiler that binds USQL queries directly to the Go search engine.

### [NEW] [registry.go](file:///Users/ramcharan/Desktop/UltraSearch/registry.go)
Manages Skill Book cataloging, live in-memory registry, frontmatter parsing, and the dynamic TF-IDF semantic query router.

### [NEW] [contribution.go](file:///Users/ramcharan/Desktop/UltraSearch/contribution.go)
Implements the secure AI-Human validation pipeline, dry-run safety auditor, and human promotion staging endpoints.

### [MODIFY] [main.go](file:///Users/ramcharan/Desktop/UltraSearch/main.go)
Integrate the USQL CLI and server interface so developers can run raw USQL queries or trigger skill integrations directly from the terminal.

---

## Verification Plan

### Automated Compiler Tests
- Compile and scan complex USQL statements and verify that the AST maps precisely.
- Place a mock Skill Book in `/ai_skills/unverified/` and run the contribution intake script, validating that the dry-run audits flag toxic patterns (e.g. trying to search `pastebin.com/raw` for `db_password`).
- Query the Semantic Router with natural terms like "acquisitions", "funding", "ceos", and verify that it correctly maps to `pe_deal_intelligence` with high similarity scores.
