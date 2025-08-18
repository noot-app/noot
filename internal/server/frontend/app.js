const micBtn = document.getElementById('micBtn');
const statusEl = document.getElementById('status');
const resultEl = document.getElementById('result');
const transcriptEl = document.getElementById('transcript');
const itemsEl = document.getElementById('items');
const summaryEl = document.getElementById('summary');

let mediaRecorder;
let chunks = [];

async function setupStream() {
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
    setStatus('Uploading…');
    const blob = new Blob(chunks, { type: mediaRecorder.mimeType || 'audio/webm' });
    chunks = [];
    try {
      const fd = new FormData();
      fd.append('audio', blob, 'audio.webm');
      const resp = await fetch('/api/ingest', { method: 'POST', body: fd });
      if (!resp.ok) throw new Error('Server error');
      const data = await resp.json();
      renderResult(data);
      setStatus('');
    } catch (err) {
      setStatus('Failed: ' + err.message);
    }
  };
}

function setStatus(msg) {
  statusEl.textContent = msg;
}

function renderResult(data) {
  resultEl.classList.remove('hidden');
  transcriptEl.textContent = data.transcript || '';

  // Items
  itemsEl.innerHTML = '';
  const list = document.createElement('div');
  (data.items || []).forEach((it) => {
    const div = document.createElement('div');
    div.className = 'item';
    div.innerHTML = `
      <div><strong>${escapeHtml((it.item && it.item.name) || '')}</strong> ${
        it.item && it.item.brand ? '(' + escapeHtml(it.item.brand) + ')' : ''
      }</div>
      <div>Match: ${it.fdc ? escapeHtml(it.fdc.description || '') : '—'}</div>
      <div>Nutrients: ${
        it.nutrients
          ? `kcal ${num(it.nutrients.energy_kcal)}, P ${num(it.nutrients.protein_g)}g, F ${num(it.nutrients.fat_g)}g, C ${num(it.nutrients.carbs_g)}g, Fi ${num(it.nutrients.fiber_g)}g, S ${num(it.nutrients.sugar_g)}g`
          : 'n/a'
      }</div>
      ${it.note ? `<div><em>${escapeHtml(it.note)}</em></div>` : ''}
    `;
    list.appendChild(div);
  });
  itemsEl.appendChild(list);

  // Summary
  if (data.summary) {
    const t = data.summary.totals || {};
    const p = data.summary.percent_of_daily || {};
    summaryEl.innerHTML = `
      <h3>Totals</h3>
      <div>Calories: ${num(t.energy_kcal)} kcal (${pct(p['energy.kcal'])})</div>
      <div>Protein: ${num(t.protein_g)} g (${pct(p['protein.g'])})</div>
      <div>Fat: ${num(t.fat_g)} g (${pct(p['fat.g'])})</div>
      <div>Carbs: ${num(t.carbs_g)} g (${pct(p['carbs.g'])})</div>
      <div>Fiber: ${num(t.fiber_g)} g (${pct(p['fiber.g'])})</div>
      <div>Sugar: ${num(t.sugar_g)} g (${pct(p['sugar.g'])})</div>
    `;
  } else {
    summaryEl.textContent = '';
  }
}

function num(v) {
  const n = Number(v ?? 0);
  if (!isFinite(n)) return '0';
  return n.toFixed(0);
}
function pct(v) {
  return (typeof v === 'number' ? v : 0) + '%';
}
function escapeHtml(s) {
  return String(s).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
}

// Press-and-hold recording UX
micBtn.addEventListener('mousedown', startRecording);
micBtn.addEventListener('touchstart', startRecording);
micBtn.addEventListener('mouseup', stopRecording);
micBtn.addEventListener('mouseleave', stopRecording);
micBtn.addEventListener('touchend', stopRecording);

async function startRecording(ev) {
  ev.preventDefault();
  if (!mediaRecorder) {
    try { await setupStream(); } catch (e) { setStatus('Mic access denied'); return; }
  }
  if (mediaRecorder.state === 'recording') return;
  chunks = [];
  mediaRecorder.start();
  micBtn.classList.add('recording');
  setStatus('Recording… release to stop');
}

function stopRecording(ev) {
  ev.preventDefault();
  if (!mediaRecorder || mediaRecorder.state !== 'recording') return;
  mediaRecorder.stop();
  micBtn.classList.remove('recording');
}

(async () => {
  try { await setupStream(); } catch {}
})();
