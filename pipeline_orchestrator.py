import sqlite3
import json
import os
import argparse
import time
import requests
from urllib.parse import quote_plus

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
DB_PATH = os.path.join(SCRIPT_DIR, "query_research/execution_state.db")
QUERIES_FILE = os.path.join(SCRIPT_DIR, "query_research/queries/all_queries.jsonl")
API_URL = "http://localhost:8082/search"

def init_db():
    conn = sqlite3.connect(DB_PATH)
    c = conn.cursor()
    c.execute('''
        CREATE TABLE IF NOT EXISTS queries (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            query_id TEXT,
            domain TEXT,
            query_size TEXT,
            query_text TEXT,
            status TEXT DEFAULT 'PENDING',
            response_json TEXT,
            error_msg TEXT,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    ''')
    conn.commit()
    
    # Check if empty, populate if so
    c.execute("SELECT COUNT(*) FROM queries")
    if c.fetchone()[0] == 0:
        print(f"Populating database from {QUERIES_FILE}...")
        with open(QUERIES_FILE, 'r', encoding='utf-8') as f:
            for line in f:
                line = line.strip()
                if not line: continue
                q_obj = json.loads(line)
                c.execute('''
                    INSERT INTO queries (query_id, domain, query_size, query_text)
                    VALUES (?, ?, ?, ?)
                ''', (
                    q_obj.get("id"), 
                    q_obj.get("metadata", {}).get("domain"),
                    q_obj.get("metadata", {}).get("query_size"),
                    q_obj.get("query")
                ))
        conn.commit()
        print("Database populated.")
    return conn

def run_pipeline(shard_index, total_shards, limit=None, query_size=None, domain=None):
    conn = init_db()
    c = conn.cursor()
    
    # Loop for retry passes
    max_passes = 3
    for pass_num in range(1, max_passes + 1):
        # Select pending queries for this shard
        query = "SELECT id, query_id, query_text FROM queries WHERE status = 'PENDING' AND (id % ?) = ?"
        params = [total_shards, shard_index]
        
        if query_size:
            query += " AND query_size = ?"
            params.append(query_size)
            
        if domain:
            query += " AND domain = ?"
            params.append(domain)
            
        query += " ORDER BY id ASC"
        c.execute(query, params)
        
        pending = c.fetchall()
        
        if limit:
            pending = pending[:limit]
            
        if len(pending) == 0:
            # Check if there are failed queries to retry in the next pass
            if pass_num < max_passes:
                c.execute("SELECT COUNT(*) FROM queries WHERE status = 'FAILED' AND domain = ?", (domain,))
                failed_count = c.fetchone()[0]
                if failed_count > 0:
                    print(f"🔄 Pass {pass_num} complete. Found {failed_count} FAILED queries. Resetting them to PENDING for Retry Pass {pass_num + 1}...")
                    c.execute("UPDATE queries SET status = 'PENDING', error_msg = NULL WHERE status = 'FAILED' AND domain = ?", (domain,))
                    conn.commit()
                    continue
            break
            
        print(f"Shard {shard_index}/{total_shards} has {len(pending)} pending queries (Pass {pass_num}/{max_passes}).")
        
        for row in pending:
            db_id, query_id, query_text = row
            print(f"Executing [{query_id}]: {query_text[:50]}...")
        
            url = f"{API_URL}?q={quote_plus(query_text)}&limit=5&ai_mode=only"
            
            success = False
            response_data = None
            error_msg = None
            
            # Simple retry logic
            for attempt in range(3):
                try:
                    resp = requests.get(url, timeout=180)
                    if resp.status_code == 200:
                        data = resp.json()
                        results = data.get("results", [])
                        if results and len(results) > 0 and results[0].get("rank") == 0:
                            response_data = results[0].get("snippet")
                            success = True
                            break
                        elif data.get("error"):
                            error_msg = data.get("error")
                            if error_msg in ["blocked_by_captcha", "timeout waiting for page load and settle"]:
                                print(f"  -> Transient error '{error_msg}' on attempt {attempt+1}. Retrying...")
                                time.sleep(5)
                                continue
                            break # Don't retry logic errors immediately
                        else:
                            error_msg = "No AI Overview generated (or refused)."
                            break
                    else:
                        error_msg = f"HTTP {resp.status_code}: {resp.text}"
                except Exception as e:
                    error_msg = str(e)
                    time.sleep(2 * (attempt + 1)) # Backoff
            
            if success:
                c.execute('''
                    UPDATE queries SET status = 'SUCCESS', response_json = ?, updated_at = CURRENT_TIMESTAMP
                    WHERE id = ?
                ''', (response_data, db_id))
                print("  -> SUCCESS")
            else:
                c.execute('''
                    UPDATE queries SET status = 'FAILED', error_msg = ?, updated_at = CURRENT_TIMESTAMP
                    WHERE id = ?
                ''', (error_msg, db_id))
                print(f"  -> FAILED: {error_msg}")
                
            conn.commit()
            
            # --- SUDDEN LARGE BREAKS (Human Behavior Emulation) ---
            import random
            if random.random() < 0.06:
                breaks = [5 * 60, 10 * 60, 15 * 60, 20 * 60, 30 * 60, 45 * 60, 60 * 60] # seconds
                break_dur = random.choice(breaks)
                print(f"   💤 Emulating human break, sleeping for {break_dur // 60} minutes...")
                time.sleep(break_dur)
                print("   ⏰ Woke up from break! Resuming queries...")
            else:
                time.sleep(1) # Rate limit protection between queries
        
if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--shard-index", type=int, default=0, help="0-indexed shard number")
    parser.add_argument("--total-shards", type=int, default=1, help="Total number of VMs/shards")
    parser.add_argument("--limit", type=int, default=None, help="Limit execution for testing")
    parser.add_argument("--query-size", type=str, default=None, help="Filter by query size (e.g. medium, large)")
    parser.add_argument("--domain", type=str, default=None, help="Filter by domain")
    args = parser.parse_args()
    
    run_pipeline(args.shard_index, args.total_shards, args.limit, args.query_size, args.domain)
