import json

with open("query_research/final_results.jsonl") as f:
    lines = f.readlines()

count = 0
with open("temp_display.txt", "w") as out:
    for line in reversed(lines):
        obj = json.loads(line)
        if obj.get("status") == "SUCCESS":
            out.write(f"### Query: `{obj['query']}`\n")
            out.write(f"**Size:** `{obj.get('query_size', 'unknown')}`\n\n")
            out.write(f"```text\n{obj['ai_overview'][:1500]}...\n```\n\n---\n")
            count += 1
            if count == 4:
                break
