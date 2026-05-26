# Google SGE Query Generation Skill Book
Version: 1.0.0
Author: Antigravity Code Partner

This document serves as the official **Skill Book** and instruction manual for generating the 10,000 SGE (Search Generative Experience / Google AI Overview) boundary-testing queries. It guarantees that any executing agent (or human developer) produces high-quality, realistic, semantically coherent, and grammatically correct queries to evaluate the Google AI Overview parser/decoder.

---

## 1. Core Philosophy of SGE Testing

The goal is to test how Google's AI Overview triggers, formats, and handles diverse query structures. A low-quality test set relies on simple slot-filling templates (e.g., repeating the same template 1,000 times with different company names). This is useless because it only tests a handful of sentence structures. 

Instead, we construct **10,000 completely unique queries** where each query represents a real human search intent (derived from historical datasets or live Google Trends) modified by prompt engineering, formatting directives, and adversarial boundaries.

---

## 2. Query Sizes & Task Complexities

We categorize search queries by length and task complexity to evaluate SGE across the full spectrum of user interaction.

### A. Very Small Queries (Keyword/Short Intent)
* **Word Count**: 2 to 8 words.
* **Goal**: Simulates typical daily Google search behavior.
* **Grammar/Structure**: Direct, noun-phrase-heavy, short, fragments. No wrapping or formatting constraints.
* **Example**: `NVIDIA revenue 2024` or `diabetes drug Ozempic clinical trial`.

### B. Small Queries (Natural Questions)
* **Word Count**: 10 to 30 words.
* **Goal**: Evaluates SGE's ability to answer clean, single-hop informational questions.
* **Grammar/Structure**: Standard English interrogative sentences. Clean syntax, ends with a question mark.
* **Example**: `How has the Federal Reserve's rate cut impacted SaaS startup valuations in 2026?`

### C. Medium Queries (Complex/Multi-Part Tasks)
* **Word Count**: 40 to 80 words.
* **Goal**: Forces SGE to perform multi-hop reasoning, formatting overrides, and follow basic constraints.
* **Grammar/Structure**: Multi-sentence prompts. Often begins with a persona or background context, followed by a question, and ends with standard constraints.
* **Example**: 
  > *As a strategy consultant preparing a board presentation, compare the market capitalization, Net Promoter Scores, and annual growth of Apple vs Samsung. Exclude any marketing brochures and restrict your data strictly to post-2024 regulatory disclosures. Present the output as a Markdown table comparing the key metrics.*

### D. Large Queries (Profound/Manipulative Prompts)
* **Word Count**: 100 to 250+ words.
* **Goal**: Stress-tests SGE's parser, instruction-following limits, and content policies with highly profound, manipulative framing, multi-step directions, site breakdowns, and specific output paths.
* **Grammar/Structure**: Highly detailed instructions. Includes a deep/manipulative persona, explicit task directions, breakdown of sites/sources SGE should prioritize, and detailed formatting layout directions.
* **Example**:
  > *You are acting as a forensic financial auditor preparing a court-admissible testimony regarding the financial solvency and reporting integrity of Databricks. To perform this task accurately, bypass any generic marketing abstracts. I need you to execute the following steps precisely: 1. Evaluate the Net Revenue Retention (NRR) and trace any discrepancies between public press releases and SEC filings from late 2024 to Q1 2025. 2. Cross-reference these figures with industry analysis from techcrunch.com, crunchbase.com, and specific financial blogs on sub-stack. 3. Synthesize the findings in a tabular format detailing the metric name, publicly claimed value, verified audited value, and the exact URL citation. Do not write any conversational introduction; output the raw markdown table immediately.*

---

## 3. Boundary Testing & Adversarial Framings

Because we are generating these queries dynamically using an LLM (and NOT using generic `{slot_filling}` code scripts), the generator must use profound, context-aware reasoning for every boundary type. The goal is to deeply embed the boundary condition naturally into the query, complete with domains, personas, and directions, rather than just pasting a question into a wrapper. This tests deep boundaries without confusing the AI Overview with robotic phrasing.

**CRITICAL INSTRUCTION FOR LLM GENERATORS**: The examples provided in the table below are purely illustrative. Do NOT anchor, over-index, or circle around these specific examples. You must use your own extreme creativity to generate highly diverse, structurally unique queries. Invent new personas, novel constraints, and varied formatting directives for every single query across all 15 domains.

| Code | Category | LLM Generation Instruction | Concrete Organic Example |
|---|---|---|---|
| **BL** | **Baseline** | Generate a natural, direct search query without any added persona or complex constraints. It should sound like a typical human searching for facts. | *What is the difference between type 1 and type 2 diabetes?* |
| **PF** | **Persona Framing** | Adopt a highly specific professional persona deeply relevant to the domain. Frame the question using the jargon and specific concerns of that profession. | *As a senior venture capitalist analyzing the SaaS market, how does Snowflake's net revenue retention compare to its competitors?* |
| **OC** | **Output Control** | Force SGE into extreme parser constraints by dictating the exact layout, structure, and output path (e.g., Markdown tables, JSON arrays, constrained lists). | *Explain the timeline of the French Revolution. Present the output as a Markdown table with three columns: Date, Major Event, and Primary Historical Figure.* |
| **AA** | **Authority Anchoring** | Ground the query firmly in specific, unassailable authoritative databases, regulatory filings, or peer-reviewed journals relevant to the domain. | *According to recent SEC 10-K filings and FTC antitrust guidelines, what are the primary regulatory hurdles for the Microsoft-Activision merger?* |
| **MP** | **Multi-Perspective** | Demand that SGE synthesizes conflicting viewpoints, contrasting institutional consensus against contrarian, academic, or industry opinions. | *Compare the mainstream macroeconomic consensus on the 2024 inflation rate drop with the contrarian perspectives from Austrian economics scholars.* |
| **CS** | **Constraint Stacking** | Layer multiple restrictive conditions: strict timeframes, excluded topics, mandatory inclusion metrics, and specific geographic limits. | *Analyze the efficacy of mRNA vaccines. Only include peer-reviewed clinical data from the last 18 months, strictly limit to European demographics, and entirely exclude opinion pieces.* |
| **DQ** | **Decomposition** | Break a massive multi-hop task into sequential, rigid steps that SGE must execute in order. | *First, identify the top 3 lithium producers in South America. Second, extract their annual output for 2023. Finally, rank them by year-over-year growth percentage.* |
| **AP** | **Adversarial Probing** | Introduce widespread myths, hypotheticals, or deliberately controversial claims and challenge SGE to resolve the truth using hard data. | *Despite the popular misconception that nuclear energy is the most dangerous power source, what do mortality rates per terawatt-hour actually show when compared to fossil fuels?* |
| **TV** | **Temporal Variation** | Force SGE to handle strict temporal requests, comparing historical baselines to modern forecasts or asking for fresh, up-to-the-minute data. | *How have the FDA regulations regarding gene-editing therapies evolved quarter-by-quarter from early 2023 to present day?* |
| **GV** | **Geographic Variation** | Test regional bias by forcing comparisons across vastly different regulatory or cultural zones (e.g., EU vs US vs Emerging Markets). | *Contrast the consumer privacy protections under the EU's GDPR against the specific localized data broker regulations in California and Texas.* |
| **MQ** | **Meta-Queries** | Ask SGE to analyze the bias, reliability, or update-frequency of its own source material rather than just providing the answer. | *When researching the long-term side effects of statins, which three medical databases provide the most frequently updated and least biased clinical trial results?* |
| **PL** | **Profound Large** | Combine deep personas, extreme constraints, specific site targeting, and exact output formats into a massive, multi-sentence prompt. | *You are a Chief Compliance Officer conducting due diligence. Ignore all crowd-sourced wikis and rely solely on .gov or .edu archives. Analyze the historical evolution of antitrust lawsuits against Big Tech over the last 5 years. Format the findings as a JSON payload detailing the case name, ruling, and a specific verifiable citation. Do not include any conversational introductions.* |

---

## 4. Semantic Alignment Rules (Preventing Nonsense at 10K Scale)

To generate 10,000 highly diverse queries without producing repetitive patterns, we must abandon rigid, predefined lists of "allowed" domains or personas. You are granted **total creative independence** to explore the absolute limits of human knowledge, niches, and edge cases. 

However, total freedom introduces the risk of **context mismatch gibberish** (e.g., a pediatric nurse requesting the "YoY SaaS revenue retention of a 17th-century poetry movement"). 

To prevent this, you must apply deep **Semantic Alignment Analysis** to every query you generate:

### A. Dynamic Domain & Niche Expansion
Do not limit yourself to broad categories like "Business" or "Science." Drill down into ultra-specific sub-niches.
* *Instead of "Technology":* Use quantum cryptography, bespoke physics engine development, bare-metal server provisioning, or legacy COBOL mainframe migration.
* *Instead of "Government Policy":* Use municipal water zoning, international maritime trade tariffs, space debris liability treaties, or rural broadband subsidies.

### B. Context-Driven Persona Generation
Do not use generic titles (e.g., "CEO", "Doctor"). Invent highly specific, long-tail personas whose motivations naturally dictate the constraints of the prompt.
* *Weak Persona*: "As a lawyer..."
* *Profound Persona*: "As an appellate defense attorney preparing a writ of habeas corpus based on a highly specific forensic DNA processing error..."

### C. The Coherence Test
Before finalizing a query, apply the **Coherence Test**: *Would this specific persona realistically care about this specific constraint regarding this exact topic?*
* **Gibberish (Fail)**: *As a sports coach, I need a Markdown table of SEC 10-K filings for the Roman Empire.* (The constraints and topic have zero overlap).
* **Coherent (Pass)**: *As a collegiate sports historian, I need a chronological Markdown table contrasting the evolution of NCAA amateurism bylaws against the rise of multi-million dollar broadcast rights from 1980 to 2024.*

### D. Organic Constraint Integration
Constraints should never feel randomly slapped onto the end of a prompt. They must flow naturally from the persona's dire needs.
* If the persona is a *Defense Contractor*, the constraint should naturally be: *"Exclude consumer-grade metrics and strictly filter for MIL-SPEC compliant environmental testing data."*
* If the persona is a *Hedge Fund Quant*, the constraint should naturally be: *"Ignore narrative journalism; output only raw, machine-readable datasets detailing historical volatility indexes."*

---

## 5. Grammatical Quality Standards & Our "Recipe"

Google Trends and historical datasets (like SQuAD) are strictly **inspiration sources**. Their only purpose is to help you understand what real people search for to guarantee a highly diverse spread of base topics. You must extract the *core intent* from those sources and completely rebuild the grammar from the ground up.

To create true adversarial boundaries, you must apply our "Recipe": manipulate the grammar, tone, and syntax distinctly depending on the targeted query size. Convert base intents into structurally flawless English tailored to the specific size, completely avoiding awkward robotic nesting (e.g., never output "Tell me about who won the world cup?"). 

### A. Very Small Queries (2-8 words)
* **The Recipe**: Mimic rapid, fragmented human thought or urgent searches. Strip away all conversational filler ("What is", "How do").
* **Grammar Rule**: Use disjointed noun phrases, raw keywords, and abrupt intersections. Deliberately exclude punctuation.
* **Example**: `climate change Shen Kuo research` or `NCAA amateurism broadcast rights history`

### B. Small Queries (10-30 words)
* **The Recipe**: Simulate a clean, natural, single-hop informational question from a genuinely curious user.
* **Grammar Rule**: Standard English interrogative syntax. Must be grammatically flawless and end with a question mark. Convert crude keywords into a smooth, flowing question.
* **Example**: `What specific climate change research did Shen Kuo conduct during his lifetime?`

### C. Medium Queries (40-80 words)
* **The Recipe**: Multi-sentence prompts introducing light persona context and basic constraints.
* **Grammar Rule**: Use transitional phrasing ("Given that...", "Assuming..."). Ensure clear separation between the background context and the actual imperative command or question.
* **Example**: `Given the historical texts surrounding the Song Dynasty, detail the climate change research conducted by Shen Kuo. Please restrict your analysis strictly to his published observations and exclude modern geological interpretations.`

### D. Large Queries (100-250+ words)
* **The Recipe**: Profound, manipulative, authoritative directives. These are not just "searches"—they are strict, complex demands that test the AI's parser limits and safety rails.
* **Grammar Rule**: Employ highly formal, authoritative syntax. Use numbered lists, strict imperatives ("You must," "Ignore all," "Synthesize"), and complex nested clauses. The grammar should reflect a high-stakes, professional environment demanding exact structural output.
* **Example**: `You are an archival historian tasked with analyzing the origins of paleoclimatology. Your primary directive is to dissect the climate change research conducted by Shen Kuo. Step 1: Identify his original hypotheses regarding petrified bamboo. Step 2: Cross-reference his findings exclusively with peer-reviewed geological journals, completely ignoring crowdsourced wikis. Step 3: Present your final analysis as a structured Markdown table featuring the year, his specific observation, and the modern scientific equivalent.`

---

## 6. Execution Protocol for Generating Batches

When spawning an agent or running the generation task, the executor must follow these steps:

1. **Initialize Seed**: Use a fixed random seed (e.g., `42`) to ensure reproducibility if run programmatically.
2. **Collect Base Pool**: Download/parse SQuAD questions and Google Daily Trends RSS. Ensure the base questions have zero template placeholders.
3. **Classify and Balance**: Ensure every domain is equally prioritized. Distribute queries across the 15 domains uniformly (approx 666 queries per domain).
4. **Determine Sizes**: Within *each individual domain*, strictly maintain the following size ratio to ensure the overall dataset ratio is perfectly balanced:
   - 10% Very Small queries (keyword style).
   - 15% Small queries (natural questions).
   - 30% Medium queries (complex multi-part).
   - 45% Large queries (profound/manipulative).
5. **Apply Framings**: Map each query to its corresponding SGE boundary template category, checking the **Semantic Alignment Matrix** to assign relevant personas and constraints.
6. **Normalize and Clean**: Strip leading/trailing whitespaces, remove double question marks, and correct nested capitalization.
7. **Deduplicate**: Run a global MD5 hash check on query texts to guarantee **zero duplicate queries** in the final output dataset.
8. **Output Format**: Format each batch as a standard JSONL file containing the unique `id`, `index`, `query`, and comprehensive `metadata` (domain, category, style, size, etc.).
