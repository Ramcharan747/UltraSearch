# 📓 UltraSearch Usage Log & Failure Tracker

> **Core Philosophy**: Don't fix as you go. Just log it here. At the end of the week, this will serve as a prioritized bug list based on actual frequency, allowing for highly targeted architectural fixes rather than gut-feel patches.

---

## Example Entry (Do not delete)
- **Date/Time**: 2026-05-18 14:30
- **Target URL/Domain**: `crunchbase.com/organization/example`
- **Tier Assigned**: Tier 4 (Login/Domain Persistence)
- **Failure Mode**: Browser hung indefinitely after Turnstile check. `main.go` timed out after 30s.
- **Surprising Behavior**: The silent `fetch()` returned a 403 instead of the expected payload, indicating the `cf_clearance` cookie might not have attached correctly to background XHRs on this specific sub-domain.
- **Frequency**: 1

---

## 🛑 Active Log Entries

### Entry 1
- **Date/Time**: 
- **Target URL/Domain**: 
- **Tier Assigned**: 
- **Failure Mode**: 
- **Surprising Behavior**: 
- **Frequency**: 

### Entry 2
- **Date/Time**: 
- **Target URL/Domain**: 
- **Tier Assigned**: 
- **Failure Mode**: 
- **Surprising Behavior**: 
- **Frequency**: 

*(Copy and paste the template block above for new entries)*
