# 🤖 UltraSearch v4.0: AI Developer & Senior Systems Handbook

This manual serves as a high-fidelity architectural reference and troubleshooting log. If you are an AI assistant, compiler agent, or system engineer task-staged to modify, debug, or scale this codebase globally, read this document to understand the codebase structure, compilation paths, safety guidelines, and error-prevention steps.

---

## 📂 1. Directory Anatomy & Architecture Map

The entire UltraSearch edge engine is written in **native Go** as a zero-dependency, single-binary architecture. All core Go source files reside in the root directory under `package main`, allowing direct function resolution across compiler contexts without internal package import loops.

```
/UltraSearch
├── main.go               # Kernel bootstrap, CLI flags, ChromeDP executor, JSON printer
├── usql.go               # Hand-written Lexer, AST nodes, recursive-descent Parser, Go Std Library
├── registry.go           # Thread-safe RWMutex Skill Catalog, background hot-loader goroutine
├── contribution.go       # HTTPS ZIP downloader, quarantine sandbox, adversarial safety filter
├── vortex.go             # Vortex gateway: Prompt injection defense, forensic logger, telemetry client
├── http_search.go        # HTTP API Server endpoints for AI agents and external tools
├── classifier.go         # Native search categorizer and prefix keyword tries
├── ai_skills/            # Active declarative Markdown Skill Books
│   ├── pe_deal_intelligence.md
│   ├── academic_scholarly_mapping.md
│   └── government_fiscal_scrutiny.md
└── ai_skills/unverified/ # Quarantine staging area for community template package downloads
```

### 🔩 Component Roles & Interactions
1. **The Ingestion Pipeline:** `main.go` parses the CLI input (e.g. `-usql` or `-serve`).
2. **The Structured Compiler:** `usql.go` lexes the statement, checks parentheses and braces, and builds a recursive, typed AST query.
3. **The Semantic Router:** `registry.go` matches target keywords using cosine TF-IDF similarity to route queries to the correct declarative Skill Book in `ai_skills/` at scale.
4. **The Search Executor:** `main.go` compiles the AST into a Google dork query and initiates parallel browser threads (`runSearchPipeline`).
5. **The Security Gatekeeper:** `vortex.go` audits SGE responses, sanitizes null bytes/script tags, checks domain reputation overrides (bypassing `.gov` or `.edu`), and neutralizes prompt injections using cosine vector distance checks.
6. **The Telemetry Forensics:** `vortex.go` logs append-only verdicts in `usage_telemetry.jsonl` and transmits raw payloads under explicit `"opt_in_raw"` user consent.
7. **The Package Manager:** `contribution.go` pulls remote templates, runs sanitization, and queues them in `unverified/` for human verification.

---

## 📈 2. Core Upgrades Deep Dive (v4.0 Spec)

### 2.1 Recursive USQL AST & Grammar Upgrades
USQL v4.0 handles multi-level nested JSON schema formatting and custom paren function calls in the `RETURN` block.
*   **New Tokens:** `TokLeftParen` (`(`) and `TokRightParen` (`)`).
*   **Recursive Value Parser (`parseValue`):** Recursively resolves structural nested schema definitions:
    ```go
    type FuncExpr struct {
        Name string
        Args []interface{}
    }
    type NestedSchema struct {
        Fields map[string]interface{}
    }
    type ArraySchema struct {
        ValueType interface{}
    }
    ```
*   **Dual-Tier Function Resolution standard:** 
    *   **Tier 1 (Go Std Library):** Local high-speed functions (`UPPER`, `LOWER`, `TITLE`, `TRIM`, `CLEAN_PII`, `CONVERT_CURRENCY`, `ESTIMATE_ARR`) execute natively in Go on SGE responses using `EvaluateUSQLFunctions()`.
    *   **Tier 2 (Cognitive SGE Fallback):** Any custom community function (e.g. `CALCULATE_GROWTH`) is flattened and embedded into SGE's prompt instructions to be processed dynamically in the cloud.

### 2.2 Concurrency Concurrency & Registry Watcher
To support dynamic hot-reloading without dropping search connections, `registry.go` incorporates an in-memory thread lock standard:
*   **Read-Write Mutex:** `registryMutex sync.RWMutex` protects `GlobalRegistry` reads and writes.
*   **Acquisition Rules:**
    *   **Read (Semantic Router):** Acquire `registryMutex.RLock()` and `defer registryMutex.RUnlock()`.
    *   **Write (Hot-Loader):** Acquire `registryMutex.Lock()` and `defer registryMutex.Unlock()`.
*   **Zero-Dependency Directory Watcher:** `StartRegistryWatcher` runs a background polling loop checking the file count and sum of modification times in `ai_skills/` every 2 seconds, triggering `LoadSkillBookRegistry` seamlessly on changes.

### 2.3 Automated ZIP GitHub Installer
`DownloadCommunitySkill` retrieves repository archives via native Go streams:
1.  **Direct Download:** Ends with `.md` -> streams the raw markdown file directly.
2.  **Repository Extract:** Points to GitHub -> fetches `main.zip` or `master.zip`, extracts all `.md` files in-memory using `archive/zip`, runs sandbox validations, and stages safe books in `unverified/`.

---

## ⚠️ 3. Troubleshooting & Friction Log (AI Gating Guide)

If you are an AI attempting to modify or debug this engine, follow these strict directives to prevent compilations or search breakdowns:

### 3.1 Common Compilation Errors
*   **Undefined regexp/json/zip in files:**
    *   *Cause:* Go standard imports are declared per file. Adding code that uses `json.Marshal`, `regexp.MustCompile`, or `zip.NewReader` without updating the top `import` block of that specific file causes compilation failures.
    *   *Solution:* Always ensure target file headers contain:
        ```go
        import (
            "encoding/json"
            "regexp"
            "archive/zip"
            "bytes"
            "io"
            "net/http"
            // ...
        )
        ```
*   **Interface Type Mismatches in ReturnFields:**
    *   *Cause:* In older v3.0 specs, `ReturnFields` was a `map[string]string`. In v4.0, it is a `map[string]interface{}` to support nested structures.
    *   *Solution:* Do not treat `ReturnFields[key]` as a string. Assert types using `switch val := v.(type)` or `v.(string)`.

### 3.2 Thread-Safety Rules
*   **Never modify `GlobalRegistry` without locking:** Doing so will trigger concurrency race crashes during search queries.
*   *Incorrect:*
    ```go
    GlobalRegistry = append(GlobalRegistry, sb)
    ```
*   *Correct:*
    ```go
    registryMutex.Lock()
    GlobalRegistry = append(GlobalRegistry, sb)
    registryMutex.Unlock()
    ```

### 3.3 Quarantine Sandboxing Guidelines
*   The Contribution Gateway safety validator (`SandboxValidateSkillBook`) scans files for adversarial keywords (e.g. `pastebin.com/raw`, `attacker.com`, `ignore previous instructions`).
*   If your community template fails staging, verify that SGE dork prompts do not contain raw exfiltration URLs or blacklisted credentials terms.
*   Standard READMEs and markdown texts will fail staging if they do not contain valid YAML frontmatter blocks starting with `---`. This is by design to prevent non-SkillBook clutter from entering staging.

---

## 🚀 4. How to Safely Scale and Verify Changes

When executing modifications, follow this three-step validation pipeline:

1.  **Compile Check:**
    ```bash
    go build -o ultrasearch
    ```
    Ensure no syntax, undefined variable, or import errors.
2.  **Diagnostics Execution:**
    ```bash
    ./ultrasearch -vortex-diag
    ```
    This runs our robust end-to-end check block, validating:
    *   Dual-vector security anomaly and cohesion algorithms.
    *   Recursive token parsing of parenthesized function schemas.
    *   Dynamic standard library local evaluations.
    *   TF-IDF Semantic query routing and matching.
    *   Sandbox quarantining and human promotions.
3.  **Run Complex USQL Queries:**
    Test nested schema parsing via standard shell flags:
    ```bash
    ./ultrasearch -usql "SEARCH company:'Databricks' RETURN { founders: array({ name: UPPER(string), title: string }) }"
    ```
