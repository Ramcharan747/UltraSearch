---
name: government_fiscal_scrutiny
version: 1.0.0
author: Ramcharan
trust_tier: core
domains: [government_procurement, defense_spending, federal_contracts]
---

# Government Fiscal Scrutiny Schema

This Skill Book defines the structured schema, reliability prioritizations, and query dorking patterns for auditing federal procurement contracts, defense allocations, and unclassified agency justification documents legally.

## 1. Schema Properties

Every query matching this skill book must return a JSON object conforming to the following schema:

*   **`federal_award_id_piid`** (string, required): The unique Procurement Instrument Identifier (PIID) for the award.
    *   *Source Priority:* `sam_gov` -> `usaspending`
*   **`winning_contractor`** (string, required): The official legal name of the corporation or entity awarded the contract.
    *   *Source Priority:* `sam_gov` -> `usaspending`
*   **`contractor_duns_or_uei`** (string, optional): The Dun & Bradstreet (DUNS) or Unique Entity ID (UEI) number of the contractor.
    *   *Source Priority:* `sam_gov`
*   **`total_award_value_usd`** (number, required): The total obligated dollar amount of the contract award.
    *   *Source Priority:* `usaspending` -> `sam_gov`
*   **`funding_agency`** (string, required): The sponsoring department, agency, or military command awarding the contract.
    *   *Source Priority:* `sam_gov` -> `usaspending`
*   **`naics_code`** (string, optional): The North American Industry Classification System code representing the sector.
    *   *Source Priority:* `sam_gov`

## 2. Ingestion & Routing Drivers

When compiling USQL read calls, the routing layer maps target fields using these operators:

*   **VFS Target Path:** `/web/sam_gov/opportunities/{contract_id}`
*   **Contract Search Operator:** `site:sam.gov/opp/{contract_id} award winning contractor obligates`
*   **Agency Spending Operator:** `site:usaspending.gov award history value agency`
