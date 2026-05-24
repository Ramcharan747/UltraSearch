# Implementation Plan: Native Go Vortex Security & Telemetry Migration

This plan details the migration of the UltraSearch Security Gateway (Immunizer) and Telemetry Client from Python to native Go. It also implements the high-fidelity, unstripped telemetry data policy when explicit user consent is provided, ensuring zero third-party exposure and strict security controls.

## User Review Required

> [!IMPORTANT]
> **Complete Python Removal:** By migrating the Vortex modules to native Go, we eliminate Python as a runtime dependency for the main UltraSearch engine, maximizing execution speed and simplifying deployment down to a single binary.
> 
> **Opt-In Raw Telemetry (No Data Stripping):** When a user grants explicit consent (`"telemetry_consent": "opt_in_raw"`), UltraSearch will bypass all client-side data stripping (such as anonymization or PII filtering). The raw user search query, generative SGE responses, exact content strings, and cited URLs will be transmitted 100% intact. This is critical to building a high-fidelity dataset for ML search optimization.
> 
> **Zero-Sharing Privacy Commitment:** All transmitted telemetry is forwarded exclusively over modern, secure TLS 1.2/1.3 contexts to a private database inside a secure VPC, solely used for improving search accuracy and optimizing system performance. No data will ever be shared with third parties.

---

## Strategic Stack Architecture: Who, Where, and Why?

To scale UltraSearch into a robust platform integrated across CLI, spreadsheets, extensions, and the cloud, we partition our engineering stacks based on the unique strengths of each technology:

| Technology / Language | Where It is Used | Why This Choice? |
| :--- | :--- | :--- |
| **Go** | Backend Core Daemon (`main.go`, `vortex.go`, `http_search.go`) | **Concurrency & Simplicity:** Goroutines allow executing hundreds of parallel dorks, scrapes, and probes with minimal memory overhead. Warmed connection pools and headless Chrome CDP control via `chromedp` execute at edge speed. It compiles into a single, dependency-free binary for easy multi-platform deployment. |
| **Rust** | OS-Hook Input Capture Daemon (`cursor_capture/src/`) | **Deterministic Microsecond Precision:** Rust allows manual memory control with zero Garbage Collection (GC) pauses. This guarantees the input hook captures HID coordinates at true sub-millisecond precision without losing mouse events. |
| **JavaScript / TypeScript** | Chrome/Firefox Extensions, Excel/Sheets Add-ins, VS Code/Cursor Plugins | **Native Integration:** JS/TS is the native language of browser environments, Microsoft Office Add-in host contexts, and developer editor APIs. It acts as a lightweight "spoke" wrapper that communicates with the local Go core. |
| **SQL (ClickHouse)** | Telemetry Storage Vault & Analytics | **High-Throughput Analytics:** Columnar databases compress raw search vectors, queries, and SGE telemetry logs by up to 10x, enabling sub-millisecond aggregations over billions of data rows. |

---

## Proposed Changes

We will delete the legacy Python scripts under `scripts/vortex` and create a native Go sub-package `vortex` (or incorporate it directly into the Go module) to manage security validation and telemetry log shipping.

### [DELETE] [immunizer.py](file:///Users/ramcharan/Desktop/UltraSearch/scripts/vortex/immunizer.py)
### [DELETE] [telemetry.py](file:///Users/ramcharan/Desktop/UltraSearch/scripts/vortex/telemetry.py)

### [NEW] [vortex.go](file:///Users/ramcharan/Desktop/UltraSearch/vortex.go)
We will implement a clean, lightweight Go package file:
1. **Ingestion Sandbox:** Strips script/HTML tags and null bytes.
2. **Dynamic Domain Reputation Audit:** Computes average domain trust. High-trust domains bypass heavy scans, minimizing execution latency.
3. **TF-IDF Cosine Similarity Vector Engine:** Pure Go implementation of vector correlation to compute cohesion and anomaly scores.
4. **Secondary AI/Rule-based Verification Gateway:** Flags adversarial prompts/injections.
5. **Secure Telemetry Client:**
   - Reads `telemetry_config.json`.
   - If `consent == "opt_in_raw"`, packages the raw `usql_query`, `raw_sge_response`, and associated cited URLs **with absolutely zero data stripping**.
   - Dispatches a secure HTTPS POST to the cloud endpoint using a strict `crypto/tls` config enforcing TLS 1.2/1.3.

### [MODIFY] [main.go](file:///Users/ramcharan/Desktop/UltraSearch/main.go)
Integrate the native `vortex` processing immediately after SGE responses are retrieved or scraped.

---

## Verification Plan

### Automated Tests
- Run unit test blocks inside `vortex.go` to verify:
  1. Cosine similarity calculation matches Python-level TF-IDF results.
  2. Consent parsing logic correctly switches between silent ignore (`opt_out`) and unstripped transmission (`opt_in_raw`).
  3. Sanitization correctly strips `<script>` blocks but retains fully intact raw payloads when opt-in is enabled.
- Verify the Go compilation: `go build -o ultrasearch` compiles cleanly with no errors.
