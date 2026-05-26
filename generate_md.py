import json

with open("query_research/final_results.jsonl") as f:
    lines = f.readlines()

output = ["# UltraSearch Extracted AI Overviews\n\nHere are the 12 successful queries from our test run:\n"]

for line in lines:
    obj = json.loads(line)
    if obj.get('status') == 'SUCCESS':
        output.append(f"### Query: `{obj['query']}`\n**Domain:** `{obj['domain']}` | **Status:** `{obj['status']}`\n\n```text\n{obj['ai_overview']}\n```\n")

with open("/Users/ramcharan/.gemini/antigravity/brain/6c3d95e1-fc0b-4beb-9e3e-5b4d328e5231/sample_extracted_results.md", "w") as out:
    out.write("\n---\n".join(output))
