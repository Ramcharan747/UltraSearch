import subprocess
import json
import os

CLI_DIR = "/Users/ramcharan/Library/Mobile Documents/com~apple~CloudDocs/Desktop/UltraSearch"
OUTPUT_FILE = os.path.join(CLI_DIR, "scratch/crunchbase_test.json")

prompt = (
    "As a Quantitative Data Engineer for an M&A intelligence firm, I am configuring our automated ingestion pipeline. "
    "Your task is to parse the complete public profile data for the company 'OpenAI' from business directories like Crunchbase, PitchBook, or CB Insights, "
    "and reconstruct the entire profile into a single, highly detailed, valid JSON object.\n\n"
    "The JSON MUST include all available data points typically found on a Crunchbase page, including but not limited to:\n"
    "- 'company_name'\n"
    "- 'website'\n"
    "- 'industries'\n"
    "- 'headquarters_location'\n"
    "- 'short_description'\n"
    "- 'long_description'\n"
    "- 'funding_history': (total funding amount, number of rounds, latest round type, date, and key lead investors)\n"
    "- 'key_executives_and_founders': (names and titles)\n"
    "- 'acquisitions' (if any)\n"
    "- 'contact_information' (emails, phone, or LinkedIn links)\n\n"
    "Ensure the output is pure, valid JSON format starting with '{' and ending with '}'. Do not include markdown formatting or conversational text. "
    "Do not use generic placeholders; extract the actual indexed data for OpenAI."
)

print(f"🚀 Running Crunchbase JSON Extraction Test...", flush=True)
cmd = ["./ultrasearch", "-query", prompt, "-only-ai", "-output", OUTPUT_FILE]

try:
    result = subprocess.run(cmd, cwd=CLI_DIR, capture_output=True, text=True, timeout=60)
    print(f"  Stderr (last 200 chars): ...{result.stderr[-200:]}", flush=True)
    
    if os.path.exists(OUTPUT_FILE):
        with open(OUTPUT_FILE, "r") as f:
            data = json.load(f)
        print("\n✅ Extracted JSON from SGE:")
        
        # SGE output is inside data[0]["results"][0]["snippet"]
        if len(data) > 0 and len(data[0]["results"]) > 0:
            snippet = data[0]["results"][0]["snippet"]
            print(snippet)
            
            # Save the raw text to a markdown file for the user to review
            with open(os.path.join(CLI_DIR, "scratch/crunchbase_result.md"), "w") as out_md:
                out_md.write("### Extracted Crunchbase Data (SGE Snippet)\n\n")
                out_md.write(snippet)
        else:
            print("No results array found in JSON.")
    else:
        print("❌ Error: No output JSON file generated")
except subprocess.TimeoutExpired:
    print(f"⚠️ Timeout expired (60s)!", flush=True)
except Exception as e:
    print(f"⚠️ Exception: {str(e)}", flush=True)
