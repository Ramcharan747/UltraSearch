# Implementation Plan: Adaptive Query Routing & Sourcing Optimizer

This document establishes the architecture for the **Adaptive Query Routing & Sourcing Optimizer (AQRSO)** inside the UltraSearch compiler. It eliminates keyword-sorting bottlenecks and intent classification latency by implementing a local, deterministic **Entity-Attribute Semantic Switchboard** that intelligently rephrases and anchors search queries in a single pass. It also details the high-fidelity telemetry schema to support future local model training.

## User Review Required

> [!IMPORTANT]
> **Deterministic Entity-Attribute Switchboard:** We are replacing the keyword switch statements with a unified POS/Entity classifier inside `usql.go`. This classifier parses the user query, extracts the main Entity (e.g., "Databricks") and target Attributes (e.g., "CEO", "valuation"), and dynamically binds the optimal search anchors.
> 
> **Soft-Constraint Prompt Anchoring:** Instead of hard `site:` exclusions, we inject semantic anchors (e.g., `"according to official listings on Crunchbase"`) into SGE dorks. This guides Google's ranker to bubble up authoritative sources without causing hard search failures if a domain is unreachable or lacks a specific page.
> 
> **Supervised Training Logging (Telemetry):** To support future offline training of our own local models, we enforce a strict high-fidelity log structure inside `usage_telemetry.jsonl` containing:
> 1. Raw User Query
> 2. Output Format (JSON vs. Text)
> 3. Compiled Query Dork
> 4. Raw SGE Response (Ground Truth)
> 5. Parsed/Extracted Output (Labels)
> 
> This provides a complete dataset of optimized input-output pairs.

---

## Technical Specifications & Architecture

```
                  ┌──────────────────────────────┐
                  │ User Query + Output Format   │
                  │ (JSON or Text)               │
                  └──────────────┬───────────────┘
                                 │
                                 ▼
                  ┌──────────────────────────────┐
                  │ Entity-Attribute Switchboard │
                  │ (POS / Noun-Phrase Parser)   │
                  └──────────────┬───────────────┘
                                 │
                                 ├──────────────────────────────┐
                                 ▼                              ▼
                         [Format == JSON]               [Format == Text]
                       ┌───────────────────┐          ┌───────────────────┐
                       │ Inject Formatting │          │ Inject Summarizer │
                       │ Force Headers     │          │ Headers           │
                       └─────────┬─────────┘          └─────────┬─────────┘
                                 │                              │
                                 └──────────────┬───────────────┘
                                                │
                                                ▼
                               ┌────────────────────────────────┐
                               │ Soft-Constraint Sourcing       │
                               │ (Domain & Anchor Injector)     │
                               └────────────────┬───────────────┘
                                                │
                                                ▼
                                 [Google Search / SGE Page]
                                                │
                                                ▼
                               ┌────────────────────────────────┐
                               │ Edge-Side Extraction & Cleanup │
                               └────────────────┬───────────────┘
                                                │
                                                ▼
                               ┌────────────────────────────────┐
                               │ Telemetry Logger (JSONL)       │
                               └────────────────────────────────┘
```

### 1. Dynamic Routing & Query Expansion
We implement a lightweight rule registry in `usql.go` that maps general categories of intent to domains and prompt anchors:
* **Category: Business/Finance** (triggered by *valuation*, *funding*, *revenue*, *round*, *investor*, *ceo*, *founder*)
  * *Anchor:* `"according to profiles on Crunchbase, PitchBook, or SEC filings"`
* **Category: Academic/Science** (triggered by *paper*, *research*, *study*, *authors*, *arxiv*, *abstract*)
  * *Anchor:* `"as documented in peer-reviewed publications on arXiv, PubMed, or Nature"`
* **Category: Tech/Developer** (triggered by *library*, *package*, *code*, *compile*, *api*, *struct*)
  * *Anchor:* `"verified in developer docs, GitHub repositories, or StackOverflow"`

### 2. File Changes

#### [MODIFY] [usql.go](file:///Users/ramcharan/Desktop/UltraSearch/usql.go)
* Add `SourcingRule` struct and `SourcingRegistry` directory mapping category trigger keywords to query anchors.
* Implement `ExtractEntityAndAttributes(query string) (entity string, attributes []string)` to dynamically isolate proper nouns and intent markers from user inputs.
* Implement `CompileAdaptiveQuery(query string, format string) string` which combines the extracted entity, user attributes, selected profile prompt wrappers, and semantic search anchors.
* Integrate period trimming and horizontal space limits inside `extractNameWithTemplates` to clean up text matches cleanly.

#### [MODIFY] [main.go](file:///Users/ramcharan/Desktop/UltraSearch/main.go)
* Integrate `CompileAdaptiveQuery` into the CLI search pipeline and the API server endpoint.
* Update `LogQueryFailure` and the telemetry writer to dump the high-fidelity `usage_telemetry.jsonl` log containing user inputs, dorks, raw context, and final labels.

---

## Verification Plan

### Automated Verification
* Compile the binary: `go build -o ultrasearch`
* Run test queries representing different output profiles and verify they extract correct attributes in a single pass:
  * `./ultrasearch -usql "who is the founder of company:Databricks"` (Verify it outputs the name cleanly and logs the entry in `usage_telemetry.jsonl`).
  * `./ultrasearch -usql "valuation of company:Databricks"` (Verify it extracts the currency figures correctly).
* Inspect the telemetry logs to verify all fields are populated correctly.
