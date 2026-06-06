#!/bin/bash
# Kill existing session if it exists
tmux kill-session -t pipeline 2>/dev/null

# Start a new detached session
tmux new-session -d -s pipeline -n rotator

# Start rotator in the first window
tmux send-keys -t pipeline:rotator "cd ~/UltraSearch && sudo ip netns exec vpn_ns python3 rotate_accounts.py" C-m

# Create a second window for the server
tmux new-window -t pipeline -n server
tmux send-keys -t pipeline:server "cd ~/UltraSearch && sudo ip netns exec vpn_ns ./ultrasearch -serve -port 8082 -workers 5" C-m

# Create a third window for the orchestrator
tmux new-window -t pipeline -n orchestrator
tmux send-keys -t pipeline:orchestrator "sleep 15 && cd ~/UltraSearch && sudo ip netns exec vpn_ns python3 pipeline_orchestrator.py --domain business_finance && sudo ip netns exec vpn_ns python3 pipeline_orchestrator.py --domain science_research" C-m
