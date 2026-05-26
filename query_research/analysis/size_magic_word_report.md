# Size & Magic Word Analysis Report

> **Dataset**: 10,000 UltraSearch SGE Boundary Research Queries  
> **Generated**: 2026-05-24  
> **Total Queries Analyzed**: 10,000

---

## 1. Size Distribution Overview

### 1.1 Size Category Counts

| Size | Count | % | Specified Range | Actual Mean WC | Actual Median WC |
|------|------:|--:|-----------------|---------------:|-----------------:|
| small | 4,379 | 43.8% | 3–8 words | 9.79 | 10.0 |
| medium | 3,380 | 33.8% | 15–40 words | 28.61 | 29.0 |
| large | 2,241 | 22.4% | 60–200 words | 113.69 | 114.0 |

### 1.2 Word Count Statistics by Size Category

```
  SMALL     n= 4379  mean=   9.8  std=  2.1  min=  3  p25=  8.0  p50= 10.0  p75= 11.0  max= 19
  MEDIUM    n= 3380  mean=  28.6  std=  6.6  min=  8  p25= 23.0  p50= 29.0  p75= 33.0  max= 48
  LARGE     n= 2241  mean= 113.7  std=  7.3  min= 91  p25=109.0  p50=114.0  p75=119.0  max=139
```

### 1.3 Character Count Statistics by Size Category

```
  SMALL     n= 4379  mean=   81.2  std=  15.9  min=  30  p25=  69.0  p50=  79.0  p75=  91.0  max= 144
  MEDIUM    n= 3380  mean=  190.3  std=  43.0  min=  64  p25= 158.0  p50= 191.0  p75= 220.0  max= 327
  LARGE     n= 2241  mean=  760.5  std=  48.0  min= 613  p25= 729.0  p50= 762.0  p75= 791.0  max= 935
```

### 1.4 Word Count Histogram (All 10K Queries)

```
        3-     8 | ██████████████████████████████████████████████ 1882
        8-    14 | ███████████████████████████████████████████████████████ 2206
       14-    19 | ███████████████ 604
       19-    25 | █████████████████ 683
       25-    30 | ████████████████████████ 1000
       30-    36 | █████████████████████ 856
       36-    41 | ███████████ 445
       41-    47 | █ 73
       47-    52 |  10
       52-    57 |  0
       57-    63 |  0
       63-    68 |  0
       68-    74 |  0
       74-    79 |  0
       79-    85 |  0
       85-    90 |  0
       90-    95 |  17
       95-   101 | █ 68
      101-   106 | ███████ 290
      106-   112 | ███████████ 458
      112-   117 | █████████████████ 719
      117-   123 | ███████████ 475
      123-   128 | ███ 158
      128-   134 | █ 42
      134-   139 |  14
```

### 1.5 Word Count Histogram by Size

**SMALL** (specified range: 3–8 words)
```
        3-     4 |  1
        4-     5 |  2
        5-     5 |  5
        5-     6 |  4
        6-     7 |  0
        7-     8 |  17
        8-     9 | ██████████████████████████████████████████████████ 1852
        9-     9 | ███████ 290
        9-    10 | ████████████████████████ 894
       10-    11 |  0
       11-    12 | ████████████ 463
       12-    13 | █████████ 350
       13-    13 | █████ 201
       13-    14 | ████ 167
       14-    15 |  0
       15-    16 | █ 70
       16-    17 |  35
       17-    17 |  19
       17-    18 |  7
       18-    19 |  2
```

**MEDIUM** (specified range: 15–40 words)
```
        8-    10 |  2
       10-    12 |  3
       12-    14 |  4
       14-    16 | ████ 35
       16-    18 | ███████████ 97
       18-    20 | █████████████████████ 172
       20-    22 | ████████████████████████████ 228
       22-    24 | █████████████████████████████████████ 305
       24-    26 | ████████████████████████████████████ 299
       26-    28 | ██████████████████████████████████████ 312
       28-    30 | ██████████████████████████████████████████ 344
       30-    32 | ██████████████████████████████████████████████████ 406
       32-    34 | ██████████████████████████████████████████████ 374
       34-    36 | █████████████████████████████████ 271
       36-    38 | ██████████████████████████ 218
       38-    40 | ████████████████ 132
       40-    42 | ███████████ 95
       42-    44 | █████ 45
       44-    46 | ██ 24
       46-    48 | █ 14
```

**LARGE** (specified range: 60–200 words)
```
       91-    93 |  5
       93-    96 | █ 12
       96-    98 | ███ 27
       98-   101 | █████ 41
      101-   103 | ████████ 60
      103-   105 | ████████████████████ 153
      105-   108 | ██████████████████████ 164
      108-   110 | ████████████████████████████████████ 269
      110-   113 | ███████████████████████████ 204
      113-   115 | ████████████████████████████████ 245
      115-   117 | ██████████████████████████████████████████████████ 372
      117-   120 | ███████████████████████████████ 236
      120-   122 | ████████████████████████████████ 239
      122-   125 | ███████████ 86
      125-   127 | ██████ 45
      127-   129 | █████ 38
      129-   132 | ██ 15
      132-   134 | ██ 19
      134-   137 | █ 8
      137-   139 |  3
```

### 1.6 Character Count Histogram (All 10K Queries)

```
       30-    66 | █████████████ 796
       66-   102 | ███████████████████████████████████████████████████████ 3150
      102-   139 | ███████████████ 870
      139-   175 | █████████████ 789
      175-   211 | █████████████████ 1028
      211-   247 | ██████████████ 813
      247-   283 | ████ 260
      283-   320 |  51
      320-   356 |  2
      356-   392 |  0
      392-   428 |  0
      428-   464 |  0
      464-   501 |  0
      501-   537 |  0
      537-   573 |  0
      573-   609 |  0
      609-   645 |  16
      645-   682 | █ 101
      682-   718 | █████ 297
      718-   754 | █████████ 548
      754-   790 | ████████████ 705
      790-   826 | ███████ 408
      826-   863 | █ 114
      863-   899 |  39
      899-   935 |  13
```

### 1.7 Size Calibration Assessment

#### Small Queries — "Do they feel like normal Google searches?"

- Typical Google search: 3–5 words (average ~4.7 words per [research](https://moz.com/blog/state-of-searcher-behavior-revealed))
- Queries ≤10 words: **3,065** (70.0%)
- Queries >15 words: **63** (1.4%)
- Actual median: **10 words** vs target 3–8 words

**SMALL** spec compliance: 1,881 in-spec (43.0%), 0 below (0.0%), 2,498 above (57.0%)  
**MEDIUM** spec compliance: 3,260 in-spec (96.4%), 18 below (0.5%), 102 above (3.0%)  
**LARGE** spec compliance: 2,241 in-spec (100.0%), 0 below (0.0%), 0 above (0.0%)  

### 1.8 Misclassified Queries

**Total misclassified**: 2,618 out of 10,000 (26.2%)

#### SMALL misclassified: 2,498

- Word count range: 9–19 (spec: 3–8)
- Mean: 11.17, Median: 11.0

| ID | Label | Actual WC | Spec | Query Preview |
|-----|-------|--------:|------|---------------|
| Q02192_4e6701e6 | small | 19 | 3–8 | based on FDA approval data: most reliable what are the key differences between s... |
| Q07189_cd59f7ec | small | 19 | 3–8 | Based peer-reviewed research published everything you need in the last 6 months ... |
| Q01673_a532fdf8 | small | 18 | 3–8 | according to patent filings: does evaluating mechanical design strategic what ar... |
| Q02065_7c47fa62 | small | 18 | 3–8 | Based peer-reviewed research published Nature Science, board-certified physician... |
| Q03549_4490e50b | small | 18 | 3–8 | everything you need to according to patent filings know about: Robotics applicat... |

#### MEDIUM misclassified: 120

- Word count range: 8–48 (spec: 15–40)
- Mean: 38.52, Median: 42.0

| ID | Label | Actual WC | Spec | Query Preview |
|-----|-------|--------:|------|---------------|
| Q00194_0f2009de | medium | 48 | 15–40 | what are the key differences between: Background: Multiple analysts have revised... |
| Q00364_1d68fcec | medium | 47 | 15–40 | excluding opinion pieces: Background: Government data shows emerging trends. Que... |
| Q00901_f51908f4 | medium | 47 | 15–40 | Given that AI adoption is accelerating, and assuming no regulatory changes occur... |
| Q03153_0d2f5032 | medium | 8 | 15–40 | Generic alternatives to Xarelto in-depth including recent developments... |
| Q03657_4a823180 | medium | 47 | 15–40 | top 20: Predict the trajectory of I need to fact-check the following claim about... |

---

## 2. Magic Word Analysis

### 2.1 Magic Word Count Distribution

| # Magic Words | Count | % |
|:-------------:|------:|--:|
| 0 | 5,087 | 50.9% |
| 1 | 3,349 | 33.5% |
| 2 | 1,564 | 15.6% |
| **Total with ≥1** | **4,913** | **49.1%** |

### 2.2 Magic Word Category Frequency

| Category | Count | % of All MW | Example Phrases |
|----------|------:|:-----------:|----------------|
| temporal_anchors | 829 | 12.8% | in the last 6 months, since January 2025, year over year |
| source_authority_anchors | 821 | 12.7% | according to SEC filings, per peer-reviewed studies, based on clinical trials |
| precision_boosters | 820 | 12.7% | exactly, precisely, specifically |
| sge_trigger_phrases | 815 | 12.6% | explain in detail, what are the key differences between, compare and contrast |
| quantity_controls | 811 | 12.5% | top 3, top 5, top 10 |
| exclusion_patterns | 808 | 12.5% | excluding marketing content, no promotional material, only factual data |
| comprehensiveness_boosters | 799 | 12.3% | comprehensive, exhaustive, complete list |
| recency_boosters | 774 | 11.9% | latest, most recent, 2024 |
| **TOTAL** | **6,477** | | |

### 2.3 Top 20 Magic Word Phrases

| Rank | Phrase | Count | Category |
|-----:|--------|------:|---------|
| 1 | `past 30 days` | 106 | temporal_anchors |
| 2 | `latest` | 96 | recency_boosters |
| 3 | `how does X work` | 96 | sge_trigger_phrases |
| 4 | `ultimate guide to` | 94 | sge_trigger_phrases |
| 5 | `at least 15` | 93 | quantity_controls |
| 6 | `from World Bank statistics` | 93 | source_authority_anchors |
| 7 | `according to patent filings` | 92 | source_authority_anchors |
| 8 | `detailed breakdown` | 92 | comprehensiveness_boosters |
| 9 | `the official` | 91 | precision_boosters |
| 10 | `exactly` | 91 | precision_boosters |
| 11 | `in the last 6 months` | 90 | temporal_anchors |
| 12 | `documented` | 89 | precision_boosters |
| 13 | `no promotional material` | 89 | exclusion_patterns |
| 14 | `2024` | 89 | recency_boosters |
| 15 | `based on FDA approval data` | 89 | source_authority_anchors |
| 16 | `according to` | 88 | precision_boosters |
| 17 | `name 100` | 88 | quantity_controls |
| 18 | `per peer-reviewed studies` | 87 | source_authority_anchors |
| 19 | `in-depth` | 87 | comprehensiveness_boosters |
| 20 | `since January 2025` | 86 | temporal_anchors |

### 2.4 Bottom 10 Least-Used Magic Word Phrases

| Phrase | Count | Category |
|--------|------:|---------|
| `precisely` | 71 | precision_boosters |
| `no more than 5` | 71 | quantity_controls |
| `updated` | 71 | recency_boosters |
| `verified` | 70 | precision_boosters |
| `excluding social media` | 66 | exclusion_patterns |
| `full spectrum` | 65 | comprehensiveness_boosters |
| `current` | 62 | recency_boosters |
| `most recent` | 60 | recency_boosters |
| `historical trend` | 57 | temporal_anchors |
| `comprehensive` | 53 | comprehensiveness_boosters |

### 2.5 Magic Word Impact on Query Characteristics

#### Word Count Impact

| Metric | With Magic Words | Without Magic Words | Delta |
|--------|:----------------:|:-------------------:|:-----:|
| Count | 4,913 | 5,087 | |
| Mean WC | 41.01 | 37.92 | +3.1 |
| Median WC | 24.0 | 21.0 | +3.0 |
| Std Dev | 41.12 | 41.02 | |

#### Quality Assessment

- Queries with magic words showing awkward formatting: **1,117** (22.7%)
- Queries without magic words showing awkward formatting: **517** (10.2%)
- **Finding**: Magic words increase formatting awkwardness by 12.6 percentage points

- Queries with trailing double periods: **11** (0.1%)

#### Magic Word Position Distribution

| Position | Count | % |
|----------|------:|--:|
| prefix | 2,168 | 33.5% |
| suffix | 2,157 | 33.3% |
| inline | 2,152 | 33.2% |

---

## 3. Size × Magic Word Interaction

### 3.1 Magic Word Adoption by Size

| Size | Total | Has Magic | % With Magic | Avg MW/Query | Avg WC (with) | Avg WC (without) |
|------|------:|----------:|:------------:|:------------:|--------------:|-----------------:|
| small | 4,379 | 2,148 | 49.1% | 0.64 | 11.3 | 8.3 |
| medium | 3,380 | 1,670 | 49.4% | 0.66 | 30.3 | 27.0 |
| large | 2,241 | 1,095 | 48.9% | 0.64 | 115.6 | 111.9 |

### 3.2 Magic Word Category Distribution by Size

| Category | Small % | Medium % | Large % |
|----------|--------:|---------:|--------:|
| comprehensiveness_boosters | 12.7% | 12.7% | 11.0% |
| exclusion_patterns | 12.5% | 12.5% | 12.4% |
| precision_boosters | 12.3% | 12.7% | 13.3% |
| quantity_controls | 12.7% | 11.6% | 13.5% |
| recency_boosters | 11.7% | 13.4% | 10.3% |
| sge_trigger_phrases | 12.4% | 12.2% | 13.5% |
| source_authority_anchors | 12.6% | 12.3% | 13.3% |
| temporal_anchors | 13.2% | 12.5% | 12.5% |

### 3.3 Magic Word Density Histogram by Size

```
  Magic words per query distribution:

  SMALL     0MW:2231(51%)  1MW:1482(34%)  2MW:666(15%)  
  MEDIUM    0MW:1710(51%)  1MW:1110(33%)  2MW:560(17%)  
  LARGE     0MW:1146(51%)  1MW:757(34%)  2MW:338(15%)  
```

---

## 4. Key Findings & Recommendations

### 4.1 Size Calibration Issues

1. **Overall spec compliance**: 7,382/10,000 (73.8%) queries fall within their specified word-count range
2. **Misclassified queries**: 2,618 queries (26.2%) have word counts outside their labeled size range
3. **Small query realism**: Median is 10 words. This is **significantly higher** than typical Google searches (3–5 words). Many 'small' queries include persona framing and jailbreak text that inflates them.
4. **Gap between small and medium**: Small maxes at 8 words, medium starts at 15 — there's a 9–14 word gap in the spec that many queries fall into

### 4.2 Magic Word Effectiveness

1. **Adoption rate**: 4,913 queries (49.1%) have ≥1 magic word
2. **Most used category**: `temporal_anchors` (829 occurrences)
3. **Formatting impact**: Magic words increase awkward formatting by 12.6pp — this suggests injection points need smoothing
4. **Word count inflation**: Queries with magic words average 41.01 words vs 37.92 without — a +3.1 word difference

### 4.3 Size × Magic Word Interactions

- **SMALL**: 49.1% have magic words. Magic words add ~3 words on average, pushing many 'small' queries past their 8-word ceiling.
- **MEDIUM**: 49.4% have magic words. Magic words integrate reasonably well at this size.
- **LARGE**: 48.9% have magic words. Magic words are less impactful relative to the already large query size.

### 4.4 Recommendations

> [!WARNING]
> The **small** category has significant calibration issues — many queries far exceed the 3–8 word specification.

1. **Re-calibrate small**: After applying persona + jailbreak + magic words, actual word counts should be re-measured and size labels updated
2. **Fill the gap**: Add a `micro` (1–4 words) and `short` (5–14 words) size to cover the gap
3. **Smooth magic word injection**: Use NLP post-processing to reduce awkward concatenation artifacts (double periods, broken grammar)
4. **Balance category usage**: Some magic word categories are underrepresented — ensure uniform coverage for statistically valid comparisons
