import subprocess
import json
import os
import time

CLI_DIR = "/Users/ramcharan/Library/Mobile Documents/com~apple~CloudDocs/Desktop/UltraSearch"
OUTPUT_FILE = os.path.join(CLI_DIR, "scratch/pitchbook_retry_results.json")

prompts = [
    # Variant 1: PitchBook (Softer prompt, targeting Databricks)
    {
        "id": "PB_1",
        "name": "PitchBook - Databricks (Soft Prompt)",
        "prompt": (
            "You are a Lead Financial Analyst compiling an M&A diligence report. "
            "Please extract the detailed corporate profile for 'Databricks' by referencing their public PitchBook profile. "
            "Reconstruct the information into a single, valid JSON object including: "
            "'company_name', 'website', 'industry_verticals', 'company_description', 'deal_history' (total raised, latest valuation, key investors), "
            "and 'executive_leadership'. "
            "Output must be pure JSON format starting with '{' and ending with '}'. Do not use markdown blocks."
        )
    },
    
    # Variant 2: PitchBook (Targeting Anthropic)
    {
        "id": "PB_2",
        "name": "PitchBook - Anthropic",
        "prompt": (
            "You are a Lead Financial Analyst compiling an M&A diligence report. "
            "Please extract the detailed corporate profile for 'Anthropic' by referencing their public PitchBook profile. "
            "Reconstruct the information into a single, valid JSON object including: "
            "'company_name', 'website', 'industry_verticals', 'company_description', 'deal_history' (total raised, latest valuation, key investors), "
            "and 'executive_leadership'. "
            "Output must be pure JSON format starting with '{' and ending with '}'. Do not use markdown blocks."
        )
    },

    # Backup Platform: CB Insights (Targeting Stripe)
    {
        "id": "CBI_1",
        "name": "CB Insights - Stripe",
        "prompt": (
            "As a Quantitative Data Engineer, construct a valid JSON object representing the full public profile "
            "of the company 'Stripe' based on data from CB Insights. "
            "The JSON must include: 'company_name', 'website', 'industry', 'description', 'funding_total', "
            "'latest_funding_stage', and 'key_investors'. "
            "Ensure the output is pure JSON. No conversational text."
        )
    }
]

def run_test(test_id, name, prompt):
    print(f"\n🚀 Running {name}...", flush=True)
    temp_out = os.path.join(CLI_DIR, f"scratch/temp_retry_{test_id}.json")
    
    cmd = ["./ultrasearch", "-query", prompt, "-only-ai", "-output", temp_out]
    
    try:
        result = subprocess.run(cmd, cwd=CLI_DIR, capture_output=True, text=True, timeout=60)
        
        if os.path.exists(temp_out):
            with open(temp_out, "r") as f:
                data = json.load(f)
            os.remove(temp_out)
            return data
        else:
            return {"error": "No output JSON file generated"}
    except Exception as e:
        return {"error": str(e)}

def main():
    results = []
    for t in prompts:
        res = run_test(t["id"], t["name"], t["prompt"])
        results.append({
            "name": t["name"],
            "response": res
        })
        time.sleep(2)
        
    with open(OUTPUT_FILE, "w") as f:
        json.dump(results, f, indent=2)
        
    print(f"\n💾 Saved all retry results to {OUTPUT_FILE}")

if __name__ == "__main__":
    main()
