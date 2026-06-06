import re
import requests
import time
import sys

URL = "https://proxy22.iitd.ac.in/cgi-bin/proxy.cgi"
CREDENTIALS_FILE = "credentials.txt"

def load_credentials():
    creds = []
    try:
        with open(CREDENTIALS_FILE, 'r') as f:
            for line in f:
                line = line.strip()
                if line and ':' in line:
                    u, p = line.split(':', 1)
                    creds.append((u.strip(), p.strip()))
    except FileNotFoundError:
        pass
    return creds

def authenticate(username, password):
    s = requests.Session()
    try:
        r = s.get(URL, verify=False, timeout=10)
        match = re.search(r'name="sessionid"\s+type="hidden"\s+value="([^"]+)"', r.text)
        if not match:
            print(f"[{username}] Could not find sessionid on proxy portal.")
            return False
            
        sessionid = match.group(1)
        data = {
            "sessionid": sessionid,
            "action": "Validate",
            "userid": username,
            "pass": password
        }
        
        r2 = s.post(URL, data=data, verify=False, timeout=10)
        if r2.status_code == 200:
            print(f"[{time.strftime('%H:%M:%S')}] Successfully authenticated active proxy IP using: {username}")
            return True
        else:
            print(f"[{username}] Auth failed with status {r2.status_code}")
            return False
    except Exception as e:
        print(f"[{username}] Error during authentication: {e}")
        return False

def main():
    import urllib3
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
    
    interval = 300 # Swap accounts every 5 minutes
    
    print(f"--- IITD Proxy Account Rotator ---")
    while True:
        creds = load_credentials()
        if not creds:
            print(f"No credentials found in {CREDENTIALS_FILE}. Create it formatted as username:password")
            time.sleep(60)
            continue
            
        for username, password in creds:
            print(f"[*] Swapping active proxy session to {username}...")
            if authenticate(username, password):
                print(f"[*] Sleeping for {interval} seconds before rotating to next account...")
                time.sleep(interval)
            else:
                print(f"[*] Failed. Retrying next account in 10s...")
                time.sleep(10)

if __name__ == "__main__":
    main()
