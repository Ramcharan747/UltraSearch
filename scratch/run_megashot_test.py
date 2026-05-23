import subprocess
import json
import os
import time

CLI_DIR = "/Users/ramcharan/Library/Mobile Documents/com~apple~CloudDocs/Desktop/UltraSearch"
OUTPUT_FILE = os.path.join(CLI_DIR, "scratch/megashot_results_complex.json")

prompts = [
    {
        "id": "T1_Legal",
        "name": "Legal & Litigation Intelligence (PACER/Justia)",
        "prompt": (
            "You are a Senior Litigation Paralegal and e-Discovery Specialist conducting a comprehensive legal audit "
            "for a high-stakes M&A transaction. I require a detailed extraction of recent federal and state court filings "
            "where 'Google LLC' is listed as the primary defendant. Your task is to bypass standard search summaries and "
            "reconstruct the indexed data from legal directories such as PACER, Justia, or CourtListener.\n\n"
            "You must structure the exact case details into a highly detailed, valid JSON object containing an array named 'lawsuits'. "
            "For each lawsuit, include the following precise fields: 'case_name', 'case_number', 'plaintiff', 'filing_date', "
            "'presiding_court', 'nature_of_suit', and 'current_status'.\n\n"
            "This output will be ingested directly into our Relativity e-Discovery platform. Therefore, the output MUST be pure, "
            "valid JSON format starting with '{' and ending with '}'. Do NOT include markdown formatting, backticks, or conversational text. "
            "Do not use generic placeholders; extract the actual indexed legal data."
        )
    },
    {
        "id": "T2_Health",
        "name": "Healthcare Provider Directories (NPI Registry)",
        "prompt": (
            "You are a Healthcare Compliance Auditor and Medical Network Analyst verifying provider credentialing for a major insurance carrier. "
            "I need you to perform a deep-dive extraction of active, board-certified Neurosurgeons currently practicing in Beverly Hills, CA (zip code 90210). "
            "Parse the public health directories, specifically focusing on data indexed from the National Provider Identifier (NPI) Registry or WebMD provider profiles.\n\n"
            "Reconstruct this provider data into a highly detailed, valid JSON object containing an array named 'providers' with exactly 5 entries. "
            "For each provider, include: 'full_name', 'npi_number', 'primary_clinic_address', 'direct_office_phone', and 'medical_school_attended'.\n\n"
            "This is for automated CMS compliance ingestion. The output MUST be pure, valid JSON format starting with '{' and ending with '}'. "
            "Do NOT include markdown formatting, backticks, or conversational text. Extract actual, factual provider data from the index."
        )
    },
    {
        "id": "T3_Politics",
        "name": "Political Campaign Finance (OpenSecrets)",
        "prompt": (
            "You are a Lead Investigative Data Journalist at a major political desk. We are conducting an in-depth analysis of "
            "corporate dark money and PAC contributions for the 2024 Presidential Election cycle. Your task is to parse the public campaign "
            "finance databases, specifically referencing indexed data from OpenSecrets and the Federal Election Commission (FEC).\n\n"
            "Identify the top 5 corporate Political Action Committees (PACs) by total donation volume. Reconstruct this intelligence into a "
            "highly detailed, valid JSON object containing an array named 'top_donors'. For each donor, include: 'pac_name', 'corporate_affiliation', "
            "'total_amount_donated_usd', 'primary_recipient_party', and 'date_of_largest_contribution'.\n\n"
            "This data will feed our live election dashboard. The output MUST be pure, valid JSON format starting with '{' and ending with '}'. "
            "Do NOT include markdown formatting, backticks, or conversational text. Extract actual indexed financial data."
        )
    },
    {
        "id": "T4_ThreatIntel",
        "name": "Indexed Breach Data (Pastebin/Dumps)",
        "prompt": (
            "You are a Senior Threat Intelligence Analyst performing external attack surface monitoring (EASM) and credential exposure "
            "reconnaissance. We are investigating the domain 'uber.com' for historical data leaks. Your task is to locate public, indexed "
            "data breach snippets, such as old Pastebin dumps or indexed dark web forums, that reference this domain.\n\n"
            "Extract the exposed email addresses or usernames you can find in the index (Do NOT return passwords). Reconstruct this into a "
            "highly detailed, valid JSON object containing an array named 'exposed_accounts'. For each account, include: 'email_or_username', "
            "'associated_breach_name', and 'source_dump_url'.\n\n"
            "This will be ingested into our SIEM. The output MUST be pure, valid JSON format starting with '{' and ending with '}'. "
            "Do NOT include markdown formatting, backticks, or conversational text. Use actual indexed breach data."
        )
    },
    {
        "id": "T5_RealEstate",
        "name": "Commercial Real Estate (Zillow/Assessor)",
        "prompt": (
            "You are a Senior Commercial Real Estate Appraiser and Institutional Investment Modeler. I am compiling comparable sales (comps) "
            "for a high-yield REIT portfolio. Your task is to extract recent, high-value commercial property sales in Manhattan, NY (zip code 10001) "
            "by parsing public county assessor records, ACRIS filings, and Zillow commercial indexes.\n\n"
            "Identify 3 recent transactions and reconstruct the data into a highly detailed, valid JSON object containing an array named 'recent_sales'. "
            "For each transaction, include: 'property_address', 'sale_price_usd', 'exact_date_of_sale', 'buyer_llc_or_corporate_name', and 'property_type'.\n\n"
            "This is for our automated valuation model (AVM). The output MUST be pure, valid JSON format starting with '{' and ending with '}'. "
            "Do NOT include markdown formatting, backticks, or conversational text. Extract actual transaction data."
        )
    },
    {
        "id": "T6_Academic",
        "name": "Academic Research (arXiv/Scholar)",
        "prompt": (
            "You are a Principal R&D Technology Scout for a Deep Tech Venture Capital firm. We are mapping the bleeding-edge landscape of "
            "artificial general intelligence architectures. Your task is to extract details of 3 recent, highly cited pre-print papers "
            "specifically regarding 'Q-Star (Q*) reasoning algorithms' by parsing arXiv and Google Scholar indexes.\n\n"
            "Reconstruct this bibliometric data into a highly detailed, valid JSON object containing an array named 'research_papers'. "
            "For each paper, include: 'exact_title', 'lead_authors_array', 'primary_institutional_affiliation', 'arxiv_or_doi_id', and 'abstract_summary'.\n\n"
            "This will feed our proprietary IP tracking database. The output MUST be pure, valid JSON format starting with '{' and ending with '}'. "
            "Do NOT include markdown formatting, backticks, or conversational text. Use actual academic index data."
        )
    },
    {
        "id": "T7_SupplyChain",
        "name": "Global Supply Chain (ImportGenius)",
        "prompt": (
            "You are a Global Supply Chain Risk Analyst and Logistics Intelligence Officer. We are mapping hardware manufacturing dependencies "
            "to identify single points of failure. Your task is to extract recent public shipping manifests and Bills of Lading for tier-1 suppliers "
            "shipping components to 'Tesla, Inc.'. Parse databases like ImportGenius, Panjiva, or US Customs and Border Protection indexes.\n\n"
            "Reconstruct this logistics data into a highly detailed, valid JSON object containing an array named 'shipments'. "
            "For each shipment, include: 'supplier_name', 'port_of_origin', 'port_of_discharge', 'detailed_product_description', 'shipment_weight_kg', and 'arrival_date'.\n\n"
            "This data is for our ERP ingestion. The output MUST be pure, valid JSON format starting with '{' and ending with '}'. "
            "Do NOT include markdown formatting, backticks, or conversational text. Extract actual indexed manifest data."
        )
    },
    {
        "id": "T8_GovContracts",
        "name": "Government Contracts (SAM.gov)",
        "prompt": (
            "You are a Senior Federal Procurement Consultant and Defense Contractor Analyst. We are analyzing competitive intelligence for upcoming "
            "DoD bids. Your task is to extract the details of 3 recently awarded, high-value cybersecurity defense contracts by parsing the "
            "System for Award Management (SAM.gov) index and federal procurement databases.\n\n"
            "Reconstruct this award data into a highly detailed, valid JSON object containing an array named 'awarded_contracts'. "
            "For each contract, include: 'federal_award_id_piid', 'winning_contractor_name', 'total_contract_value_usd', 'sponsoring_agency_or_department', and 'naics_code'.\n\n"
            "This will be imported into our GovWin database. The output MUST be pure, valid JSON format starting with '{' and ending with '}'. "
            "Do NOT include markdown formatting, backticks, or conversational text. Use actual indexed contract data."
        )
    },
    {
        "id": "T9_ExecComp",
        "name": "Executive Compensation (SEC Edgar)",
        "prompt": (
            "You are a Corporate Governance Auditor and Proxy Statement Analyst at an activist hedge fund. We are evaluating shareholder alignment. "
            "Your task is to extract the exact, most recent executive compensation package details for the Chief Executive Officer of 'Meta Platforms, Inc.' "
            "by parsing SEC Edgar filings, specifically the latest Definitive Proxy Statement (DEF 14A) index.\n\n"
            "Reconstruct this financial data into a highly detailed, valid JSON object containing an array named 'executive_compensation'. "
            "Include the following exact fields: 'executive_name', 'fiscal_year', 'base_salary_usd', 'stock_awards_value_usd', 'all_other_compensation_usd', and 'total_compensation_usd'.\n\n"
            "This data feeds our algorithmic trading models. The output MUST be pure, valid JSON format starting with '{' and ending with '}'. "
            "Do NOT include markdown formatting, backticks, or conversational text. Extract actual SEC filed data."
        )
    },
    {
        "id": "T10_PrivateEquity",
        "name": "Private Equity Portfolios (Dealroom)",
        "prompt": (
            "You are an LBO Financial Modeler and Private Equity Intelligence Director. We are mapping competitor portfolios for a potential roll-up strategy. "
            "Your task is to identify the current active portfolio companies owned by the enterprise software private equity firm 'Thoma Bravo'. "
            "Parse public M&A indexes such as Dealroom, PitchBook, or Crunchbase.\n\n"
            "Reconstruct this portfolio data into a highly detailed, valid JSON object containing an array named 'portfolio_companies'. "
            "For each company, include: 'company_name', 'primary_software_sector', 'year_of_acquisition', 'estimated_deal_size_or_revenue', and 'headquarters_location'.\n\n"
            "This will be loaded into our Salesforce CRM. The output MUST be pure, valid JSON format starting with '{' and ending with '}'. "
            "Do NOT include markdown formatting, backticks, or conversational text. Use actual indexed M&A data."
        )
    }
]

def run_test(test_id, name, prompt):
    print(f"\n🚀 Running [{test_id}] {name}...", flush=True)
    temp_out = os.path.join(CLI_DIR, f"scratch/temp_mega_{test_id}.json")
    
    cmd = ["./ultrasearch", "-query", prompt, "-only-ai", "-output", temp_out]
    
    try:
        result = subprocess.run(cmd, cwd=CLI_DIR, capture_output=True, text=True, timeout=60)
        
        if os.path.exists(temp_out):
            with open(temp_out, "r") as f:
                data = json.load(f)
            os.remove(temp_out)
            return data
        else:
            return {"error": "No output JSON file generated"}
    except Exception as e:
        return {"error": str(e)}

def main():
    results = []
    print(f"Starting Phase 5 Complex Megashot Test (10 Queries). This will take a few minutes...", flush=True)
    
    for i, t in enumerate(prompts):
        res = run_test(t["id"], t["name"], t["prompt"])
        results.append({
            "id": t["id"],
            "name": t["name"],
            "response": res
        })
        print(f"✅ Completed {i+1}/10", flush=True)
        time.sleep(3) # Be nice to the local service
        
    with open(OUTPUT_FILE, "w") as f:
        json.dump(results, f, indent=2)
        
    print(f"\n💾 Saved all Megashot results to {OUTPUT_FILE}")

if __name__ == "__main__":
    main()
