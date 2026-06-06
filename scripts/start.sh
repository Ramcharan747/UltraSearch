#!/bin/bash
# ============================================================================
#  UltraSearch — One-Click Startup Script (Single Server Mode)
#  Starts backend Go server (serving both API & static files), establishes
#  a single secure tunnel, and auto-patches config files.
# ============================================================================

set -e

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
EXT_DIR="$ROOT_DIR/spreadsheet_extension"
MANIFEST="$EXT_DIR/excel/manifest.xml"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

log()  { echo -e "${GREEN}[✓]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }
err()  { echo -e "${RED}[✗]${NC} $1"; }
info() { echo -e "${CYAN}[→]${NC} $1"; }

echo ""
echo -e "${BOLD}═══════════════════════════════════════════════════${NC}"
echo -e "${BOLD}  🚀 UltraSearch Startup (Single Server)${NC}"
echo -e "${BOLD}═══════════════════════════════════════════════════${NC}"
echo ""

# ── 1. Cleanup: kill any previous instances ──────────────────────────────────
info "Cleaning up previous processes..."

pkill -f "cloudflared tunnel" 2>/dev/null && warn "Killed old Cloudflare tunnels" || true
pkill -f "ngrok start" 2>/dev/null && warn "Killed old Ngrok processes" || true
sleep 1

for PORT in 8082 8085; do
  PID=$(lsof -ti :$PORT 2>/dev/null || true)
  if [ -n "$PID" ]; then
    kill $PID 2>/dev/null || true
    warn "Killed process on port $PORT (PID: $PID)"
  fi
done
sleep 1

# ── 2. Start Go Server (Hosts API & Static Files on 8082) ───────────────────
info "Starting Go server on port 8082..."
cd "$ROOT_DIR"
go build -o ultrasearch ./cmd/ultrasearch
./ultrasearch -serve -port 8082 -workers 5 -headless=false > /tmp/ultrasearch_api.log 2>&1 &
API_PID=$!
sleep 2

if kill -0 $API_PID 2>/dev/null; then
  log "Go server started successfully (PID: $API_PID)"
else
  err "Go server failed to start! Check /tmp/ultrasearch_api.log"
  exit 1
fi

# ── 3. Start Tunnel Service ──────────────────────────────────────────────────
TUNNEL_URL=""
TUNNEL_PROVIDER=""

# Check if ngrok is installed and authenticated
HAS_NGROK=false
if command -v ngrok &>/dev/null; then
  NGROK_CONFIG_DIR=""
  if [ -d "$HOME/Library/Application Support/ngrok" ]; then
    NGROK_CONFIG_DIR="$HOME/Library/Application Support/ngrok"
  elif [ -d "$HOME/.config/ngrok" ]; then
    NGROK_CONFIG_DIR="$HOME/.config/ngrok"
  fi
  
  if [ -n "$NGROK_CONFIG_DIR" ] && grep -q "authtoken" "$NGROK_CONFIG_DIR/ngrok.yml" 2>/dev/null; then
    HAS_NGROK=true
  fi
fi

if [ "$HAS_NGROK" = true ]; then
  info "Ngrok detected and configured. Starting Ngrok tunnel..."
  ngrok start --all > /tmp/ngrok_tunnels.log 2>&1 &
  NGROK_PID=$!
  sleep 4
  
  # Fetch URL from ngrok local API
  if curl -s http://localhost:4040/api/tunnels &>/dev/null; then
    TUNNEL_URL=$(curl -s http://localhost:4040/api/tunnels | jq -r '.tunnels[] | select(.name=="ultrasearch") | .public_url' 2>/dev/null || true)
    
    if [ -n "$TUNNEL_URL" ] && [ "$TUNNEL_URL" != "null" ]; then
      TUNNEL_PROVIDER="Ngrok"
      log "Ngrok tunnel established successfully!"
    else
      warn "Failed to retrieve public URL from Ngrok. Falling back to Cloudflare..."
      kill $NGROK_PID 2>/dev/null || true
    fi
  else
    warn "Ngrok local API unreachable. Falling back to Cloudflare..."
    kill $NGROK_PID 2>/dev/null || true
  fi
fi

# Fallback to Cloudflare if Ngrok wasn't used/available
if [ -z "$TUNNEL_PROVIDER" ]; then
  info "Starting Cloudflare tunnel (Cloudflared)..."
  cloudflared tunnel --url http://localhost:8082 > /tmp/cf_tunnel_8082.log 2>&1 &
  CF_PID=$!
  
  info "Waiting for Cloudflare to assign tunnel URL..."
  MAX_WAIT=30
  ELAPSED=0
  while [ $ELAPSED -lt $MAX_WAIT ]; do
    TUNNEL_URL=$(grep -o 'https://[a-z0-9-]*\.trycloudflare\.com' /tmp/cf_tunnel_8082.log 2>/dev/null | head -1 || true)
    if [ -n "$TUNNEL_URL" ] && [ "$TUNNEL_URL" != "null" ]; then
      break
    fi
    sleep 2
    ELAPSED=$((ELAPSED + 2))
    printf "."
  done
  echo ""
  
  if [ -n "$TUNNEL_URL" ] && [ "$TUNNEL_URL" != "null" ]; then
    TUNNEL_PROVIDER="Cloudflare"
  fi
fi

# Verify we got a URL
if [ -z "$TUNNEL_URL" ] || [ "$TUNNEL_URL" = "null" ]; then
  err "Failed to establish secure tunnel!"
  err "Check tunnel logs:"
  err "  Ngrok log: /tmp/ngrok_tunnels.log"
  err "  Cloudflare log: /tmp/cf_tunnel_8082.log"
  echo ""
  warn "Server is still running locally:"
  echo "  Address: http://localhost:8082/sidebar.html"
  exit 1
fi

log "$TUNNEL_PROVIDER Public Tunnel: $TUNNEL_URL"

# ── 4. Auto-patch configuration files ───────────────────────────────────────
info "Auto-patching config files with the new tunnel URL..."
cd "$ROOT_DIR"
python3 patch_urls.py "$MANIFEST" "$EXT_DIR" "$TUNNEL_URL" "$TUNNEL_URL"

# Copy manifest to Excel's WEF folder for sideloading
WEF_DIR="$HOME/Library/Containers/com.microsoft.Excel/Data/Documents/wef"
if [ -d "$WEF_DIR" ]; then
  cp "$MANIFEST" "$WEF_DIR/manifest.xml"
  log "Copied updated manifest.xml to Excel WEF folder: $WEF_DIR"
else
  warn "Excel WEF folder not found at: $WEF_DIR"
fi

# ── 5. Summary ───────────────────────────────────────────────────────────────
echo ""
echo -e "${BOLD}═══════════════════════════════════════════════════${NC}"
echo -e "${BOLD}  ✅ UltraSearch is LIVE (${TUNNEL_PROVIDER})${NC}"
echo -e "${BOLD}═══════════════════════════════════════════════════${NC}"
echo ""
echo -e "  ${CYAN}Frontend URL:${NC}  $TUNNEL_URL/sidebar.html"
echo -e "  ${CYAN}API URL:${NC}       $TUNNEL_URL"
echo ""
echo -e "  ${CYAN}Local URL:${NC}     http://localhost:8082/sidebar.html"
echo ""
echo -e "  ${GREEN}Now reload the Excel add-in (close & reopen sidebar).${NC}"
echo -e "  ${GREEN}The manifest has been auto-updated with the new URL.${NC}"
echo ""

# Keep running
info "Press Ctrl+C to stop all services."
trap "echo ''; info 'Shutting down...'; ./stop.sh; exit 0" INT TERM
wait
