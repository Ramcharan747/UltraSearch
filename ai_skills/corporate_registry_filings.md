---
name: corporate_registry_filings
version: 1.0.0
author: UltraSearch Core
trust_tier: core
domains: [corporate_disclosures, sec_filings, public_records]
---

# Corporate Registry Filings Schema

This Skill Book defines the structured schema and query routing patterns for parsing public corporate filings and regulatory SEC declarations.

## 1. Schema Properties

*   **`cik_number`** (string, required): The Central Index Key issued by the SEC.
    *   *Source Priority:* `sec_edgar` -> `corporate_website`
*   **`incorporation_state`** (string, required): The legal state or country of incorporation.
    *   *Source Priority:* `sec_edgar` -> `business_registries`
*   **`fiscal_year_end`** (string, required): The date representing the end of the corporation's fiscal year.
    *   *Source Priority:* `sec_edgar`
*   **`annual_revenue`** (number, optional): Total annual revenue from the latest Form 10-K, in USD.
    *   *Source Priority:* `sec_edgar`
*   **`principal_executive_office`** (string, optional): The official corporate headquarters mailing address.
    *   *Source Priority:* `sec_edgar`

## 2. Ingestion & Routing Drivers

*   **SEC Edgar Selector:** `site:sec.gov/cgi-bin/browse-edgar CIK "{company}"`
*   **Form 10-K Selector:** `"{company}" fiscal year end Form 10-K annual revenue site:sec.gov/Archives`
*   **Business Registry Selector:** `"{company}" incorporation details registry file number site:opencorporates.com`
