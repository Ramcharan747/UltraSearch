#!/bin/bash
# Kill existing session if it exists
tmux kill-session -t pipeline 2>/dev/null

# Start a new detached session
tmux new-session -d -s pipeline -n main

# Start the server in the first pane
tmux send-keys -t pipeline "cd ~/UltraSearch && ./ultrasearch -serve -port 8082 -workers 5" C-m

# Split the window to create a second pane
tmux split-window -h -t pipeline

# Start the orchestrator in the second pane after waiting for the server to boot
tmux send-keys -t pipeline:0.1 "sleep 15 && cd ~/UltraSearch && python3 pipeline_orchestrator.py --domain business_finance && python3 pipeline_orchestrator.py --domain science_research" C-m
