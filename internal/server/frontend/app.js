// DOM Elements
const micBtn = document.getElementById('micBtn');
const statusEl = document.getElementById('status');
const resultEl = document.getElementById('result');
const transcriptEl = document.getElementById('transcript-text');
const summaryEl = document.getElementById('summary-content');
const itemsEl = document.getElementById('items-content');
const staticMic = document.getElementById('staticMic');
const siriWaves = document.getElementById('siriWaves');
const themeToggle = document.getElementById('themeToggle');
const devBanner = document.getElementById('dev-banner');
const simulateBtn = document.getElementById('simulate-btn');

// Theme management
function initTheme() {
  // Check for saved theme preference or default to dark mode
  const savedTheme = localStorage.getItem('theme') || 'dark';
  setTheme(savedTheme);
}

function setTheme(theme) {
  document.body.setAttribute('data-theme', theme);
  localStorage.setItem('theme', theme);
}

function toggleTheme() {
  const currentTheme = document.body.getAttribute('data-theme');
  const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
  setTheme(newTheme);
}

// Audio feedback
let audioContext;
let startSound, endSound;

// MediaRecorder Setup
let mediaRecorder;
let chunks = [];

// Development mode detection - check if we're on localhost or have dev indicator
const isDevelopmentMode = window.location.hostname === 'localhost' || 
                         window.location.hostname === '127.0.0.1' ||
                         window.location.port === '8080' ||
                         document.body.hasAttribute('data-dev-mode');

// Initialize audio feedback
async function initAudioFeedback() {
  try {
    audioContext = new (window.AudioContext || window.webkitAudioContext)();
    
    // Create simple beep sounds
    startSound = createBeepSound(800, 0.1); // Higher pitch for start
    endSound = createBeepSound(400, 0.15);  // Lower pitch for end
  } catch (error) {
    console.warn('Audio feedback initialization failed:', error);
  }
}

// Create beep sound
function createBeepSound(frequency, duration) {
  const oscillator = audioContext.createOscillator();
  const gainNode = audioContext.createGain();
  
  oscillator.connect(gainNode);
  gainNode.connect(audioContext.destination);
  
  oscillator.frequency.setValueAtTime(frequency, audioContext.currentTime);
  oscillator.type = 'sine';
  
  gainNode.gain.setValueAtTime(0, audioContext.currentTime);
  gainNode.gain.linearRampToValueAtTime(0.1, audioContext.currentTime + 0.01);
  gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + duration);
  
  return { oscillator, gainNode, duration };
}

// Play sound
function playSound(soundConfig) {
  if (!audioContext || !soundConfig) return;
  
  try {
    const { oscillator, gainNode, duration } = createBeepSound(
      soundConfig === startSound ? 800 : 400,
      soundConfig === startSound ? 0.1 : 0.15
    );
    
    oscillator.start(audioContext.currentTime);
    oscillator.stop(audioContext.currentTime + duration);
  } catch (error) {
    console.warn('Failed to play sound:', error);
  }
}

// Load simulation data from JSON file
async function loadSimulationData() {
  try {
    const response = await fetch('/simulation-data.json');
    if (!response.ok) {
      throw new Error(`Failed to load simulation data: ${response.status}`);
    }
    return await response.json();
  } catch (error) {
    console.error('Error loading simulation data:', error);
    // Fallback to minimal data if JSON fails to load
    return {
      transcript: "For breakfast I had Bob's Red Mill organic rolled oats, Greek yogurt, honey, goji berries, cacao nibs, and Trader Joe's dried blueberries.",
      items: []
    };
  }
}

// Generate simulation response with calculated totals
async function getSimulationData() {
  const data = await loadSimulationData();
  
  if (data.items.length === 0) {
    return data; // Return fallback data as-is
  }

  // Calculate totals from individual items
  const totals = data.items.reduce((sum, item) => {
    const nutrients = item.nutrients;
    Object.keys(nutrients).forEach(key => {
      sum[key] = (sum[key] || 0) + nutrients[key];
    });
    return sum;
  }, {});

  // Convert items to the expected format
  const formattedItems = data.items.map(item => ({
    item: item
  }));

  return {
    transcript: data.transcript,
    items: formattedItems,
    summary: {
      totals: totals,
      percent_of_daily: {
        calories: Math.round((totals.calories / 2000) * 100),
        protein: Math.round((totals.protein_g / 50) * 100),
        total_carbs: Math.round((totals.total_carbs_g / 300) * 100),
        total_fat: Math.round((totals.total_fat_g / 65) * 100),
        dietary_fiber: Math.round((totals.dietary_fiber_g / 25) * 100),
        sodium: Math.round((totals.sodium_mg / 2300) * 100),
        calcium: Math.round((totals.calcium_mg / 1000) * 100),
        iron: Math.round((totals.iron_mg / 18) * 100),
        potassium: Math.round((totals.potassium_mg / 3500) * 100)
      }
    }
  };
}

// Initialize MediaRecorder
async function setupMediaRecorder() {
  // Skip audio setup in development mode only if we're explicitly simulating
  if (isDevelopmentMode) {
    console.log('Running in development mode - audio setup available');
  }
  
  const stream = await navigator.mediaDevices.getUserMedia({ 
    audio: {
      channelCount: 1,
      sampleRate: 44100,
    } 
  });
  
  mediaRecorder = new MediaRecorder(stream, {
    mimeType: 'audio/webm;codecs=opus'
  });

  mediaRecorder.ondataavailable = (event) => {
    if (event.data.size > 0) {
      chunks.push(event.data);
    }
  };

  mediaRecorder.onstop = async () => {
    const blob = new Blob(chunks, { type: 'audio/webm' });
    chunks = [];
    await sendAudioToAPI(blob);
  };
}

// Send audio to API
async function sendAudioToAPI(audioBlob) {
  try {
    setStatus('⏳ Processing...');
    
    const formData = new FormData();
    formData.append('audio', audioBlob);

    const response = await fetch('/api/ingest', {
      method: 'POST',
      body: formData
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    displayResults(data);
  } catch (error) {
    console.error('Error:', error);
    setStatus('❌ Error processing audio. Please try again.');
  }
}

// Simulate API response with predefined data
async function simulateResponse() {
  setStatus('🔬 Simulating...');
  try {
    const data = await getSimulationData();
    setTimeout(() => {
      displayResults(data);
    }, 500); // Simulate processing delay
  } catch (error) {
    console.error('Simulation failed:', error);
    setStatus('❌ Simulation failed. Please try again.');
  }
}

// Create pie chart SVG
function createPieChart(percentage, color = '#3b82f6') {
  const radius = 16;
  const circumference = 2 * Math.PI * radius;
  const progress = Math.min(Math.max(percentage, 0), 100);
  
  return `
    <svg class="pie-chart" viewBox="0 0 40 40">
      <circle class="pie-bg" cx="20" cy="20" r="${radius}"></circle>
      <circle 
        class="pie-fill" 
        cx="20" 
        cy="20" 
        r="${radius}"
        style="--progress: ${progress}; stroke: ${color}; stroke-dasharray: ${progress} ${100 - progress};"
      ></circle>
    </svg>
  `;
}

// Get color based on percentage
function getPercentageColor(percentage) {
  if (percentage < 25) return '#ef4444'; // red
  if (percentage < 50) return '#f59e0b'; // amber
  if (percentage < 75) return '#eab308'; // yellow
  if (percentage < 100) return '#22c55e'; // green
  return '#3b82f6'; // blue (over 100%)
}

// Display results
function displayResults(data) {
  setStatus('');
  
  // Show transcript
  transcriptEl.textContent = data.transcript || 'No transcript available';
  
  // Show summary first (as requested) with enhanced nutrition data
  if (data.summary && data.summary.totals) {
    displaySummary(data.summary);
  }
  
  // Show individual items breakdown
  if (data.items && data.items.length > 0) {
    displayItems(data.items);
  }
  
  // Show results section
  resultEl.classList.remove('hidden');
  
  // Smooth scroll to results
  setTimeout(() => {
    resultEl.scrollIntoView({ behavior: 'smooth' });
  }, 100);
}

// Display comprehensive nutrition summary organized by categories
function displaySummary(summary) {
  // Enable scrolling when content is shown
  document.body.classList.add('content-visible');
  
  const totals = summary.totals || {};
  const percentDaily = summary.percent_of_daily || {};
  
  // Calculate sugar percentages based on WHO/FDA guidelines
  const totalSugarsG = totals.total_sugars_g || 0;
  const addedSugarsG = totals.added_sugars_g || 0;
  const naturalSugarsG = Math.max(0, totalSugarsG - addedSugarsG);
  
  // Added sugar limit: 50g for 2000-calorie diet (10% of calories)
  const addedSugarLimit = 50; 
  const addedSugarPercent = Math.min((addedSugarsG / addedSugarLimit) * 100, 100);
  
  // Macronutrients section
  const macros = [
    { 
      label: 'Calories', 
      value: formatNumber(totals.calories), 
      unit: 'kcal', 
      percent: Math.min(percentDaily['calories'] || 0, 100),
      category: 'macro'
    },
    { 
      label: 'Protein', 
      value: formatNumber(totals.protein_g), 
      unit: 'g', 
      percent: Math.min(percentDaily['protein'] || 0, 100),
      category: 'macro'
    },
    { 
      label: 'Carbs', 
      value: formatNumber(totals.total_carbs_g), 
      unit: 'g', 
      percent: Math.min(percentDaily['total_carbs'] || 0, 100),
      category: 'macro'
    },
    { 
      label: 'Fat', 
      value: formatNumber(totals.total_fat_g), 
      unit: 'g', 
      percent: Math.min(percentDaily['total_fat'] || 0, 100),
      category: 'macro'
    }
  ];
  
  // Sugar breakdown section
  const sugarBreakdown = [
    { 
      label: 'Added Sugar', 
      value: formatNumber(addedSugarsG), 
      unit: 'g', 
      percent: addedSugarPercent,
      limit: addedSugarLimit,
      category: 'sugar',
      warning: addedSugarPercent > 80
    },
    { 
      label: 'Natural Sugar', 
      value: formatNumber(naturalSugarsG), 
      unit: 'g', 
      percent: 0, // No daily limit for natural sugars
      category: 'sugar',
      note: 'from fruits & dairy'
    }
  ].filter(item => parseFloat(item.value) > 0);
  
  // Fiber section
  const fiber = [
    { 
      label: 'Fiber', 
      value: formatNumber(totals.dietary_fiber_g), 
      unit: 'g', 
      percent: Math.min(percentDaily['dietary_fiber'] || 0, 100),
      category: 'fiber'
    }
  ].filter(item => parseFloat(item.value) > 0);
  
  // Vitamins & Minerals section
  const vitaminsAndMinerals = [
    { 
      label: 'Vitamin C', 
      value: formatNumber(totals.vitamin_c_mg || 0), 
      unit: 'mg', 
      percent: Math.min(percentDaily['vitamin_c'] || 0, 100),
      category: 'vitamin'
    },
    { 
      label: 'Vitamin D', 
      value: formatNumber(totals.vitamin_d_mcg || 0), 
      unit: 'μg', 
      percent: Math.min(percentDaily['vitamin_d'] || 0, 100),
      category: 'vitamin'
    },
    { 
      label: 'Calcium', 
      value: formatNumber(totals.calcium_mg || 0), 
      unit: 'mg', 
      percent: Math.min(percentDaily['calcium'] || 0, 100),
      category: 'mineral'
    },
    { 
      label: 'Iron', 
      value: formatNumber(totals.iron_mg || 0), 
      unit: 'mg', 
      percent: Math.min(percentDaily['iron'] || 0, 100),
      category: 'mineral'
    },
    { 
      label: 'Sodium', 
      value: formatNumber(totals.sodium_mg || 0), 
      unit: 'mg', 
      percent: Math.min(percentDaily['sodium'] || 0, 100),
      category: 'mineral'
    },
    { 
      label: 'Potassium', 
      value: formatNumber(totals.potassium_mg || 0), 
      unit: 'mg', 
      percent: Math.min(percentDaily['potassium'] || 0, 100),
      category: 'mineral'
    }
  ].filter(item => parseFloat(item.value) > 0);

  // Create sections HTML
  const createSection = (title, items, emoji) => {
    if (items.length === 0) return '';
    
    return `
      <div class="nutrition-section">
        <h3 class="nutrition-section-title">${emoji} ${title}</h3>
        <div class="nutrition-section-grid">
          ${items.map(item => {
            const color = item.warning ? '#ef4444' : getPercentageColor(item.percent);
            const showPieChart = item.percent > 0 && !item.note; // Don't show pie chart for items with notes (like natural sugar)
            const pieChart = showPieChart ? createPieChart(item.percent, color) : '';
            const dailyText = item.note ? item.note : 
                             item.limit ? `of ${item.limit}g limit` : 
                             `${Math.round(item.percent)}% daily`;
            
            return `
              <div class="summary-item ${item.warning ? 'warning' : ''}">
                ${showPieChart ? `<div style="display: flex; align-items: center; justify-content: center; margin-bottom: 0.5rem;">${pieChart}</div>` : '<div style="height: 0.5rem;"></div>'}
                <div class="nutrition-label">${item.label}</div>
                <div class="nutrition-value">${item.value}${item.unit}</div>
                <div class="nutrition-daily">${dailyText}</div>
              </div>
            `;
          }).join('')}
        </div>
      </div>
    `;
  };

  summaryEl.innerHTML = `
    <div class="nutrition-summary">
      ${createSection('Macronutrients', macros, '⚡')}
      ${createSection('Sugar Breakdown', sugarBreakdown, '🍯')}
      ${createSection('Fiber', fiber, '🌾')}
      ${createSection('Vitamins & Minerals', vitaminsAndMinerals, '💊')}
    </div>
  `;
  
  // Trigger pie chart animations
  setTimeout(() => {
    const pieCharts = summaryEl.querySelectorAll('.pie-fill');
    pieCharts.forEach(chart => {
      chart.style.animationPlayState = 'running';
    });
  }, 200);
}

// Display individual items breakdown
function displayItems(items) {
  if (!items || items.length === 0) {
    itemsEl.innerHTML = '<p style="color: var(--text-secondary); text-align: center; padding: 2rem;">No items available</p>';
    return;
  }

  itemsEl.innerHTML = items.map(itemData => {
    const item = itemData.item;
    const nutrients = item.nutrients;
    
    if (!nutrients) {
      return `
        <div class="item-card">
          <div class="item-header">
            <h3 class="item-name">${item.name}</h3>
            <div class="item-quantity">${formatQuantity(item.quantity, item.unit)}</div>
          </div>
          <div class="item-note">Nutrition data unavailable</div>
        </div>
      `;
    }

    // Key nutrition facts to display for each item
    const keyNutrients = [
      { label: 'Calories', value: nutrients.calories, unit: 'kcal' },
      { label: 'Protein', value: nutrients.protein_g, unit: 'g' },
      { label: 'Carbs', value: nutrients.total_carbs_g, unit: 'g' },
      { label: 'Fat', value: nutrients.total_fat_g, unit: 'g' }
    ];

    // Additional nutrients (only show if significant values)
    const additionalNutrients = [
      { label: 'Fiber', value: nutrients.dietary_fiber_g, unit: 'g', threshold: 0.5 },
      { label: 'Sugar', value: nutrients.total_sugars_g, unit: 'g', threshold: 0.5 },
      { label: 'Vitamin C', value: nutrients.vitamin_c_mg, unit: 'mg', threshold: 1 },
      { label: 'Calcium', value: nutrients.calcium_mg, unit: 'mg', threshold: 5 },
      { label: 'Iron', value: nutrients.iron_mg, unit: 'mg', threshold: 0.1 },
      { label: 'Sodium', value: nutrients.sodium_mg, unit: 'mg', threshold: 1 },
      { label: 'Potassium', value: nutrients.potassium_mg, unit: 'mg', threshold: 5 }
    ].filter(n => n.value >= n.threshold);

    return `
      <div class="item-card">
        <div class="item-header">
          <h3 class="item-name">${item.name}</h3>
          <div class="item-quantity">${formatQuantity(item.quantity, item.unit)}</div>
        </div>
        <div class="item-nutrition">
          <div class="nutrition-main">
            ${keyNutrients.map(n => `
              <div class="nutrition-item">
                <span class="nutrition-label">${n.label}</span>
                <span class="nutrition-value">${formatNumber(n.value)}${n.unit}</span>
              </div>
            `).join('')}
          </div>
          ${additionalNutrients.length > 0 ? `
            <div class="nutrition-additional">
              ${additionalNutrients.map(n => `
                <div class="nutrition-item-small">
                  <span class="nutrition-label-small">${n.label}</span>
                  <span class="nutrition-value-small">${formatNumber(n.value)}${n.unit}</span>
                </div>
              `).join('')}
            </div>
          ` : ''}
        </div>
      </div>
    `;
  }).join('');
}

// Format quantity display
function formatQuantity(quantity, unit) {
  if (!quantity || !unit) return '';
  const num = formatNumber(quantity);
  return `${num} ${unit}${quantity > 1 ? 's' : ''}`;
}

// Set status message
function setStatus(message) {
  statusEl.textContent = message;
}

// Format numbers
function formatNumber(value) {
  const num = Number(value ?? 0);
  if (!isFinite(num)) return '0';
  return num < 10 ? num.toFixed(1) : num.toFixed(0);
}

// Recording Functions
async function startRecording(event) {
  event.preventDefault();
  
  // Regular mode - only try to setup media recorder if not already done
  if (!mediaRecorder) {
    try {
      await setupMediaRecorder();
    } catch (error) {
      console.error('Media setup failed:', error);
      setStatus('❌ Microphone access denied. Please allow microphone access and refresh.');
      return;
    }
  }
  
  if (mediaRecorder.state === 'recording') return;
  
  // Play start sound
  playSound(startSound);
  
  chunks = [];
  mediaRecorder.start();
  micBtn.classList.add('recording');
  setStatus('🎤 Recording... Release to stop');
}

function stopRecording(event) {
  event.preventDefault();
  
  if (!mediaRecorder || mediaRecorder.state !== 'recording') return;
  
  // Play end sound
  playSound(endSound);
  
  mediaRecorder.stop();
  micBtn.classList.remove('recording');
}

// Event Listeners
micBtn.addEventListener('mousedown', startRecording);
micBtn.addEventListener('mouseup', stopRecording);
micBtn.addEventListener('mouseleave', stopRecording);

// Theme toggle
themeToggle.addEventListener('click', toggleTheme);

// Simulate button (development mode only)
if (simulateBtn) {
  simulateBtn.addEventListener('click', simulateResponse);
}

// Touch events for mobile
micBtn.addEventListener('touchstart', (e) => {
  e.preventDefault();
  startRecording(e);
});

micBtn.addEventListener('touchend', (e) => {
  e.preventDefault();
  stopRecording(e);
});

micBtn.addEventListener('touchcancel', (e) => {
  e.preventDefault();
  stopRecording(e);
});

// Initialize
(async () => {
  try {
    // Initialize theme first
    initTheme();
    
    // Show development banner if in development mode
    if (isDevelopmentMode) {
      devBanner.classList.remove('hidden');
      document.body.classList.add('dev-mode-active');
      console.log('Running in development mode');
    }
    
    // Setup audio feedback and media recorder
    await initAudioFeedback();
    await setupMediaRecorder();
    
  } catch (error) {
    console.warn('Initial setup failed:', error);
    // Still show development banner even if audio setup fails
    if (isDevelopmentMode) {
      devBanner.classList.remove('hidden');
      document.body.classList.add('dev-mode-active');
    }
  }
})();
