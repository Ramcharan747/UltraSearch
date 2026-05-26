// Excel Custom Functions Logic: Live Streaming & Text Animations Support

// Helper to cycle text animation frames (4 distinct states)
function startTextAnimation(invocation, baseText, intervalMs = 400) {
  const frames = [
    `${baseText}    `,
    `${baseText}.   `,
    `${baseText}..  `,
    `${baseText}... `
  ];
  let frameIdx = 0;
  
  // Set initial frame
  invocation.setResult(frames[frameIdx]);
  frameIdx = 1;

  const timer = setInterval(() => {
    invocation.setResult(frames[frameIdx]);
    frameIdx = (frameIdx + 1) % frames.length;
  }, intervalMs);
  
  return timer;
}

/**
 * Executes a custom Excel formula to fetch UltraSearch overview.
 * @customfunction
 * @param {string} query The search target query.
 * @param {CustomFunctions.StreamingInvocation<string>} invocation Excel streaming invocation.
 */
function ULTRA_SEARCH(query, invocation) {
  if (!query || query.trim() === "") {
    invocation.setResult("Error: Empty query parameter");
    return;
  }

  // Step 1: Connecting State
  invocation.setResult("#CONNECTING");

  // Process asynchronously
  setTimeout(async () => {
    let animationTimer = null;
    try {
      const cacheKey = "us_cache_" + query.toLowerCase().trim();
      let cachedVal = null;
      let apiUrl = "https://april-aorta-sandal.ngrok-free.dev";

      if (typeof OfficeRuntime !== 'undefined' && OfficeRuntime.storage) {
        try {
          cachedVal = await OfficeRuntime.storage.getItem(cacheKey);
          const storedUrl = await OfficeRuntime.storage.getItem("ULTRA_SEARCH_API_URL");
          if (storedUrl) {
            apiUrl = storedUrl;
          }
        } catch (e) {
          // Ignore storage get error
        }
      }

      if (cachedVal) {
        invocation.setResult(cachedVal);
        return;
      }

      // Step 2: Start 4-state Searching Animation
      animationTimer = startTextAnimation(invocation, "🔍 Searching", 400);

      const response = await fetch(`${apiUrl}/search?q=${encodeURIComponent(query)}&limit=5&ai_mode=only`);
      
      // Stop Searching animation
      clearInterval(animationTimer);

      if (!response.ok) {
        invocation.setResult(`HTTP Error ${response.status}: ${response.statusText}`);
        return;
      }

      // Step 3: Start 4-state Finalizing Animation
      animationTimer = startTextAnimation(invocation, "✍️ Finalizing", 400);
      
      const data = await response.json();
      
      // Stop Finalizing animation
      clearInterval(animationTimer);

      const results = data.results || [];
      if (results.length > 0 && results[0].rank === 0) {
        const snippet = results[0].snippet || "";
        if (typeof OfficeRuntime !== 'undefined' && OfficeRuntime.storage) {
          try {
            await OfficeRuntime.storage.setItem(cacheKey, snippet);
          } catch (e) {
            // Ignore storage set error
          }
        }
        // Step 4: Final Value
        invocation.setResult(snippet);
        return;
      }

      invocation.setResult(data.error || "No AI Overview returned.");
    } catch (err) {
      if (animationTimer) clearInterval(animationTimer);
      invocation.setResult(`Connection Error: ${err.message}`);
    }
  }, 100);
}

// Associate function definition with Excel manifest namespace
CustomFunctions.associate("ULTRA_SEARCH", ULTRA_SEARCH);
