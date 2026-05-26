// Unified Spreadsheet Adapter Layer
// Exposes a common promise-based interface for grid interactions.

class SpreadsheetAdapter {
  getSelectedRows() {
    return Promise.resolve([]);
  }

  writeCell(rowIndex, colIndex, value) {
    return Promise.resolve();
  }

  writeRow(rowIndex, startColIndex, colValues) {
    return Promise.resolve();
  }

  showNotification(msg, type) {
    console.log(`[Spreadsheet] ${type}: ${msg}`);
  }

  // Register cell change listeners for real-time monitoring
  registerChangeListener(handler) {
    return Promise.resolve();
  }

  deregisterChangeListener() {
    return Promise.resolve();
  }

  // Highlight active processing row
  setRowHighlight(rowIndex, isHighlighted) {
    return Promise.resolve();
  }

  // Get first 100 rows of a column for checklist display
  getAllRows(colIndex) {
    return Promise.resolve([]);
  }
}

// 1. Mock Harness Adapter (Local Testing)
class MockAdapter extends SpreadsheetAdapter {
  constructor() {
    super();
    this.api = null;
    if (window.parent && window.parent.harnessApi) {
      this.api = window.parent.harnessApi;
    } else if (window.harnessApi) {
      this.api = window.harnessApi;
    }
  }

  getSelectedRows() {
    if (this.api) {
      return Promise.resolve(this.api.getSelectedRows());
    }
    return Promise.resolve([]);
  }

  getHeaders() {
    if (this.api && this.api.getHeaders) {
      return Promise.resolve(this.api.getHeaders());
    }
    return Promise.resolve(["query", "status", "result"]);
  }

  readCell(rowIndex, colIndex) {
    if (this.api && this.api.readCell) {
      return Promise.resolve(this.api.readCell(rowIndex, colIndex));
    }
    return Promise.resolve("");
  }

  writeCell(rowIndex, colIndex, value) {
    if (this.api) {
      this.api.writeCell(rowIndex, colIndex, value);
    }
    return Promise.resolve();
  }

  writeRow(rowIndex, startColIndex, colValues) {
    if (this.api) {
      colValues.forEach((val, idx) => {
        this.api.writeCell(rowIndex, startColIndex + idx, val);
      });
    }
    return Promise.resolve();
  }

  showNotification(msg, type) {
    if (this.api) {
      this.api.showNotification(msg, type);
    } else {
      super.showNotification(msg, type);
    }
  }

  registerChangeListener(handler) {
    if (this.api) {
      this.api.onChangedHandler = (rowIndex, value) => {
        handler({
          row: rowIndex,
          column: 0,
          value: value
        });
      };
    }
    return Promise.resolve();
  }

  deregisterChangeListener() {
    if (this.api) {
      this.api.onChangedHandler = null;
    }
    return Promise.resolve();
  }

  setRowHighlight(rowIndex, isHighlighted) {
    if (this.api) {
      this.api.setRowHighlight(rowIndex, isHighlighted);
    }
    return Promise.resolve();
  }

  getAllRows(colIndex) {
    if (this.api && this.api.getAllRows) {
      return Promise.resolve(this.api.getAllRows(colIndex));
    }
    return Promise.resolve([
      { index: 1, value: "Founders Fund" },
      { index: 2, value: "White Whale Group" },
      { index: 3, value: "Sipares" },
      { index: 4, value: "HNF Holding" },
      { index: 5, value: "AD Startup" }
    ]);
  }
}

// 2. Google Sheets Adapter (Apps Script)
class GSheetsAdapter extends SpreadsheetAdapter {
  constructor() {
    super();
    this.savedStyles = new Map(); // rowIndex -> {backgrounds, fontColors}
  }

  getSelectedRows() {
    return new Promise((resolve, reject) => {
      if (typeof google === 'undefined' || !google.script || !google.script.run) {
        reject("Google Apps Script context not found.");
        return;
      }
      google.script.run
        .withSuccessHandler(resolve)
        .withFailureHandler(reject)
        .getSelectedRangeData();
    });
  }

  getHeaders() {
    return new Promise((resolve, reject) => {
      if (typeof google === 'undefined' || !google.script || !google.script.run) {
        reject("Google Apps Script context not found.");
        return;
      }
      google.script.run
        .withSuccessHandler(resolve)
        .withFailureHandler(reject)
        .getHeadersVal();
    });
  }

  readCell(rowIndex, colIndex) {
    return new Promise((resolve, reject) => {
      if (typeof google === 'undefined' || !google.script || !google.script.run) {
        reject("Google Apps Script context not found.");
        return;
      }
      google.script.run
        .withSuccessHandler(resolve)
        .withFailureHandler(reject)
        .readCellVal(rowIndex, colIndex);
    });
  }

  getAllRows(colIndex) {
    return new Promise((resolve, reject) => {
      if (typeof google === 'undefined' || !google.script || !google.script.run) {
        reject("Google Apps Script context not found.");
        return;
      }
      google.script.run
        .withSuccessHandler(resolve)
        .withFailureHandler(reject)
        .getAllRowsVal(colIndex);
    });
  }

  writeCell(rowIndex, colIndex, value) {
    return new Promise((resolve, reject) => {
      google.script.run
        .withSuccessHandler(resolve)
        .withFailureHandler(reject)
        .writeCellVal(rowIndex, colIndex, value);
    });
  }

  writeRow(rowIndex, startColIndex, colValues) {
    return new Promise((resolve, reject) => {
      google.script.run
        .withSuccessHandler(resolve)
        .withFailureHandler(reject)
        .writeRowVals(rowIndex, startColIndex, colValues);
    });
  }

  showNotification(msg, type) {
    if (typeof google !== 'undefined' && google.script && google.script.run) {
      google.script.run.showToast(msg, type);
    } else {
      super.showNotification(msg, type);
    }
  }

  setRowHighlight(rowIndex, isHighlighted) {
    return new Promise((resolve, reject) => {
      if (typeof google === 'undefined' || !google.script || !google.script.run) {
        resolve();
        return;
      }
      
      if (isHighlighted) {
        // Step 1: Read and store original styling state
        google.script.run
          .withSuccessHandler(async (styles) => {
            this.savedStyles.set(rowIndex, styles);
            
            // Step 2: Apply the highlight
            google.script.run
              .withSuccessHandler(resolve)
              .withFailureHandler(reject)
              .setRowHighlightVal(rowIndex, true);
          })
          .withFailureHandler(reject)
          .getRowStyles(rowIndex);
      } else {
        const storedStyles = this.savedStyles.get(rowIndex);
        if (storedStyles) {
          // Step 3: Restore original styling state
          google.script.run
            .withSuccessHandler(() => {
              this.savedStyles.delete(rowIndex);
              resolve();
            })
            .withFailureHandler(reject)
            .restoreRowStyles(rowIndex, storedStyles);
        } else {
          // Fallback reset
          google.script.run
            .withSuccessHandler(resolve)
            .withFailureHandler(reject)
            .setRowHighlightVal(rowIndex, false);
        }
      }
    });
  }
}

// 3. Microsoft Excel Adapter (Office.js)
class ExcelAdapter extends SpreadsheetAdapter {
  constructor() {
    super();
    this.excelEventResult = null;
    this.savedStyles = new Map(); // rowIndex -> {fillColor, fontColor}
  }

  getSelectedRows(queryCols) {
    return new Promise((resolve, reject) => {
      if (typeof Excel === 'undefined') {
        reject("Excel JS context not found.");
        return;
      }
      Excel.run(async (context) => {
        const range = context.workbook.getSelectedRange();
        range.load(["rowIndex", "rowCount", "columnIndex"]);
        await context.sync();
        
        const startRow = range.rowIndex;
        const rowCount = range.rowCount;
        const startCol = range.columnIndex;
        
        const sheet = context.workbook.worksheets.getActiveWorksheet();
        
        // Resolve columns to read
        let colsToRead = queryCols;
        if (!colsToRead || colsToRead.length === 0) {
          colsToRead = [startCol];
        }
        
        // Fetch header names
        const headerRange = sheet.getRangeByIndexes(0, 0, 1, 100);
        headerRange.load("values");
        await context.sync();
        const headers = headerRange.values[0].map(v => v === null || v === undefined ? "" : String(v).trim());
        const headerCount = headers.filter(h => h !== "").length;
        
        // Read cells for all columns up to headerCount
        const maxCols = Math.max(headerCount, 1);
        const dataRange = sheet.getRangeByIndexes(startRow, 0, rowCount, maxCols);
        dataRange.load("values");
        await context.sync();
        
        const result = [];
        for (let i = 0; i < rowCount; i++) {
          const colValues = dataRange.values[i].map(v => v === null || v === undefined ? "" : String(v).trim());
          let parts = [];
          for (let colIdx of colsToRead) {
            const val = colValues[colIdx];
            if (val !== undefined && val !== null && String(val).trim() !== "") {
              if (colsToRead.length > 1) {
                const headerName = headers[colIdx] || `Col ${String.fromCharCode(65 + colIdx)}`;
                parts.push(`${headerName}: ${String(val).trim()}`);
              } else {
                parts.push(String(val).trim());
              }
            }
          }
          
          result.push({
            index: startRow + i,
            query: parts.join(" | "),
            colValues: colValues
          });
        }
        resolve(result);
      }).catch(reject);
    });
  }

  getHeaders() {
    return new Promise((resolve, reject) => {
      if (typeof Excel === 'undefined') {
        reject("Excel JS context not found.");
        return;
      }
      Excel.run(async (context) => {
        const sheet = context.workbook.worksheets.getActiveWorksheet();
        const range = sheet.getRangeByIndexes(0, 0, 1, 100);
        range.load("values");
        await context.sync();
        const headers = range.values[0].map(v => String(v || "").trim());
        resolve(headers);
      }).catch(reject);
    });
  }

  readCell(rowIndex, colIndex) {
    return new Promise((resolve, reject) => {
      if (typeof Excel === 'undefined') {
        reject("Excel JS context not found.");
        return;
      }
      Excel.run(async (context) => {
        const sheet = context.workbook.worksheets.getActiveWorksheet();
        const cell = sheet.getRangeByIndexes(rowIndex, colIndex, 1, 1);
        cell.load("values");
        await context.sync();
        const val = cell.values[0][0];
        resolve(val ? String(val).trim() : "");
      }).catch(reject);
    });
  }

  getAllRows(colIndex) {
    return new Promise((resolve, reject) => {
      if (typeof Excel === 'undefined') {
        reject("Excel JS context not found.");
        return;
      }
      Excel.run(async (context) => {
        const sheet = context.workbook.worksheets.getActiveWorksheet();
        const range = sheet.getUsedRange();
        range.load(["rowCount", "rowIndex"]);
        await context.sync();
        
        const rowCount = Math.min(range.rowCount, 100);
        const startRow = range.rowIndex;
        
        const colRange = sheet.getRangeByIndexes(startRow, colIndex, rowCount, 1);
        colRange.load("values");
        await context.sync();
        
        const result = [];
        // Skip header row if it is row 0 and startRow is 0
        const firstDataRow = (startRow === 0) ? 1 : 0;
        
        for (let i = firstDataRow; i < rowCount; i++) {
          const val = colRange.values[i][0];
          result.push({
            index: startRow + i,
            value: val ? String(val).trim() : ""
          });
        }
        resolve(result);
      }).catch(reject);
    });
  }

  writeCell(rowIndex, colIndex, value) {
    return new Promise((resolve, reject) => {
      Excel.run(async (context) => {
        const sheet = context.workbook.worksheets.getActiveWorksheet();
        const cell = sheet.getRangeByIndexes(rowIndex, colIndex, 1, 1);
        cell.values = [[value]];
        await context.sync();
        resolve();
      }).catch(reject);
    });
  }

  writeRow(rowIndex, startColIndex, colValues) {
    return new Promise((resolve, reject) => {
      Excel.run(async (context) => {
        const sheet = context.workbook.worksheets.getActiveWorksheet();
        const range = sheet.getRangeByIndexes(rowIndex, startColIndex, 1, colValues.length);
        range.values = [colValues];
        await context.sync();
        resolve();
      }).catch(reject);
    });
  }

  showNotification(msg, type) {
    super.showNotification(msg, type);
  }

  registerChangeListener(handler) {
    return new Promise((resolve, reject) => {
      Excel.run(async (context) => {
        const sheet = context.workbook.worksheets.getActiveWorksheet();
        
        this.excelEventResult = sheet.onChanged.add(async (eventArgs) => {
          if (eventArgs.changeType === "ValueChange") {
            const range = eventArgs.getRange();
            range.load(["rowIndex", "columnIndex", "values"]);
            await context.sync();
            
            if (range.columnIndex === 0 && range.rowIndex > 0) {
              handler({
                row: range.rowIndex,
                column: range.columnIndex,
                value: range.values[0][0]
              });
            }
          }
        });
        
        await context.sync();
        resolve();
      }).catch(reject);
    });
  }

  deregisterChangeListener() {
    return new Promise((resolve, reject) => {
      if (!this.excelEventResult) {
        resolve();
        return;
      }
      Excel.run(async (context) => {
        this.excelEventResult.remove();
        await context.sync();
        this.excelEventResult = null;
        resolve();
      }).catch(reject);
    });
  }

  setRowHighlight(rowIndex, isHighlighted) {
    return new Promise((resolve, reject) => {
      if (typeof Excel === 'undefined') {
        resolve();
        return;
      }
      Excel.run(async (context) => {
        const sheet = context.workbook.worksheets.getActiveWorksheet();
        const range = sheet.getRangeByIndexes(rowIndex, 0, 1, 6); // columns A-F
        
        if (isHighlighted) {
          // Load and store original cell styling
          range.format.fill.load("color");
          range.format.font.load("color");
          await context.sync();
          
          this.savedStyles.set(rowIndex, {
            fillColor: range.format.fill.color,
            fontColor: range.format.font.color
          });
          
          // Apply highlight
          range.format.fill.color = "#1E1B4B";
          range.format.font.color = "#E0E7FF";
        } else {
          const style = this.savedStyles.get(rowIndex);
          if (style) {
            // Restore original styles
            range.format.fill.color = style.fillColor;
            range.format.font.color = style.fontColor;
            this.savedStyles.delete(rowIndex);
          } else {
            // Fallback clear
            range.format.fill.color = null;
            range.format.font.color = null;
          }
        }
        await context.sync();
        resolve();
      }).catch(reject);
    });
  }
}

// Global Factory Function
function getActiveAdapter() {
  const urlParams = new URLSearchParams(window.location.search);
  if (urlParams.get('harness') === 'true' || (window.parent && window.parent.isHarness)) {
    console.log("[AdapterFactory] Harness detected. Injecting MockAdapter.");
    return new MockAdapter();
  } else if (typeof google !== 'undefined' && google.script && google.script.run) {
    console.log("[AdapterFactory] Google Apps Script detected. Injecting GSheetsAdapter.");
    return new GSheetsAdapter();
  } else if (typeof Office !== 'undefined' && typeof Excel !== 'undefined') {
    console.log("[AdapterFactory] Excel OfficeJS detected. Injecting ExcelAdapter.");
    return new ExcelAdapter();
  } else {
    console.log("[AdapterFactory] No active context. Defaulting to MockAdapter.");
    return new MockAdapter();
  }
}
