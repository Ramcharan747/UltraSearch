---
name: academic_scholarly_mapping
version: 1.0.0
author: Ramcharan
trust_tier: core
domains: [academic_literature, scientific_research]
---

# Academic Scholarly Mapping Schema

This Skill Book defines the structured schema, reliability prioritizations, and query dorking patterns for mapping state-of-the-art computer science and artificial intelligence research papers indexed by Google Scholar and arXiv.

## 1. Schema Properties

Every query matching this skill book must return a JSON object conforming to the following schema:

*   **`exact_title`** (string, required): The official title of the academic publication or preprint.
    *   *Source Priority:* `arxiv` -> `google_scholar`
*   **`authors`** (array of strings, required): Full names of all co-authors listed on the docket.
    *   *Source Priority:* `arxiv` -> `pubmed`
*   **`lead_author_institution`** (string, optional): The university, research lab, or corporate lab affiliation of the primary author.
    *   *Source Priority:* `google_scholar` -> `arxiv`
*   **`arxiv_id`** (string, required): The unique numerical identifier in the arXiv index (e.g., `2411.08942`).
    *   *Source Priority:* `arxiv`
*   **`doi`** (string, optional): The Digital Object Identifier (DOI) if published in a peer-reviewed journal.
    *   *Source Priority:* `crossref` -> `arxiv`
*   **`abstract_summary`** (string, required): A concise, 3-sentence executive summary of the paper's findings.
    *   *Source Priority:* `arxiv`
    *   *Confidence Floor:* `0.7`
*   **`primary_methodology`** (string, optional): The core algorithms, mathematical architectures, or datasets introduced in the study.
    *   *Source Priority:* `arxiv`

## 2. Ingestion & Routing Drivers

When compiling USQL read calls, the routing layer maps target fields using these operators:

*   **VFS Target Path:** `/web/arxiv/papers/{id}`
*   **Scholarly Search Operator:** `site:arxiv.org/abs/{id} OR site:arxiv.org/pdf/{id} author abstract`
*   **Citation & Affiliation Operator:** `"{paper_title}" lead author university institution publications`
