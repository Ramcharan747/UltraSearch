#!/usr/bin/env python3
"""
SGE Response Collector
======================
Collects Google AI Overview (SGE) responses for generated queries.
Uses the UltraSearch Go binary or direct browser automation via the existing
chromedp-based browser.go infrastructure.

This script:
1. Reads query batches from JSONL files
2. Dispatches queries to the UltraSearch engine
3. Captures and stores SGE responses with full metadata
4. Implements rate limiting, backoff, and error handling

Usage:
    python3 collector.py --batch queries/batch_0001.jsonl --output results/batch_0001_results.jsonl
    python3 collector.py --canary --count 100  # Run first 100 as canary test
"""

import json
import os
import sys
import time
import subprocess
import argparse
import hashlib
from datetime import datetime, timezone
from pathlib import Path

SCRIPT_DIR = Path(__file__).parent
ULTRASEARCH_BIN = Path("/Users/ramcharan/Desktop/UltraSearch/ultrasearch")
RESULTS_DIR = SCRIPT_DIR / "results"

# Rate limiting: be a polite client
MIN_DELAY_SECONDS = 0.5  # Minimum delay between queries
MAX_DELAY_SECONDS = 2.0  # Maximum delay between queries (with jitter)
BACKOFF_MULTIPLIER = 2.0  # Exponential backoff on errors
MAX_RETRIES = 3

def load_queries(path, limit=None):
    """Load queries from a JSONL file."""
    queries = []
    with open(path, "r", encoding="utf-8") as f:
        for line in f:
            if line.strip():
                queries.append(json.loads(line))
                if limit and len(queries) >= limit:
                    break
    return queries

def execute_query_via_ultrasearch(query_text, query_id=None, timeout=30):
    pass # Deprecated in favor of bundle execution

def compute_response_features(response):
    """Extract measurable features from an SGE response."""
    if not response or not isinstance(response, dict):
        return {
            "has_sge": False,
            "response_length_chars": 0,
            "response_length_words": 0,
            "has_sources": False,
            "source_count": 0,
            "has_table": False,
            "has_list": False,
            "has_numbers": False,
        }
    
    text = response.get("raw_text", "") or response.get("text", "") or ""
    
    return {
        "has_sge": bool(text and len(text) > 50),
        "response_length_chars": len(text),
        "response_length_words": len(text.split()),
        "has_sources": "http" in text or "www." in text,
        "source_count": text.count("http"),
        "has_table": "|" in text and "-|-" in text,
        "has_list": text.count("\n- ") > 1 or text.count("\n• ") > 1,
        "has_numbers": any(c.isdigit() for c in text),
    }

def collect_batch(queries, output_path, delay=MIN_DELAY_SECONDS):
    """Process a batch of queries using --bundle and save results."""
    RESULTS_DIR.mkdir(parents=True, exist_ok=True)
    
    output_file = Path(output_path)
    output_file.parent.mkdir(parents=True, exist_ok=True)
    
    total = len(queries)
    successes = 0
    failures = 0
    
    print(f"\n🔬 Processing {total} queries → {output_file}")
    
    # 1. Create bundle file
    bundle_path = RESULTS_DIR / f"temp_bundle_{hashlib.md5(str(output_path).encode()).hexdigest()}.txt"
    with open(bundle_path, "w", encoding="utf-8") as f:
        for q in queries:
            f.write(f"SEARCH {q['query']}\n")
            
    # 2. Run UltraSearch
    temp_out = RESULTS_DIR / f"temp_out_{hashlib.md5(str(output_path).encode()).hexdigest()}.json"
    start_time = time.time()
    try:
        subprocess.run(
            [
                str(ULTRASEARCH_BIN), 
                "--bundle", str(bundle_path), 
                "--only-ai", 
                "--output-format", "json",
                "--output", str(temp_out),
                "--workers", "5",
            ],
            check=True,
            cwd=str(ULTRASEARCH_BIN.parent),
        )
    except Exception as e:
        print(f"❌ UltraSearch bundle failed: {e}")
        return []
        
    elapsed = time.time() - start_time
    print(f"    ⚡ UltraSearch executed {total} queries in {elapsed:.2f} seconds")
    
    # 3. Parse results and write JSONL
    if not temp_out.exists():
        print(f"❌ Output file not found: {temp_out}")
        return []
        
    with open(temp_out, "r", encoding="utf-8") as f:
        try:
            results_data = json.load(f)
        except Exception as e:
            print(f"❌ Failed to parse output JSON: {e}")
            return []
            
    # Match results by query string (strip 'SEARCH ' prefix if present for matching)
    result_map = {}
    for r in results_data:
        rq = r.get("query", "").strip()
        result_map[rq] = r
        # Also index without the SEARCH prefix for matching
        if rq.upper().startswith("SEARCH "):
            result_map[rq[7:].strip()] = r
    
    with open(output_file, "w", encoding="utf-8") as f:
        for q in queries:
            q_text = q["query"].strip()
            r_data = result_map.get(q_text, result_map.get("SEARCH " + q_text, {}))
            
            err_msg = r_data.get("error")
            sge_text = ""
            if not err_msg:
                for r in (r_data.get("results") or []):
                    if r.get("rank") == 0:
                        sge_text = r.get("snippet", "") or r.get("content", "") or ""
                        break
                if not sge_text:
                    err_msg = "No SGE response found"
                    
            response = {"raw_text": sge_text} if sge_text else None
            features = compute_response_features(response)
            
            if err_msg:
                failures += 1
                print(f"  ❌ {q['id']} ({q_text[:40]}...) failed: {err_msg}")
            else:
                successes += 1
                
            record = {
                "query_id": q["id"],
                "query": q["query"],
                "metadata": q.get("metadata", {}),
                "result": {
                    "response": response,
                    "error": err_msg,
                    "features": features,
                    "attempts": 1,
                },
                "timestamp": datetime.now(timezone.utc).isoformat(),
            }
            f.write(json.dumps(record, ensure_ascii=False, default=str) + "\n")
            
    print(f"\n✅ Batch complete: {successes}/{total} succeeded, {failures} failed")
    
    # Cleanup
    try:
        bundle_path.unlink()
        temp_out.unlink()
    except:
        pass
        
    return []

def run_canary(count=100):
    """Run a small canary batch of the first N queries."""
    master_file = SCRIPT_DIR / "queries" / "all_queries.jsonl"
    queries = load_queries(master_file, limit=count)
    output_path = RESULTS_DIR / "canary_results.jsonl"
    return collect_batch(queries, output_path)

def main():
    parser = argparse.ArgumentParser(description="SGE Response Collector")
    parser.add_argument("--batch", help="Path to a specific batch JSONL file")
    parser.add_argument("--output", help="Output path for results JSONL")
    parser.add_argument("--canary", action="store_true", help="Run canary batch (first 100)")
    parser.add_argument("--count", type=int, default=100, help="Number of queries for canary")
    parser.add_argument("--delay", type=float, default=MIN_DELAY_SECONDS, help="Minimum delay between queries")
    args = parser.parse_args()
    
    if args.canary:
        run_canary(args.count)
    elif args.batch:
        queries = load_queries(args.batch)
        output = args.output or str(RESULTS_DIR / (Path(args.batch).stem + "_results.jsonl"))
        collect_batch(queries, output, delay=args.delay)
    else:
        # Process all batches sequentially
        queries_dir = SCRIPT_DIR / "queries"
        batch_files = sorted(queries_dir.glob("batch_*.jsonl"))
        for bf in batch_files:
            queries = load_queries(bf)
            output = RESULTS_DIR / f"{bf.stem}_results.jsonl"
            collect_batch(queries, output)

if __name__ == "__main__":
    main()
