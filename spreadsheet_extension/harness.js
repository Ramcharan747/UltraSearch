// Spreadsheet Extension Developer Harness Logic
const COLS = ['A', 'B', 'C', 'D', 'E', 'F', 'G'];
const ROW_COUNT = 25;
const gridTable = document.getElementById('excel-grid');
const selectionRangeText = document.getElementById('selection-range-text');
const selectedCountText = document.getElementById('selected-count-text');

// State storing cell values: cellValues[row][col] (0-indexed)
let cellValues = Array.from({ length: ROW_COUNT }, () => Array(COLS.length).fill(''));
let selectedRows = new Set(); // Stores 0-indexed row indices

function initGrid() {
  gridTable.innerHTML = '';
  
  // Create Header Row
  const thead = document.createElement('thead');
  const headerRow = document.createElement('tr');
  
  const corner = document.createElement('th');
  corner.className = 'corner-header';
  headerRow.appendChild(corner);
  
  COLS.forEach(col => {
    const th = document.createElement('th');
    th.textContent = col;
    headerRow.appendChild(th);
  });
  thead.appendChild(headerRow);
  gridTable.appendChild(thead);
  
  // Create Data Rows
  const tbody = document.createElement('tbody');
  for (let r = 0; r < ROW_COUNT; r++) {
    const tr = document.createElement('tr');
    tr.dataset.row = r;
    
    // Row Number Header
    const rowNumTh = document.createElement('th');
    rowNumTh.className = 'row-num';
    rowNumTh.textContent = r + 1;
    rowNumTh.addEventListener('click', (e) => handleRowHeaderClick(r, e));
    tr.appendChild(rowNumTh);
    
    // Columns
    for (let c = 0; c < COLS.length; c++) {
      const td = document.createElement('td');
      td.dataset.row = r;
      td.dataset.col = c;
      
      const input = document.createElement('input');
      input.type = 'text';
      input.value = cellValues[r][c];
      input.addEventListener('change', (e) => {
        cellValues[r][c] = e.target.value;
        // Trigger live change event for Column A
        if (c === 0 && window.harnessApi && typeof window.harnessApi.onChangedHandler === 'function') {
          window.harnessApi.onChangedHandler(r, e.target.value);
        }
      });
      
      td.appendChild(input);
      tr.appendChild(td);
    }
    tbody.appendChild(tr);
  }
  gridTable.appendChild(tbody);
  updateSelectionUI();
}

function handleRowHeaderClick(rowIndex, event) {
  if (event.shiftKey && selectedRows.size > 0) {
    const arr = Array.from(selectedRows);
    const lastSelected = arr[arr.length - 1];
    const start = Math.min(lastSelected, rowIndex);
    const end = Math.max(lastSelected, rowIndex);
    for (let i = start; i <= end; i++) {
      selectedRows.add(i);
    }
  } else if (event.metaKey || event.ctrlKey) {
    if (selectedRows.has(rowIndex)) {
      selectedRows.delete(rowIndex);
    } else {
      selectedRows.add(rowIndex);
    }
  } else {
    selectedRows.clear();
    selectedRows.add(rowIndex);
  }
  updateSelectionUI();
}

function updateSelectionUI() {
  const rows = gridTable.querySelectorAll('tbody tr');
  rows.forEach((tr, r) => {
    if (selectedRows.has(r)) {
      tr.classList.add('selected-row');
    } else {
      tr.classList.remove('selected-row');
    }
  });

  if (selectedRows.size === 0) {
    selectionRangeText.textContent = 'Selection: None';
    selectedCountText.textContent = 'Rows Selected: 0';
  } else {
    const sorted = Array.from(selectedRows).sort((a, b) => a - b);
    const ranges = [];
    let start = sorted[0];
    let prev = sorted[0];
    
    for (let i = 1; i < sorted.length; i++) {
      if (sorted[i] === prev + 1) {
        prev = sorted[i];
      } else {
        ranges.push(start === prev ? `Row ${start+1}` : `Rows ${start+1}-${prev+1}`);
        start = sorted[i];
        prev = sorted[i];
      }
    }
    ranges.push(start === prev ? `Row ${start+1}` : `Rows ${start+1}-${prev+1}`);
    
    selectionRangeText.textContent = `Selection: ${ranges.join(', ')}`;
    selectedCountText.textContent = `Rows Selected: ${selectedRows.size}`;
  }
}

function loadMockQueries() {
  const demoQueries = [
    "Analyze the valuation, latest round, and key investors of Stripe",
    "Analyze the valuation, latest round, and key investors of OpenAI",
    "Analyze the valuation, latest round, and key investors of Scale AI",
    "Analyze the valuation, latest round, and key investors of Databricks",
    "Analyze the valuation, latest round, and key investors of Mistral AI",
    "Analyze the valuation, latest round, and key investors of Anthropic",
    "Analyze the valuation, latest round, and key investors of Groq"
  ];
  
  clearGrid();
  demoQueries.forEach((q, i) => {
    cellValues[i][0] = q; 
    selectedRows.add(i);  
  });
  initGrid();
}

function clearGrid() {
  cellValues = Array.from({ length: ROW_COUNT }, () => Array(COLS.length).fill(''));
  selectedRows.clear();
  initGrid();
}

function exportDataJson() {
  const data = cellValues.map((row, r) => {
    return {
      row: r + 1,
      values: row
    };
  }).filter(r => r.values.some(v => v !== ''));
  
  alert(JSON.stringify(data, null, 2));
}

// Expose API for the Sidebar Iframe
window.isHarness = true;
window.harnessApi = {
  getSelectedRows() {
    const result = [];
    const sorted = Array.from(selectedRows).sort((a, b) => a - b);
    sorted.forEach(r => {
      result.push({
        index: r,
        query: cellValues[r][0],
        colValues: [...cellValues[r]]
      });
    });
    return result;
  },
  
  writeCell(rowIndex, colIndex, value) {
    if (rowIndex >= 0 && rowIndex < ROW_COUNT && colIndex >= 0 && colIndex < COLS.length) {
      cellValues[rowIndex][colIndex] = value;
      const rowTr = gridTable.querySelector(`tbody tr[data-row="${rowIndex}"]`);
      if (rowTr) {
        const cellTd = rowTr.querySelector(`td[data-col="${colIndex}"]`);
        if (cellTd) {
          const input = cellTd.querySelector('input');
          if (input) {
            input.value = value;
          }
        }
      }
    }
  },

  showNotification(msg, type) {
    console.log(`[Harness Notification - ${type}] ${msg}`);
  },

  setRowHighlight(rowIndex, isHighlighted) {
    const rowTr = gridTable.querySelector(`tbody tr[data-row="${rowIndex}"]`);
    if (rowTr) {
      if (isHighlighted) {
        rowTr.classList.add('processing-row');
      } else {
        rowTr.classList.remove('processing-row');
      }
    }
  },

  onChangedHandler: null // Will be assigned by MockAdapter
};

window.addEventListener('load', () => {
  initGrid();
});
