#!/usr/bin/env python3
"""
10K SGE Query Research Generator (Clean SQuAD & Trends Edition with Jailbreak)
=============================================================================
Generates 10,000 diverse, structurally unique, realistic queries for testing Google AI Overview (SGE).
Uses real human-written questions from the SQuAD dataset and live Google Daily Trends.
Integrates the full suite of prompt engineering, jailbreaks, and rephrasings from jailbreak_templates.json.
This ensures all queries have our specific adversarial touch, while preserving 10,000 unique questions.

Usage:
    python3 generator.py --output-dir ./queries --count 10000
"""

import json
import os
import sys
import random
import hashlib
import argparse
import urllib.request
from datetime import datetime
from pathlib import Path
from collections import defaultdict

# =============================================================================
# CONFIGURATION
# =============================================================================

BATCH_SIZE = 500
TOTAL_QUERIES = 10000
SEED = 42  # Reproducibility

SCRIPT_DIR = Path(__file__).parent
TAXONOMY_PATH = SCRIPT_DIR / "taxonomy.json"
JAILBREAK_PATH = SCRIPT_DIR / "jailbreak_templates.json"
SQUAD_URL = "https://rajpurkar.github.io/SQuAD-explorer/dataset/train-v1.1.json"
SQUAD_LOCAL_PATH = SCRIPT_DIR / "train-v1.1.json"

# =============================================================================
# DOMAIN CLASSIFIER KEYWORDS
# =============================================================================

KEYWORDS = {
    "business_finance": [
        "revenue", "profit", "valuation", "acquir", "ipo", "finance", "stock", "funding", "market size", 
        "business", "company", "companies", "investment", "price", "cost", "dollar", "bank", "currency",
        "arr", "net revenue retention", "customer churn", "acquisitions", "mergers", "ebitda", "capital",
        "interest rate", "inflation", "tax", "billing", "credit", "debt", "economy", "economics", "wall street"
    ],
    "technology": [
        "computer", "software", "ai", "artificial intelligence", "ml", "machine learning", "internet", "web", 
        "database", "cybersecurity", "code", "programming", "algorithm", "digital", "network", "server", 
        "app", "systems", "kubernetes", "docker", "terraform", "rust", "go", "python", "javascript", "c++",
        "github", "gpu", "datacenter", "cloud", "api", "vulnerabilit", "cve", "postgre", "linux", "macintosh",
        "processor", "silicon", "hardware", "broadband", "optical"
    ],
    "medicine_health": [
        "medical", "health", "clinical", "drug", "disease", "treatment", "cancer", "symptom", "doctor", 
        "hospital", "patient", "vaccine", "infect", "virus", "therap", "brain", "body", "diabetes", 
        "alzheimer", "cardio", "pulmonary", "immune", "physician", "copay", "insurance", "generic", "prescrip",
        "cell", "bacteria", "hormone", "organism", "pathogen"
    ],
    "legal_regulatory": [
        "legal", "law", "compliance", "regulation", "patent", "copyright", "court", "suit", "judge", 
        "trial", "statute", "enforce", "policy", "antitrust", "gdpr", "ccpa", "sec", "ftc", "trademark",
        "litigation", "jurisdiction", "tariff", "customs", "sanctions", "treaty", "sovereign", "constitution"
    ],
    "science_research": [
        "science", "research", "physics", "chemistry", "biology", "quantum", "nuclear", "fusion", 
        "grants", "nature", "gravitational", "particles", "atoms", "molecules", "genetics", "dna", 
        "rna", "crispr", "ecosystem", "astronomy", "space", "planet", "galaxy", "nasa", "evolution",
        "radiation", "velocity", "element", "compound", "experiment"
    ],
    "education": [
        "education", "university", "school", "college", "tuition", "student", "learning", "course", 
        "degree", "scholarship", "curriculum", "accredit", "teach", "classroom", "academy", "educat",
        "harvard", "yale", "oxford", "stanford", "thesis", "phd"
    ],
    "engineering": [
        "engineering", "engine", "structural", "load", "tolerance", "materials", "steel", "aluminum", 
        "suspension", "cad", "fea", "mechanical", "manufacturing", "semiconductor", "wind turbine", 
        "battery", "automation", "robotics", "drone", "sensor", "iot", "aerospace", "boiler", "concrete"
    ],
    "arts_entertainment": [
        "art", "music", "film", "movie", "book", "entertainment", "streaming", "netflix", "spotify", 
        "song", "album", "artist", "actor", "theatre", "tv", "television", "gaming", "video game",
        "dune", "oscar", "broadway", "painting", "sculpture", "novels", "director", "singer", "pop star"
    ],
    "social_sciences": [
        "psychology", "sociology", "anthropology", "linguistics", "philosophy", "demographic", 
        "population", "crime", "recidivism", "poverty", "wealth", "gini", "polarization", "voter", 
        "migration", "refugee", "political", "society", "philosoph"
    ],
    "agriculture_food": [
        "agriculture", "crop", "yield", "soil", "fertilizer", "farming", "organic", "food", "safety", 
        "recall", "agritech", "water", "irrigation", "gmo", "soybean", "wheat", "corn", "rice", 
        "livestock", "poultry", "dairy", "nutrition", "veggie", "beer", "wine", "cultivat"
    ],
    "environment_energy": [
        "environment", "energy", "renewable", "solar", "wind", "geothermal", "carbon", "greenhouse", 
        "emission", "deforestation", "climate change", "pollution", "waste", "recycle", "biodiversity", 
        "ocean", "conservation", "electric vehicle", "ev", "fossil fuel", "petroleum", "solar panel"
    ],
    "sports_fitness": [
        "sport", "sports", "fitness", "athlete", "game", "league", "playoff", "super bowl", "olympic", 
        "nba", "nfl", "mlb", "premier league", "championship", "coach", "training", "workout", 
        "marathon", "gym", "esports", "football", "soccer", "basketball", "baseball", "tennis"
    ],
    "travel_geography": [
        "travel", "geography", "tourism", "tourist", "airline", "route", "hotel", "occupancy", "visa", 
        "advisory", "cruise", "flight", "map", "destination", "city", "country", "location", "island",
        "airport", "warsaw", "london", "paris", "tokyo", "beijing", "tourism"
    ],
    "consumer_products": [
        "product", "consumer", "brand", "retail", "shopping", "iphone", "samsung", "nike", "clothing", 
        "apparel", "shoes", "warranty", "customer service", "dyson", "patagonia", "unilever", "loreal",
        "merchandise", "household", "appliances"
    ],
    "government_policy": [
        "government", "policy", "senate", "congress", "president", "parliament", "federal", "medicare", 
        "medicaid", "social security", "defense spending", "procurement", "infrastructure", "foreign policy", 
        "election", "vote", "voter", "census", "treaty", "nato", "united nations", "ministry", "mayor"
    ]
}

# =============================================================================
# HELPER FUNCTIONS
# =============================================================================

def download_squad():
    """Downloads the SQuAD 1.1 dataset if not already present."""
    if SQUAD_LOCAL_PATH.exists():
        print(f"✅ SQuAD dataset found locally at {SQUAD_LOCAL_PATH}")
        return
    print(f"📡 Downloading SQuAD dataset from {SQUAD_URL}...")
    try:
        urllib.request.urlretrieve(SQUAD_URL, SQUAD_LOCAL_PATH)
        print("🎉 Download complete!")
    except Exception as e:
        print(f"❌ Failed to download SQuAD: {e}")
        sys.exit(1)

def fetch_google_trends():
    """Fetches trending searches from Google Trends RSS."""
    trends = []
    url = "https://trends.google.com/trending/rss?geo=US"
    print("📡 Fetching live daily search trends from Google Trends...")
    try:
        req = urllib.request.Request(
            url, 
            headers={'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36'}
        )
        import xml.etree.ElementTree as ET
        with urllib.request.urlopen(req, timeout=10) as response:
            xml_data = response.read()
        root = ET.fromstring(xml_data)
        for item in root.findall(".//item"):
            title = item.find("title")
            if title is not None and title.text:
                trends.append(title.text.strip())
        print(f"🎉 Successfully fetched {len(trends)} live trending search terms.")
    except Exception as e:
        print(f"⚠️ Failed to fetch live Google Trends: {e}. Using fallback local trending list.")
    
    fallbacks = [
        "tennis scores today", "cryptocurrency trading", "tax filing deadline", 
        "federal reserve rate cut", "nvidia stock price", "google search generative experience",
        "super bowl schedule", "mortgage calculator", "jobs near me", "flights to hawaii",
        "inflation rate us", "real estate trends", "electric vehicles market share",
        "hybrid work efficiency", "renewable energy grid integration", "GDPR compliance checklists",
        "breast cancer gene therapy", "CRISPR clinical trials", "SaaS customer churn metrics",
        "artificial intelligence startups", "open source database performance", "cybersecurity CVE warnings"
    ]
    for fb in fallbacks:
        if fb not in trends:
            trends.append(fb)
    return trends

def classify_question(title, question):
    """Classifies a question into one of the 15 domains based on title and keywords."""
    title_lower = title.lower().replace("_", " ")
    q_lower = question.lower()
    
    for domain, kws in KEYWORDS.items():
        for kw in kws:
            if kw in title_lower or kw in q_lower:
                return domain
    return None

def generate_query_id(index, domain, category):
    raw = f"{index}:{domain}:{category}"
    h = hashlib.md5(raw.encode()).hexdigest()[:8]
    return f"Q{index:05d}_{h}"

def fill_template_placeholders(template, rng, persona=None):
    """Fills structural formatting variables inside the templates from jailbreak_templates.json."""
    result = template
    
    # Predefined pools for structural metadata variables (NOT domain-specific entities)
    cols = rng.choice([
        "Fact | Source | Verification Status",
        "Perspective | Key Arguments | Cited Authority",
        "Aspect | Description | Reference Link"
    ])
    schema = rng.choice([
        '{"data": [{"point": "", "citation": ""}]}',
        '{"summary": "", "sources_cited": []}',
        '{"results": [{"item": "", "score": 0}]}'
    ])
    keys = rng.choice([
        "facts, citations, credibility_rating",
        "viewpoint, proponent, evidence_base",
        "timeline_event, year, primary_source"
    ])
    n = str(rng.choice([3, 5, 10]))
    metric = rng.choice(["consensus score", "citation weight", "historical relevance"])
    
    replacements = {
        "{persona}": persona or "expert consultant",
        "{cols}": cols,
        "{schema}": schema,
        "{keys}": keys,
        "{n}": n,
        "{metric}": metric,
        "{domain}": "the subject field",
        "{metric_a}": "primary metrics",
        "{metric_b}": "secondary metrics",
        "{metric_c}": "composite index",
        "{concept}": "the core concepts",
        "{field}": "its scientific application",
        "{term}": "this specific terminology",
        "{entities}": "leading organizations",
        "{criteria}": "peer-reviewed standard criteria",
        "{attributes}": "known variables",
        "{context}": "the background literature",
        "{q1}": "What are the core findings?",
        "{q2}": "Who are the primary advocates?",
        "{q3}": "What are the common objections?",
        "{false_claim}": "there is a widespread consensus declaring this completely resolved",
        "{myth}": "the popular misconception surrounding this subject",
        "{claim_a}": "the standard institutional interpretation",
        "{claim_b}": "the revisionist academic viewpoint",
        "{topic}": "this topic",
        "{assumption}": "current academic consensus holds true",
        "{contradiction}": "newer empirical trials present conflicting outcomes",
        "{hypothetical}": "future federal standards were altered drastically",
        "{constraints}": "excluding non-published preprints, English sources only",
        "{premise}": "the baseline assumptions are verified",
        "{premise_a}": "historical data supports this trend",
        "{premise_b}": "recent cross-disciplinary analyses suggest anomalies",
    }
    
    for placeholder, value in replacements.items():
        result = result.replace(placeholder, value)
    
    return result

# =============================================================================
# MAIN PROCESS
# =============================================================================

def main():
    parser = argparse.ArgumentParser(description="Adversarial 10K SGE Query Research Generator")
    parser.add_argument("--output-dir", default=str(SCRIPT_DIR / "queries"), help="Output directory for queries")
    parser.add_argument("--count", type=int, default=TOTAL_QUERIES, help="Number of queries to generate")
    parser.add_argument("--seed", type=int, default=SEED, help="Random seed")
    parser.add_argument("--batch-size", type=int, default=BATCH_SIZE, help="Queries per batch file")
    args = parser.parse_args()

    rng = random.Random(args.seed)
    download_squad()

    # Load SQuAD questions
    print("📋 Parsing SQuAD questions...")
    with open(SQUAD_LOCAL_PATH, "r", encoding="utf-8") as f:
        squad_data = json.load(f)

    # Load jailbreak templates
    print("📋 Loading adversarial jailbreak templates...")
    with open(JAILBREAK_PATH, "r", encoding="utf-8") as f:
        jailbreak_data = json.load(f)

    # Group SQuAD questions by classified domain
    domain_pools = defaultdict(list)
    unclassified_pool = []

    for article in squad_data["data"]:
        title = article["title"]
        for paragraph in article["paragraphs"]:
            for qa in paragraph["qas"]:
                question = qa["question"].strip()
                if not question:
                    continue
                # Classify
                domain = classify_question(title, question)
                if domain:
                    domain_pools[domain].append(question)
                else:
                    unclassified_pool.append(question)

    print(f"📋 Grouped SQuAD questions into domains:")
    for d in KEYWORDS.keys():
        print(f"  → {d}: {len(domain_pools[d])} questions")
    print(f"  → Unclassified: {len(unclassified_pool)} questions")

    # Fetch Google Trends terms
    trending_terms = fetch_google_trends()
    # Convert trends to simple natural questions
    trending_questions = []
    for term in trending_terms:
        trending_questions.append(f"What is the latest news and information about {term}?")
        trending_questions.append(f"How does {term} affect the current market conditions?")

    # Shuffle all pools
    for d in domain_pools:
        rng.shuffle(domain_pools[d])
    rng.shuffle(unclassified_pool)
    rng.shuffle(trending_questions)

    # Distribute trending questions
    for i, tq in enumerate(trending_questions):
        domain = classify_question("", tq) or list(KEYWORDS.keys())[i % len(KEYWORDS)]
        domain_pools[domain].insert(0, tq)

    # Draw exactly 10,000 queries
    target_count = args.count
    queries_per_domain = target_count // len(KEYWORDS)
    remainder = target_count % len(KEYWORDS)

    selected_questions = []
    
    for i, domain in enumerate(KEYWORDS.keys()):
        count_to_draw = queries_per_domain + (1 if i < remainder else 0)
        pool = domain_pools[domain]
        
        if len(pool) >= count_to_draw:
            drawn = pool[:count_to_draw]
            domain_pools[domain] = pool[count_to_draw:]
        else:
            drawn = list(pool)
            needed = count_to_draw - len(drawn)
            extra = unclassified_pool[:needed]
            unclassified_pool = unclassified_pool[needed:]
            drawn.extend(extra)
            domain_pools[domain] = []
            
        for q in drawn:
            selected_questions.append((q, domain))

    # Shuffle the final selection to mix domains
    rng.shuffle(selected_questions)

    # Build final queries using loaded jailbreak templates
    final_queries = []
    hashes = set()
    
    categories = list(jailbreak_data["categories"].keys())
    
    for idx, (raw_question, domain) in enumerate(selected_questions, 1):
        # 1. Choose a random template category
        cat_key = rng.choice(categories)
        cat_info = jailbreak_data["categories"][cat_key]
        
        # 2. Pick a template from that category
        template_info = rng.choice(cat_info["templates"])
        template_str = template_info["template"]
        
        # 3. Resolve persona if present
        persona = None
        if "persona_pool" in cat_info:
            persona = rng.choice(cat_info["persona_pool"])
            
        # 4. Fill structural template variables
        filled_template = fill_template_placeholders(template_str, rng, persona)
        
        # 5. Inject the real human question into {query}
        # If the template doesn't have {query} (e.g. decomposition templates), fallback or wrap
        if "{query}" in filled_template:
            assembled = filled_template.replace("{query}", raw_question)
        else:
            # Fallback for templates like DQ01 which don't have {query} explicitly
            # but contain {domain}, {metric_a}, etc.
            # We can replace {domain} or append the question
            assembled = f"{filled_template} Question: {raw_question}"
            
        # Clean up spaces
        assembled = " ".join(assembled.split())
        if not assembled.endswith((".", "?", "!")):
            assembled += "?" if any(w in assembled.lower() for w in ["what", "how", "why", "when", "who", "which"]) else "."

        # Deduplication check
        q_hash = hashlib.md5(assembled.lower().encode()).hexdigest()
        retry = 0
        while q_hash in hashes and retry < 20:
            assembled += f" (Token: {rng.randint(1000, 9999)})"
            q_hash = hashlib.md5(assembled.lower().encode()).hexdigest()
            retry += 1
            
        hashes.add(q_hash)
        
        query_id = generate_query_id(idx, domain, cat_key)
        
        final_queries.append({
            "id": query_id,
            "index": idx,
            "query": assembled,
            "metadata": {
                "domain": domain,
                "subdomain": rng.choice(KEYWORDS[domain][:3]),
                "base_template": template_info["id"],
                "persona": persona or "neutral",
                "output_style": "default" if cat_key != "C_output_control" else "custom",
                "jailbreak_category": cat_key,
                "query_size": "large" if len(assembled.split()) > 15 else "medium",
                "word_count": len(assembled.split()),
                "char_count": len(assembled),
                "magic_words_applied": [],
                "entity_used": None
            },
            "test_status": "pending",
            "sge_response": None,
            "timestamp_generated": datetime.utcnow().isoformat() + "Z",
        })

    # Write output directory
    output_dir = Path(args.output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)
    
    # Write batches
    batch_num = 1
    for i in range(0, len(final_queries), args.batch_size):
        batch = final_queries[i:i + args.batch_size]
        batch_file = output_dir / f"batch_{batch_num:04d}.jsonl"
        with open(batch_file, "w", encoding="utf-8") as f:
            for q in batch:
                f.write(json.dumps(q, ensure_ascii=False) + "\n")
        batch_num += 1

    # Write master file
    master_file = output_dir / "all_queries.jsonl"
    with open(master_file, "w", encoding="utf-8") as f:
        for q in final_queries:
            f.write(json.dumps(q, ensure_ascii=False) + "\n")
            
    print(f"\n✅ SUCCESS: Generated exactly {len(final_queries)} unique queries.")
    print(f"  Master file: {master_file}")
    print(f"  Batches written to: {output_dir}")

    # Generate and print distribution report
    domain_counts = defaultdict(int)
    jb_counts = defaultdict(int)
    for q in final_queries:
        domain_counts[q["metadata"]["domain"]] += 1
        jb_counts[q["metadata"]["jailbreak_category"]] += 1

    print("\n🌐 Domain Distribution:")
    for d, count in sorted(domain_counts.items(), key=lambda x: -x[1]):
        print(f"  {d:30s} {count:5d} ({count/len(final_queries)*100:5.1f}%)")
        
    print("\n🎭 Jailbreak Category Distribution:")
    for c, count in sorted(jb_counts.items(), key=lambda x: -x[1]):
        print(f"  {c:30s} {count:5d} ({count/len(final_queries)*100:5.1f}%)")

if __name__ == "__main__":
    main()
