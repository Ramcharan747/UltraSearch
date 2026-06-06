// Google Apps Script Backend for UltraSearch

// Add Custom Menu on Spreadsheet Load
function onOpen() {
  const ui = SpreadsheetApp.getUi();
  ui.createMenu('UltraSearch')
    .addItem('Open Research Sidebar', 'showSidebar')
    .addItem('Setup Live Auto-Enrichment Trigger', 'setupInstallableTrigger')
    .addToUi();
}

// Show Sidebar Panel
function showSidebar() {
  const html = HtmlService.createTemplateFromFile('sidebar')
    .evaluate()
    .setTitle('UltraSearch Research Assistant')
    .setWidth(300);
  SpreadsheetApp.getUi().showSidebar(html);
}

// Helper to include CSS/JS contents inside template
function include(filename) {
  return HtmlService.createHtmlOutputFromFile(filename).getContent();
}

// Get selected range cell values (Column A acts as query, other columns as values)
function getSelectedRangeData() {
  const sheet = SpreadsheetApp.getActiveSheet();
  const range = sheet.getActiveRange();
  
  if (!range) return [];
  
  const startRow = range.getRow(); 
  const numRows = range.getNumRows();
  
  const values = range.getValues();
  const result = [];
  
  for (let i = 0; i < numRows; i++) {
    const absoluteRowIndex = startRow + i - 1; 
    const queryVal = values[i][0];
    
    result.push({
      index: absoluteRowIndex,
      query: queryVal ? queryVal.toString().trim() : "",
      colValues: values[i].map(v => v ? v.toString() : "")
    });
  }
  
  return result;
}

// Get Row 1 Headers
function getHeadersVal() {
  const sheet = SpreadsheetApp.getActiveSheet();
  const range = sheet.getRange(1, 1, 1, sheet.getLastColumn() || 26);
  const values = range.getValues();
  return values[0].map(v => v ? v.toString().trim() : "");
}

// Read single cell value
function readCellVal(rowIndex, colIndex) {
  const sheet = SpreadsheetApp.getActiveSheet();
  const cell = sheet.getRange(rowIndex + 1, colIndex + 1);
  const val = cell.getValue();
  return val ? val.toString().trim() : "";
}

// Get all rows for a column
function getAllRowsVal(colIndex) {
  const sheet = SpreadsheetApp.getActiveSheet();
  const lastRow = Math.min(sheet.getLastRow(), 100);
  if (lastRow <= 1) return [];
  const range = sheet.getRange(2, colIndex + 1, lastRow - 1, 1);
  const values = range.getValues();
  return values.map((r, i) => ({
    index: i + 1, // row index is 1-based in sheets (excluding header row 1, so row 2 is index 1)
    value: r[0] ? r[0].toString().trim() : ""
  }));
}

// Write single cell value
function writeCellVal(rowIndex, colIndex, value) {
  const sheet = SpreadsheetApp.getActiveSheet();
  const cell = sheet.getRange(rowIndex + 1, colIndex + 1);
  cell.setValue(value);
}

// Write row values starting from a column index
function writeRowVals(rowIndex, startColIndex, colValues) {
  const sheet = SpreadsheetApp.getActiveSheet();
  const range = sheet.getRange(rowIndex + 1, startColIndex + 1, 1, colValues.length);
  range.setValues([colValues]);
}

// Display Google Sheets toast
function showToast(message, type) {
  const title = type ? `UltraSearch: ${type.toUpperCase()}` : "UltraSearch Status";
  SpreadsheetApp.getActiveSpreadsheet().toast(message, title, 5);
}

// Hashing Helper for Cache Keys
function MD5(str) {
  const digest = Utilities.computeDigest(Utilities.DigestAlgorithm.MD5, str);
  let hash = "";
  for (let i = 0; i < digest.length; i++) {
    let byteVal = digest[i];
    if (byteVal < 0) byteVal += 256;
    let byteString = byteVal.toString(16);
    if (byteString.length == 1) byteString = "0" + byteString;
    hash += byteString;
  }
  return hash;
}

/**
 * Executes a custom Google Sheets formula to fetch UltraSearch overview.
 * Example: =ULTRA_SEARCH(A2, "http://20.41.112.252:8082")
 * 
 * @param {string} query The search target query
 * @param {string} apiUrl The endpoint URL of the running UltraSearch server (Defaults to VM IP)
 * @return {string} The AI generated overview or parsed snippet
 * @customfunction
 */
function ULTRA_SEARCH(query, apiUrl) {
  if (!query || query.toString().trim() === "") {
    return "Error: Empty query parameter";
  }
  
  // Cache check to avoid duplicate calls during sheet recalculations
  const cache = CacheService.getDocumentCache();
  const cacheKey = "us_" + MD5(query.toString().toLowerCase().trim());
  const cachedVal = cache.get(cacheKey);
  
  if (cachedVal) {
    return cachedVal;
  }
  
  const url = apiUrl || "http://20.41.112.252:8082";
  const requestUrl = `${url}/search?q=${encodeURIComponent(query)}&limit=5&ai_mode=only`;
  
  try {
    const response = UrlFetchApp.fetch(requestUrl, {
      method: 'get',
      muteHttpExceptions: true,
      timeoutInSeconds: 30
    });
    
    const statusCode = response.getResponseCode();
    if (statusCode !== 200) {
      return `HTTP Error ${statusCode}: ${response.getContentText()}`;
    }
    
    const data = JSON.parse(response.getContentText());
    const results = data.results || [];
    if (results.length > 0 && results[0].rank === 0) {
      const snippet = results[0].snippet || "";
      // Cache result for 6 hours (21600 seconds)
      cache.put(cacheKey, snippet, 21600);
      return snippet;
    }
    
    return data.error || "No AI Overview returned.";
  } catch (err) {
    return `Connection Error: ${err.message}`;
  }
}

// Set up the edit trigger programmatically
function setupInstallableTrigger() {
  const triggers = ScriptApp.getProjectTriggers();
  for (let i = 0; i < triggers.length; i++) {
    if (triggers[i].getHandlerFunction() === 'onEditInstallable') {
      ScriptApp.deleteTrigger(triggers[i]);
    }
  }
  
  ScriptApp.newTrigger('onEditInstallable')
    .forSpreadsheet(SpreadsheetApp.getActiveSpreadsheet())
    .onEdit()
    .create();
    
  SpreadsheetApp.getActiveSpreadsheet().toast("Installable edit trigger set up successfully!", "UltraSearch Setup");
}

// Handle edit events asynchronously (requires installable trigger credentials)
function onEditInstallable(e) {
  const range = e.range;
  const sheet = range.getSheet();
  
  // Filter: column A, row > 1, single cell edits
  if (range.getColumn() === 1 && range.getRow() > 1 && range.getNumRows() === 1) {
    const query = range.getValue().toString().trim();
    if (query === "") return;
    
    const row = range.getRow();
    
    // Store original row cell styles before highlighting
    const formatRange = sheet.getRange(row, 1, 1, 6);
    const originalBackgrounds = formatRange.getBackgrounds();
    const originalFontColors = formatRange.getFontColors();
    
    // Highlight Row (Columns A-F)
    setRowHighlightVal(row - 1, true);
    
    // Step 1: Stream Connecting State
    sheet.getRange(row, 2).setValue("#CONNECTING");
    SpreadsheetApp.flush(); // Force cell update immediately
    
    const props = PropertiesService.getDocumentProperties();
    const apiUrl = props.getProperty("ULTRA_SEARCH_API_URL") || "http://20.41.112.252:8082";
    
    // Step 2: Stream Searching State
    sheet.getRange(row, 2).setValue("🔍 Searching...");
    SpreadsheetApp.flush();
    
    try {
      const overview = ULTRA_SEARCH(query, apiUrl);
      
      // Step 3: Stream Summarizing State
      sheet.getRange(row, 2).setValue("✍️ Summarizing...");
      SpreadsheetApp.flush();
      
      if (overview.startsWith("Error:") || overview.startsWith("Connection Error:") || overview.startsWith("HTTP Error")) {
        sheet.getRange(row, 2).setValue("FAILED");
        sheet.getRange(row, 3).setValue(overview);
      } else {
        sheet.getRange(row, 2).setValue("SUCCESS");
        
        // Parse JSON if possible
        let parsed = null;
        const match = overview.match(/(\{[\s\S]*\}|\[[\s\S]*\])/);
        if (match) {
          try {
            parsed = JSON.parse(match[1]);
          } catch (err) {
            try {
              let cleaned = match[1].replace(/,\s*([\]}])/g, '$1');
              parsed = JSON.parse(cleaned);
            } catch (e2) {}
          }
        }
        
        if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
          const keys = Object.keys(parsed);
          const values = keys.map(k => {
            const val = parsed[k];
            return typeof val === 'object' ? JSON.stringify(val) : val.toString();
          });
          sheet.getRange(row, 3, 1, values.length).setValues([values]);
        } else {
          sheet.getRange(row, 3).setValue(overview);
        }
      }
    } catch (err) {
      sheet.getRange(row, 2).setValue("FAILED");
      sheet.getRange(row, 3).setValue(err.message);
    }
    
    // Restore original row formatting instead of clearing to default
    formatRange.setBackgrounds(originalBackgrounds);
    formatRange.setFontColors(originalFontColors);
    SpreadsheetApp.flush();
  }
}

// Get Row Styles (Backgrounds and Font Colors)
function getRowStyles(rowIndex) {
  const sheet = SpreadsheetApp.getActiveSheet();
  const range = sheet.getRange(rowIndex + 1, 1, 1, 6);
  return {
    backgrounds: range.getBackgrounds(),
    fontColors: range.getFontColors()
  };
}

// Restore Row Styles
function restoreRowStyles(rowIndex, styles) {
  const sheet = SpreadsheetApp.getActiveSheet();
  const range = sheet.getRange(rowIndex + 1, 1, 1, 6);
  range.setBackgrounds(styles.backgrounds);
  range.setFontColors(styles.fontColors);
  SpreadsheetApp.flush();
}

// Set Row Highlight Value
function setRowHighlightVal(rowIndex, isHighlighted) {
  const sheet = SpreadsheetApp.getActiveSheet();
  const range = sheet.getRange(rowIndex + 1, 1, 1, 6); // Highlight columns A-F
  if (isHighlighted) {
    range.setBackground("#1e1b4b"); // Soft dark indigo
    range.setFontColor("#e0e7ff");
  } else {
    range.setBackground(null); // Reset
    range.setFontColor(null); // Reset
  }
  SpreadsheetApp.flush();
}

// Allow configuring API URL in document properties
function setApiUrl(url) {
  PropertiesService.getDocumentProperties().setProperty("ULTRA_SEARCH_API_URL", url);
  return "API URL saved successfully.";
}
