---
name: pe_deal_intelligence
version: 1.0.0
author: Ramcharan
trust_tier: core
domains: [venture_capital, private_equity, startup_funding]
---

# PE Deal Intelligence Schema

This Skill Book defines the structured schema, reliability prioritizations, and query dorking patterns for venture capital and private equity deal sourcing. It is designed to aggregate public, legally accessible profile details from Crunchbase and PitchBook.

## 1. Schema Properties

Every query matching this skill book must return a JSON object conforming to the following schema:

*   **`company_name`** (string, required): The official trade name of the corporation.
    *   *Source Priority:* `crunchbase` -> `pitchbook` -> `wikipedia`
*   **`total_funding_usd`** (number, required): The total venture or private equity funding raised, denominated in USD.
    *   *Source Priority:* `crunchbase` -> `pitchbook`
*   **`latest_valuation_usd`** (number, optional): The post-money valuation of the latest private transaction.
    *   *Source Priority:* `pitchbook` -> `dealroom`
*   **`key_investors`** (array of strings, required): Lead and corporate venture capital investors associated with the funding rounds.
    *   *Source Priority:* `crunchbase` -> `pitchbook`
*   **`board_of_directors`** (array of objects, optional): Roster of active directors and their corporate affiliations.
    *   *Source Priority:* `company_website` -> `pitchbook`
    *   *Fields:* `name` (string), `title` (string), `affiliation` (string)
*   **`revenue_estimate`** (string, optional): Estimated annual recurring revenue (ARR) or general revenue ranges.
    *   *Source Priority:* `pitchbook` -> `financial_news`
    *   *Confidence Floor:* `0.5` (flag if SGE's confidence is low)

## 2. Ingestion & Routing Drivers

When compiling USQL read calls, the routing layer maps target fields using these operators:

*   **VFS Target Path:** `/web/crunchbase/organizations/{slug}`
*   **Funding & Investors Operator:** `site:crunchbase.com/organization/{slug} total funding raised investors`
*   **Valuation Operator:** `"{company}" valuation post-money rounds site:pitchbook.com OR site:dealroom.co`
*   **Leadership Operator:** `"{company}" key executives leadership team board of directors site:{company_domain}`
