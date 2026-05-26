# 10K SGE Query Boundary Research Study

## Objective
Systematically generate and test 10,000 queries against Google's AI Overview (SGE) to map:
- Which prompt formulations trigger SGE vs. standard SERP
- How query size (small/medium/large) changes response quality
- Which "magic words" or phrases alter SGE output structure
- How jailbreaking/persona/role-play prompt wrappers affect SGE behavior
- Cross-domain coverage (business, tech, medicine, legal, science, education, etc.)

## Directory Structure
```
query_research/
├── README.md                    # This file
├── taxonomy.json                # Master taxonomy of all dimensions
├── entity_seeds.json            # Real-world entities extracted from CSVs
├── jailbreak_templates.json     # Prompt engineering wrapper structures
├── generator.py                 # Main 10K query generator script
├── queries/                     # Output directory
│   ├── batch_0001.jsonl         # Queries 1-500
│   ├── batch_0002.jsonl         # Queries 501-1000
│   ├── ...
│   └── batch_0020.jsonl         # Queries 9501-10000
├── results/                     # SGE response storage
│   └── (populated after testing)
└── analysis/                    # Statistical analysis outputs
    └── (populated after testing)
```

## Methodology
1. **Combinatorial Expansion**: Cross-product of domains × personas × sizes × styles × jailbreak wrappers
2. **Entity Grounding**: Each query references real entities from Crustdata/Latka CSVs
3. **Size Calibration**: Small (3-8 words), Medium (15-40 words), Large (60-200 words)
4. **Quality Dimensions**: Precision, recall, freshness, source diversity, format compliance
5. **Version A Compliance**: No CAPTCHA solvers, no proxy rotation, only public SGE output
