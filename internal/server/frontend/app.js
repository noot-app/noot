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

    // Quantity/unit display
    let quantityStr = '';
    if (itemData.quantity != null && itemData.unit) {
      quantityStr = `<span class="item-quantity">${num(itemData.quantity)} ${escapeHtml(itemData.unit)}</span>`;
    } else if (itemData.quantity != null) {
      quantityStr = `<span class="item-quantity">${num(itemData.quantity)}</span>`;
    } else if (itemData.unit) {
      quantityStr = `<span class="item-quantity">${escapeHtml(itemData.unit)}</span>`;
    }

    div.innerHTML = `
      <div class="item-name">${escapeHtml(itemData.name || 'Unknown item')}</div>
      ${quantityStr ? `<div class="item-brand">Amount: ${quantityStr}</div>` : ''}
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
    { 
      label: 'Calories', 
      value: num(totals.energy_kcal), 
      unit: 'kcal', 
      percent: Math.min(percentDaily['energy.kcal'] || 0, 100),
      color: 'rgba(255, 255, 255, 0.9)'
    },
    { 
      label: 'Protein', 
      value: num(totals.protein_g), 
      unit: 'g', 
      percent: Math.min(percentDaily['protein.g'] || 0, 100),
      color: 'rgba(255, 255, 255, 0.9)'
    },
    { 
      label: 'Fat', 
      value: num(totals.fat_g), 
      unit: 'g', 
      percent: Math.min(percentDaily['fat.g'] || 0, 100),
      color: 'rgba(255, 255, 255, 0.9)'
    },
    { 
      label: 'Carbs', 
      value: num(totals.carbs_g), 
      unit: 'g', 
      percent: Math.min(percentDaily['carbs.g'] || 0, 100),
      color: 'rgba(255, 255, 255, 0.9)'
    },
    { 
      label: 'Fiber', 
      value: num(totals.fiber_g), 
      unit: 'g', 
      percent: Math.min(percentDaily['fiber.g'] || 0, 100),
      color: 'rgba(255, 255, 255, 0.9)'
    },
    { 
      label: 'Sugar', 
      value: num(totals.sugar_g), 
      unit: 'g', 
      percent: Math.min(percentDaily['sugar.g'] || 0, 100),
      color: 'rgba(255, 255, 255, 0.9)'
    }
  ];
  
  summaryEl.innerHTML = `
    <div class="summary-grid">
      ${summaryData.map((item, index) => `
        <div class="summary-item" data-percent="${item.percent}">
          <div class="pie-chart-container">
            <div class="pie-chart">
              <div class="pie-chart-background"></div>
              <div class="pie-chart-fill" data-percentage="${item.percent}"></div>
              <div class="pie-chart-center">${Math.round(item.percent)}%</div>
            </div>
          </div>
          <div class="summary-label">${item.label}</div>
          <div class="summary-value">${item.value}${item.unit}</div>
          <div class="summary-percentage">${item.percent}% of daily goal</div>
        </div>
      `).join('')}
    </div>
  `;
  
  // Animate pie charts after a short delay to ensure DOM is ready
  setTimeout(() => {
    animatePieCharts();
  }, 100);
}

function animatePieCharts() {
  const pieCharts = document.querySelectorAll('.pie-chart-fill');
  
  pieCharts.forEach((chart, index) => {
    const percentage = parseFloat(chart.dataset.percentage);
    const degrees = (percentage / 100) * 360;
    
    // Start animation after staggered delay
    setTimeout(() => {
      // Use CSS custom property for smooth animation
      chart.style.setProperty('--percentage', '0deg');
      
      // Trigger animation
      requestAnimationFrame(() => {
        chart.style.setProperty('--percentage', `${degrees}deg`);
      });
      
      // Animate the percentage counter in the center
      const centerElement = chart.parentElement.querySelector('.pie-chart-center');
      animateCounter(centerElement, 0, Math.round(percentage), 1500);
      
    }, index * 200); // Stagger each pie chart by 200ms
  });
}

function animateCounter(element, start, end, duration) {
  const startTime = performance.now();
  
  function updateCounter(currentTime) {
    const elapsed = currentTime - startTime;
    const progress = Math.min(elapsed / duration, 1);
    
    // Use easing function for smooth animation
    const easedProgress = easeOutCubic(progress);
    const current = Math.round(start + (end - start) * easedProgress);
    
    element.textContent = `${current}%`;
    
    if (progress < 1) {
      requestAnimationFrame(updateCounter);
    }
  }
  
  requestAnimationFrame(updateCounter);
}

function easeOutCubic(t) {
  return 1 - Math.pow(1 - t, 3);
}

// Enhanced result rendering with staggered animations
function renderResult(data) {
  resultEl.classList.remove('hidden');
  
  // Render transcript
  transcriptEl.textContent = data.transcript || 'No transcript available';
  
  // Render items with animation
  renderItems(data.items || []);
  
  // Render summary with pie charts
  if (data.summary) {
    renderSummary(data.summary);
  }
  
  // Add floating animation to cards
  setTimeout(() => {
    addFloatingAnimation();
  }, 500);
  
  // Smooth scroll to results
  setTimeout(() => {
    resultEl.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }, 200);
}

function addFloatingAnimation() {
  const cards = document.querySelectorAll('.result-card');
  
  cards.forEach((card, index) => {
    // Add subtle floating animation
    card.style.animation = `float 6s ease-in-out infinite`;
    card.style.animationDelay = `${index * 0.5}s`;
  });
}

// Add floating keyframes to CSS via JavaScript (since we can't modify CSS from here)
function addFloatingKeyframes() {
  const style = document.createElement('style');
  style.textContent = `
    @keyframes float {
      0%, 100% { transform: translateY(0px); }
      50% { transform: translateY(-6px); }
    }
    
    @keyframes gentleBob {
      0%, 100% { transform: translateY(0px) rotateZ(0deg); }
      25% { transform: translateY(-2px) rotateZ(0.5deg); }
      75% { transform: translateY(-1px) rotateZ(-0.5deg); }
    }
    
    .summary-item:hover {
      animation: gentleBob 0.6s ease-in-out;
    }
  `;
  document.head.appendChild(style);
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
  addFloatingKeyframes();
  try { 
    await setupStream(); 
  } catch (e) {
    console.warn('Initial media setup failed, will try again on first recording');
  }
})();
