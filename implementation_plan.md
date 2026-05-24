# Implementation Plan: USQL v4.0 & WI-OS Kernel v4.0 Architecture

This plan establishes the comprehensive architecture of **USQL v4.0 (Enterprise Structured Query Protocol)** and the **WI-OS Kernel v4.0 (Dynamic Search Operating System)**. It scales the language to handle complex nested schemas and dynamic function execution, and introduces a modular package manager for hot-reloading community-contributed templates.

## User Review Required

> [!IMPORTANT]
> **Dynamic Function Expressions (Grammar Expansion):** We are upgrading the USQL parser to handle nested function calls within the `RETURN` block—e.g., `RETURN { ceo: UPPER(string), arr: ESTIMATE_ARR(CONVERT_CURRENCY(number, 'EUR')) }`. By implementing a **Global Function Registry** in Go, developers can add custom functions, scaling the language standard library to 100,000+ functions without bloating the compiler.
> 
> **Recursive Nested Schema Parsing:** USQL v4.0 supports deep, multi-level nested JSON schema declarations natively—e.g., `RETURN { key_personnel: array({ name: string, title: string, background: { school: string } }) }`. The Go parser will recursively scan and validate these structures.
> 
> **Dynamic Package Manager & Hot-Loader:** Bypasses manual cataloging entirely. WI-OS Kernel v4.0 runs a live background directory watcher. When a developer drops a new `.md` Skill Book or packages are downloaded via the community client (`./ultrasearch -install <github-url>`), the kernel dynamically hot-reloads them in-memory without connection dropouts.
> 
> **AI-Human Sandbox Quarantine:** Every community package is quarantined in `unverified/`. The AI Sandbox checks dorks, domains, and properties for security breaches before queuing them for a 1-click human promotion command.

---

## Technical Specifications & Architecture

### 1. Extended USQL AST with Function Call Nodes
```go
type USQLQuery struct {
    SearchEntity string
    Sources      []string
    ReturnFields map[string]interface{} // Support nested maps/fields
    Confidence   float64
    CacheTTL     time.Duration
    Format       string
    Language     string
    Country      string
    SafeSearch   string
}

type FuncExpr struct {
	Name string
	Args []interface{}
	Type string // Return type
}
```

### 2. Lexical Grammar Upgrades
- **Symbols added:** `(`, `)`
- **Grammar structure:** `FieldName: FunctionName(Arg1, Arg2)`

---

## Proposed Changes

### [MODIFY] [usql.go](file:///Users/ramcharan/Desktop/UltraSearch/usql.go)
- Update Lexer to tokenize parentheses `(` and `)`.
- Refactor the `RETURN` parser block to execute **recursive parsing** of nested braces `{}` and parentheses `()` for function expressions.
- Implement the `GlobalFunctionRegistry` and pre-load common enterprise primitives (`UPPER`, `LOWER`, `CONVERT_CURRENCY`, `ESTIMATE_ARR`, `CLEAN_PII`).

### [MODIFY] [registry.go](file:///Users/ramcharan/Desktop/UltraSearch/registry.go)
- Implement background folder scanning that detects filesystem writes in `ai_skills/` and dynamically triggers `LoadSkillBookRegistry` in-memory.
- Partition the semantic index into partitioned prefix buckets to handle highly scalable queries across thousands of active Skill Books.

### [MODIFY] [contribution.go](file:///Users/ramcharan/Desktop/UltraSearch/contribution.go)
- Implement an automated HTTPS package downloader (`DownloadCommunitySkill(gitURL string)`) that pulls templates directly from public Git repository archives.
- Integrate the safety scanner to block any packages using arbitrary network requests, pastebin credentials mining, or shell commands inside SGE dorks.

### [MODIFY] [main.go](file:///Users/ramcharan/Desktop/UltraSearch/main.go)
- Expose the dynamic package installer flag `-install` and register the extended nested JSON-schema formatting printer.

---

## Verification Plan

### Compiler & Integration Tests
- Run unit scans on a massive query including nested function calls:
  `SEARCH company:"TLC Capital" RETURN { board: array({ name: UPPER(string) }), arr: ESTIMATE_ARR(number) }`
  Validate that the AST parses both the nested array schema and the `UPPER` function call.
- Trigger `./ultrasearch -install "https://github.com/Ramcharan747/UltraSearch-Skills-Archive"` to verify the HTTPS downloader stages packages cleanly in staging, runs the safety scanner, and promotes them upon approval.
