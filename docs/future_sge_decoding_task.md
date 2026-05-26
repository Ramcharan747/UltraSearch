# Future Task: 10K Query Google SGE Decoding & Retrieval Analysis

This document outlines the roadmap for executing a large-scale SGE retrieval research study once the core UltraSearch product is fully built.

## Objective
Decode Google Search Generative Experience (SGE / AI Overviews) trigger thresholds, safety boundaries, and retrieval mechanisms by executing a mass diagnostic query suite across multiple domains and query styles.

## Operational Plan

### 1. Agent Cluster Setup
* **Scale:** Spawn a swarm of 20 to 50 concurrent agents.
* **Specialization:** Each agent is assigned to a specific query domain (e.g., medical, financial, technical, culinary, shopping, corporate news).
* **Role:** Each agent dynamically generates specialized prompt variants (SGE-friendly, interrogative prefixes, implicit synthesis prompts, and "soft jailbreaks" that test boundary limits without crossing Google's actual abuse or legal lines).

### 2. Mass Query Execution
* **Dataset:** Execute a standardized bank of **10,000 distinct queries** through the UltraSearch CLI.
* **Retriever Configuration:** Force `aiMode: "only"` to target SGE/AI Overviews exclusively.
* **Worker Allocation:** Scale up to 10-20 parallel ChromeDP worker threads with rotating stealth sessions from our pool to prevent rate limits.

### 3. Response Analysis & Decoding
For every query, the system will log:
1.  **Raw Input Query** vs. **Normalized SGE Query**.
2.  **SGE Response Status:** Triggered (Success), Blank (Null), Refusal (Safety Block), or General Search Fallback.
3.  **Synthesized Text vs. Citations:** Which domains were cited and how the content correlates with the source.
4.  **Formatting Alignment:** Did SGE output the exact structural JSON/Markdown requested?

## Success Metrics & Outcome
Develop a complete **SGE Synthesis Trigger Matrix** mapping:
* Exact word patterns that guarantee AI Overview synthesis.
* The exact YMYL boundaries for finance and health topics.
* Optimal prompt constraints for forcing structured data output.
