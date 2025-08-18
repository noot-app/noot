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

// Demo mode for testing
const DEMO_MODE = false;

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

// Demo data for testing
function getDemoData() {
  return {
    transcript: "I had a large chicken Caesar salad with croutons, two slices of whole wheat bread, a cup of strawberries, and a glass of orange juice for lunch.",
    items: [
      {
        item: {
          name: "Large chicken Caesar salad with croutons",
          quantity: 1,
          unit: "serving",
          nutrients: {
            calories: 320,
            protein_g: 25,
            total_carbs_g: 15,
            total_fat_g: 18,
            dietary_fiber_g: 3,
            total_sugars_g: 4,
            vitamin_c_mg: 8,
            calcium_mg: 80,
            iron_mg: 2.1,
            sodium_mg: 680,
            potassium_mg: 180
          }
        }
      },
      {
        item: {
          name: "Whole wheat bread",
          quantity: 2,
          unit: "slices",
          nutrients: {
            calories: 160,
            protein_g: 6,
            total_carbs_g: 24,
            total_fat_g: 3,
            dietary_fiber_g: 4,
            total_sugars_g: 2,
            vitamin_c_mg: 0,
            calcium_mg: 60,
            iron_mg: 1.8,
            sodium_mg: 240,
            potassium_mg: 120
          }
        }
      },
      {
        item: {
          name: "Fresh strawberries",
          quantity: 1,
          unit: "cup",
          nutrients: {
            calories: 50,
            protein_g: 1,
            total_carbs_g: 12,
            total_fat_g: 0.5,
            dietary_fiber_g: 3,
            total_sugars_g: 7,
            vitamin_c_mg: 85,
            calcium_mg: 20,
            iron_mg: 0.4,
            sodium_mg: 2,
            potassium_mg: 150
          }
        }
      },
      {
        item: {
          name: "Orange juice",
          quantity: 8,
          unit: "fl oz",
          nutrients: {
            calories: 120,
            protein_g: 2,
            total_carbs_g: 28,
            total_fat_g: 0.5,
            dietary_fiber_g: 0,
            total_sugars_g: 21,
            vitamin_c_mg: 124,
            calcium_mg: 40,
            iron_mg: 0.5,
            sodium_mg: 5,
            potassium_mg: 450
          }
        }
      }
    ],
    summary: {
      totals: {
        calories: 650,
        protein_g: 35,
        total_carbs_g: 45,
        total_fat_g: 28,
        dietary_fiber_g: 8,
        total_sugars_g: 25,
        vitamin_c_mg: 120,
        vitamin_d_mcg: 2.5,
        calcium_mg: 200,
        iron_mg: 4.2,
        sodium_mg: 980,
        potassium_mg: 650
      },
      percent_of_daily: {
        calories: 32,
        protein: 70,
        total_carbs: 15,
        total_fat: 36,
        dietary_fiber: 29,
        total_sugars: 28,
        vitamin_c: 133,
        vitamin_d: 17,
        calcium: 20,
        iron: 23,
        sodium: 43,
        potassium: 14
      }
    }
  };
}

// Initialize MediaRecorder
async function setupMediaRecorder() {
  if (DEMO_MODE) {
    console.log('Running in demo mode');
    return;
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
  if (DEMO_MODE) {
    // Simulate API delay and response
    setStatus('⏳ Processing your meal...');
    setTimeout(() => {
      displayResults(getDemoData());
    }, 2000);
    return;
  }
  
  try {
    setStatus('⏳ Processing your meal...');
    
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

// Display comprehensive nutrition summary
function displaySummary(summary) {
  const totals = summary.totals || {};
  const percentDaily = summary.percent_of_daily || {};
  
  // Enhanced nutrition data including vitamins and minerals
  const nutritionData = [
    // Macronutrients
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
    },
    
    // Fiber and sugars
    { 
      label: 'Fiber', 
      value: formatNumber(totals.dietary_fiber_g), 
      unit: 'g', 
      percent: Math.min(percentDaily['dietary_fiber'] || 0, 100),
      category: 'fiber'
    },
    { 
      label: 'Sugar', 
      value: formatNumber(totals.total_sugars_g), 
      unit: 'g', 
      percent: Math.min(percentDaily['total_sugars'] || 0, 100),
      category: 'sugar'
    },
    
    // Vitamins (if available in API response)
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
    
    // Minerals (if available in API response)
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
  ];

  // Filter out items with no data and create HTML
  const validNutrients = nutritionData.filter(item => 
    parseFloat(item.value) > 0 || ['Calories', 'Protein', 'Carbs', 'Fat'].includes(item.label)
  );

  summaryEl.innerHTML = validNutrients.map(item => {
    const color = getPercentageColor(item.percent);
    const pieChart = createPieChart(item.percent, color);
    
    return `
      <div class="summary-item">
        <div style="display: flex; align-items: center; justify-content: center; margin-bottom: 0.5rem;">
          ${pieChart}
        </div>
        <div style="font-size: 0.875rem; color: var(--text-secondary); margin-bottom: 0.25rem;">${item.label}</div>
        <div style="font-size: 1.125rem; font-weight: 700; color: var(--text-primary); margin-bottom: 0.25rem;">${item.value}${item.unit}</div>
        <div style="font-size: 0.75rem; color: var(--text-tertiary);">${Math.round(item.percent)}% daily</div>
      </div>
    `;
  }).join('');
  
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
  
  if (DEMO_MODE) {
    // Demo mode - show recording animation and process demo data
    // Skip audio feedback in demo mode to avoid issues
    console.log('Demo mode: Starting recording simulation');
    micBtn.classList.add('recording');
    setStatus('🎤 Recording... Release to stop');
    return;
  }
  
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
  
  if (DEMO_MODE) {
    // Demo mode - stop recording animation and process demo data
    console.log('Demo mode: Stopping recording simulation');
    micBtn.classList.remove('recording');
    sendAudioToAPI(null); // Will use demo data
    return;
  }
  
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
    
    // Only try audio feedback if not in demo mode (to avoid errors in headless environment)
    if (!DEMO_MODE) {
      await initAudioFeedback();
      await setupMediaRecorder();
    } else {
      console.log('Running in demo mode - skipping audio setup');
    }
    
    // Show demo message for testing
    if (DEMO_MODE) {
      setTimeout(() => {
        setStatus('👋 Click and hold the button to see a demo');
      }, 1000);
    }
  } catch (error) {
    console.warn('Initial setup failed:', error);
    if (DEMO_MODE) {
      setTimeout(() => {
        setStatus('👋 Click and hold the button to see a demo');
      }, 1000);
    }
  }
})();
