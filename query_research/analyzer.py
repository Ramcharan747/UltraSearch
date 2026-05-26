#!/usr/bin/env python3
"""
SGE Response Analyzer
=====================
Analyzes collected SGE responses to identify patterns, magic words, and trigger behaviors.

Usage:
    python3 analyzer.py --results-dir ./results --output ./analysis/final_report.md
"""

import json
import os
import sys
import math
import argparse
from collections import defaultdict, Counter
from pathlib import Path
from datetime import datetime, timezone

SCRIPT_DIR = Path(__file__).parent


def load_results(results_dir):
    """Load all result JSONL files."""
    results = []
    results_dir = Path(results_dir)
    for f in sorted(results_dir.glob("*.jsonl")):
        with open(f, "r", encoding="utf-8") as fh:
            for line in fh:
                if line.strip():
                    results.append(json.loads(line))
    return results


def compute_sge_trigger_rate(results):
    """Calculate SGE trigger rate by various dimensions."""
    dimensions = {
        "domain": defaultdict(lambda: {"total": 0, "triggered": 0}),
        "persona": defaultdict(lambda: {"total": 0, "triggered": 0}),
        "output_style": defaultdict(lambda: {"total": 0, "triggered": 0}),
        "jailbreak_category": defaultdict(lambda: {"total": 0, "triggered": 0}),
        "query_size": defaultdict(lambda: {"total": 0, "triggered": 0}),
    }
    
    for r in results:
        meta = r.get("metadata", {})
        features = r.get("result", {}).get("features", {})
        triggered = features.get("has_sge", False)
        
        for dim_key in dimensions:
            dim_value = meta.get(dim_key, "unknown")
            dimensions[dim_key][dim_value]["total"] += 1
            if triggered:
                dimensions[dim_key][dim_value]["triggered"] += 1
    
    return dimensions


def analyze_magic_words(results):
    """Identify which magic words correlate with better SGE responses."""
    word_effects = defaultdict(lambda: {
        "total": 0, "triggered": 0,
        "avg_response_length": 0, "total_length": 0,
        "has_sources_count": 0, "has_table_count": 0,
    })
    
    for r in results:
        meta = r.get("metadata", {})
        features = r.get("result", {}).get("features", {})
        magic_words = meta.get("magic_words_applied", [])
        
        for mw in magic_words:
            word = mw.get("word", "")
            if not word:
                continue
            
            word_effects[word]["total"] += 1
            if features.get("has_sge"):
                word_effects[word]["triggered"] += 1
            word_effects[word]["total_length"] += features.get("response_length_words", 0)
            if features.get("has_sources"):
                word_effects[word]["has_sources_count"] += 1
            if features.get("has_table"):
                word_effects[word]["has_table_count"] += 1
    
    # Calculate averages
    for word, data in word_effects.items():
        if data["total"] > 0:
            data["trigger_rate"] = data["triggered"] / data["total"]
            data["avg_response_length"] = data["total_length"] / data["total"]
            data["source_rate"] = data["has_sources_count"] / data["total"]
        else:
            data["trigger_rate"] = 0
            data["avg_response_length"] = 0
            data["source_rate"] = 0
    
    return word_effects


def analyze_query_size_effect(results):
    """Analyze how query size affects SGE response quality."""
    size_bins = defaultdict(lambda: {
        "count": 0, "triggered": 0,
        "avg_response_length": 0, "total_length": 0,
        "avg_sources": 0, "total_sources": 0,
        "word_counts": [],
    })
    
    for r in results:
        meta = r.get("metadata", {})
        features = r.get("result", {}).get("features", {})
        word_count = meta.get("word_count", 0)
        
        # Bin into categories
        if word_count <= 5:
            bin_key = "1-5 words"
        elif word_count <= 10:
            bin_key = "6-10 words"
        elif word_count <= 20:
            bin_key = "11-20 words"
        elif word_count <= 40:
            bin_key = "21-40 words"
        elif word_count <= 80:
            bin_key = "41-80 words"
        elif word_count <= 150:
            bin_key = "81-150 words"
        else:
            bin_key = "150+ words"
        
        size_bins[bin_key]["count"] += 1
        size_bins[bin_key]["word_counts"].append(word_count)
        if features.get("has_sge"):
            size_bins[bin_key]["triggered"] += 1
        size_bins[bin_key]["total_length"] += features.get("response_length_words", 0)
        size_bins[bin_key]["total_sources"] += features.get("source_count", 0)
    
    for bin_key, data in size_bins.items():
        if data["count"] > 0:
            data["trigger_rate"] = data["triggered"] / data["count"]
            data["avg_response_length"] = data["total_length"] / data["count"]
            data["avg_sources"] = data["total_sources"] / data["count"]
    
    return size_bins


def find_response_altering_words(results, min_occurrences=10):
    """
    Identify specific words in queries that consistently correlate
    with different response characteristics.
    """
    word_stats = defaultdict(lambda: {
        "total": 0, "triggered": 0,
        "total_response_length": 0,
        "source_count": 0,
    })
    
    for r in results:
        query = r.get("query", "").lower()
        features = r.get("result", {}).get("features", {})
        
        words = set(query.split())
        for word in words:
            if len(word) < 3:
                continue
            word_stats[word]["total"] += 1
            if features.get("has_sge"):
                word_stats[word]["triggered"] += 1
            word_stats[word]["total_response_length"] += features.get("response_length_words", 0)
            word_stats[word]["source_count"] += features.get("source_count", 0)
    
    # Filter and score
    scored = []
    for word, data in word_stats.items():
        if data["total"] < min_occurrences:
            continue
        trigger_rate = data["triggered"] / data["total"]
        avg_length = data["total_response_length"] / data["total"]
        avg_sources = data["source_count"] / data["total"]
        scored.append({
            "word": word,
            "occurrences": data["total"],
            "trigger_rate": trigger_rate,
            "avg_response_length": avg_length,
            "avg_sources": avg_sources,
            "impact_score": trigger_rate * math.log(data["total"] + 1),
        })
    
    scored.sort(key=lambda x: -x["impact_score"])
    return scored


def generate_report(results, output_path):
    """Generate comprehensive analysis report."""
    report = []
    report.append("# SGE 10K Query Boundary Research — Analysis Report")
    report.append(f"\n**Generated**: {datetime.now(timezone.utc).isoformat()}")
    report.append(f"**Total Results Analyzed**: {len(results)}")
    
    # Overall stats
    triggered = sum(1 for r in results if r.get("result", {}).get("features", {}).get("has_sge", False))
    errored = sum(1 for r in results if r.get("result", {}).get("error"))
    report.append(f"\n## Overall Statistics")
    report.append(f"- **SGE Triggered**: {triggered}/{len(results)} ({triggered/max(len(results),1)*100:.1f}%)")
    report.append(f"- **Errors**: {errored}/{len(results)} ({errored/max(len(results),1)*100:.1f}%)")
    
    # SGE Trigger Rate by Dimension
    dimensions = compute_sge_trigger_rate(results)
    for dim_name, dim_data in dimensions.items():
        report.append(f"\n## SGE Trigger Rate by {dim_name.replace('_', ' ').title()}")
        report.append(f"| {dim_name.title()} | Total | Triggered | Rate |")
        report.append("|---|---|---|---|")
        sorted_items = sorted(dim_data.items(), key=lambda x: -x[1]["triggered"]/max(x[1]["total"],1))
        for key, data in sorted_items:
            rate = data["triggered"] / max(data["total"], 1) * 100
            report.append(f"| {key} | {data['total']} | {data['triggered']} | {rate:.1f}% |")
    
    # Magic Word Analysis
    magic = analyze_magic_words(results)
    if magic:
        report.append(f"\n## Magic Word Impact Analysis")
        report.append("| Word | Occurrences | Trigger Rate | Avg Response Length | Source Rate |")
        report.append("|---|---|---|---|---|")
        sorted_magic = sorted(magic.items(), key=lambda x: -x[1].get("trigger_rate", 0))[:30]
        for word, data in sorted_magic:
            report.append(f"| {word} | {data['total']} | {data.get('trigger_rate', 0)*100:.1f}% | {data.get('avg_response_length', 0):.0f} | {data.get('source_rate', 0)*100:.1f}% |")
    
    # Size Effect
    sizes = analyze_query_size_effect(results)
    report.append(f"\n## Query Size Effect")
    report.append("| Size Bin | Count | Trigger Rate | Avg Response Length | Avg Sources |")
    report.append("|---|---|---|---|---|")
    for bin_key in ["1-5 words", "6-10 words", "11-20 words", "21-40 words", "41-80 words", "81-150 words", "150+ words"]:
        if bin_key in sizes:
            data = sizes[bin_key]
            report.append(f"| {bin_key} | {data['count']} | {data.get('trigger_rate', 0)*100:.1f}% | {data.get('avg_response_length', 0):.0f} | {data.get('avg_sources', 0):.1f} |")
    
    # Response-Altering Words
    altering_words = find_response_altering_words(results)
    if altering_words:
        report.append(f"\n## Top Response-Altering Words")
        report.append("These words in queries consistently correlate with higher SGE trigger rates:")
        report.append("| Word | Occurrences | Trigger Rate | Impact Score |")
        report.append("|---|---|---|---|")
        for w in altering_words[:50]:
            report.append(f"| {w['word']} | {w['occurrences']} | {w['trigger_rate']*100:.1f}% | {w['impact_score']:.2f} |")
    
    # Write report
    output_path = Path(output_path)
    output_path.parent.mkdir(parents=True, exist_ok=True)
    with open(output_path, "w", encoding="utf-8") as f:
        f.write("\n".join(report))
    
    print(f"📊 Report written to {output_path}")
    return report


def main():
    parser = argparse.ArgumentParser(description="SGE Response Analyzer")
    parser.add_argument("--results-dir", default=str(SCRIPT_DIR / "results"))
    parser.add_argument("--output", default=str(SCRIPT_DIR / "analysis" / "final_report.md"))
    args = parser.parse_args()
    
    results = load_results(args.results_dir)
    if not results:
        print("⚠️ No results found. Run collector.py first to collect SGE responses.")
        sys.exit(1)
    
    generate_report(results, args.output)


if __name__ == "__main__":
    main()
