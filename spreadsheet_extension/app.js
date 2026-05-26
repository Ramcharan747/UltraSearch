// UltraSearch Sidebar Application Core
let adapter = null;
let isRunning = false;
let stopRequested = false;
let activeAbortController = null;
let connectionTimer = null;
let activeAnimations = new Map(); // rowIndex -> intervalTimer
let globalHeaders = [];

// UI Elements
const variableChips = document.getElementById('variable-chips');
const autocompletePopup = document.getElementById('autocomplete-popup');
const statusDot = document.getElementById('status-dot');
const statusText = document.getElementById('status-text');
const serverUrlInput = document.getElementById('server-url');
const goalPromptInput = document.getElementById('goal-prompt');
const chkAutoEnrich = document.getElementById('chk-auto-enrich');
const batchSizeGroup = document.getElementById('batch-size-group');
const batchSizeInput = document.getElementById('batch-size');
const btnRun = document.getElementById('btn-run');
const btnStop = document.getElementById('btn-stop');
const progressPercentage = document.getElementById('progress-percentage');
const progressFill = document.getElementById('progress-fill');
const activeRowDetails = document.getElementById('active-row-details');
const activeRowText = document.getElementById('active-row-text');
const streamEmpty = document.getElementById('stream-empty');
const streamList = document.getElementById('stream-list');
const debugTerminal = document.getElementById('debug-terminal');
const btnClearDebug = document.getElementById('btn-clear-debug');
const chkUseCache = document.getElementById('chk-use-cache');
const btnClearCache = document.getElementById('btn-clear-cache');
const manualQueryCol = document.getElementById('manual-query-col');
const btnRefreshRows = document.getElementById('btn-refresh-rows');
const multiRowChecklist = document.getElementById('multi-row-checklist');
const outputColDropdown = document.getElementById('output-col-dropdown');
const queryColsChecklist = document.getElementById('query-cols-checklist');
const btnReloadSidebar = document.getElementById('btn-reload-sidebar');

// New UI Elements
const btnSettingsToggle = document.getElementById('btn-settings-toggle');
const settingsPanel = document.getElementById('settings-panel');
const mainUiPanel = document.getElementById('main-ui-panel');
const progressPanel = document.getElementById('progress-panel');
const btnBack = document.getElementById('btn-back');
const chatHistoryContainer = document.getElementById('chat-history-container');
const activeSourceGroup = document.getElementById('active-source-group');

// Default goal prompt
const DEFAULT_GOAL = "Extract the valuation, latest funding round, and key investors.";
goalPromptInput.value = DEFAULT_GOAL;

// Initialize
// Initialize Application
async function initApp() {
  let savedUrl = localStorage.getItem("ULTRA_SEARCH_API_URL");
  if (!savedUrl || savedUrl.startsWith("http://localhost") || savedUrl.startsWith("http://127.0.0.1") || !savedUrl.includes("trycloudflare") || savedUrl.includes("proc-rolling-kent-commercial")) {
    savedUrl = "https://april-aorta-sandal.ngrok-free.dev";
    localStorage.setItem("ULTRA_SEARCH_API_URL", savedUrl);
  }

  try {
    if (typeof OfficeRuntime !== 'undefined' && OfficeRuntime.storage) {
      await OfficeRuntime.storage.setItem("ULTRA_SEARCH_API_URL", savedUrl);
    }
  } catch (e) {
    // Ignore error
  }

  serverUrlInput.value = savedUrl;

  const savedGoal = localStorage.getItem("ULTRA_SEARCH_GOAL_PROMPT");
  if (savedGoal) {
    goalPromptInput.value = savedGoal;
  }

  const autoEnrichSaved = localStorage.getItem("ULTRA_SEARCH_AUTO_ENRICH") === "true";
  chkAutoEnrich.checked = autoEnrichSaved;
  
  const useCacheSaved = localStorage.getItem("ULTRA_SEARCH_USE_CACHE") !== "false";
  chkUseCache.checked = useCacheSaved;
  
  if (autoEnrichSaved) {
    logDebug("Re-registering auto-enrichment listener...");
    if (adapter.constructor.name === "GSheetsAdapter") {
      google.script.run.setupInstallableTrigger();
    } else {
      await adapter.registerChangeListener(handleLiveChange);
    }
  }

  // Set up event listeners
  document.querySelectorAll('input[name="search-mode"]').forEach(radio => {
    radio.addEventListener('change', (e) => {
      batchSizeGroup.style.display = e.target.value === 'batch' ? 'block' : 'none';
    });
  });

  // Persist AI/HTTP search mode selection
  const savedAiMode = localStorage.getItem("ULTRA_SEARCH_AI_MODE");
  if (savedAiMode) {
    const radio = document.querySelector(`input[name="ai-search-mode"][value="${savedAiMode}"]`);
    if (radio) radio.checked = true;
  }
  document.querySelectorAll('input[name="ai-search-mode"]').forEach(radio => {
    radio.addEventListener('change', (e) => {
      localStorage.setItem("ULTRA_SEARCH_AI_MODE", e.target.value);
      logDebug(`Search engine mode changed to: ${e.target.value === 'only' ? 'AI Overview' : 'HTTP'}`);
    });
  });

  document.querySelectorAll('input[name="source-mode"]').forEach(radio => {
    radio.addEventListener('change', (e) => {
      const isManual = e.target.value === 'manual';
      document.getElementById('manual-source-group').style.display = isManual ? 'block' : 'none';
      document.getElementById('live-selection-panel').style.display = isManual ? 'none' : 'flex';
      // Context columns hidden from UI — always keep it hidden
      activeSourceGroup.style.display = 'none';
      
      if (isManual) {
        refreshRowList();
      }
    });
  });

  // Slide-out Drawer Elements
  const drawerPanel = document.getElementById('drawer-panel');
  const drawerBackdrop = document.getElementById('drawer-backdrop');
  const btnDrawerClose = document.getElementById('btn-drawer-close');
  const btnDrawerSettings = document.getElementById('btn-drawer-settings');
  const drawerChevron = btnDrawerSettings ? btnDrawerSettings.querySelector('.drawer-chevron') : null;

  // Navigation Panel Views Elements
  const navViewsPanel = document.getElementById('nav-views-panel');
  const navViewTitle = document.getElementById('nav-view-title');
  const btnNavViewBack = document.getElementById('btn-nav-view-back');

  // Toggle Drawer visibility
  function openDrawer() {
    if (drawerPanel && drawerBackdrop) {
      drawerPanel.classList.add('active');
      drawerBackdrop.classList.add('active');
    }
  }

  function closeDrawer() {
    if (drawerPanel && drawerBackdrop) {
      drawerPanel.classList.remove('active');
      drawerBackdrop.classList.remove('active');
    }
  }

  if (btnSettingsToggle) {
    btnSettingsToggle.addEventListener('click', openDrawer);
  }
  if (btnDrawerClose) {
    btnDrawerClose.addEventListener('click', closeDrawer);
  }
  if (drawerBackdrop) {
    drawerBackdrop.addEventListener('click', closeDrawer);
  }

  // Settings Accordion inside Drawer
  if (btnDrawerSettings) {
    btnDrawerSettings.addEventListener('click', () => {
      const isHidden = settingsPanel.style.display === 'none';
      settingsPanel.style.display = isHidden ? 'flex' : 'none';
      if (drawerChevron) {
        drawerChevron.classList.toggle('rotated', !isHidden);
      }
    });
  }

  // Navigation Routing inside Drawer
  document.querySelectorAll('.drawer-nav-item[data-section]').forEach(button => {
    button.addEventListener('click', () => {
      const section = button.getAttribute('data-section');
      if (section) {
        showNavViewSection(section);
      }
    });
  });

  // Function to show a specific navigation view section (SPA transition / Popping Modal overlay)
  function showNavViewSection(sectionId) {
    closeDrawer();
    
    // Show backdrop for the modal view
    if (drawerBackdrop) {
      drawerBackdrop.classList.add('active');
    }

    // Show nav views panel
    if (navViewsPanel) navViewsPanel.style.display = 'flex';

    // Hide all section views inside the nav views panel
    document.querySelectorAll('.nav-section-view').forEach(view => {
      view.style.display = 'none';
    });

    // Show selected section view
    const selectedView = document.getElementById(`section-${sectionId}`);
    if (selectedView) {
      selectedView.style.display = 'flex';
    }

    // Set Title
    let titleText = sectionId.charAt(0).toUpperCase() + sectionId.slice(1);
    if (sectionId === 'templates') titleText = 'Saved Templates';
    if (sectionId === 'support') titleText = 'Report / Support';
    if (sectionId === 'guide') titleText = 'Documents / Guide';
    if (navViewTitle) {
      navViewTitle.textContent = titleText;
    }

    // Load data for specific sections
    if (sectionId === 'history') {
      loadNavHistoryList();
    } else if (sectionId === 'templates') {
      loadNavTemplatesList();
    } else if (sectionId === 'styles') {
      loadNavStylesList();
    } else if (sectionId === 'marketplace') {
      loadNavMarketplaceList();
    }

    // Highlight active item in drawer
    document.querySelectorAll('.drawer-nav-item').forEach(item => {
      if (item.getAttribute('data-section') === sectionId) {
        item.classList.add('active');
      } else {
        item.classList.remove('active');
      }
    });
  }

  // Go Back from nav views panel
  if (btnNavViewBack) {
    btnNavViewBack.addEventListener('click', () => {
      if (navViewsPanel) navViewsPanel.style.display = 'none';
      if (drawerBackdrop) {
        drawerBackdrop.classList.remove('active');
      }
      
      // Remove active highlight from drawer items
      document.querySelectorAll('.drawer-nav-item').forEach(item => {
        item.classList.remove('active');
      });
    });
  }

  // Back button in progress panel
  btnBack.addEventListener('click', () => {
    progressPanel.style.display = 'none';
    mainUiPanel.style.display = 'flex';
  });

  serverUrlInput.addEventListener('change', async (e) => {
    const newUrl = e.target.value.trim();
    localStorage.setItem("ULTRA_SEARCH_API_URL", newUrl);
    try {
      if (typeof OfficeRuntime !== 'undefined' && OfficeRuntime.storage) {
        await OfficeRuntime.storage.setItem("ULTRA_SEARCH_API_URL", newUrl);
      }
    } catch (err) {
      // Ignore error
    }
    logDebug("API server URL updated.");
  });

  goalPromptInput.addEventListener('change', (e) => {
    localStorage.setItem("ULTRA_SEARCH_GOAL_PROMPT", e.target.value.trim());
  });

  chkAutoEnrich.addEventListener('change', async (e) => {
    const isChecked = e.target.checked;
    localStorage.setItem("ULTRA_SEARCH_AUTO_ENRICH", isChecked);
    
    if (isChecked) {
      logDebug("Live Auto-Enrichment enabled.");
      if (adapter.constructor.name === "GSheetsAdapter") {
        google.script.run.setupInstallableTrigger();
      } else {
        await adapter.registerChangeListener(handleLiveChange);
      }
      adapter.showNotification("Live Auto-Enrichment enabled. Edits to Column A will enrich automatically.", "info");
    } else {
      logDebug("Live Auto-Enrichment disabled.");
      if (adapter.constructor.name !== "GSheetsAdapter") {
        await adapter.deregisterChangeListener();
      }
      adapter.showNotification("Live Auto-Enrichment disabled.", "info");
    }
  });

  btnRun.addEventListener('click', startPipeline);
  btnStop.addEventListener('click', stopPipeline);
  btnClearDebug.addEventListener('click', () => { debugTerminal.textContent = ''; });
  btnReloadSidebar.addEventListener('click', () => { window.location.reload(true); });

  chkUseCache.addEventListener('change', (e) => {
    localStorage.setItem("ULTRA_SEARCH_USE_CACHE", e.target.checked);
  });

  btnClearCache.addEventListener('click', () => {
    let clearedCount = 0;
    const keys = Object.keys(localStorage);
    keys.forEach(k => {
      if (k.startsWith("us_cache_")) {
        localStorage.removeItem(k);
        clearedCount++;
      }
    });
    logDebug(`Cleared ${clearedCount} cached search results from local storage.`);
    adapter.showNotification(`Cleared ${clearedCount} cached entries!`, "success");
  });

  manualQueryCol.addEventListener('change', refreshRowList);
  outputColDropdown.addEventListener('change', refreshRowList);
  btnRefreshRows.addEventListener('click', () => {
    refreshHeaders().then(refreshRowList);
  });

  // Show connecting state immediately on load
  statusDot.className = 'indicator-dot connecting';
  statusText.textContent = 'CONNECTING...';

  checkServerConnection();
  connectionTimer = setInterval(checkServerConnection, 5000);
  
  // Live Selection Polling
  setInterval(pollSelection, 1500);
  
  // Initial dropdown and checklist population
  refreshHeaders().then(refreshRowList);
  
  // Load History
  loadHistory();
  autoGrowTextarea();
}

// Window Load Listener + OfficeJS ready wait
window.addEventListener('load', () => {
  if (typeof Office !== 'undefined') {
    Office.onReady(async (info) => {
      adapter = getActiveAdapter();
      logDebug("Initializing adapter: " + adapter.constructor.name);
      await initApp();
    });
  } else {
    // Fallback for mock harness
    adapter = getActiveAdapter();
    logDebug("Initializing adapter (Mock fallback): " + adapter.constructor.name);
    initApp();
  }
});

// Load Query Column headers into dropdown
async function refreshHeaders() {
  try {
    const headers = await adapter.getHeaders();
    manualQueryCol.innerHTML = '';
    outputColDropdown.innerHTML = '';
    queryColsChecklist.innerHTML = '';
    
    headers.forEach((h, idx) => {
      // Only include columns that have valid header data
      if (h && h.trim() !== '') {
        const opt = document.createElement('option');
        opt.value = idx;
        opt.textContent = `${String.fromCharCode(65 + idx)} - ${h.trim()}`;
        
        manualQueryCol.appendChild(opt);
        outputColDropdown.appendChild(opt.cloneNode(true));
        
        // Add a checkbox to queryColsChecklist
        const itemDiv = document.createElement('div');
        itemDiv.style.display = 'flex';
        itemDiv.style.alignItems = 'center';
        itemDiv.style.gap = '6px';
        
        const chk = document.createElement('input');
        chk.type = 'checkbox';
        chk.className = 'query-col-checkbox';
        chk.value = idx;
        chk.id = `qcol-${idx}`;
        chk.style.width = 'auto';
        chk.style.margin = '0';
        chk.style.cursor = 'pointer';
        
        // Leave checkboxes unchecked by default so ExcelAdapter falls back to active selected column
        
        const label = document.createElement('label');
        label.htmlFor = `qcol-${idx}`;
        label.textContent = `${String.fromCharCode(65 + idx)} - ${h.trim()}`;
        label.style.cursor = 'pointer';
        label.style.fontSize = '11px';
        label.style.margin = '0';
        
        itemDiv.appendChild(chk);
        itemDiv.appendChild(label);
        queryColsChecklist.appendChild(itemDiv);
      }
    });
    
    // Default output column to the first option, or a sensible default
    if (outputColDropdown.options.length > 0) {
      // If there's an option for Column K (index 10), default to it, otherwise default to the last column
      let defaultIdx = outputColDropdown.options.length - 1;
      for (let i = 0; i < outputColDropdown.options.length; i++) {
        if (parseInt(outputColDropdown.options[i].value, 10) === 10) {
          defaultIdx = i;
          break;
        }
      }
      outputColDropdown.selectedIndex = defaultIdx;
    }
    
    // Save to globalHeaders and render chips
    globalHeaders = headers;
    variableChips.innerHTML = '';
    headers.forEach((h) => {
      if (h && h.trim() !== '') {
        const cleanH = h.trim();
        const chip = document.createElement('span');
        chip.className = 'variable-chip';
        chip.textContent = `/${cleanH.replace(/\s+/g, '')}`;
        chip.title = `Click to insert variable for ${cleanH}`;
        chip.addEventListener('click', () => {
          insertVariableInPrompt(`/${cleanH.replace(/\s+/g, '')}`);
        });
        variableChips.appendChild(chip);
      }
    });
  } catch (e) {
    logDebug("Error loading headers: " + e);
  }
}

function insertVariableInPrompt(varText) {
  const start = goalPromptInput.selectionStart;
  const end = goalPromptInput.selectionEnd;
  const text = goalPromptInput.value;
  const before = text.substring(0, start);
  const after = text.substring(end, text.length);
  goalPromptInput.value = before + varText + after;
  goalPromptInput.selectionStart = goalPromptInput.selectionEnd = start + varText.length;
  goalPromptInput.focus();
}

// Autocomplete logic for "/" and "\" triggers
let activeSuggestionIdx = 0;
let autocompleteQueryStart = -1;

goalPromptInput.addEventListener('input', (e) => {
  autoGrowTextarea();
  const welcome = document.getElementById('chat-welcome-placeholder');
  if (welcome) welcome.style.display = 'none';

  const text = goalPromptInput.value;
  const caretPos = goalPromptInput.selectionStart;
  const textBeforeCaret = text.substring(0, caretPos);
  
  const lastSlashIdx = textBeforeCaret.lastIndexOf('/');
  const lastBackslashIdx = textBeforeCaret.lastIndexOf('\\');
  
  let triggerIdx = -1;
  let triggerChar = '';
  
  if (lastSlashIdx !== -1 && lastSlashIdx >= textBeforeCaret.lastIndexOf(' ')) {
    triggerIdx = lastSlashIdx;
    triggerChar = '/';
  }
  
  if (lastBackslashIdx !== -1 && lastBackslashIdx >= textBeforeCaret.lastIndexOf(' ') && lastBackslashIdx > lastSlashIdx) {
    triggerIdx = lastBackslashIdx;
    triggerChar = '\\';
  }
  
  if (triggerIdx !== -1) {
    const query = textBeforeCaret.substring(triggerIdx + 1);
    autocompleteQueryStart = triggerIdx;
    showAutocompleteSuggestions(query, triggerChar);
  } else {
    hideAutocompleteSuggestions();
  }
});

goalPromptInput.addEventListener('focus', () => {
  const welcome = document.getElementById('chat-welcome-placeholder');
  if (welcome) welcome.style.display = 'none';
});

goalPromptInput.addEventListener('click', () => {
  const welcome = document.getElementById('chat-welcome-placeholder');
  if (welcome) welcome.style.display = 'none';
});

goalPromptInput.addEventListener('keydown', (e) => {
  if (autocompletePopup.style.display === 'block') {
    const items = autocompletePopup.querySelectorAll('.autocomplete-item');
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      activeSuggestionIdx = (activeSuggestionIdx + 1) % items.length;
      highlightActiveSuggestion(items);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      activeSuggestionIdx = (activeSuggestionIdx - 1 + items.length) % items.length;
      highlightActiveSuggestion(items);
    } else if (e.key === 'Enter' || e.key === 'Tab') {
      e.preventDefault();
      const activeItem = items[activeSuggestionIdx];
      if (activeItem) activeItem.click();
    } else if (e.key === 'Escape') {
      e.preventDefault();
      hideAutocompleteSuggestions();
    }
  }
});

function showAutocompleteSuggestions(query, triggerChar = '/') {
  const headers = globalHeaders || [];
  const cleanQuery = query.toLowerCase();
  const matches = headers.filter(h => h && h.trim() !== '' && h.trim().toLowerCase().replace(/\s+/g, '').includes(cleanQuery));
  
  if (matches.length === 0) {
    hideAutocompleteSuggestions();
    return;
  }
  
  autocompletePopup.innerHTML = '';
  activeSuggestionIdx = 0;
  
  matches.forEach((m, idx) => {
    const item = document.createElement('div');
    item.className = 'autocomplete-item';
    item.style.padding = '6px 10px';
    item.style.fontSize = '11px';
    item.style.color = '#e2e8f0';
    item.style.cursor = 'pointer';
    item.style.borderRadius = '4px';
    
    const cleanVar = m.trim().replace(/\s+/g, '');
    item.textContent = triggerChar === '\\' ? `\\${cleanVar}` : `/${cleanVar}`;
    
    item.addEventListener('click', () => {
      const varText = triggerChar === '\\' ? `\\${cleanVar}` : `/${cleanVar}`;
      const text = goalPromptInput.value;
      const before = text.substring(0, autocompleteQueryStart);
      const after = text.substring(goalPromptInput.selectionStart);
      goalPromptInput.value = before + varText + after;
      goalPromptInput.selectionStart = goalPromptInput.selectionEnd = autocompleteQueryStart + varText.length;
      goalPromptInput.focus();
      hideAutocompleteSuggestions();
    });
    
    autocompletePopup.appendChild(item);
  });
  
  const rect = goalPromptInput.getBoundingClientRect();
  autocompletePopup.style.top = `${rect.bottom + window.scrollY}px`;
  autocompletePopup.style.left = `${rect.left + window.scrollX}px`;
  autocompletePopup.style.width = `${rect.width}px`;
  autocompletePopup.style.display = 'block';
  
  highlightActiveSuggestion(autocompletePopup.querySelectorAll('.autocomplete-item'));
}

function highlightActiveSuggestion(items) {
  items.forEach((item, idx) => {
    if (idx === activeSuggestionIdx) {
      item.style.background = '#4338ca';
      item.style.color = '#ffffff';
    } else {
      item.style.background = 'none';
      item.style.color = '#e2e8f0';
    }
  });
}

function hideAutocompleteSuggestions() {
  autocompletePopup.style.display = 'none';
}

document.addEventListener('click', (e) => {
  if (e.target !== goalPromptInput && !autocompletePopup.contains(e.target)) {
    hideAutocompleteSuggestions();
  }
});

// Load Row Checklist from chosen Column (Filtering out empty rows)
async function refreshRowList() {
  const colIdx = parseInt(manualQueryCol.value, 10);
  if (isNaN(colIdx)) return;
  
  multiRowChecklist.innerHTML = '<div style="padding: 12px; text-align: center; color: var(--text-muted); font-size: 11px;">Scanning rows...</div>';
  
  try {
    const queryRows = await adapter.getAllRows(colIdx);
    
    // Also fetch the output column to check for blanks
    const outColIdx = parseInt(outputColDropdown.value, 10);
    const outRows = isNaN(outColIdx) ? [] : await adapter.getAllRows(outColIdx);
    
    multiRowChecklist.innerHTML = '';
    
    // Filter out rows that do not have data in the query column, OR already have data in the output column
    const validRows = queryRows.filter(r => {
      const hasQueryData = r.value && r.value.trim() !== '';
      const outRow = outRows.find(out => out.index === r.index);
      const isOutputEmpty = !outRow || !outRow.value || outRow.value.trim() === '';
      return hasQueryData && isOutputEmpty;
    });
    
    if (validRows.length === 0) {
      multiRowChecklist.innerHTML = '<div style="padding: 12px; text-align: center; color: var(--text-muted); font-size: 11px;">No valid rows found (all output cells may be full).</div>';
      return;
    }
    
    validRows.forEach(r => {
      const div = document.createElement('label');
      div.className = 'checklist-item';
      
      const chk = document.createElement('input');
      chk.type = 'checkbox';
      chk.value = r.index;
      chk.className = 'row-checkbox';
      chk.checked = true; // Checked by default since it has data
      
      const text = document.createElement('span');
      text.className = 'row-val';
      text.innerHTML = `Row ${r.index + 1}: <strong>${escapeHtml(r.value)}</strong>`;
      
      div.appendChild(chk);
      div.appendChild(text);
      multiRowChecklist.appendChild(div);
    });
  } catch (e) {
    logDebug("Error loading row checklist: " + e);
    multiRowChecklist.innerHTML = `<div style="padding: 12px; text-align: center; color: #f87171; font-size: 11px;">Error loading rows: ${e}</div>`;
  }
}

// Poll active spreadsheet selection and update sidebar UI
async function pollSelection() {
  if (isRunning) return;
  try {
    const rows = await adapter.getSelectedRows();
    const indicator = document.getElementById('selection-indicator');
    const text = document.getElementById('selection-text');
    
    if (rows && rows.length > 0) {
      const firstQuery = rows[0].query.trim();
      if (firstQuery) {
        indicator.className = 'selection-indicator active';
        text.textContent = rows.length === 1 
          ? `Selected: Row ${rows[0].index + 1} (${firstQuery})` 
          : `Selected: ${rows.length} Rows (from ${firstQuery})`;
      } else {
        indicator.className = 'selection-indicator empty';
        text.textContent = `Selected: Row ${rows[0].index + 1} (Col A is empty!)`;
      }
    } else {
      indicator.className = 'selection-indicator empty';
      text.textContent = 'No selection found.';
    }
  } catch (e) {
    // Ignore silent polling errors
  }
}

// Write to terminal
function logDebug(message) {
  const timestamp = new Date().toLocaleTimeString();
  debugTerminal.textContent += `[${timestamp}] ${message}\n`;
  debugTerminal.scrollTop = debugTerminal.scrollHeight;
}

// Ping UltraSearch Server (with fallback + Cloudflare interstitial bypass)
let lastConnectionLog = 0; // throttle debug logs
async function checkServerConnection() {
  const configuredUrl = serverUrlInput.value.trim();
  const urls = [configuredUrl, "http://localhost:8082"].filter(Boolean);
  const shouldLog = Date.now() - lastConnectionLog > 30000; // log every 30s max
  
  for (const baseUrl of urls) {
    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 4000);
      
      const resp = await fetch(`${baseUrl}/ping`, { 
        signal: controller.signal,
        mode: 'cors',
        cache: 'no-store',
        headers: { 'Accept': 'application/json' }
      });
      
      clearTimeout(timeoutId);
      
      if (resp.ok) {
        // Check that we got actual JSON back, not a Cloudflare interstitial HTML page
        const contentType = resp.headers.get('content-type') || '';
        if (contentType.includes('application/json')) {
          statusDot.className = 'indicator-dot online';
          statusText.textContent = 'ONLINE';
          
          // If we connected via localhost fallback, auto-update the URL field
          if (baseUrl !== configuredUrl && baseUrl.includes('localhost')) {
            if (shouldLog) logDebug(`[Connection] Tunnel unreachable — using localhost fallback.`);
          }
          lastConnectionLog = Date.now();
          return;
        } else {
          // Got a response but it's not JSON (likely Cloudflare interstitial)
          if (shouldLog) logDebug(`[Connection] ${baseUrl} returned non-JSON (possible Cloudflare interstitial). Trying next...`);
          continue;
        }
      } else if (resp.status === 400 || resp.status === 404) {
        // Server is reachable but endpoint may differ — still counts as online
        statusDot.className = 'indicator-dot online';
        statusText.textContent = 'ONLINE';
        lastConnectionLog = Date.now();
        return;
      }
    } catch (e) {
      if (shouldLog && urls.indexOf(baseUrl) === urls.length - 1) {
        logDebug(`[Connection] All endpoints unreachable: ${e.message || e}`);
      }
      continue;
    }
  }
  
  // All URLs failed
  statusDot.className = 'indicator-dot offline';
  statusText.textContent = 'DISCONNECTED';
  lastConnectionLog = Date.now();
}

// 4-state Loading Animations
function startStatusAnimation(rowIndex, colIndex, baseText) {
  stopStatusAnimation(rowIndex);
  
  const frames = [
    `${baseText}    `,
    `${baseText}.   `,
    `${baseText}..  `,
    `${baseText}... `
  ];
  let frameIdx = 0;
  
  const writeCol = colIndex !== undefined ? colIndex : 1;
  
  // Set initial status
  adapter.writeCell(rowIndex, writeCol, frames[frameIdx]);
  const rowItem = document.getElementById(`row-item-${rowIndex}`);
  
  const timer = setInterval(() => {
    frameIdx = (frameIdx + 1) % frames.length;
    const text = frames[frameIdx];
    
    // Only write live frames to cell if NOT Google Sheets to avoid Google Apps Script quota limits
    if (adapter.constructor.name !== "GSheetsAdapter") {
      adapter.writeCell(rowIndex, writeCol, text);
    }
    
    if (rowItem) {
      const badge = rowItem.querySelector('.status-badge');
      if (badge) {
        badge.textContent = text;
      }
    }
  }, 400);
  
  activeAnimations.set(rowIndex, timer);
}

function stopStatusAnimation(rowIndex) {
  if (activeAnimations.has(rowIndex)) {
    clearInterval(activeAnimations.get(rowIndex));
    activeAnimations.delete(rowIndex);
  }
}

// Start research loop
async function startPipeline() {
  if (isRunning) return;
  
  const sourceMode = document.querySelector('input[name="source-mode"]:checked').value;
  let selectedRows = [];
  
  if (sourceMode === 'manual') {
    const colIndex = parseInt(manualQueryCol.value, 10);
    const checkboxes = document.querySelectorAll('.row-checkbox:checked');
    if (checkboxes.length === 0) {
      adapter.showNotification("Please check at least one row from the checklist.", "error");
      return;
    }
    
    logDebug(`[Manual Mode] Fetching queries for ${checkboxes.length} selected checklist rows...`);
    for (let chk of checkboxes) {
      const rowIndex = parseInt(chk.value, 10);
      try {
        const val = await adapter.readCell(rowIndex, colIndex);
        selectedRows.push({
          index: rowIndex,
          query: val,
          colValues: []
        });
      } catch (e) {
        logDebug(`[Manual Mode] Error reading Row ${rowIndex + 1}: ${e}`);
      }
    }
  } else {
    logDebug("Fetching selected spreadsheet rows...");
    try {
      const checkedQueryCols = Array.from(document.querySelectorAll('.query-col-checkbox:checked')).map(chk => parseInt(chk.value, 10));
      const queryHeadersIndices = getHeadersInQuery();
      const mergedQueryColsSet = new Set([...checkedQueryCols, ...queryHeadersIndices]);
      const finalQueryCols = Array.from(mergedQueryColsSet);
      
      logDebug(`Selected context columns: [${finalQueryCols.join(', ')}]`);
      selectedRows = await adapter.getSelectedRows(finalQueryCols);
    } catch (e) {
      logDebug("Error reading selection: " + e);
      adapter.showNotification("Could not read row selection. Make sure you select cells.", "error");
      return;
    }
  }
  
  const validRows = selectedRows.filter(r => r.query && r.query.trim() !== '');
  
  if (validRows.length === 0) {
    logDebug("No query target found. Please ensure the target query cell is not empty.");
    adapter.showNotification("Please select or specify a cell that contains a search query.", "warning");
    return;
  }
  
  try {
    isRunning = true;
    stopRequested = false;
    
    // UI Transitions
    mainUiPanel.style.display = 'none';
    progressPanel.style.display = 'flex';
    btnBack.style.display = 'none';
    btnStop.style.display = 'block';
    btnStop.textContent = "Stop Execution";
    btnStop.disabled = false;
    
    streamEmpty.style.display = 'none';
    streamList.innerHTML = '';
    activeRowDetails.style.display = 'flex';
    
    saveToHistory(goalPromptInput.value); // Save to history
    
    logDebug(`Starting execution queue for ${validRows.length} rows...`);
    
    validRows.forEach(row => {
      const li = document.createElement('li');
      li.className = 'stream-item pending';
      li.id = `row-item-${row.index}`;
      li.innerHTML = `
        <div class="item-left">
          <span class="item-row-num">ROW ${row.index + 1}</span>
          <span class="item-query">${escapeHtml(row.query)}</span>
        </div>
        <div class="item-right">
          <span class="status-badge">PENDING</span>
        </div>
      `;
      streamList.appendChild(li);
    });
    
    const searchMode = document.querySelector('input[name="search-mode"]:checked').value;
    
    if (searchMode === 'single') {
      await runSingleMode(validRows);
    } else {
      const size = parseInt(batchSizeInput.value) || 3;
      await runBatchMode(validRows, size);
    }
  } catch (err) {
    logDebug("Execution encountered an error: " + err);
  } finally {
    isRunning = false;
    btnStop.style.display = 'none';
    btnStop.textContent = "Stop Execution";
    btnStop.disabled = false;
    btnBack.style.display = 'block';
    activeRowDetails.style.display = 'none';
    
    // Clear any active animations
    for (const rowIndex of activeAnimations.keys()) {
      stopStatusAnimation(rowIndex);
    }
    
    if (stopRequested) {
      logDebug("Pipeline execution stopped by user.");
      adapter.showNotification("Execution Stopped.", "info");
    } else {
      logDebug("Pipeline execution completed.");
    }
  }
}

// Stop execution
function stopPipeline() {
  if (!isRunning) return;
  stopRequested = true;
  btnStop.textContent = "Stopping...";
  btnStop.disabled = true;
  if (activeAbortController) {
    activeAbortController.abort();
    activeAbortController = null;
  }
  logDebug("Stop requested. Terminating current query...");
}

// ESCAPE HTML helper
function escapeHtml(text) {
  if (!text) return '';
  return text.toString()
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

// Regular Expression JSON Extractor
function extractJson(text) {
  if (!text) return null;
  const match = text.match(/(\{[\s\S]*\}|\[[\s\S]*\])/);
  if (match) {
    try {
      return JSON.parse(match[1]);
    } catch (e) {
      try {
        let cleaned = match[1]
          .replace(/,\s*([\]}])/g, '$1')
          .replace(/[\u201C\u201D]/g, '"');
        return JSON.parse(cleaned);
      } catch (err) {
        console.error("JSON parse failure even after clean-up", err);
      }
    }
  }
  return null;
}

// Handle Auto-Enrichment cell edit triggers
async function handleLiveChange(changeInfo) {
  const { row, value } = changeInfo;
  if (!value || value.trim() === "") return;

  logDebug(`[Auto-Enrich Trigger] Column A cell at Row ${row + 1} edited to "${value}". Processing...`);

  let UIItem = document.getElementById(`row-item-${row}`);
  if (!UIItem) {
    streamEmpty.style.display = 'none';
    UIItem = document.createElement('li');
    UIItem.className = 'stream-item pending';
    UIItem.id = `row-item-${row}`;
    UIItem.innerHTML = `
      <div class="item-left">
        <span class="item-row-num">ROW ${row + 1} [AUTO-ENRICH]</span>
        <span class="item-query">${escapeHtml(value)}</span>
      </div>
      <div class="item-right">
        <span class="status-badge">PENDING</span>
      </div>
    `;
    streamList.appendChild(UIItem);
  }

  const colValues = [];
  const headers = globalHeaders || [];
  for (let c = 0; c < headers.length; c++) {
    try {
      colValues.push(await adapter.readCell(row, c));
    } catch (e) {
      colValues.push("");
    }
  }

  await runSingleRowDirect(row, value, UIItem, colValues);
}

// Shared Single Row Execution Pipeline
async function runSingleRowDirect(rowIndex, queryText, UIItem, colValues) {
  if (stopRequested) {
    logDebug(`Row ${rowIndex + 1}: Skipped (stop requested).`);
    if (UIItem) {
      UIItem.className = 'stream-item pending';
      const badge = UIItem.querySelector('.status-badge');
      if (badge) badge.textContent = 'PENDING';
    }
    return;
  }
  UIItem.className = 'stream-item running';
  const startColIndex = parseInt(outputColDropdown.value, 10);
  const writeCol = !isNaN(startColIndex) ? startColIndex : 2; // Default to Col C
  
  // Highlight Row
  try {
    await adapter.setRowHighlight(rowIndex, true);
  } catch (highlightErr) {
    logDebug(`Row ${rowIndex + 1}: Styling highlight error (ignored): ${highlightErr}`);
  }
  
  // Step 1: Connect
  logDebug(`Row ${rowIndex + 1}: Connecting...`);
  await adapter.writeCell(rowIndex, writeCol, "⏳ Connecting...");
  
  const goal = goalPromptInput.value.trim();
  
  // Extract backslash column mappings BEFORE variable replacements
  const mappingRegex = /\{\{\s*([^}]+?)\s*\}\}\\([a-zA-Z0-9_\-]+)/g;
  let match;
  const mappings = [];
  const headers = globalHeaders || [];
  
  while ((match = mappingRegex.exec(goal)) !== null) {
    const fieldName = match[1].trim();
    const colName = match[2].trim();
    
    // Find matching column index in headers (case-insensitive, ignoring spaces)
    let matchedColIdx = -1;
    const cleanColName = colName.toLowerCase().replace(/[^a-z0-9]/g, '');
    for (let idx = 0; idx < headers.length; idx++) {
      if (headers[idx]) {
        const cleanHeader = headers[idx].toLowerCase().replace(/[^a-z0-9]/g, '');
        if (cleanHeader === cleanColName) {
          matchedColIdx = idx;
          break;
        }
      }
    }
    
    mappings.push({
      fullMatch: match[0],
      fieldName: fieldName,
      colName: colName,
      colIdx: matchedColIdx
    });
  }

  // Remove the mapping syntax so the AI gets a plain descriptive prompt
  let resolvedGoal = goal;
  mappings.forEach(m => {
    resolvedGoal = resolvedGoal.replace(m.fullMatch, m.fieldName);
  });

  if (colValues && colValues.length > 0) {
    headers.forEach((h, colIdx) => {
      if (h && h.trim() !== '') {
        const cleanHeader = h.trim();
        const val = colValues[colIdx] || "";
        
        // Replace {{HeaderName}}
        const braceRegex = new RegExp(`\\{\\{\\s*${escapeRegExp(cleanHeader)}\\s*\\}\\}`, 'gi');
        resolvedGoal = resolvedGoal.replace(braceRegex, val);
        
        // Replace /HeaderName (no spaces)
        const cleanVar = cleanHeader.replace(/\s+/g, '');
        const slashRegex = new RegExp(`/${escapeRegExp(cleanVar)}\\b`, 'gi');
        resolvedGoal = resolvedGoal.replace(slashRegex, val);

        // Replace /Header Name (exact spaces)
        const slashRegexSpaces = new RegExp(`/${escapeRegExp(cleanHeader)}\\b`, 'gi');
        resolvedGoal = resolvedGoal.replace(slashRegexSpaces, val);
      }
    });
    if (resolvedGoal !== goal) {
      logDebug(`[Row ${rowIndex + 1} Context] Resolved Goal: "${resolvedGoal}"`);
    }
  }

  const baseUrl = serverUrlInput.value.trim() || "http://localhost:8082";
  
  // Step 2: Start 4-state Searching Animation
  logDebug(`Row ${rowIndex + 1}: Searching...`);
  startStatusAnimation(rowIndex, writeCol, "🔍 Searching");
  
  // Write a single update to Google Sheets to indicate searching state
  if (adapter.constructor.name === "GSheetsAdapter") {
    await adapter.writeCell(rowIndex, writeCol, "🔍 Searching...");
  }
  
  let jsonInstructions = "";
  if (mappings.length > 0) {
    const schemaObj = {};
    mappings.forEach(m => {
      schemaObj[m.colName] = `extracted ${m.fieldName}`;
    });
    jsonInstructions = `Respond ONLY with a valid JSON object matching the requested attributes: ${JSON.stringify(schemaObj)}. ` +
      `Ensure you return only raw JSON. Do not include markdown code block backticks (\`\`\`) or introductory conversational text.`;
  } else {
    jsonInstructions = `Respond ONLY with a valid JSON object matching the requested attributes (e.g. {"attribute_name": "extracted_info"}). ` +
      `Ensure you return only raw JSON. Do not include markdown code block backticks (\`\`\`) or introductory conversational text.`;
  }
  
  const aiSearchMode = document.querySelector('input[name="ai-search-mode"]:checked')?.value || "none";
  logDebug(`Row ${rowIndex + 1}: ai_mode=${aiSearchMode}`);
  let rephrasedQuery = "";
  if (aiSearchMode === "none") {
    // HTTP mode: build a concise Google-friendly query instead of sending the whole system prompt
    if (mappings.length > 0) {
      // Extract just the field names from mappings for a focused search
      const fieldTerms = mappings.map(m => m.fieldName).join(' ');
      rephrasedQuery = `${queryText} ${fieldTerms}`.trim().replace(/\s+/g, ' ').substring(0, 150);
    } else {
      rephrasedQuery = `${queryText}`.trim().replace(/\s+/g, ' ').substring(0, 150);
    }
  } else {
    // AI Overview mode: send a concise, natural Google query — NOT a verbose LLM prompt
    if (mappings.length > 0) {
      const fieldTerms = mappings.map(m => m.fieldName).join(', ');
      // Build a natural question
      rephrasedQuery = `what is the ${fieldTerms} for ${queryText}?`;
      
      const schemaObj = {};
      mappings.forEach(m => { schemaObj[m.colName] = `extracted ${m.fieldName}`; });
      rephrasedQuery += ` format the answer as a JSON object like ${JSON.stringify(schemaObj)}`;
    } else {
      rephrasedQuery = `${queryText} (format answer as a JSON object)`;
    }
  }
  const url = `${baseUrl}/search?q=${encodeURIComponent(rephrasedQuery)}&limit=5&ai_mode=${aiSearchMode}`;
  
  let success = false;
  let parsedData = null;
  let errorMsg = "";
  
  const cacheKey = "us_cache_" + queryText.toLowerCase().trim();
  const cachedVal = chkUseCache.checked ? localStorage.getItem(cacheKey) : null;
  
  if (cachedVal) {
    logDebug(`Row ${rowIndex + 1}: Found result in local cache! Skipping API...`);
    try {
      parsedData = JSON.parse(cachedVal);
      success = true;
    } catch (e) {
      logDebug(`Row ${rowIndex + 1}: Cache parse error. Re-fetching.`);
    }
  }
  
  if (!success) {
    try {
      activeAbortController = new AbortController();
      const resp = await fetch(url, { signal: activeAbortController.signal });
      activeAbortController = null;
    
    // Stop Searching animation, start Finalizing animation
    stopStatusAnimation(rowIndex);
    startStatusAnimation(rowIndex, writeCol, "✍️ Finalizing");
    
    if (adapter.constructor.name === "GSheetsAdapter") {
      await adapter.writeCell(rowIndex, writeCol, "✍️ Finalizing...");
    }

    if (resp.ok) {
      const data = await resp.json();
      const results = data.results || [];
      if (results.length > 0 && results[0].rank === 0) {
        const aiResponse = results[0].snippet || "";
        parsedData = extractJson(aiResponse);
        
        if (parsedData) {
          success = true;
        } else {
          parsedData = { "raw_overview": aiResponse };
          success = true; 
          logDebug(`Row ${rowIndex + 1}: Failed parsing JSON. Writing raw overview as fallback.`);
        }
      } else if (results.length > 0) {
        // Fallback for HTTP search or organic results
        const textResponse = results[0].snippet || results[0].title || "";
        parsedData = extractJson(textResponse);
        if (parsedData) {
          success = true;
        } else {
          parsedData = { "result": textResponse };
          success = true;
          logDebug(`Row ${rowIndex + 1}: Writing first organic snippet/title as fallback.`);
        }
      } else {
        errorMsg = data.error || "No search results returned.";
      }
    } else {
      errorMsg = `HTTP Error ${resp.status}: ${resp.statusText}`;
    }
  } catch (err) {
    activeAbortController = null;
    if (err.name === 'AbortError') {
      errorMsg = "Stopped by user.";
    } else {
      errorMsg = err.message || "Network Error connecting to UltraSearch.";
    }
  }
  }
  
  // Stop all status animations for this row
  stopStatusAnimation(rowIndex);

  if (success && parsedData) {
    UIItem.className = 'stream-item success';
    UIItem.querySelector('.status-badge').textContent = 'SUCCESS';
    logDebug(`Row ${rowIndex + 1}: SUCCESS. Writing results back to spreadsheet.`);
    
    const cacheKey = "us_cache_" + queryText.toLowerCase().trim();
    const rawVal = parsedData.raw_overview || JSON.stringify(parsedData);
    localStorage.setItem(cacheKey, rawVal);
    
    delete parsedData.raw_overview;
    
    if (mappings.length > 0) {
      // Write mapped fields specifically to their target columns
      for (const m of mappings) {
        if (m.colIdx !== -1) {
          let val = undefined;
          const normColName = m.colName.toLowerCase().replace(/[^a-z0-9]/g, '');
          const normFieldName = m.fieldName.toLowerCase().replace(/[^a-z0-9]/g, '');
          
          for (const key of Object.keys(parsedData)) {
            const normKey = key.toLowerCase().replace(/[^a-z0-9]/g, '');
            if (normKey === normColName || normKey === normFieldName) {
              val = parsedData[key];
              break;
            }
          }
          if (val === undefined || val === null) {
            // Fallback: if no JSON key matched, use the raw snippet/result text
            val = parsedData.result || parsedData.raw_overview || "";
            if (val) {
              logDebug(`Row ${rowIndex + 1}: No exact JSON key match for "${m.fieldName}". Using organic snippet as fallback.`);
            }
          }
          let strVal = typeof val === 'object' ? JSON.stringify(val) : val.toString();
          if (strVal.startsWith("=")) {
            strVal = "'" + strVal;
          }
          try {
            await adapter.writeCell(rowIndex, m.colIdx, strVal);
            logDebug(`Row ${rowIndex + 1}: Wrote "${strVal}" to mapped column "${m.colName}" (Col index ${m.colIdx})`);
          } catch (writeErr) {
            logDebug(`Row ${rowIndex + 1}: Error writing to mapped column "${m.colName}": ${writeErr.message || writeErr}`);
          }
        } else {
          logDebug(`Row ${rowIndex + 1}: Target column "${m.colName}" was not found in sheet headers.`);
        }
      }
    } else {
      // Excel formula safety sanitation
      const keys = Object.keys(parsedData);
      const values = keys.map(k => {
        let val = parsedData[k];
        if (val === null || val === undefined) return "";
        let strVal = typeof val === 'object' ? JSON.stringify(val) : val.toString();
        // Prevent formula injection in Excel by prepending a single quote if it starts with '='
        if (strVal.startsWith("=")) {
          strVal = "'" + strVal;
        }
        return strVal;
      });
      
      try {
        await adapter.writeRow(rowIndex, writeCol, values);
      } catch (writeErr) {
        logDebug(`Row ${rowIndex + 1}: Excel writeRow error: ${writeErr.message || writeErr}. Trying fallback writeCell...`);
        try {
          const fallbackVal = values.join(" | ");
          await adapter.writeCell(rowIndex, writeCol, fallbackVal);
        } catch (cellErr) {
          logDebug(`Row ${rowIndex + 1}: FAILED writing cells: ${cellErr.message || cellErr}`);
        }
      }
    }
  } else {
    UIItem.className = 'stream-item failed';
    UIItem.querySelector('.status-badge').textContent = 'FAILED';
    logDebug(`Row ${rowIndex + 1}: FAILED: ${errorMsg}`);
    
    let safeError = errorMsg || "Unknown Error";
    if (safeError.startsWith("=")) {
      safeError = "'" + safeError;
    }
    try {
      await adapter.writeCell(rowIndex, writeCol, `FAILED: ${safeError}`);
    } catch (cellErr) {
      logDebug(`Row ${rowIndex + 1}: FAILED writing error cell: ${cellErr}`);
    }
  }
  
  // Remove Row Highlight
  try {
    await adapter.setRowHighlight(rowIndex, false);
  } catch (highlightErr) {
    logDebug(`Row ${rowIndex + 1}: Styling unhighlight error (ignored): ${highlightErr}`);
  }
}

// 1. Single Row Mode Execution (Queue runner)
async function runSingleMode(rows) {
  for (let i = 0; i < rows.length; i++) {
    if (stopRequested) break;
    
    const row = rows[i];
    const rowItem = document.getElementById(`row-item-${row.index}`);
    
    activeRowText.textContent = `Processing Row ${row.index + 1}: ${row.query}`;
    
    await runSingleRowDirect(row.index, row.query, rowItem, row.colValues);
    
    const progress = Math.round(((i + 1) / rows.length) * 100);
    progressPercentage.textContent = `${progress}%`;
    progressFill.style.width = `${progress}%`;
    
    await new Promise(resolve => setTimeout(resolve, 800));
  }
}

// 2. Batched Mode Execution (Combines 2-5 rows in one query)
async function runBatchMode(rows, batchSize) {
  const goal = goalPromptInput.value.trim();
  const baseUrl = serverUrlInput.value.trim() || "http://localhost:8082";
  
  // Extract backslash column mappings BEFORE variable replacements
  const mappingRegex = /\{\{\s*([^}]+?)\s*\}\}\\([a-zA-Z0-9_\-]+)/g;
  let match;
  const mappings = [];
  const headers = globalHeaders || [];
  
  while ((match = mappingRegex.exec(goal)) !== null) {
    const fieldName = match[1].trim();
    const colName = match[2].trim();
    
    // Find matching column index in headers
    let matchedColIdx = -1;
    const cleanColName = colName.toLowerCase().replace(/[^a-z0-9]/g, '');
    for (let idx = 0; idx < headers.length; idx++) {
      if (headers[idx]) {
        const cleanHeader = headers[idx].toLowerCase().replace(/[^a-z0-9]/g, '');
        if (cleanHeader === cleanColName) {
          matchedColIdx = idx;
          break;
        }
      }
    }
    
    mappings.push({
      fullMatch: match[0],
      fieldName: fieldName,
      colName: colName,
      colIdx: matchedColIdx
    });
  }

  // Remove the mapping syntax so the AI gets a plain descriptive prompt
  let resolvedGoal = goal;
  mappings.forEach(m => {
    resolvedGoal = resolvedGoal.replace(m.fullMatch, m.fieldName);
  });

  let processedCount = 0;
  
  for (let i = 0; i < rows.length; i += batchSize) {
    if (stopRequested) break;
    
    const batch = rows.slice(i, i + batchSize);
    logDebug(`Creating batch of ${batch.length} rows...`);
    
    batch.forEach(row => {
      const rowItem = document.getElementById(`row-item-${row.index}`);
      rowItem.className = 'stream-item running';
      rowItem.querySelector('.status-badge').textContent = 'RUNNING';
    });
    
    activeRowText.textContent = `Batch Processing Rows: ${batch.map(r => r.index + 1).join(', ')}`;
    
    const entitiesList = batch.map((r, index) => `${index + 1}. "${r.query}"`).join('\n');
    
    const aiSearchMode = document.querySelector('input[name="ai-search-mode"]:checked')?.value || "none";
    let batchedQuery = "";
    if (aiSearchMode === "none") {
      // For HTTP search, keep the query extremely clean and short to prevent search failures
      batchedQuery = `companies:\n${batch.map(r => r.query).join(', ')}\nGoal: ${resolvedGoal}`.substring(0, 150);
    } else {
      if (mappings.length > 0) {
        const exampleObj = { "entity": "Company1" };
        mappings.forEach(m => {
          exampleObj[m.colName] = `extracted ${m.fieldName}`;
        });
        batchedQuery = `Extract details for these companies/topics:\n${entitiesList}\n\n` +
          `Extraction Goal: ${resolvedGoal}\n\n` +
          `Respond ONLY with a valid JSON array of objects, one for each entity in the exact order requested. ` +
          `Example format:\n${JSON.stringify([exampleObj, { ...exampleObj, entity: "Company2" }], null, 2)}\n` +
          `Do not include markdown code block block backticks (\`\`\`) or introductory conversational text.`;
      } else {
        batchedQuery = `Extract details for these companies/topics:\n${entitiesList}\n\n` +
          `Extraction Goal: ${resolvedGoal}\n\n` +
          `Respond ONLY with a valid JSON array of objects, one for each entity in the exact order requested. ` +
          `Example format:\n[\n  {"entity": "Company1", "result_field1": "val1", "result_field2": "val2"},\n  {"entity": "Company2", "result_field1": "val3", "result_field2": "val4"}\n]\n` +
          `Do not include markdown code block backticks (\`\`\`) or introductory conversational text.`;
      }
    }
      
    logDebug("Sending batched query to UltraSearch...");
    
    for (let row of batch) {
      try {
        await adapter.setRowHighlight(row.index, true);
      } catch (highlightErr) {
        logDebug(`Row ${row.index + 1}: Batch Styling highlight error (ignored): ${highlightErr}`);
      }
      try {
        await adapter.writeCell(row.index, 1, "#CONNECTING");
      } catch (e) {}
    }
    
    // Start searching animation for all batch items
    batch.forEach(row => {
      startStatusAnimation(row.index, "🔍 Searching");
    });
    
    const url = `${baseUrl}/search?q=${encodeURIComponent(batchedQuery)}&limit=5&ai_mode=${aiSearchMode}`;
    
    let success = false;
    let parsedArray = null;
    let errorMsg = "";
    
    try {
      activeAbortController = new AbortController();
      const resp = await fetch(url, { signal: activeAbortController.signal });
      activeAbortController = null;
      
      // Stop searching animation, start finalizing animation
      batch.forEach(row => {
        stopStatusAnimation(row.index);
        startStatusAnimation(row.index, "✍️ Finalizing");
      });
      
      if (adapter.constructor.name === "GSheetsAdapter") {
        for (let row of batch) {
          await adapter.writeCell(row.index, 1, "✍️ Finalizing...");
        }
      }

      if (resp.ok) {
        const data = await resp.json();
        const results = data.results || [];
        if (results.length > 0 && results[0].rank === 0) {
          const aiResponse = mapSnippetToBatch(results[0].snippet || "", batch);
          parsedArray = extractJson(aiResponse);
          
          if (parsedArray && Array.isArray(parsedArray)) {
            success = true;
          } else {
            errorMsg = "AI returned text overview but it did not parse as a JSON array.";
          }
        } else if (results.length > 0) {
          // Fallback for HTTP search or organic results
          const textResponse = results[0].snippet || results[0].title || "";
          parsedArray = extractJson(textResponse);
          if (parsedArray && Array.isArray(parsedArray)) {
            success = true;
          } else {
            // Attempt to build a mock array mapping to the batch entities
            parsedArray = batch.map((row, idx) => {
              if (idx === 0) {
                return { "entity": row.query, "result": textResponse };
              }
              return { "entity": row.query, "result": "Unavailable" };
            });
            success = true;
            logDebug(`Row ${batch[0].index + 1}: Writing first organic snippet/title as batch fallback.`);
          }
        } else {
          errorMsg = data.error || "No search results returned.";
        }
      } else {
        errorMsg = `HTTP Error ${resp.status}: ${resp.statusText}`;
      }
    } catch (err) {
      activeAbortController = null;
      if (err.name === 'AbortError') {
        errorMsg = "Stopped by user.";
      } else {
        errorMsg = err.message || "Network Error connecting to UltraSearch.";
      }
    }
    
    // Stop all animations for the batch
    batch.forEach(row => {
      stopStatusAnimation(row.index);
    });

    // Write back results
    for (let bIdx = 0; bIdx < batch.length; bIdx++) {
      const row = batch[bIdx];
      const rowItem = document.getElementById(`row-item-${row.index}`);
      
      let rowSuccess = false;
      let rowData = null;
      
      if (success && parsedArray && parsedArray[bIdx]) {
        rowData = parsedArray[bIdx];
        rowSuccess = true;
      }
      
      if (rowSuccess && rowData) {
        rowItem.className = 'stream-item success';
        rowItem.querySelector('.status-badge').textContent = 'SUCCESS';
        
        const cacheKey = "us_cache_" + row.query.toLowerCase().trim();
        localStorage.setItem(cacheKey, JSON.stringify(rowData));
        
        try {
          await adapter.writeCell(row.index, 1, "SUCCESS");
        } catch (e) {}
        
        delete rowData.entity;
        
        if (mappings.length > 0) {
          // Write mapped fields specifically to their target columns
          for (const m of mappings) {
            if (m.colIdx !== -1) {
              let val = undefined;
              const normColName = m.colName.toLowerCase().replace(/[^a-z0-9]/g, '');
              const normFieldName = m.fieldName.toLowerCase().replace(/[^a-z0-9]/g, '');
              
              for (const key of Object.keys(rowData)) {
                const normKey = key.toLowerCase().replace(/[^a-z0-9]/g, '');
                if (normKey === normColName || normKey === normFieldName) {
                  val = rowData[key];
                  break;
                }
              }
              if (val === undefined || val === null) {
                // Fallback: use raw result text if no JSON key matched
                val = rowData.result || rowData.raw_overview || "";
                if (val) {
                  logDebug(`Row ${row.index + 1}: No exact JSON key match for "${m.fieldName}". Using organic snippet as fallback.`);
                }
              }
              let strVal = typeof val === 'object' ? JSON.stringify(val) : val.toString();
              if (strVal.startsWith("=")) {
                strVal = "'" + strVal;
              }
              try {
                await adapter.writeCell(row.index, m.colIdx, strVal);
              } catch (writeErr) {
                logDebug(`Row ${row.index + 1}: Batch Excel write mapped column error: ${writeErr.message || writeErr}`);
              }
            }
          }
        } else {
          const keys = Object.keys(rowData);
          const values = keys.map(k => {
            let val = rowData[k];
            if (val === null || val === undefined) return "";
            let strVal = typeof val === 'object' ? JSON.stringify(val) : val.toString();
            if (strVal.startsWith("=")) {
              strVal = "'" + strVal;
            }
            return strVal;
          });
          
          try {
            await adapter.writeRow(row.index, 2, values);
          } catch (writeErr) {
            logDebug(`Row ${row.index + 1}: Batch Excel writeRow error: ${writeErr}. Trying writeCell fallback...`);
            try {
              await adapter.writeCell(row.index, 2, values.join(" | "));
            } catch (cellErr) {
              logDebug(`Row ${row.index + 1}: Failed writing cells: ${cellErr}`);
            }
          }
        }
      } else {
        rowItem.className = 'stream-item failed';
        rowItem.querySelector('.status-badge').textContent = 'FAILED';
        
        let failReason = errorMsg || `No matching index in response array.`;
        if (failReason.startsWith("=")) {
          failReason = "'" + failReason;
        }
        try {
          await adapter.writeCell(row.index, 1, `FAILED: ${failReason}`);
        } catch (e) {
          logDebug(`Row ${row.index + 1}: Failed writing error: ${e}`);
        }
      }
      
      processedCount++;
      const progress = Math.round((processedCount / rows.length) * 100);
      progressPercentage.textContent = `${progress}%`;
      progressFill.style.width = `${progress}%`;
      
      // Remove Row Highlight
      try {
        await adapter.setRowHighlight(row.index, false);
      } catch (highlightErr) {
        logDebug(`Row ${row.index + 1}: Batch Styling unhighlight error (ignored): ${highlightErr}`);
      }
    }
    
    logDebug(`Batch processed rows: ${batch.map(r => r.index + 1).join(', ')}`);
    await new Promise(resolve => setTimeout(resolve, 1500));
  }
}

// Map raw AI overview snippets to batch entities in case of format mismatches
function mapSnippetToBatch(snippet, batch) {
  if (!snippet) return "[]";
  
  // Try to find a JSON block in the snippet first
  const match = snippet.match(/(\[[\s\S]*\])/);
  if (match) {
    return match[1];
  }
  
  // If raw text overview is returned instead of JSON array, build a mapped mock array
  try {
    const lines = snippet.split("\n").filter(l => l.trim() !== "");
    const mapped = batch.map((row, idx) => {
      const lineContent = lines[idx] || (lines.length > 0 ? lines[0] : snippet);
      return {
        "entity": row.query,
        "result": lineContent.substring(0, 100).trim()
      };
    });
    return JSON.stringify(mapped);
  } catch (e) {
    return "[]";
  }
}

function escapeRegExp(string) {
  return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function getHeadersInQuery() {
  const goalText = goalPromptInput.value;
  const headers = globalHeaders || [];
  const foundIndices = new Set();
  
  headers.forEach((h, idx) => {
    if (h && h.trim() !== '') {
      const cleanHeader = h.trim();
      const cleanVar = cleanHeader.replace(/\s+/g, '');
      
      // Check for /HeaderName or {{HeaderName}}
      const slashRegex = new RegExp(`/${escapeRegExp(cleanVar)}\\b`, 'i');
      const slashRegexSpaces = new RegExp(`/${escapeRegExp(cleanHeader)}\\b`, 'i');
      const braceRegex = new RegExp(`\\{\\{\\s*${escapeRegExp(cleanHeader)}\\s*\\}\\}`, 'i');
      
      if (slashRegex.test(goalText) || slashRegexSpaces.test(goalText) || braceRegex.test(goalText)) {
        foundIndices.add(idx);
      }
    }
  });
  return Array.from(foundIndices);
}

// ==========================================
// Chat History Logic
// ==========================================

function loadHistory() {
  chatHistoryContainer.innerHTML = '';
  
  const placeholderHtml = `
    <div id="chat-welcome-placeholder" class="chat-welcome-placeholder">
      <div class="welcome-divider"></div>
      <h3 class="welcome-title">Extract Web Data</h3>
      <p class="welcome-subtitle">Fill spreadsheet rows using AI</p>
      <div class="welcome-divider"></div>
    </div>
  `;
  
  try {
    const saved = localStorage.getItem("ULTRA_SEARCH_HISTORY");
    const history = saved ? JSON.parse(saved) : [];
    
    if (history.length === 0) {
      chatHistoryContainer.innerHTML = placeholderHtml;
    } else {
      history.forEach(item => {
        const bubble = document.createElement('div');
        bubble.className = 'history-bubble';
        bubble.textContent = item;
        bubble.style.cursor = 'pointer';
        bubble.title = 'Click to use this goal again';
        bubble.addEventListener('click', () => {
          goalPromptInput.value = item;
          autoGrowTextarea();
          const welcome = document.getElementById('chat-welcome-placeholder');
          if (welcome) welcome.style.display = 'none';
        });
        chatHistoryContainer.appendChild(bubble);
      });
      // Scroll to bottom
      chatHistoryContainer.scrollTop = chatHistoryContainer.scrollHeight;
    }
  } catch (e) {
    logDebug("Error loading history: " + e);
  }
}

function autoGrowTextarea() {
  if (!goalPromptInput) return;
  goalPromptInput.style.height = 'auto';
  goalPromptInput.style.height = (goalPromptInput.scrollHeight) + 'px';
}

function saveToHistory(promptText) {
  if (!promptText || promptText.trim() === '') return;
  try {
    const saved = localStorage.getItem("ULTRA_SEARCH_HISTORY");
    let history = saved ? JSON.parse(saved) : [];
    
    // Don't save duplicate consecutive
    if (history[history.length - 1] === promptText) return;
    
    history.push(promptText);
    
    // Keep last 10
    if (history.length > 10) {
      history = history.slice(history.length - 10);
    }
    localStorage.setItem("ULTRA_SEARCH_HISTORY", JSON.stringify(history));
    loadHistory();
  } catch (e) {
    logDebug("Error saving history: " + e);
  }
}

// ==========================================
// SPA Drawer Navigation Helper Functions
// ==========================================

const PREDEFINED_TEMPLATES = [
  {
    title: "Venture Solvency Audit",
    prompt: "You are acting as a forensic financial auditor. Dissect the target entity's capitalization structure. Step 1: Trace discrepancy between public press releases and official SEC filings from late 2024 to Q1 2026. Step 2: Retrieve exact Net Revenue Retention (NRR) and logo churn rates. Step 3: Present the output as a Markdown table detailing the metric name, publicly claimed value, verified SEC value, and source URL citation.",
    desc: "For audit due diligence on target companies in spreadsheets."
  },
  {
    title: "SaaS Multi-Metric Benchmarking",
    prompt: "As a senior venture capitalist analyzing the SaaS market, compare the target firm against its top 3 direct competitors. You must extract and tabulate: YoY revenue growth rate, Net Promoter Score (NPS), Average Contract Value (ACV), and customer acquisition cost (CAC). Exclude any marketing brochures and restrict your search strictly to post-2025 financial filings. Present the output as a Markdown table comparing these metrics.",
    desc: "Detailed competitive SaaS analysis metrics."
  },
  {
    title: "B2B Outreach Social Profiler",
    prompt: "Given the target company name, find the official corporate email domain, corporate phone number, LinkedIn company page URL, and crunchbase URL. Step 1: Search for these coordinates across their corporate website and official directories. Step 2: Cross-reference with Crunchbase. Step 3: Output as a raw JSON object containing these keys: corporate_email, telephone, linkedin_url, and crunchbase_url. Do not include introductory text.",
    desc: "Extract corporate social handles and contact profiles."
  },
  {
    title: "Product Spec Sheet Parser",
    prompt: "Extract product specifications, exact retail price, manufacturer name, weight, and average customer rating. Step 1: Scan official retail listings and distributor spec sheets. Step 2: Exclude consumer blogs; rely solely on authorized manufacturer datasheets. Step 3: Present the findings as a clean Markdown list detailing each spec. Answer in a brief, concise format.",
    desc: "Build technical specifications databases from product names."
  },
  {
    title: "E-Commerce Competitor Pricing",
    prompt: "You are an e-commerce category manager conducting pricing audits. Step 1: Identify the exact retail prices of this product on Amazon, Walmart, and Target. Step 2: Exclude any third-party reseller markups; rely only on direct corporate prices. Step 3: Format the output as a Markdown table with columns: Retailer, Current Price, Availability Status, and Source URL.",
    desc: "Scrape pricing indexes for retail products across platforms."
  },
  {
    title: "Lead Intelligence Profiler",
    prompt: "As a growth marketing analyst, build a firmographic profile for the target company. Extract: estimated annual revenue range, total headcount, corporate headquarters city, and year of incorporation. Filter out third-party job boards; rely only on zoominfo.com, linkedin.com, and crunchbase.com. Present the output in a clean bulleted list using dashes.",
    desc: "Gather essential firmographics for sales leads."
  },
  {
    title: "Corporate ESG Compliance Audit",
    prompt: "As a corporate compliance officer, evaluate this enterprise's ESG metrics. Step 1: Extract their reported Scope 1 and Scope 2 carbon emissions from their latest sustainability report. Step 2: Determine if they have a formal Diversity, Equity, and Inclusion (DEI) policy. Step 3: Present output as a JSON object containing keys: scope_1_emissions, scope_2_emissions, and dei_policy_status.",
    desc: "Automated ESG compliance audits from enterprise columns."
  },
  {
    title: "App Store Competitor Performance",
    prompt: "Compare the App Store performance of this mobile app against its primary competitor. Extract: overall rating, total review count, date of last update, and size in megabytes. Exclude personal tech blogs; query only the official Apple App Store and Google Play Store listings. Format the output as a Markdown table contrasting the two apps.",
    desc: "Competitor mobile app store performance monitoring."
  },
  {
    title: "NCAA Broadcast Rights Speculator",
    prompt: "As a collegiate sports historian preparing an article, trace the history of broadcast rights revenues for this university athletic conference. Step 1: Extract the total broadcast rights payout for the fiscal years 2022, 2023, and 2024. Step 2: Restrict calculations strictly to official university athletic disclosures. Step 3: Present the payouts in a chronological Markdown table with columns: Year, Total Payout, and Source Link.",
    desc: "Historical athletic conference broadcast payouts history."
  },
  {
    title: "Gene-Editing Regulatory Tracker",
    prompt: "As a regulatory affairs specialist, trace the quarter-by-quarter FDA regulatory status changes for this gene-editing therapeutic drug. Limit your search strictly to the period between early 2024 and Q1 2026. Ignore all generic news abstracts; consult only official clinicaltrials.gov and fda.gov databases. Present the status history as a chronological bulleted list.",
    desc: "Trace FDA clinical trial regulatory changes over time."
  },
  {
    title: "Lithium Output Regional Audit",
    prompt: "Step 1: Identify the top 3 lithium extraction sites operated by this mining company in South America. Step 2: Extract their reported annual output for the years 2023 and 2024. Step 3: Exclude speculative investment newsletters; rely only on audited corporate reports. Step 4: Rank the sites by percentage growth YoY and present them in a Markdown table.",
    desc: "Dissect regional mining company outputs and YoY growth."
  },
  {
    title: "Real Estate Property Due Diligence",
    prompt: "You are a real estate investment analyst performing due diligence on the target property address. Step 1: Extract total square footage, number of bedrooms, number of bathrooms, and year built. Step 2: Locate the most recent listing price or municipal tax assessment valuation. Step 3: Exclude Zillow estimates; rely only on official county tax assessor databases. Format the output as a raw JSON object.",
    desc: "Property assessments and due diligence from addresses."
  },
  {
    title: "Clinical Trial Eligibility Audit",
    prompt: "As a clinical research coordinator, extract the participant eligibility parameters for this clinical trial NCT ID. Specifically retrieve: primary inclusion criteria, primary exclusion criteria, age limit boundaries, and enrollment target size. Restrict data strictly to official clinicaltrials.gov regulatory registries. Present findings in a bulleted list using dashes.",
    desc: "Extract precise clinical trial eligibility rules."
  },
  {
    title: "GitHub Repository Health Specs",
    prompt: "Dissect the repository health of this open-source project. Step 1: Extract the total star count, current open issue count, open pull request count, and date of the latest commit. Step 2: Exclude community discussion forums; query only api.github.com directly. Step 3: Format the output as a JSON payload detailing these metrics.",
    desc: "Open-source codebase health status scraping."
  },
  {
    title: "Macroeconomic Inflation Divergence",
    prompt: "Compare the mainstream federal consensus on this region's Q3 2025 CPI index against contrarian private economic indices. Step 1: Extract the official government-reported CPI percentage. Step 2: Extract alternative inflation percentages reported by independent analysts. Step 3: Present these perspectives side-by-side in a Markdown table detailing the reporting agency, reported rate, and database source.",
    desc: "Analyze government vs independent inflation metrics."
  },
  {
    title: "Sovereign Debt Solvency Risk",
    prompt: "As a global macro sovereign debt trader, analyze the default risk of this nation. Step 1: Extract the debt-to-GDP ratio, current credit rating, and yield on 10-year sovereign bonds. Step 2: Exclude opinion articles; rely only on the World Bank and IMF database disclosures. Step 3: Synthesize the indicators in a raw markdown table.",
    desc: "Compile country-level macroeconomic debt default risk."
  },
  {
    title: "SGE Parser Authority Benchmark",
    prompt: "You are evaluating the data reliability of medical search engines. Step 1: Retrieve the recommended daily dosage and potential major side effects for this compound. Step 2: Compare the drug monograph in the FDA Orange Book with the consensus on WebMD. Step 3: Highlight discrepancies between the official monograph and popular advice in a Markdown table.",
    desc: "Medical compound dosage safety discrepancy benchmarks."
  },
  {
    title: "Bespoke Physics Engine Tech Specs",
    prompt: "As a lead systems engineer, analyze the technical performance architecture of this game engine. Step 1: Extract the default garbage collection model, maximum thread utilization overhead, and primary rendering API support. Step 2: Limit research to official developer documentation and code repository readme files. Step 3: Output as a bulleted list.",
    desc: "Technical specs for game/physics rendering engines."
  },
  {
    title: "Municipal Water Compliance Spec",
    prompt: "You are a civil environmental inspector checking regional water safety. Step 1: Extract the latest reported parts-per-million (ppm) levels of lead, arsenic, and copper for this municipal water district. Step 2: Exclude consumer advocacy columns; rely strictly on official EPA Safe Drinking Water Information System reports. Step 3: Present in a Markdown table.",
    desc: "Water safety safety compliance audits."
  },
  {
    title: "Space Debris Liability Audit",
    prompt: "As a maritime and aerospace legal compliance auditor, review this satellite mission's orbital profile. Step 1: Retrieve the orbital altitude, inclination angle, and planned retirement de-orbit protocol. Step 2: Cross-reference with the UN Register of Objects Launched into Outer Space. Step 3: Present as a JSON object containing keys: altitude_km, inclination_deg, and deorbit_protocol.",
    desc: "Space debris regulatory registration checks."
  },
  {
    title: "Patent Ownership Discrepancy",
    prompt: "Trace the current ownership chain for this USPTO patent number. Step 1: Extract the original assignee, date of filing, and current active assignee. Step 2: Exclude general news websites; consult only the official USPTO Patent Assignment Database. Step 3: Format the output as a Markdown table showing the transaction history, date, assignor, and assignee.",
    desc: "Patent history and corporate IP audits."
  },
  {
    title: "Semiconductor Foundry Lead Times",
    prompt: "As a global supply chain director, audit the lead times for this semiconductor microarchitecture. Step 1: Extract current estimated fabrication lead times, average yield percentage, and primary foundry locations. Step 2: Rely exclusively on industrial distributor advisories and foundry corporate disclosures. Step 3: Present the metrics in a bulleted list.",
    desc: "Track chip production timelines and yields."
  },
  {
    title: "Vulnerability CVE Exploit Assessment",
    prompt: "You are an enterprise cybersecurity threat analyst. Dissect this CVE ID vulnerability. Step 1: Extract the CVSS base score, affected software versions, and standard remediation patch status. Step 2: Bypass generic news summaries; query only NVD NIST and official vendor security advisories. Step 3: Present the analysis in a structured JSON payload.",
    desc: "Cybersecurity vulnerability risk assessment."
  },
  {
    title: "Sovereign Rare Earth Reserves",
    prompt: "Step 1: Retrieve the total estimated metric tonnage of rare earth elements (REE) reserves in this country. Step 2: Exclude corporate sales pitches; rely strictly on USGS Mineral Commodity Summaries from 2024 through 2026. Step 3: Format the output as a Markdown table detailing the reserve size, major mineral types, and global production rank.",
    desc: "Analyze country-level raw mineral reserves."
  },
  {
    title: "Cryptographic Algorithm Strengths",
    prompt: "As a cryptographer auditing system protocols, analyze this cryptographic algorithm. Step 1: Extract the recommended key sizes, known mathematical vulnerabilities, and typical execution latency overhead. Step 2: Rely strictly on NIST publication guidelines and academic consensus. Step 3: Detail findings in a concise bulleted list.",
    desc: "Compare cryptography algorithm safety constraints."
  },
  {
    title: "Corporate Carbon Offsets Auditing",
    prompt: "You are a carbon credit auditor inspecting corporate offset registries. Step 1: Extract the project name, registry origin, total metric tons of CO2 offset, and verification agency for this offset credit ID. Step 2: Limit research to Gold Standard or Verra registry databases. Step 3: Present the output as a Markdown table.",
    desc: "Certify carbon offsets from registration columns."
  },
  {
    title: "Retail Supply Chain Freight Tariffs",
    prompt: "As a logistics analyst, review current freight tariffs for this international shipping route. Step 1: Extract the average cost per FEU, typical maritime transit times, and primary port congestion indicators. Step 2: Restrict search to Drewry World Container Index data. Step 3: Output findings as a JSON object.",
    desc: "Analyze container logistics prices and times."
  },
  {
    title: "Appellate DNA Processing Error Audit",
    prompt: "You are an appellate defense attorney preparing a writ of habeas corpus. Dissect the forensic DNA protocol used in this specific case ID. Step 1: Extract the amplification cycle count, baseline threshold settings, and reported mixture ratio. Step 2: Restrict research strictly to official court transcripts and lab disclosures. Step 3: Present findings in a Markdown table.",
    desc: "Legal forensics and DNA protocol review."
  },
  {
    title: "Rural Broadband Subsidy Allocation",
    prompt: "Trace the FCC subsidy allocations for this rural broadband provider. Step 1: Extract the total funding received under the Rural Digital Opportunity Fund (RDOF). Step 2: Retrieve the committed broadband deployment speed and total target household counts. Step 3: Format output as a JSON object containing keys: total_funding_usd, target_households, and committed_speed_mbps.",
    desc: "Track FCC rural internet funding allocations."
  },
  {
    title: "Clinical Trial Adverse Events Audit",
    prompt: "As a clinical safety auditor, check the reported adverse events for this pharmaceutical drug. Step 1: Extract the percentage occurrence of primary side effects during Phase III clinical trials. Step 2: Rely strictly on official FDA approval documents and peer-reviewed journals. Step 3: Format the metrics as a Markdown table.",
    desc: "Extract Phase III drug side-effects data."
  }
];

const PREDEFINED_STYLES = [
  {
    title: "Concise (Single sentence)",
    suffix: " Answer in a brief, concise sentence.",
    desc: "Saves grid space by limiting output to one line."
  },
  {
    title: "Bulleted List",
    suffix: " Format the output as a clean bulleted list using dashes.",
    desc: "Provides readable, structured points."
  },
  {
    title: "Strict JSON Object",
    suffix: " Format the output strictly as a JSON object containing the requested data.",
    desc: "Makes downstream parsing or API integrations seamless."
  },
  {
    title: "Raw Numbers Only",
    suffix: " Extract only the raw numerical values, numbers, or currencies. Remove all other text.",
    desc: "Perfect for financial and analytical calculations."
  },
  {
    title: "Executive Summary",
    suffix: " Write a formal 2-sentence executive summary.",
    desc: "Provides a professional overview of the findings."
  }
];

function loadNavHistoryList() {
  const historyItemsList = document.getElementById('history-items-list');
  const navViewsPanel = document.getElementById('nav-views-panel');
  const mainUiPanel = document.getElementById('main-ui-panel');
  if (!historyItemsList) return;
  historyItemsList.innerHTML = '';
  
  try {
    const saved = localStorage.getItem("ULTRA_SEARCH_HISTORY");
    const history = saved ? JSON.parse(saved) : [];
    
    if (history.length === 0) {
      historyItemsList.innerHTML = '<div class="loading-text">No history prompts yet.</div>';
      return;
    }
    
    history.slice().reverse().forEach((promptText) => {
      const item = document.createElement('div');
      item.className = 'history-item';
      
      const textSpan = document.createElement('span');
      textSpan.className = 'history-item-text';
      textSpan.textContent = promptText;
      textSpan.addEventListener('click', () => {
        goalPromptInput.value = promptText;
        autoGrowTextarea();
        navViewsPanel.style.display = 'none';
        
        // Hide backdrop
        const drawerBackdrop = document.getElementById('drawer-backdrop');
        if (drawerBackdrop) {
          drawerBackdrop.classList.remove('active');
        }
        
        const welcome = document.getElementById('chat-welcome-placeholder');
        if (welcome) welcome.style.display = 'none';
      });
      
      const deleteBtn = document.createElement('button');
      deleteBtn.className = 'btn-item-delete';
      deleteBtn.title = 'Delete';
      deleteBtn.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>`;
      
      deleteBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        deleteHistoryItem(promptText);
      });
      
      item.appendChild(textSpan);
      item.appendChild(deleteBtn);
      historyItemsList.appendChild(item);
    });
  } catch (e) {
    logDebug("Error loading history list: " + e);
  }
}

function deleteHistoryItem(promptText) {
  try {
    const saved = localStorage.getItem("ULTRA_SEARCH_HISTORY");
    let history = saved ? JSON.parse(saved) : [];
    history = history.filter(item => item !== promptText);
    localStorage.setItem("ULTRA_SEARCH_HISTORY", JSON.stringify(history));
    loadNavHistoryList();
    loadHistory(); // reload welcome/chat page history too
  } catch (e) {
    logDebug("Error deleting history item: " + e);
  }
}

function loadNavTemplatesList() {
  const templatesItemsList = document.getElementById('templates-items-list');
  const navViewsPanel = document.getElementById('nav-views-panel');
  const mainUiPanel = document.getElementById('main-ui-panel');
  if (!templatesItemsList) return;
  templatesItemsList.innerHTML = '';
  
  PREDEFINED_TEMPLATES.forEach(tpl => {
    const card = document.createElement('div');
    card.className = 'custom-card item-card';
    card.innerHTML = `
      <div class="item-card-title">${tpl.title}</div>
      <p class="item-card-desc">${tpl.desc}</p>
    `;
    card.addEventListener('click', () => {
      goalPromptInput.value = tpl.prompt;
      autoGrowTextarea();
      navViewsPanel.style.display = 'none';
      
      // Hide backdrop
      const drawerBackdrop = document.getElementById('drawer-backdrop');
      if (drawerBackdrop) {
        drawerBackdrop.classList.remove('active');
      }
      
      const welcome = document.getElementById('chat-welcome-placeholder');
      if (welcome) welcome.style.display = 'none';
    });
    templatesItemsList.appendChild(card);
  });
}

function loadNavStylesList() {
  const stylesItemsList = document.getElementById('styles-items-list');
  const navViewsPanel = document.getElementById('nav-views-panel');
  const mainUiPanel = document.getElementById('main-ui-panel');
  if (!stylesItemsList) return;
  stylesItemsList.innerHTML = '';
  
  PREDEFINED_STYLES.forEach(st => {
    const card = document.createElement('div');
    card.className = 'custom-card item-card';
    card.innerHTML = `
      <div class="item-card-title">${st.title}</div>
      <p class="item-card-desc">${st.desc}</p>
    `;
    card.addEventListener('click', () => {
      let currentPrompt = goalPromptInput.value.trim();
      if (currentPrompt !== '') {
        if (!currentPrompt.includes(st.suffix.trim())) {
          goalPromptInput.value = currentPrompt + st.suffix;
        }
      } else {
        goalPromptInput.value = st.suffix.trim();
      }
      autoGrowTextarea();
      navViewsPanel.style.display = 'none';
      
      // Hide backdrop
      const drawerBackdrop = document.getElementById('drawer-backdrop');
      if (drawerBackdrop) {
        drawerBackdrop.classList.remove('active');
      }
      
      const welcome = document.getElementById('chat-welcome-placeholder');
      if (welcome) welcome.style.display = 'none';
    });
    stylesItemsList.appendChild(card);
  });
}

// Hook up support view send button
const btnSendSupport = document.getElementById('btn-send-support');
const supportMessage = document.getElementById('support-message');
if (btnSendSupport && supportMessage) {
  btnSendSupport.addEventListener('click', () => {
    const msg = supportMessage.value.trim();
    if (msg === '') {
      adapter.showNotification("Please enter a message before sending.", "info");
      return;
    }
    logDebug("Sending feedback: " + msg);
    adapter.showNotification("Feedback sent successfully! Thank you.", "success");
    supportMessage.value = '';
    const navViewsPanel = document.getElementById('nav-views-panel');
    const mainUiPanel = document.getElementById('main-ui-panel');
    if (navViewsPanel) {
      navViewsPanel.style.display = 'none';
    }
    const drawerBackdrop = document.getElementById('drawer-backdrop');
    if (drawerBackdrop) {
      drawerBackdrop.classList.remove('active');
    }
  });
}

// ==========================================
// Marketplace Templates Dynamic Loader
// ==========================================

const PREDEFINED_MARKETPLACE_TEMPLATES = [
  // Financial & Valuation
  {
    title: "Forensic Balance Sheet Audit",
    prompt: "You are acting as a forensic financial auditor. Dissect the target entity's capitalization structure. Step 1: Trace discrepancy between public press releases and official SEC filings from late 2024 to Q1 2026. Step 2: Retrieve exact Net Revenue Retention (NRR) and logo churn rates. Step 3: Present the output as a Markdown table detailing the metric name, publicly claimed value, verified SEC value, and source URL citation.",
    desc: "Compare public press statements against regulatory filings."
  },
  {
    title: "Venture Capital SaaS Benchmarking",
    prompt: "As a senior venture capitalist analyzing the SaaS market, compare the target firm against its top 3 direct competitors. You must extract and tabulate: YoY revenue growth rate, Net Promoter Score (NPS), Average Contract Value (ACV), and customer acquisition cost (CAC). Exclude any marketing brochures and restrict your search strictly to post-2025 financial filings. Present the output as a Markdown table comparing these metrics.",
    desc: "VC-grade competitive SaaS comparisons."
  },
  {
    title: "Sovereign Debt Default Risk Profile",
    prompt: "As a global macro sovereign debt trader, analyze the default risk of this nation. Step 1: Extract the debt-to-GDP ratio, current credit rating, and yield on 10-year sovereign bonds. Step 2: Exclude opinion articles; rely only on the World Bank and IMF database disclosures. Step 3: Synthesize the indicators in a raw markdown table.",
    desc: "Macro country debt indicators."
  },
  {
    title: "Global Supply Chain Freight Tariffs",
    prompt: "As a logistics analyst, review current freight tariffs for this international shipping route. Step 1: Extract the average cost per FEU, typical maritime transit times, and primary port congestion indicators. Step 2: Restrict search to Drewry World Container Index data. Step 3: Output findings as a JSON object.",
    desc: "Freight shipping costs and logistics times."
  },
  {
    title: "Semiconductor Foundry Production Speeds",
    prompt: "As a global supply chain director, audit the lead times for this semiconductor microarchitecture. Step 1: Extract current estimated fabrication lead times, average yield percentage, and primary foundry locations. Step 2: Rely exclusively on industrial distributor advisories and foundry corporate disclosures. Step 3: Present the metrics in a bulleted list.",
    desc: "Track chip production timelines and yields."
  },
  // Real Estate & Property
  {
    title: "County Tax Assessor Evaluation",
    prompt: "You are a real estate investment analyst performing due diligence on the target property address. Step 1: Extract total square footage, number of bedrooms, number of bathrooms, and year built. Step 2: Locate the most recent listing price or municipal tax assessment valuation. Step 3: Exclude Zillow estimates; rely only on official county tax assessor databases. Format the output as a raw JSON object.",
    desc: "Property tax assessor verification."
  },
  {
    title: "Commercial Zone Zoning Assessment",
    prompt: "As a commercial real estate zoning officer, analyze this parcel's land use restrictions. Step 1: Retrieve the zoning designation code (e.g., C-2, M-1), maximum permitted floor-area ratio (FAR), and building height limits. Step 2: Rely only on municipal government zoning archives. Step 3: Format the output as a JSON payload.",
    desc: "Check parcel zoning codes and limits."
  },
  {
    title: "Homeowner Association Bylaw Compliance",
    prompt: "Dissect the HOA bylaws for this subdivision name. Step 1: Extract the monthly assessment fee, pet restrictions (weight/count limits), and leasing constraints (short-term rental policies). Step 2: Restrict your search to verified community declaration documents. Step 3: List these rules in a clear Markdown table.",
    desc: "Extract HOA fees and renting limits."
  },
  {
    title: "Environmental Property Assessment",
    prompt: "As a phase 1 environmental site assessor, analyze the soil/water risk history of this property location. Step 1: Search for registered underground storage tanks (USTs) or toxic spills on the EPA Envirofacts database. Step 2: List the date of occurrence, contaminant type, and status of cleanup. Format the output as a Markdown table.",
    desc: "Phase 1 environmental property audits."
  },
  {
    title: "Construction Permit Audit",
    prompt: "Retrieve the active municipal building permits for this address. Step 1: Extract the permit number, issue date, status, contractor name, and estimated valuation. Step 2: Consult municipal permit registries. Step 3: Output as a JSON payload.",
    desc: "Municipal building permit searches."
  },
  // Lead Gen & Firmographics
  {
    title: "Firmographic Lead Profiler",
    prompt: "As a growth marketing analyst, build a firmographic profile for the target company. Extract: estimated annual revenue range, total headcount, corporate headquarters city, and year of incorporation. Filter out third-party job boards; rely only on zoominfo.com, linkedin.com, and crunchbase.com. Present the output in a clean bulleted list using dashes.",
    desc: "Gather essential firmographics for leads."
  },
  {
    title: "B2B Executive Contact Finder",
    prompt: "Given the target company name and executive title (e.g., VP of Sales), search for the executive's full name, corporate email address, and LinkedIn profile URL. Step 1: Search corporate newsrooms and public registries. Step 2: Format the results as a Markdown table.",
    desc: "Find decision-maker names and emails."
  },
  {
    title: "Corporate Funding Timeline Tracker",
    prompt: "Dissect the funding history of this startup. Step 1: Extract all funding rounds, round amount, date, and lead investor names. Step 2: Rely on Crunchbase and official funding press releases. Step 3: Present in chronological order as a Markdown table.",
    desc: "startup venture round timelines."
  },
  {
    title: "Enterprise Technology Stack Audit",
    prompt: "As a B2B sales engineer, audit the software tools used by this company. Step 1: Identify their CRM platform, cloud provider, and marketing automation suite. Step 2: Use job postings, public case studies, and stackshare.io. Step 3: Output as a JSON object.",
    desc: "Identify enterprise CRM/cloud tech stack."
  },
  {
    title: "Corporate Subsidiary Structure",
    prompt: "Analyze the corporate parentage of this entity. Step 1: Identify the ultimate parent company, immediate parent company, and any known active subsidiaries. Step 2: Restrict data strictly to SEC Exhibit 21 disclosures. Step 3: Format the relations in a Markdown list.",
    desc: "Corporate hierarchy and ownership chain."
  },
  // E-Commerce & Retail
  {
    title: "Retail Price Comparison Scraper",
    prompt: "You are an e-commerce category manager conducting pricing audits. Step 1: Identify the exact retail prices of this product on Amazon, Walmart, and Target. Step 2: Exclude any third-party reseller markups; rely only on direct corporate prices. Step 3: Format the output as a Markdown table with columns: Retailer, Current Price, Availability Status, and Source URL.",
    desc: "Amazon vs Walmart price tracking."
  },
  {
    title: "Amazon Product Review Synthesis",
    prompt: "As a product manager, synthesize reviews for this ASIN. Step 1: Identify the top 3 recurring customer complaints and the top 3 praised features. Step 2: Extract the aggregate rating and total rating count. Step 3: Exclude sponsored reviews; list findings in a Markdown table.",
    desc: "Aggregate complaints and features of products."
  },
  {
    title: "E-Commerce Product Spec Parser",
    prompt: "Extract product specifications, exact retail price, manufacturer name, weight, and average customer rating. Step 1: Scan official retail listings and distributor spec sheets. Step 2: Exclude consumer blogs; rely solely on authorized manufacturer datasheets. Step 3: Present the findings as a clean Markdown list detailing each spec. Answer in a brief, concise format.",
    desc: "Convert product names into detailed spec sheets."
  },
  {
    title: "Retail Inventory Stock Tracker",
    prompt: "Check the local store availability for this item UPC code. Step 1: Retrieve stock levels, unit price, and pickup options at the closest stores. Step 2: Rely only on official store inventory systems. Step 3: Output as a JSON array.",
    desc: "Check local retail store availability."
  },
  {
    title: "Brand Counterfeit Monitoring",
    prompt: "You are a brand protection specialist. Step 1: Search for unauthorized listings of this product brand name on eBay and DHgate. Step 2: Retrieve listing title, seller rating, and item price. Step 3: Present potential counterfeits in a Markdown table.",
    desc: "Track unauthorized brand distribution."
  },
  // Academic & Research
  {
    title: "Scientific Consensus Synthesis",
    prompt: "Compare the mainstream consensus on this scientific phenomenon (e.g. mRNA vaccine efficacy) with alternative hypotheses. Step 1: Extract the core claim of the scientific consensus. Step 2: Detail contrarian hypotheses published in peer-reviewed journals. Step 3: Format as a Markdown table detailing the hypothesis, supporting paper, and year.",
    desc: "Compare conflicting scientific theories."
  },
  {
    title: "Academic Paper Citation Audit",
    prompt: "For this research paper DOI, retrieve the citation metrics. Step 1: Extract the total citation count on Google Scholar, publication date, journal impact factor, and primary funder. Step 2: Restrict data to official Crossref records. Step 3: Output as a JSON payload.",
    desc: "Verify academic citations and impact."
  },
  {
    title: "Evolution of Historical Theory",
    prompt: "Trace the evolution of historical interpretations regarding this event (e.g. Song Dynasty paleoclimatology). Step 1: Compare observations written during the period against modern geological studies. Step 2: Detail major disagreements. Step 3: Present as a Markdown table.",
    desc: "Historical consensus evolution analysis."
  },
  {
    title: "Government Funding Grants Check",
    prompt: "Identify federal research grants awarded for this technology name. Step 1: Extract the awarding agency (e.g., NSF, NIH), total grant amount, award date, and principal investigator. Step 2: Query only USAspending.gov and NIH RePORTER. Step 3: Output in a Markdown table.",
    desc: "Scrape federal research grant awards."
  },
  {
    title: "Patent Prior Art Searcher",
    prompt: "As an IP paralegal, conduct a prior art search for this invention description. Step 1: Identify the closest 3 active patents filed before 2024. Step 2: Extract patent number, filing date, and claim summary. Format the output as a Markdown table.",
    desc: "Prior art checks for patent drafts."
  },
  // Medical & FDA Trials
  {
    title: "FDA Clinical Trial Audit",
    prompt: "As a clinical research coordinator, extract the participant eligibility parameters for this clinical trial NCT ID. Specifically retrieve: primary inclusion criteria, primary exclusion criteria, age limit boundaries, and enrollment target size. Restrict data strictly to official clinicaltrials.gov regulatory registries. Present findings in a bulleted list using dashes.",
    desc: "FDA trial inclusion/exclusion criteria."
  },
  {
    title: "Drug Monograph Safety Review",
    prompt: "You are evaluating the data reliability of medical search engines. Step 1: Retrieve the recommended daily dosage and potential major side effects for this chemical compound. Step 2: Compare the drug monograph in the FDA Orange Book with the consensus on WebMD. Step 3: Highlight discrepancies between the official monograph and popular advice in a Markdown table.",
    desc: "Compare drug monographs against public sites."
  },
  {
    title: "Clinical Trial Adverse Events Audit",
    prompt: "As a clinical safety auditor, check the reported adverse events for this pharmaceutical drug. Step 1: Extract the percentage occurrence of primary side effects during Phase III clinical trials. Step 2: Rely strictly on official FDA approval documents and peer-reviewed journals. Step 3: Format the metrics as a Markdown table.",
    desc: "Phase III clinical trial safety analysis."
  },
  {
    title: "Medical Device FDA Recalls",
    prompt: "Verify the safety recall status of this medical device. Step 1: Search the FDA Medical Device Recalls database. Step 2: Retrieve recall class (e.g., Class I, II), recall date, reason for recall, and quantity affected. Step 3: Format output as a JSON object.",
    desc: "FDA safety recall checks for medical hardware."
  },
  {
    title: "Orphan Drug Status Tracker",
    prompt: "Confirm the Orphan Drug designation details for this compound. Step 1: Retrieve date of designation, targeted rare disease, and sponsor name. Step 2: Restrict data strictly to the FDA Orphan Drug Product Designation database. Step 3: Output findings in a Markdown list.",
    desc: "FDA orphan drug status checks."
  },
  // Technical & Cybersecurity
  {
    title: "Vulnerability CVE Exploit Assessment",
    prompt: "You are an enterprise cybersecurity threat analyst. Dissect this CVE ID vulnerability. Step 1: Extract the CVSS base score, affected software versions, and standard remediation patch status. Step 2: Bypass generic news summaries; query only NVD NIST and official vendor security advisories. Step 3: Present the analysis in a structured JSON payload.",
    desc: "CVE vulnerability risk scores and patches."
  },
  {
    title: "API Endpoint Specs Scraper",
    prompt: "Given this public API name, document the primary authentication protocol, default rate limits, and base URL endpoints. Step 1: Scan official developer documentation pages. Step 2: Exclude community discussion boards. Step 3: Present metrics in a structured Markdown table.",
    desc: "Extract API developer technical specifications."
  },
  {
    title: "Server SSL Certificate Audit",
    prompt: "As a network security administrator, audit the SSL certificate status of this hostname URL. Step 1: Retrieve certificate issuer, expiration date, and supported TLS versions. Step 2: Ignore commercial sales pitches. Step 3: Present findings as a JSON payload.",
    desc: "Verify SSL certificate issues and TLS versions."
  },
  {
    title: "Game/Physics Engine Tech Specs",
    prompt: "As a lead systems engineer, analyze the technical performance architecture of this game engine. Step 1: Extract the default garbage collection model, maximum thread utilization overhead, and primary rendering API support. Step 2: Limit research to official developer documentation and code repository readme files. Step 3: Output as a bulleted list.",
    desc: "Detailed technical profiles of game engines."
  },
  {
    title: "Cloud Subnet Architecture Spec",
    prompt: "Identify the default subnet design and IP address limits for this cloud virtual network protocol. Step 1: Extract subnet mask range, reserved IP counts, and security group behaviors. Step 2: Query official cloud documentation. Step 3: Format output as a JSON object.",
    desc: "Cloud networking IP limits and subnets."
  },
  // Legal & Municipal
  {
    title: "Patent Assignment Chain Tracker",
    prompt: "Trace the current ownership chain for this USPTO patent number. Step 1: Extract the original assignee, date of filing, and current active assignee. Step 2: Exclude general news websites; consult only the official USPTO Patent Assignment Database. Step 3: Format the output as a Markdown table showing the transaction history, date, assignor, and assignee.",
    desc: "Audit patent ownership transactions."
  },
  {
    title: "Municipal Water Quality Compliance",
    prompt: "You are a civil environmental inspector checking regional water safety. Step 1: Extract the latest reported parts-per-million (ppm) levels of lead, arsenic, and copper for this municipal water district. Step 2: Exclude consumer advocacy columns; rely strictly on official EPA Safe Drinking Water Information System reports. Step 3: Present in a Markdown table.",
    desc: "Inspect EPA tap water safety reports."
  },
  {
    title: "State Lobbying Expenditure Audit",
    prompt: "Check the registered lobbying expenses for this corporation in this state. Step 1: Extract total annual lobbying expenditure, registered lobbyist names, and key legislative files targeted. Step 2: Restrict data to official Secretary of State lobby disclosures. Step 3: Format as a Markdown table.",
    desc: "Expose corporate lobbying spends."
  },
  {
    title: "Federal OSHA Safety Citations",
    prompt: "Audit the workplace safety records of this business name. Step 1: Search the OSHA Enforcement database for active or historical health and safety citations. Step 2: Extract citation date, fine amount, classification (e.g. Serious, Willful), and status. Step 3: Output in a Markdown table.",
    desc: "OSHA safety citation and fine tracking."
  },
  {
    title: "Corporate Carbon Offsets Auditing",
    prompt: "You are a carbon credit auditor inspecting corporate offset registries. Step 1: Extract the project name, registry origin, total metric tons of CO2 offset, and verification agency for this offset credit ID. Step 2: Limit research to Gold Standard or Verra registry databases. Step 3: Present the output as a Markdown table.",
    desc: "Inspect environmental carbon offset claims."
  }
];

function loadNavMarketplaceList() {
  const marketplaceItemsList = document.getElementById('marketplace-items-list');
  const navViewsPanel = document.getElementById('nav-views-panel');
  const searchInput = document.getElementById('marketplace-search-input');
  if (!marketplaceItemsList) return;

  // Clear search input on each open
  if (searchInput) {
    searchInput.value = '';
  }

  // Render helper function
  const renderList = (templates) => {
    marketplaceItemsList.innerHTML = '';
    templates.forEach(tpl => {
      const card = document.createElement('div');
      card.className = 'custom-card item-card';
      card.innerHTML = `
        <div class="item-card-title">${tpl.title}</div>
        <p class="item-card-desc">${tpl.desc}</p>
        <div class="item-card-prompt-preview">${tpl.prompt}</div>
      `;
      card.addEventListener('click', () => {
        goalPromptInput.value = tpl.prompt;
        autoGrowTextarea();
        
        // Close nav view panel
        if (navViewsPanel) navViewsPanel.style.display = 'none';
        
        // Hide backdrop
        const drawerBackdrop = document.getElementById('drawer-backdrop');
        if (drawerBackdrop) {
          drawerBackdrop.classList.remove('active');
        }
        
        const welcome = document.getElementById('chat-welcome-placeholder');
        if (welcome) welcome.style.display = 'none';
      });
      marketplaceItemsList.appendChild(card);
    });
  };

  // Initial render of all templates
  renderList(PREDEFINED_MARKETPLACE_TEMPLATES);

  // Setup input listener for dynamic search
  if (searchInput && !searchInput.dataset.listenerBound) {
    searchInput.dataset.listenerBound = 'true';
    searchInput.addEventListener('input', () => {
      const q = searchInput.value.trim().toLowerCase();
      if (!q) {
        renderList(PREDEFINED_MARKETPLACE_TEMPLATES);
        return;
      }

      // Filter and prioritize:
      // score 2 = match in title (heading)
      // score 1 = match in prompt or description
      const matched = [];
      PREDEFINED_MARKETPLACE_TEMPLATES.forEach(tpl => {
        const titleMatch = tpl.title.toLowerCase().includes(q);
        const descMatch = tpl.desc.toLowerCase().includes(q);
        const promptMatch = tpl.prompt.toLowerCase().includes(q);

        if (titleMatch) {
          matched.push({ tpl, score: 2 });
        } else if (promptMatch || descMatch) {
          matched.push({ tpl, score: 1 });
        }
      });

      // Sort: score 2 items first, then score 1 items
      matched.sort((a, b) => b.score - a.score);

      // Render the matching and sorted list
      renderList(matched.map(item => item.tpl));
    });
  }
}
