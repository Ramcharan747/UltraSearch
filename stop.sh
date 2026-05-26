#!/bin/bash
# ============================================================================
#  UltraSearch — Stop All Services
# ============================================================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo ""
echo -e "${RED}Stopping UltraSearch services...${NC}"

pkill -f "cloudflared tunnel" 2>/dev/null && echo -e "${YELLOW}[!] Killed Cloudflare tunnels${NC}" || true
pkill -f "ngrok start" 2>/dev/null && echo -e "${YELLOW}[!] Killed Ngrok tunnels${NC}" || true
pkill -f "ultrasearch -serve" 2>/dev/null && echo -e "${YELLOW}[!] Killed API server${NC}" || true

for PORT in 8082 8085; do
  PID=$(lsof -ti :$PORT 2>/dev/null || true)
  if [ -n "$PID" ]; then
    kill $PID 2>/dev/null || true
    echo -e "${YELLOW}[!] Killed process on port $PORT${NC}"
  fi
done

echo -e "${GREEN}[✓] All services stopped.${NC}"
echo ""
