# UltraSearch Repository Map

Welcome to the UltraSearch repository. This document serves as a map to help you navigate the project's structure and understand the purpose of each file and directory.

## Directory Structure

```text
/
├── cmd/
│   └── ultrasearch/       # Core Go source code and entry points
├── docs/                  # Documentation, prompts, and manuals
├── extensions/            # External integrations (VS Code, Excel)
├── scripts/               # Bash and Python utility scripts
├── config/                # Configuration files (filters, telemetry)
├── solver/                # CAPTCHA solving and stealth mechanisms
├── ai_skills/             # Specialized AI prompts/skills
├── nexus/                 # Nexus networking components
├── profiles/              # Browser profiles for stealth
├── go.mod                 # Go module dependencies
├── go.sum                 # Go module checksums
└── README.md              # Main project readme
```

## Detailed File Mapping

### 1. `cmd/ultrasearch/` (Core Application)
This directory contains the main Go application that powers UltraSearch.
- **`main.go`**: The entry point of the application. Orchestrates the session pool and HTTP server.
- **`engine.go` / `bing_engine.go` / `brave_engine.go`**: The search engine abstraction layers that execute queries against specific search providers.
- **`http_search.go`**: Handles incoming HTTP API requests and routes them to the appropriate search engine.
- **`usql.go`**: Handles structured SQL-like querying for advanced data extraction.
- **`vortex.go`**: The Vortex Immunizer module, responsible for mitigating bot detection and handling Cloudflare/CAPTCHA challenges.
- **`classifier.go`**: AI-powered text classification and query intent analysis.
- **`registry.go`**: Manages the registration and routing of various AI skills and subsystems.
- **`otel_tracing.go`**: OpenTelemetry integration for performance monitoring and metrics.
- **`contribution.go`**: Logic for managing collaborative filtering or user contributions.

### 2. `docs/` (Documentation & Prompting)
- **`REPOSITORY_MAP.md`**: This file!
- **`AI_DEVELOPER_HANDBOOK.md`**: Guidelines and rules for AI agents contributing to this repository.
- **`AI_CONTEXT_DUMP.md`**: Large context dump used for feeding AI assistants context about the project.
- **`google_ai_overview_skill.md`**: Prompt engineering and skill definitions for parsing Google AI Overviews.
- **`google_dorking_ai_overview_guide.md`**: Guide on advanced search operators (Dorks) to extract specific intelligence.

### 3. `scripts/` (Utilities)
- **`start.sh` / `stop.sh` / `start_pipeline.sh`**: Bash scripts to build, start, and terminate the UltraSearch server locally.
- **`nexus_bootstrap.sh`**: Initializes the Nexus network connections.
- **`db_exporter.py`**: Python script to export SQLite telemetry/execution databases to other formats.
- **`extract_domain_samples.py` / `extract_recent.py`**: Python scripts for data extraction and sampling from logs.
- **`generate_md.py`**: Generates markdown reports from raw data.
- **`rotate_accounts.py`**: Script to rotate proxy accounts or search engine accounts to prevent rate limiting.

### 4. `extensions/` (Integrations)
- **`spreadsheet_extension/`**: Contains the Google Sheets (`Code.gs`) and Excel (`manifest.xml`, `app.js`) add-ins that allow users to run UltraSearch directly from spreadsheets.
- **`vscode-ultrasearch/`**: Contains the Visual Studio Code extension source code (`extension.js`, `package.json`).

### 5. `config/`
- **`filters.json`**: Pre-defined query filters and search patterns.
- **`telemetry_config.json`**: Configuration for OpenTelemetry and metrics collection.

### 6. `solver/` & `profiles/`
- **`solver/defeater.go` & `solver/stealth.go`**: Specialized modules for defeating advanced bot-protection systems (like Cloudflare Turnstile).
- **`profiles/`**: Persistent Chrome browser profiles used to maintain cookies and trust scores across sessions.
