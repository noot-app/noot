// DOM Elements
const micBtn = document.getElementById('micBtn');
const statusEl = document.getElementById('status');
const resultEl = document.getElementById('result');
const transcriptEl = document.getElementById('transcript-text');
const summaryEl = document.getElementById('summary-content');

// MediaRecorder Setup
let mediaRecorder;
let chunks = [];

// Initialize MediaRecorder
async function setupMediaRecorder() {
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

// Display results
function displayResults(data) {
  setStatus('');
  
  // Show transcript
  transcriptEl.textContent = data.transcript || 'No transcript available';
  
  // Show summary first (as requested)
  if (data.summary && data.summary.totals) {
    displaySummary(data.summary);
  }
  
  // Show results section
  resultEl.classList.remove('hidden');
  
  // Smooth scroll to results
  setTimeout(() => {
    resultEl.scrollIntoView({ behavior: 'smooth' });
  }, 100);
}

// Display nutrition summary
function displaySummary(summary) {
  const totals = summary.totals || {};
  const percentDaily = summary.percent_of_daily || {};
  
  const summaryData = [
    { label: 'Calories', value: formatNumber(totals.calories), unit: 'kcal', percent: Math.min(percentDaily['calories'] || 0, 100) },
    { label: 'Protein', value: formatNumber(totals.protein_g), unit: 'g', percent: Math.min(percentDaily['protein'] || 0, 100) },
    { label: 'Carbs', value: formatNumber(totals.total_carbs_g), unit: 'g', percent: Math.min(percentDaily['total_carbs'] || 0, 100) },
    { label: 'Fat', value: formatNumber(totals.total_fat_g), unit: 'g', percent: Math.min(percentDaily['total_fat'] || 0, 100) },
    { label: 'Fiber', value: formatNumber(totals.dietary_fiber_g), unit: 'g', percent: Math.min(percentDaily['dietary_fiber'] || 0, 100) },
    { label: 'Sugar', value: formatNumber(totals.total_sugars_g), unit: 'g', percent: Math.min(percentDaily['total_sugars'] || 0, 100) }
  ];

  summaryEl.innerHTML = summaryData.map(item => `
    <div class="summary-item">
      <div class="summary-label">${item.label}</div>
      <div class="summary-value">${item.value}${item.unit}</div>
      <div class="summary-percentage">${Math.round(item.percent)}% daily</div>
    </div>
  `).join('');
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

// Escape HTML to prevent XSS
function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

// Recording Functions
async function startRecording(event) {
  event.preventDefault();
  
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
  
  chunks = [];
  mediaRecorder.start();
  micBtn.classList.add('recording');
  setStatus('🎤 Recording... Release to stop');
}

function stopRecording(event) {
  event.preventDefault();
  
  if (!mediaRecorder || mediaRecorder.state !== 'recording') return;
  
  mediaRecorder.stop();
  micBtn.classList.remove('recording');
}

// Event Listeners
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

micBtn.addEventListener('touchcancel', (e) => {
  e.preventDefault();
  stopRecording(e);
});

// Initialize
(async () => {
  try {
    await setupMediaRecorder();
  } catch (error) {
    console.warn('Initial media setup failed, will try again on first recording');
  }
})();