<script lang="ts">
  import { onMount, onDestroy } from 'svelte'
  
  // Props
  export let isRecording: boolean = false
  export let isProcessing: boolean = false
  export let disabled: boolean = false
  export let onClick: () => void = () => {}
  export let ariaLabel: string = 'Record button'
  
  // Audio analysis variables
  let audioContext: AudioContext | null = null
  let analyser: AnalyserNode | null = null
  let microphone: MediaStreamAudioSourceNode | null = null
  let dataArray: Uint8Array | null = null
  let animationFrame: number | null = null
  
  // Reactive loudness and smoothing
  let loudness: number = 0
  let smoothedLoudness: number = 0
  let isAudioInitialized: boolean = false
  
  // Smoothing parameters
  const SMOOTHING_FACTOR = 0.85 // Higher = more smoothing
  const LOUDNESS_THRESHOLD = 0.02 // Minimum loudness to register
  const SCALE_MULTIPLIER = 0.4 // How much the button scales with loudness
  const GLOW_MULTIPLIER = 0.6 // How much the glow effect scales
  
  // Initialize audio context and mic input
  async function initializeAudio() {
    if (isAudioInitialized || !isRecording) return
    
    try {
      // Create audio context
      const AudioContextConstructor = window.AudioContext || (window as any).webkitAudioContext
      if (!AudioContextConstructor) {
        console.warn('Web Audio API not supported')
        return
      }
      
      audioContext = new AudioContextConstructor()
      
      // Get microphone access (reuse existing stream if possible)
      const stream = await navigator.mediaDevices.getUserMedia({ 
        audio: { 
          echoCancellation: false,
          noiseSuppression: false,
          autoGainControl: false,
          sampleRate: 44100
        } 
      })
      
      // Create audio analysis nodes
      microphone = audioContext.createMediaStreamSource(stream)
      analyser = audioContext.createAnalyser()
      
      // Configure analyser for better responsiveness
      analyser.fftSize = 512 // Higher resolution
      analyser.smoothingTimeConstant = 0.3 // Less smoothing for more responsiveness
      analyser.minDecibels = -90
      analyser.maxDecibels = -10
      
      // Connect nodes
      microphone.connect(analyser)
      
      // Create data array for frequency analysis
      const bufferLength = analyser.frequencyBinCount
      dataArray = new Uint8Array(bufferLength)
      
      isAudioInitialized = true
      
      // Start analysis loop
      analyzeAudio()
      
    } catch (error) {
      console.error('Failed to initialize audio analysis:', error)
      isAudioInitialized = false
    }
  }
  
  // Clean up audio resources
  function cleanupAudio() {
    if (animationFrame) {
      cancelAnimationFrame(animationFrame)
      animationFrame = null
    }
    
    if (microphone) {
      microphone.disconnect()
      microphone = null
    }
    
    if (audioContext) {
      audioContext.close()
      audioContext = null
    }
    
    analyser = null
    dataArray = null
    isAudioInitialized = false
    loudness = 0
    smoothedLoudness = 0
  }
  
  // Analyze audio and update loudness
  function analyzeAudio() {
    if (!analyser || !dataArray || !isRecording) {
      return
    }
    
    // Get frequency data
    analyser.getByteFrequencyData(dataArray)
    
    // Focus on human speech frequency range (85Hz - 8kHz)
    const sampleRate = audioContext?.sampleRate || 44100
    const nyquist = sampleRate / 2
    const binWidth = nyquist / dataArray.length
    
    const speechStartBin = Math.floor(85 / binWidth)
    const speechEndBin = Math.floor(8000 / binWidth)
    
    let sum = 0
    let count = 0
    let peak = 0
    
    // Calculate both average and peak for more dynamic response
    for (let i = speechStartBin; i < Math.min(speechEndBin, dataArray.length); i++) {
      const value = dataArray[i]
      sum += value
      count++
      peak = Math.max(peak, value)
    }
    
    // Combine average and peak for more responsive visualization
    const avgLoudness = count > 0 ? (sum / count) / 255 : 0
    const peakLoudness = peak / 255
    
    // Use weighted combination: 70% average, 30% peak
    const rawLoudness = (avgLoudness * 0.7) + (peakLoudness * 0.3)
    
    // Apply threshold and boost sensitivity
    loudness = rawLoudness > LOUDNESS_THRESHOLD ? Math.pow(rawLoudness, 0.5) : 0
    
    // Smooth the loudness using exponential moving average
    smoothedLoudness = smoothedLoudness * SMOOTHING_FACTOR + loudness * (1 - SMOOTHING_FACTOR)
    
    // Continue analysis
    animationFrame = requestAnimationFrame(analyzeAudio)
  }
  
  // Watch for recording state changes
  $: if (isRecording && !isAudioInitialized) {
    initializeAudio()
  } else if (!isRecording && isAudioInitialized) {
    cleanupAudio()
  }
  
  // Calculate button scale based on smoothed loudness
  $: buttonScale = isRecording 
    ? 1 + (smoothedLoudness * SCALE_MULTIPLIER)
    : 1
  
  // Calculate glow intensity
  $: glowIntensity = isRecording ? smoothedLoudness * GLOW_MULTIPLIER : 0
  
  // Cleanup on component destroy
  onDestroy(() => {
    cleanupAudio()
  })
</script>

<div class="mic-button-container">
  <!-- Outer glow rings for enhanced visual feedback -->
  {#if isRecording}
    <div 
      class="glow-ring-outer" 
      style="opacity: {glowIntensity * 0.3}; transform: scale({1 + smoothedLoudness * 0.8})"
    ></div>
    <div 
      class="glow-ring-middle" 
      style="opacity: {glowIntensity * 0.5}; transform: scale({1 + smoothedLoudness * 0.6})"
    ></div>
    <div 
      class="glow-ring-inner" 
      style="opacity: {glowIntensity * 0.8}; transform: scale({1 + smoothedLoudness * 0.3})"
    ></div>
  {/if}
  
  <button
    class="audio-reactive-mic-button {isRecording ? 'recording' : ''} {isProcessing ? 'processing' : ''}"
    style="transform: scale({buttonScale}); box-shadow: {isRecording ? `0 0 ${20 + (smoothedLoudness * 40)}px rgba(43, 38, 33, ${0.4 + (smoothedLoudness * 0.4)})` : ''}"
    on:click={onClick}
    {disabled}
    aria-label={ariaLabel}
  >
    {#if isProcessing}
      <!-- Processing spinner -->
      <svg
        class="w-16 h-16"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
        />
      </svg>
    {:else if isRecording}
      <!-- Stop icon (square) -->
      <svg class="w-20 h-20" fill="currentColor" viewBox="0 0 24 24">
        <rect x="6" y="6" width="12" height="12" rx="2" />
      </svg>
    {:else}
      <!-- Microphone icon -->
      <svg
        class="w-20 h-20"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2.5"
          d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z"
        />
      </svg>
    {/if}
  </button>
</div>

<style>
  .mic-button-container {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .audio-reactive-mic-button {
    width: 200px;
    height: 200px;
    border-radius: 50%;
    border: 4px solid;
    cursor: pointer;
    transition: box-shadow 0.1s ease-out; /* Only transition box-shadow for performance */
    position: relative;
    z-index: 10;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: bold;
    transform-origin: center;
  }

  .audio-reactive-mic-button:not(.recording):not(.processing) {
    border-color: hsl(var(--p));
    background-color: hsl(var(--p));
    color: hsl(var(--pc));
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .audio-reactive-mic-button:not(.recording):not(.processing):hover {
    transform: scale(1.05);
    box-shadow:
      0 25px 50px -12px rgba(0, 0, 0, 0.25),
      0 0 0 4px hsla(var(--p), 0.25);
  }

  .audio-reactive-mic-button:not(.recording):not(.processing):active {
    transform: scale(0.95);
  }

  .audio-reactive-mic-button.recording {
    border-color: var(--color-dark);
    background-color: var(--color-dark);
    color: var(--color-dark-content);
  }

  .audio-reactive-mic-button.processing {
    border-color: hsl(var(--wa));
    background-color: hsl(var(--wa));
    color: hsl(var(--wac));
    animation: spin 2s linear infinite;
  }

  /* Radiating glow rings for enhanced audio reactivity */
  .glow-ring-outer,
  .glow-ring-middle,
  .glow-ring-inner {
    position: absolute;
    border-radius: 50%;
    border: 2px solid;
    pointer-events: none;
    transition: all 0.1s ease-out;
    z-index: 1;
  }

  .glow-ring-outer {
    width: 280px;
    height: 280px;
    border-color: rgba(43, 38, 33, 0.15);
  }

  .glow-ring-middle {
    width: 240px;
    height: 240px;
    border-color: rgba(43, 38, 33, 0.25);
  }

  .glow-ring-inner {
    width: 220px;
    height: 220px;
    border-color: rgba(43, 38, 33, 0.4);
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
</style>