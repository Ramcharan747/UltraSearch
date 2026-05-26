---
name: developer_api_enrichment
version: 1.0.0
author: UltraSearch Core
trust_tier: core
domains: [software_development, api_documentation, open_source]
---

# Developer API Enrichment Schema

This Skill Book defines the structured schema, query routing patterns, and extraction mappings for developer API documentation, library specs, and open-source software repositories.

## 1. Schema Properties

*   **`library_name`** (string, required): The official package name or repository name.
    *   *Source Priority:* `github` -> `npm` -> `pypi`
*   **`current_version`** (string, required): The latest stable release version.
    *   *Source Priority:* `github` -> `registry`
*   **`installation_command`** (string, required): The recommended package manager install command.
    *   *Source Priority:* `official_docs` -> `readme`
*   **`dependencies`** (array of strings, optional): Main project dependencies required by this library.
    *   *Source Priority:* `package.json` -> `setup.py`
*   **`basic_usage_example`** (string, optional): A minimal code snippet demonstrating initialization or standard usage.
    *   *Source Priority:* `readme` -> `official_docs`

## 2. Ingestion & Routing Drivers

*   **Repository Selector:** `site:github.com/{org}/{repo} readme installation usage`
*   **Package Registry Selector:** `site:npmjs.com/package/{name} OR site:pypi.org/project/{name} install`
*   **API Reference Selector:** `site:docs.rs/{name} OR site:pkg.go.dev/{name} API reference documentation`
