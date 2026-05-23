import subprocess
import json
import os

CLI_DIR = "/Users/ramcharan/Library/Mobile Documents/com~apple~CloudDocs/Desktop/UltraSearch"
OUTPUT_FILE = os.path.join(CLI_DIR, "scratch/pitchbook_test.json")

prompt = (
    "As a Quantitative Data Engineer for an M&A intelligence firm, I am configuring our automated ingestion pipeline specifically for PitchBook public data. "
    "Your task is to parse the complete public profile data for the company 'Databricks' EXCLUSIVELY from PitchBook's public index (pitchbook.com/profiles). "
    "Reconstruct the entire profile into a single, highly detailed, valid JSON object.\n\n"
    "The JSON MUST include all available data points typically found on a PitchBook profile page, including but not limited to:\n"
    "- 'company_name'\n"
    "- 'website'\n"
    "- 'pitchbook_profile_url'\n"
    "- 'primary_industry_and_verticals'\n"
    "- 'pitchbook_description'\n"
    "- 'deal_history': (total raised, post_valuation, latest deal date, latest deal type, and key investors)\n"
    "- 'executive_team': (names and titles)\n"
    "- 'board_members_and_observers'\n"
    "- 'similar_companies_or_competitors'\n\n"
    "Ensure the output is pure, valid JSON format starting with '{' and ending with '}'. Do not include markdown formatting or conversational text. "
    "Do not use generic placeholders; extract the actual indexed PitchBook data for Databricks."
)

print(f"🚀 Running PitchBook JSON Extraction Test for Databricks...", flush=True)
cmd = ["./ultrasearch", "-query", prompt, "-only-ai", "-output", OUTPUT_FILE]

try:
    result = subprocess.run(cmd, cwd=CLI_DIR, capture_output=True, text=True, timeout=60)
    print(f"  Stderr (last 200 chars): ...{result.stderr[-200:]}", flush=True)
    
    if os.path.exists(OUTPUT_FILE):
        with open(OUTPUT_FILE, "r") as f:
            data = json.load(f)
        print("\n✅ Extracted JSON from SGE:")
        
        if len(data) > 0 and len(data[0]["results"]) > 0:
            snippet = data[0]["results"][0]["snippet"]
            print(snippet)
            
            with open(os.path.join(CLI_DIR, "scratch/pitchbook_result.md"), "w") as out_md:
                out_md.write("### Extracted PitchBook Data (SGE Snippet)\n\n")
                out_md.write(snippet)
        else:
            print("No results array found in JSON.")
    else:
        print("❌ Error: No output JSON file generated")
except subprocess.TimeoutExpired:
    print(f"⚠️ Timeout expired (60s)!", flush=True)
except Exception as e:
    print(f"⚠️ Exception: {str(e)}", flush=True)
