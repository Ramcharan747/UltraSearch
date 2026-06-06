import sqlite3
import json

import os
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
DB_PATH = os.path.join(SCRIPT_DIR, "query_research/execution_state.db")
OUTPUT_FILE = os.path.join(SCRIPT_DIR, "query_research/final_results.jsonl")

def export_db():
    conn = sqlite3.connect(DB_PATH)
    c = conn.cursor()
    
    c.execute('''
        SELECT query_id, domain, query_size, query_text, status, response_json, error_msg 
        FROM queries
    ''')
    
    rows = c.fetchall()
    
    with open(OUTPUT_FILE, 'w', encoding='utf-8') as f:
        for r in rows:
            obj = {
                "id": r[0],
                "domain": r[1],
                "size": r[2],
                "query": r[3],
                "status": r[4],
                "ai_overview": r[5],
                "error": r[6]
            }
            f.write(json.dumps(obj, ensure_ascii=False) + "\n")
            
    print(f"Exported {len(rows)} records to {OUTPUT_FILE}")

if __name__ == "__main__":
    export_db()
