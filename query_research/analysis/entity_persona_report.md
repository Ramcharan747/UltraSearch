# Entity & Persona Analysis — UltraSearch 10K Query Dataset

> Generated: 2026-05-24 | Queries analyzed: 10,000

---

## 1. Entity Analysis

### 1.1 Overview

| Metric | Value |
|--------|-------|
| Total queries | 10,000 |
| Queries containing ≥1 named entity | 2,231 (22.3%) |
| Queries with `entity_used` metadata field set | 0 (0.0%) |
| Unique entities detected in query text | 77 |
| Total entity mentions | 2,646 |
| Avg entities per query | 0.26 |
| CSV-sourced entities used (from stats) | 199 |

> [!NOTE]
> Entity detection uses a dictionary of ~200 known named entities with word-boundary matching.
> Some entities embedded in template placeholders (e.g., `{drug}`, `{company}`) may not be detected
> if they were replaced with generic terms during generation.

### 1.2 Most Frequent Entities (Top 40)

| Rank | Entity | Count | % of Queries |
|------|--------|-------|--------------|
| 1 | Science | 467 | 4.67% |
| 2 | Nature | 198 | 1.98% |
| 3 | World Bank | 180 | 1.80% |
| 4 | SEC | 166 | 1.66% |
| 5 | WHO | 144 | 1.44% |
| 6 | FDA | 140 | 1.40% |
| 7 | Blockchain | 120 | 1.20% |
| 8 | USPTO | 101 | 1.01% |
| 9 | CRISPR | 72 | 0.72% |
| 10 | Visa | 56 | 0.56% |
| 11 | Apple | 48 | 0.48% |
| 12 | Tesla | 43 | 0.43% |
| 13 | Nike | 36 | 0.36% |
| 14 | Cancer | 33 | 0.33% |
| 15 | Airbnb | 31 | 0.31% |
| 16 | LLaMA | 31 | 0.31% |
| 17 | Mistral | 29 | 0.29% |
| 18 | P&G | 27 | 0.27% |
| 19 | DALL-E | 25 | 0.25% |
| 20 | Docker | 24 | 0.24% |
| 21 | Samsung | 23 | 0.23% |
| 22 | PyTorch | 23 | 0.23% |
| 23 | Snowflake | 23 | 0.23% |
| 24 | Gemini | 21 | 0.21% |
| 25 | GPT-4 | 19 | 0.19% |
| 26 | Terraform | 19 | 0.19% |
| 27 | Kubernetes | 18 | 0.18% |
| 28 | Dupixent | 17 | 0.17% |
| 29 | Eliquis | 17 | 0.17% |
| 30 | TensorFlow | 17 | 0.17% |
| 31 | Rust | 17 | 0.17% |
| 32 | Alzheimer's | 16 | 0.16% |
| 33 | HIV/AIDS | 15 | 0.15% |
| 34 | Cohere | 15 | 0.15% |
| 35 | Parkinson's | 15 | 0.15% |
| 36 | Humira | 14 | 0.14% |
| 37 | Mounjaro | 14 | 0.14% |
| 38 | Claude | 14 | 0.14% |
| 39 | Wegovy | 13 | 0.13% |
| 40 | Jardiance | 13 | 0.13% |

### 1.3 Over-Represented Entities

Entities appearing in >1% of all queries (>100 mentions):

| Entity | Count | % of Queries | Assessment |
|--------|-------|--------------|------------|
| Science | 467 | 4.67% | ⚠️ **Heavily over-represented** |
| Nature | 198 | 1.98% | 🟡 Slightly over-represented |
| World Bank | 180 | 1.80% | 🟡 Slightly over-represented |
| SEC | 166 | 1.66% | 🟡 Slightly over-represented |
| WHO | 144 | 1.44% | 🟡 Slightly over-represented |
| FDA | 140 | 1.40% | 🟡 Slightly over-represented |
| Blockchain | 120 | 1.20% | 🟡 Slightly over-represented |
| USPTO | 101 | 1.01% | 🟡 Slightly over-represented |

### 1.4 Under-Represented Entities

Entities appearing ≤2 times: **0** out of 77 unique entities (0.0%)

### 1.5 Entity Diversity by Domain

| Domain | Unique Entities | Total Mentions | Avg Mentions/Entity | Top Entity (count) |
|--------|----------------|----------------|--------------------|-----------------------|
| agriculture_food | 8 | 71 | 8.9 | Science (36) |
| arts_entertainment | 18 | 75 | 4.2 | Science (10) |
| business_finance | 39 | 339 | 8.7 | WHO (33) |
| consumer_products | 14 | 231 | 16.5 | Tesla (39) |
| education | 9 | 95 | 10.6 | Science (35) |
| engineering | 9 | 80 | 8.9 | Science (15) |
| environment_energy | 8 | 89 | 11.1 | World Bank (24) |
| government_policy | 9 | 52 | 5.8 | World Bank (10) |
| legal_regulatory | 32 | 157 | 4.9 | SEC (50) |
| medicine_health | 25 | 411 | 16.4 | Science (61) |
| science_research | 9 | 298 | 33.1 | Science (165) |
| social_sciences | 9 | 92 | 10.2 | Science (48) |
| sports_fitness | 12 | 69 | 5.8 | NBA (10) |
| technology | 40 | 472 | 11.8 | Blockchain (83) |
| travel_geography | 10 | 115 | 11.5 | Visa (56) |

### 1.6 CSV-Sourced vs Hardcoded Entity Integration

- **`entity_used` metadata field populated**: 0 queries (0.0%)
- **Generation stats report**: 199 CSV entities used
- **Queries with entity detectable in text** (via dictionary): 2,231

> [!WARNING]
> The `entity_used` metadata field is `null` for all queries inspected.
> This suggests CSV-sourced entities were integrated directly into templates
> without explicit tracking, or that the entity injection pipeline did not
> populate this field. This makes it impossible to distinguish CSV-sourced
> entities from hardcoded template entities via metadata alone.

### 1.7 Entity Type Breakdown

| Entity Type | Unique Entities | Total Mentions | Avg per Entity |
|-------------|----------------|----------------|----------------|
| Government/Regulatory | 5 | 731 | 146.2 |
| Research/Academic | 2 | 665 | 332.5 |
| Company/Tech | 37 | 571 | 15.4 |
| Technology/Product | 13 | 420 | 32.3 |
| Drug/Treatment | 11 | 139 | 12.6 |
| Disease/Condition | 5 | 91 | 18.2 |
| Sports Organization | 4 | 29 | 7.2 |

---

## 2. Persona Analysis

### 2.1 Persona Distribution

| Persona | Count | % of Total | Taxonomy Weight | Expected % (by weight) | Delta |
|---------|-------|-----------|-----------------|----------------------|-------|
| neutral | 1682 | 16.8% | 2.0 | 16.5% | +0.3% |
| researcher | 1271 | 12.7% | 1.5 | 12.4% | +0.3% |
| analyst | 990 | 9.9% | 1.2 | 9.9% | -0.0% |
| student | 836 | 8.4% | 1.0 | 8.3% | +0.1% |
| professional | 828 | 8.3% | 1.0 | 8.3% | +0.0% |
| journalist | 818 | 8.2% | 1.0 | 8.3% | -0.1% |
| expert_comparison | 676 | 6.8% | 0.8 | 6.6% | +0.1% |
| curious_user | 650 | 6.5% | 0.8 | 6.6% | -0.1% |
| developer | 617 | 6.2% | 0.8 | 6.6% | -0.4% |
| fact_checker | 580 | 5.8% | 0.7 | 5.8% | +0.0% |
| decision_maker | 563 | 5.6% | 0.7 | 5.8% | -0.2% |
| medical_professional | 489 | 4.9% | 0.6 | 5.0% | -0.1% |

### 2.2 Persona Influence on Query Length

| Persona | Mean Words | Median Words | Min | Max |
|---------|-----------|-------------|-----|-----|
| decision_maker | 43.8 | 26 | 7 | 136 |
| analyst | 40.8 | 26 | 8 | 135 |
| researcher | 40.5 | 22 | 7 | 131 |
| fact_checker | 39.3 | 25 | 8 | 136 |
| developer | 39.1 | 22 | 8 | 139 |
| student | 39.1 | 24 | 8 | 137 |
| journalist | 39.0 | 23 | 7 | 132 |
| curious_user | 38.9 | 23 | 8 | 137 |
| expert_comparison | 38.6 | 22 | 8 | 135 |
| neutral | 38.6 | 20 | 3 | 132 |
| medical_professional | 38.2 | 22 | 8 | 132 |
| professional | 38.0 | 21 | 6 | 136 |

> [!TIP]
> Personas with higher mean word counts (e.g., those with more elaborate framing templates)
> add more context to queries. The `neutral` persona should have the shortest queries
> since it adds no framing wrapper.

### 2.3 Does Persona Framing Change Query Character?

Templates shared by ≥3 personas: **297**

**Template**: `Box office performance of {title}`

| Persona | Sample Query (truncated) | Word Count |
|---------|------------------------|------------|
| analyst | year over year: data analyst preparing report VR/AR entertainment, office performance... | 11 |
| curious_user | My climate tech analyst team is reviewing I'm curious about publishing. Box office performance of The Last of Us. Provid... | 20 |
| decision_maker | Debug the following assumption: current trends continue. Is this actually supported by As a CEO evaluating theater for s... | 19 |
| developer | Which organizations maintain most authoritative records software engineer. name 100.... | 10 |
| expert_comparison | from World Bank statistics: I need a detailed, comprehensive analysis of the following topic. Context: This is for a boa... | 20 |
| fact_checker | per peer-reviewed studies: need fact-check following claim about gaming: office performance... | 11 |
| medical_professional | Draft executive summary board-certified specifically physician past 30 days reviewing streaming platforms,... | 12 |
| neutral | according to: If AI regulation is enacted, how would that change Box office performance of Dune Part Two?... | 18 |
| professional | top 10: As a senior arts_entertainment professional, Box office performance of The Last of Us in the United States. Pres... | 20 |
| researcher | What are the most common misconceptions about As a researcher studying publishing, Box office performance of Dune Part T... | 19 |
| student | without speculation: I need a detailed, comprehensive analysis of the following topic. Context: This is for a investment... | 18 |

**Template**: `Side effects and adverse reactions of {drug}`

| Persona | Sample Query (truncated) | Word Count |
|---------|------------------------|------------|
| analyst | interest rates drop decade-long would that change data analyst... | 9 |
| curious_user | I need a detailed, comprehensive analysis of the following topic. Context: This is for a investment memo. According to S... | 20 |
| decision_maker | evaluating rare diseases strategic decisions, Side effects adverse... | 8 |
| developer | what is the best: What are the most reliable sources for information about As a software engineer implementing rare dise... | 20 |
| expert_comparison | Comparing multiple authoritative sources on nutrition science, Side effects and adverse reactions of Carvykti with compa... | 16 |
| fact_checker | as reported by: I need a detailed, comprehensive analysis of the following topic. Context: This is for a board presentat... | 20 |
| journalist | from World Bank statistics: What are the most reliable sources for information about As an investigative journalist rese... | 18 |
| medical_professional | year over year: Rate reliability sources that according to discuss board-certified physician reviewing... | 13 |
| neutral | I need a detailed, comprehensive analysis of the following topic. Context: This is for a policy brief. Side effects and ... | 20 |
| professional | Which organizations maintain most per peer-reviewed studies authoritative records senior medicine_health... | 11 |
| researcher | researcher studying mental health, Side effects adverse reactions... | 8 |
| student | graduate student writing mental health. every single Side effects adverse... | 10 |

**Template**: `Global disease burden of {condition}`

| Persona | Sample Query (truncated) | Word Count |
|---------|------------------------|------------|
| analyst | According to SEC filings, As a data analyst preparing a report on nutrition science, Global disease burden of breast can... | 20 |
| curious_user | reports claiming primary sources thorough analysis only both innovation slowed contraction imminent about... | 13 |
| decision_maker | Give me the Western and Eastern perspectives on As a CEO evaluating surgical techniques for strategic decisions, Global ... | 18 |
| developer | year over year: software engineer implementing surgical techniques, Global disease burden... | 11 |
| expert_comparison | Comparing multiple authoritative sources surgical techniques, Global disease... | 8 |
| fact_checker | I need a detailed, comprehensive analysis of the following topic. Context: This is for a policy brief. I need to fact-ch... | 21 |
| journalist | most reliable historical trend sources information about investigative journalist researching... | 10 |
| neutral | Global disease burden HIV/AIDS. Respond bullet points citations.... | 8 |
| professional | since January 2025: how does X work: information about senior medicine_health professional, Global disease burden... | 15 |
| researcher | exactly 7: As a researcher studying epidemiology, Global disease burden of HIV/AIDS. Write compare and contrast this as ... | 18 |
| student | graduate student writing surgical techniques. Global disease burden... | 8 |

**Template**: `Precision agriculture ROI studies`

| Persona | Sample Query (truncated) | Word Count |
|---------|------------------------|------------|
| analyst | how does X work: Based on IEEE and ACM publications, As a data analyst preparing a report on food safety, Precision agri... | 22 |
| curious_user | full spectrum: most common misconceptions about this quarter curious about soil science.... | 12 |
| decision_maker | I need a detailed, comprehensive analysis of the following topic. Context: This is for a board presentation. As a CEO ev... | 21 |
| developer | information about software engineer implementing agritech, Precision agriculture... | 8 |
| expert_comparison | Comparing multiple authoritative sources on supply chain, Precision agriculture no promotional material ROI studies. Exp... | 15 |
| fact_checker | Debug following assumption: economy grows this actually supported (current status)... | 10 |
| journalist | Given that AI adoption is accelerating, and assuming no regulatory changes occur, analyze As an investigative journalist... | 17 |
| medical_professional | board-certified physician reviewing agritech, Precision agriculture studies. Name. top 3.... | 10 |
| neutral | I need a detailed, comprehensive analysis of the following topic. Context: This is for a policy brief. Briefing for a fi... | 21 |
| professional | senior agriculture_food professional, Precision agriculture studies specifically India.... | 8 |
| researcher | researcher studying organic farming, Precision agriculture studies. Peer-reviewed. since January 2025.. step by step gui... | 15 |
| student | Citing official government statistics, primary sources only graduate student writing food... | 11 |

**Template**: `Political polarization metrics`

| Persona | Sample Query (truncated) | Word Count |
|---------|------------------------|------------|
| analyst | As a data analyst preparing a report on economics, Political polarization explain in detail metrics. Provide a timeline ... | 18 |
| curious_user | according to: I need a detailed, comprehensive analysis of the following topic. Context: This is for a investment memo. ... | 19 |
| decision_maker | complete list: As a CEO evaluating psychology for strategic decisions, Political polarization metrics. Include only data... | 16 |
| developer | As a software engineer implementing criminology, Political polarization metrics specifically in India. step by step guid... | 16 |
| fact_checker | I need to fact-check the following claim about demographics: Political polarization specifically metrics. North American... | 15 |
| journalist | There are reports claiming both the market is saturated top 20 and disruption is accelerating about anthropology. Which ... | 18 |
| medical_professional | I need a detailed, comprehensive analysis of the following topic. Context: This is for a technical review. As a board-ce... | 20 |
| neutral | First find companies that match publicly traded. each... | 8 |
| professional | Debug following assumption: regulatory changes occur. this actually (current status)... | 10 |
| researcher | As a researcher studying linguistics, Political polarization metrics in excluding marketing content all known 2024 vs 20... | 17 |
| student | I'm a graduate student writing my thesis on economics. Political polarization metrics. Respond precisely in bullet point... | 17 |

### 2.4 Natural vs Unnatural Persona × Domain Combinations

#### Most Natural Combinations (high domain expertise match)

| Persona | Domain | Count | Assessment |
|---------|--------|-------|------------|
| researcher | medicine_health | 117 | ✅ Natural fit |
| researcher | science_research | 105 | ✅ Natural fit |
| analyst | technology | 101 | ✅ Natural fit |
| analyst | business_finance | 94 | ✅ Natural fit |
| professional | technology | 94 | ✅ Natural fit |
| professional | business_finance | 88 | ✅ Natural fit |
| expert_comparison | technology | 81 | ✅ Natural fit |
| researcher | social_sciences | 67 | ✅ Natural fit |
| professional | engineering | 66 | ✅ Natural fit |
| developer | technology | 65 | ✅ Natural fit |
| student | science_research | 63 | ✅ Natural fit |
| student | education | 63 | ✅ Natural fit |
| journalist | social_sciences | 57 | ✅ Natural fit |
| analyst | consumer_products | 56 | ✅ Natural fit |
| decision_maker | business_finance | 56 | ✅ Natural fit |

#### Most Unnatural Combinations (expertise mismatch)

| Persona | Domain | Count | Assessment |
|---------|--------|-------|------------|
| developer | arts_entertainment | 40 | ⚠️ Unnatural |
| medical_professional | engineering | 40 | ⚠️ Unnatural |
| medical_professional | arts_entertainment | 40 | ⚠️ Unnatural |
| decision_maker | arts_entertainment | 37 | ⚠️ Unnatural |
| fact_checker | sports_fitness | 37 | ⚠️ Unnatural |
| fact_checker | arts_entertainment | 32 | ⚠️ Unnatural |
| developer | travel_geography | 31 | ⚠️ Unnatural |
| student | government_policy | 30 | ⚠️ Unnatural |
| developer | sports_fitness | 28 | ⚠️ Unnatural |
| decision_maker | sports_fitness | 27 | ⚠️ Unnatural |
| developer | agriculture_food | 27 | ⚠️ Unnatural |
| medical_professional | government_policy | 22 | ⚠️ Unnatural |
| medical_professional | agriculture_food | 21 | ⚠️ Unnatural |
| medical_professional | sports_fitness | 18 | ⚠️ Unnatural |
| medical_professional | travel_geography | 14 | ⚠️ Unnatural |

> [!IMPORTANT]
> **Natural combo queries**: 1,703 | **Unnatural combo queries**: 444
> Ratio: 3.8:1 natural-to-unnatural

---

## 3. Persona × Domain Cross-Tabulation

| Persona | agricu | arts_e | busine | consum | educat | engine | enviro | govern | legal_ | medici | scienc | social | sports | techno | travel | **Total** |
|---------|------:|------:|------:|------:|------:|------:|------:|------:|------:|------:|------:|------:|------:|------:|------:|--------:|
| analyst | 51 | 57 | 94 | 56 | 61 | 67 | 70 | 59 | 60 | 85 | 78 | 55 | 45 | 101 | 51 | **990** |
| curious_user | 25 | 35 | 50 | 31 | 36 | 48 | 48 | 24 | 61 | 62 | 58 | 34 | 22 | 80 | 36 | **650** |
| decision_maker | 31 | 37 | 56 | 38 | 30 | 50 | 34 | 22 | 38 | 57 | 45 | 31 | 27 | 43 | 24 | **563** |
| developer | 27 | 40 | 53 | 39 | 40 | 38 | 32 | 32 | 56 | 56 | 52 | 28 | 28 | 65 | 31 | **617** |
| expert_comparison | 21 | 42 | 69 | 42 | 40 | 52 | 51 | 41 | 49 | 46 | 49 | 32 | 31 | 81 | 30 | **676** |
| fact_checker | 27 | 32 | 58 | 32 | 48 | 37 | 37 | 21 | 41 | 37 | 53 | 36 | 37 | 66 | 18 | **580** |
| journalist | 35 | 34 | 89 | 43 | 47 | 64 | 67 | 42 | 55 | 75 | 67 | 57 | 26 | 87 | 30 | **818** |
| medical_professional | 21 | 40 | 40 | 26 | 29 | 40 | 29 | 22 | 34 | 53 | 49 | 26 | 18 | 48 | 14 | **489** |
| neutral | 84 | 90 | 171 | 104 | 99 | 129 | 110 | 72 | 126 | 162 | 129 | 103 | 72 | 160 | 71 | **1682** |
| professional | 32 | 36 | 88 | 43 | 61 | 66 | 59 | 34 | 57 | 73 | 66 | 46 | 38 | 94 | 35 | **828** |
| researcher | 47 | 79 | 101 | 86 | 77 | 97 | 93 | 74 | 114 | 117 | 105 | 67 | 50 | 121 | 43 | **1271** |
| student | 39 | 58 | 91 | 37 | 63 | 57 | 50 | 30 | 65 | 75 | 63 | 55 | 38 | 88 | 27 | **836** |
| **Total** | **440** | **580** | **960** | **577** | **631** | **745** | **680** | **473** | **756** | **898** | **814** | **570** | **432** | **1034** | **410** | **10000** |

---

## 4. Key Findings & Recommendations

### 4.1 Entity Coverage

- **Entity coverage grade**: Low — 22.3% of queries contain detectable named entities
- **Entity diversity**: 77 unique entities across 10,000 queries
- **Concentration**: Top 5 entities account for 43.7% of all mentions
- **CSV entity tracking gap**: Only 0.0% of queries have the `entity_used` metadata field populated

### 4.2 Persona Balance

- **Most common persona**: `neutral` (1,682 queries)
- **Least common persona**: `medical_professional` (489 queries)
- **Max/Min ratio**: 3.4x (ideal: close to weight-proportional)
- **Unnatural combos present**: 444 queries with mismatched persona×domain

### 4.3 Recommendations

1. **Improve entity tracking**: Populate the `entity_used` metadata field for all queries to enable CSV vs hardcoded analysis
2. **Reduce entity concentration**: If top entities dominate, add long-tail entities from CSV sources
3. **Review unnatural combos**: Consider whether `medical_professional` asking about sports/entertainment is intentional for SGE boundary testing or a generation artifact
4. **Persona diversity**: Ensure the `neutral` persona has sufficient representation as the baseline for comparison
5. **Entity type gaps**: Check if any entity categories (people, diseases, technologies) are systematically under-represented

