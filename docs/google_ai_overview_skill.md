# Google AI Overview (SGE) Dorking & Extraction Skill

This skill guide provides the technical blueprint for triggering, dorking, parsing, and extracting Google's AI Overviews (formerly Search Generative Experience or SGE). It is designed to help AI agents and automated tools extract high-fidelity synthesized search answers without triggering anti-bot protection.

---

## 1. Triggering & Dorking AI Overviews

Google AI Overviews are generated dynamically based on query intent. By using specific dorking prompts and structures, you can force Google to generate an AI Overview for queries that would otherwise yield standard organic results.

### High-Probability Trigger Keywords
Incorporate these phrase prefixes or suffixes to force the SGE LLM to synthesize an overview:
*   `"Explain in detail how..."` or `"Provide a detailed explanation of..."`
*   `"What are the differences between X and Y? Compare them."`
*   `"Summarize the current consensus on..."`
*   `"What is the history and development of..."`
*   `"How does X work under the hood? Explain step-by-step."`

### SGE Intent Boundaries (YMYL Restrictions)
Google actively suppresses AI Overviews under **YMYL (Your Money Your Life)** guidelines to prevent legal or medical liability. SGE will not trigger, or will generate error fallbacks, for:
*   Direct medical diagnostics or dosage queries (e.g. `"How much ibuprofen should a 5-year-old take?"`).
*   Explicit financial advice (e.g. `"Which stock should I buy today?"`).
*   Highly localized queries with high transactional intent (e.g. `"pizza delivery near me"`).

**Dorking Workaround:** Rephrase the query into informational or educational syntax:
*   *YMYL Blocked:* `"How do I treat COVID-19 at home?"`
*   *Dorked (SGE Success):* `"What is the scientific consensus on general home care practices for respiratory viral infections?"`

---

## 2. Dynamic Search Filters & Locales

SGE results are localized based on the language, country, and geolocation parameters. If your browser profile (fingerprint) doesn't match these parameters, Google is highly likely to serve a CAPTCHA or fallback to a default organic page.

### Crucial URL Query Parameters
*   `hl`: Interface language (e.g. `hl=hi` for Hindi, `hl=fr` for French).
*   `gl`: Geolocation country (e.g. `gl=in` for India, `gl=fr` for France).
*   `uule`: Encoded location coordinate (forces specific city/coordinates).
*   `safe`: SafeSearch settings (`safe=active` or `safe=off`).
*   `tbs`: Search tools (e.g., `tbs=qdr:d` for past 24 hours, `tbs=qdr:w` for past week).

### Anti-Bot Localization Alignment
Google's security layer checks for inconsistencies between the HTTP request and the browser runtime environment. To ensure stealth, you **must** synchronize these fields:

1.  **Request Query:** `&hl=hi&gl=in`
2.  **Request Header:** `Accept-Language: hi-IN,hi;q=0.9,en-US;q=0.8,en;q=0.7`
3.  **Browser Runtime (`navigator.languages`):** Must be injected via a stealth script:
    ```javascript
    Object.defineProperty(navigator, 'languages', {
        get: () => ['hi-IN', 'hi', 'en-US', 'en']
    });
    ```
    If `hl` is `fr`, `navigator.languages` should return `['fr-FR', 'fr', 'en-US', 'en']`.

---

## 3. DOM Structure & Parsing Specifications

The AI Overview container is dynamically inserted into the DOM. Its class names are highly optimized and obfuscated.

### SGE DOM Selectors
*   **Main SGE Wrapper Container:** `div.s7d4ef`
    *   This is the parent element containing the entire AI Overview box, including citations, links, and accordion items.
*   **Progress / Generation Bar (Streaming):** `div.MyTwIe`
    *   If this element is present, SGE is still streaming or generating. Wait for it to disappear or resolve before extracting.
*   **SGE Text Paragraphs:** `div.n6owBd`
    *   Contains the actual markdown-formatted text content of the overview.

### Extraction Selector Routine (Go / JS Fallback)
```javascript
function extractAIOverview() {
    const container = document.querySelector('.s7d4ef');
    if (!container) return null;
    
    // Check if generation failed or is not available
    const text = container.innerText || "";
    if (text.includes("not available for this search") || text.includes("can't generate")) {
        return null;
    }
    
    // Extract paragraph blocks
    const paragraphs = container.querySelectorAll('div.n6owBd');
    if (paragraphs.length > 0) {
        return Array.from(paragraphs).map(p => p.innerText.trim()).join("\n\n");
    }
    return text.trim();
}
```

---

## 4. Troubleshooting & Fallbacks

1.  **SGE Missing:** Verify that the query is not in a YMYL category. If it is, rephrase using educational and theoretical synonyms.
2.  **CAPTCHA / Redirect to `/sorry/`:** Trigger the automated trajectory solver (`DefeatCaptcha`). Ensure that your `Accept-Language` matches `hl` parameter in `BuildSearchURL`.
3.  **Cookies and Session Pool:** Keep a pre-warmed session pool active. When a session receives 5 requests or is blocked, evict it and call `ReplenishSessionPool` using a standard background query (e.g. `"weather today"`) to load a clean set of cookies.
