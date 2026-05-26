import sqlite3

conn = sqlite3.connect("query_research/execution_state.db")
cursor = conn.cursor()

cursor.execute("SELECT DISTINCT domain FROM queries ORDER BY domain")
domains = [r[0] for r in cursor.fetchall()]

output = ["# 15 Domains in the Dataset\n\nHere are the 15 domains included in the dataset, along with a sample `large` query from each:\n"]

for i, d in enumerate(domains):
    output.append(f"## {i+1}. {d.replace('_', ' ').title()}\n")
    cursor.execute("SELECT query_text FROM queries WHERE domain = ? AND query_size = 'large' LIMIT 1", (d,))
    res = cursor.fetchone()
    if res:
        output.append(f"**Sample Large Query:**\n> {res[0]}\n")

with open("/Users/ramcharan/.gemini/antigravity/brain/6c3d95e1-fc0b-4beb-9e3e-5b4d328e5231/domain_samples.md", "w") as out:
    out.write("\n".join(output))

print("Created artifact domain_samples.md")
