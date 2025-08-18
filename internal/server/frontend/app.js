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
    const nutrients = itemData.nutrients || {};
    
    // Calculate serving amount
    let servingAmount = getEstimatedServingSize(itemData.name, itemData.quantity, itemData.unit);
    
    div.innerHTML = `
      <div class="item-name">${escapeHtml(itemData.name || 'Unknown item')}</div>
      <div class="item-serving">
        <strong>Estimated serving:</strong> <span class="serving-amount">${servingAmount}g</span>
        ${itemData.quantity != null && itemData.unit ? 
          ` (from ${num(itemData.quantity)} ${escapeHtml(itemData.unit)})` : 
          itemData.quantity != null ? 
            ` (${num(itemData.quantity)} servings)` : ''
        }
      </div>
      ${itemData.brand ? `<div class="item-brand">Brand: ${escapeHtml(itemData.brand)}</div>` : ''}
      ${Object.keys(nutrients).length > 0 ? `
        <div class="item-match">Match: Nutrition data provided by AI</div>
        <div class="item-nutrients">
          <div class="nutrient-title">Per serving:</div>
          Cal: ${num(nutrients.calories)} | 
          P: ${num(nutrients.protein_g)}g | 
          F: ${num(nutrients.total_fat_g)}g | 
          C: ${num(nutrients.total_carbs_g)}g | 
          Fiber: ${num(nutrients.dietary_fiber_g)}g | 
          Sugar: ${num(nutrients.total_sugars_g)}g
        </div>
      ` : `
        <div class="item-match">Match: No nutrition data available</div>
      `}
      ${item.note ? `<div class="item-note">${escapeHtml(item.note)}</div>` : ''}
    `;
    
    itemsEl.appendChild(div);
  });
}

// Estimate serving sizes based on common food items
function getEstimatedServingSize(foodName, quantity, unit) {
  const name = (foodName || '').toLowerCase();
  
  // If we have specific measurements, try to convert
  if (quantity && unit) {
    const unitLower = unit.toLowerCase();
    
    // Weight conversions
    if (unitLower.includes('g') || unitLower.includes('gram')) {
      return Math.round(quantity);
    }
    if (unitLower.includes('kg') || unitLower.includes('kilogram')) {
      return Math.round(quantity * 1000);
    }
    if (unitLower.includes('oz') || unitLower.includes('ounce')) {
      return Math.round(quantity * 28.35);
    }
    if (unitLower.includes('lb') || unitLower.includes('pound')) {
      return Math.round(quantity * 453.592);
    }
    
    // Volume to weight estimates (very approximate)
    if (unitLower.includes('cup')) {
      if (name.includes('milk') || name.includes('yogurt')) return Math.round(quantity * 240);
      if (name.includes('seed') || name.includes('nut')) return Math.round(quantity * 120);
      if (name.includes('berry') || name.includes('fruit')) return Math.round(quantity * 150);
      return Math.round(quantity * 100); // Generic cup
    }
    if (unitLower.includes('tbsp') || unitLower.includes('tablespoon')) {
      if (name.includes('seed') || name.includes('nut')) return Math.round(quantity * 12);
      return Math.round(quantity * 15);
    }
    if (unitLower.includes('tsp') || unitLower.includes('teaspoon')) {
      return Math.round(quantity * 5);
    }
  }
  
  // Default serving sizes for common foods (in grams)
  const servingSizes = {
    'chia seed': 15,
    'seed': 15,
    'nut': 30,
    'almond': 30,
    'walnut': 30,
    'yogurt': 170,
    'milk': 240,
    'latte': 240,
    'coffee': 240,
    'espresso': 30,
    'berry': 100,
    'blueberry': 100,
    'strawberry': 150,
    'banana': 120,
    'apple': 180,
    'orange': 150,
    'cheese': 30,
    'bread': 30,
    'egg': 50,
    'oat': 40,
    'cereal': 40,
    'protein': 30
  };
  
  // Find matching food type
  for (const [food, size] of Object.entries(servingSizes)) {
    if (name.includes(food)) {
      return quantity ? Math.round(size * quantity) : size;
    }
  }
  
  // Default serving size
  return quantity ? Math.round(100 * quantity) : 100;
}

function renderSummary(summary) {
  const totals = summary.totals || {};
  const percentDaily = summary.percent_of_daily || {};
  
  const summaryData = [
    { 
      label: 'Calories', 
      value: num(totals.calories), 
      unit: 'kcal', 
      percent: Math.min(percentDaily['calories'] || 0, 100),
      color: 'rgba(255, 255, 255, 0.9)'
    },
    { 
      label: 'Protein', 
      value: num(totals.protein), 
      unit: 'g', 
      percent: Math.min(percentDaily['protein'] || 0, 100),
      color: 'rgba(255, 255, 255, 0.9)'
    },
    { 
      label: 'Fat', 
      value: num(totals.total_fat_g), 
      unit: 'g', 
      percent: Math.min(percentDaily['total_fat'] || 0, 100),
      color: 'rgba(255, 255, 255, 0.9)'
    },
    { 
      label: 'Carbs', 
      value: num(totals.total_carbs_g), 
      unit: 'g', 
      percent: Math.min(percentDaily['total_carbs'] || 0, 100),
      color: 'rgba(255, 255, 255, 0.9)'
    },
    { 
      label: 'Fiber', 
      value: num(totals.dietary_fiber_g), 
      unit: 'g', 
      percent: Math.min(percentDaily['dietary_fiber'] || 0, 100),
      color: 'rgba(255, 255, 255, 0.9)'
    },
    { 
      label: 'Sugar', 
      value: num(totals.total_sugars_g), 
      unit: 'g', 
      percent: Math.min(percentDaily['total_sugars'] || 0, 100),
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
