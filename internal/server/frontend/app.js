// DOM Elements
const micBtn = document.getElementById('micBtn');
const statusEl = document.getElementById('status');
const resultEl = document.getElementById('result');
const transcriptEl = document.getElementById('transcript');
const itemsEl = document.getElementById('items');
const summaryEl = document.getElementById('summary');
const themeToggle = document.getElementById('themeToggle');

// Theme Management
function initTheme() {
  const savedTheme = localStorage.getItem('theme');
  const systemTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  const initialTheme = savedTheme || systemTheme;
  
  applyTheme(initialTheme);
  updateThemeIcon(initialTheme);
}

function applyTheme(theme) {
  document.documentElement.setAttribute('data-theme', theme);
  localStorage.setItem('theme', theme);
}

function updateThemeIcon(theme) {
  const icon = themeToggle.querySelector('.theme-icon');
  icon.textContent = theme === 'dark' ? '☀️' : '🌙';
}

function toggleTheme() {
  const currentTheme = document.documentElement.getAttribute('data-theme');
  const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
  applyTheme(newTheme);
  updateThemeIcon(newTheme);
}

// MediaRecorder Setup
let mediaRecorder;
let chunks = [];

async function setupStream() {
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
    const options = { mimeType: 'audio/webm' };
    try {
      mediaRecorder = new MediaRecorder(stream, options);
    } catch (e) {
      mediaRecorder = new MediaRecorder(stream);
    }

    mediaRecorder.ondataavailable = (e) => {
      if (e.data && e.data.size > 0) chunks.push(e.data);
    };

    mediaRecorder.onstop = async () => {
      setStatus('🔄 Processing your meal...', 'processing');
      const blob = new Blob(chunks, { type: mediaRecorder.mimeType || 'audio/webm' });
      chunks = [];
      
      try {
        const fd = new FormData();
        fd.append('audio', blob, 'audio.webm');
        const resp = await fetch('/api/ingest', { method: 'POST', body: fd });
        
        if (!resp.ok) {
          const errorData = await resp.json().catch(() => ({ error: 'Unknown error' }));
          const errorMsg = errorData.error || `Server error (${resp.status})`;
          throw new Error(errorMsg);
        }
        
        const data = await resp.json();
        renderResult(data);
        setStatus('✅ Done! Here\'s your nutrition breakdown', 'success');
        setTimeout(() => setStatus(''), 3000);
      } catch (err) {
        setStatus('❌ ' + err.message, 'error');
        console.error('Request failed:', err);
      }
    };
  } catch (err) {
    setStatus('❌ Microphone access denied. Please allow microphone access and refresh.', 'error');
    console.error('Media access error:', err);
  }
}

// UI Helper Functions
function setStatus(msg, type = '') {
  statusEl.textContent = msg;
  statusEl.className = `status ${type}`;
}

function renderResult(data) {
  resultEl.classList.remove('hidden');
  
  // Render transcript
  transcriptEl.textContent = data.transcript || 'No transcript available';
  
  // Render items
  renderItems(data.items || []);
  
  // Render summary
  if (data.summary) {
    renderSummary(data.summary);
  }
  
  // Smooth scroll to results
  resultEl.scrollIntoView({ behavior: 'smooth', block: 'start' });
}

function renderItems(items) {
  itemsEl.innerHTML = '';
  
  items.forEach((item, index) => {
    const div = document.createElement('div');
    div.className = 'item';
    
    const itemData = item.item || {};
    const fdcData = item.fdc || {};
    const nutrients = item.nutrients || {};
    
    div.innerHTML = `
      <div class="item-name">${escapeHtml(itemData.name || 'Unknown item')}</div>
      ${itemData.brand ? `<div class="item-brand">Brand: ${escapeHtml(itemData.brand)}</div>` : ''}
      <div class="item-match">Match: ${escapeHtml(fdcData.description || 'No match found')}</div>
      ${Object.keys(nutrients).length > 0 ? `
        <div class="item-nutrients">
          Cal: ${num(nutrients.energy_kcal)} | 
          P: ${num(nutrients.protein_g)}g | 
          F: ${num(nutrients.fat_g)}g | 
          C: ${num(nutrients.carbs_g)}g | 
          Fiber: ${num(nutrients.fiber_g)}g | 
          Sugar: ${num(nutrients.sugar_g)}g
        </div>
      ` : ''}
      ${item.note ? `<div class="item-note">${escapeHtml(item.note)}</div>` : ''}
    `;
    
    itemsEl.appendChild(div);
  });
}

function renderSummary(summary) {
  const totals = summary.totals || {};
  const percentDaily = summary.percent_of_daily || {};
  
  const summaryData = [
    { label: 'Calories', value: num(totals.energy_kcal), unit: 'kcal', percent: percentDaily['energy.kcal'] },
    { label: 'Protein', value: num(totals.protein_g), unit: 'g', percent: percentDaily['protein.g'] },
    { label: 'Fat', value: num(totals.fat_g), unit: 'g', percent: percentDaily['fat.g'] },
    { label: 'Carbs', value: num(totals.carbs_g), unit: 'g', percent: percentDaily['carbs.g'] },
    { label: 'Fiber', value: num(totals.fiber_g), unit: 'g', percent: percentDaily['fiber.g'] },
    { label: 'Sugar', value: num(totals.sugar_g), unit: 'g', percent: percentDaily['sugar.g'] }
  ];
  
  summaryEl.innerHTML = `
    <div class="summary-grid">
      ${summaryData.map(item => `
        <div class="summary-item">
          <div class="summary-label">${item.label}</div>
          <div class="summary-value">${item.value}${item.unit}</div>
          <div class="summary-percentage">${item.percent || 0}% daily</div>
        </div>
      `).join('')}
    </div>
  `;
}

// Recording Functions
async function startRecording(ev) {
  ev.preventDefault();
  
  if (!mediaRecorder) {
    try { 
      await setupStream(); 
    } catch (e) { 
      setStatus('❌ Microphone setup failed', 'error');
      return; 
    }
  }
  
  if (mediaRecorder.state === 'recording') return;
  
  chunks = [];
  mediaRecorder.start();
  micBtn.classList.add('recording');
  setStatus('🎤 Recording... Release to stop', 'processing');
}

function stopRecording(ev) {
  ev.preventDefault();
  
  if (!mediaRecorder || mediaRecorder.state !== 'recording') return;
  
  mediaRecorder.stop();
  micBtn.classList.remove('recording');
  setStatus('⏳ Processing...', 'processing');
}

// Utility Functions
function num(v) {
  const n = Number(v ?? 0);
  if (!isFinite(n)) return '0';
  return n.toFixed(n < 10 ? 1 : 0);
}

function escapeHtml(s) {
  return String(s).replace(/[&<>"']/g, c => ({
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;'
  }[c]));
}

// Event Listeners
themeToggle.addEventListener('click', toggleTheme);

// Mouse events for mic button
micBtn.addEventListener('mousedown', startRecording);
micBtn.addEventListener('mouseup', stopRecording);
micBtn.addEventListener('mouseleave', stopRecording);

// Touch events for mobile
micBtn.addEventListener('touchstart', (e) => {
  e.preventDefault();
  startRecording(e);
});
micBtn.addEventListener('touchend', (e) => {
  e.preventDefault();
  stopRecording(e);
});

// Initialize
(async () => {
  initTheme();
  try { 
    await setupStream(); 
  } catch (e) {
    console.warn('Initial media setup failed, will try again on first recording');
  }
})();
