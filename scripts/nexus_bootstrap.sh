#!/bin/bash

# UltraSearch Agent State Nexus Bootstrapper
# This script initializes the centralized storage and communication
# streamline for the 14-Agent Engineering Team.

NEXUS_DIR="/Users/ramcharan/Desktop/UltraSearch/nexus"
STREAMLINE_FILE="$NEXUS_DIR/global_streamline.md"

AGENTS=(
    "agent_usql_parser"
    "agent_query_normalizer"
    "agent_sourcing_prompt"
    "agent_output_layout"
    "agent_vortex_gateway"
    "agent_stdlib_registry"
    "agent_skill_catalog"
    "agent_dom_preprocessor"
    "agent_cognitive_fallback"
    "agent_chromedp_manager"
    "agent_telemetry_logger"
    "agent_quarantine_sandbox"
    "agent_stealth_trajectory"
    "agent_template_builder"
)

mkdir -p "$NEXUS_DIR"

# Initialize the global message bus
if [ ! -f "$STREAMLINE_FILE" ]; then
    echo "# Global Streamline (Message Bus)" > "$STREAMLINE_FILE"
    echo "Welcome to the UltraSearch Engineering Nexus. All agents should append global broadcasts here." >> "$STREAMLINE_FILE"
    echo "Format: **[Timestamp] [Agent Name]:** Message" >> "$STREAMLINE_FILE"
    echo "---" >> "$STREAMLINE_FILE"
fi

for agent in "${AGENTS[@]}"; do
    AGENT_DIR="$NEXUS_DIR/$agent"
    mkdir -p "$AGENT_DIR"
    
    # Touch state files
    touch "$AGENT_DIR/memory.md"
    touch "$AGENT_DIR/logs.md"
    touch "$AGENT_DIR/chats.md"
    
    # Initialize soul.md if it doesn't exist
    if [ ! -f "$AGENT_DIR/soul.md" ]; then
        echo "# Persona: $agent" > "$AGENT_DIR/soul.md"
        echo "You are $agent, part of the UltraSearch 14-Agent Engineering Team." >> "$AGENT_DIR/soul.md"
        echo "Always check $AGENT_DIR/memory.md upon startup." >> "$AGENT_DIR/soul.md"
    fi
done

echo "Agent State Nexus initialized successfully at $NEXUS_DIR"
