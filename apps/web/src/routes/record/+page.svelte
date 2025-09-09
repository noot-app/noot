<script lang="ts">
  import { apiClient } from "$lib/api/client"
  import { toast } from "$lib/stores/toast"
  import { formatErrorForUser, handleApiCallWithAuthRedirect } from "$lib/utils/error-handling"
  import ConsumptionDisplay from "$lib/components/ConsumptionDisplay.svelte"
  import ConsumptionTextSubmit from "$lib/components/ConsumptionTextSubmit.svelte"
  import { getAppName } from "$lib/utils/app-info"
  import { onMount, onDestroy } from "svelte"
  import { browser } from "$app/environment"
  import { getStorageJSON, setStorageJSON, removeStorageItem } from "$lib/utils/secure-storage"

  // Get app name from runtime environment
  $: appName = getAppName()

  // Storage keys for draft persistence
  const DRAFT_STORAGE_KEYS = {
    MODE: 'record-draft-mode'
  }

  // Draft expiration time (24 hours)
  const DRAFT_EXPIRY_MS = 24 * 60 * 60 * 1000

  // Core recording state
  let isRecording = false
  let mediaRecorder: MediaRecorder | null = null
  let audioBlob: Blob | null = null
  let status = "Ready to record"
  let transcript = ""
  let result: any = null
  let error: string = ""
  let consumptionId: string | null = null
  
  // Text input mode - initialize to null to prevent flicker during SSR
  let isTextMode: boolean | null = null
  
  // State tracking
  let isProcessing = false
  let hasSubmitted = false
  let lastSubmissionId: string | null = null
  let lastSubmissionMode: boolean | null = null // Track the mode used for the last submission
  
  // Component references
  let textSubmitComponent: ConsumptionTextSubmit
  
  // Favorites state
  let isFavorited = false
  let isUpdatingFavorite = false
  
  // OS detection for keyboard shortcuts
  let isMac = false
  if (browser) {
    isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0 || navigator.userAgent.includes('Mac')
  }

  // Check if consumption is favorited
  async function checkFavoriteStatus() {
    if (!consumptionId) return

    try {
      const result = await handleApiCallWithAuthRedirect(async () => {
        return await apiClient.GET("/favorites")
      })

      if (result.error) {
        console.error("Failed to check favorite status:", result.error)
        return
      }

      const favorites = result.data?.favorites || []
      isFavorited = favorites.some(fav => fav.consumption_id === consumptionId)
    } catch (err) {
      console.error("Failed to check favorite status:", err)
    }
  }

  // Toggle favorite status
  async function toggleFavorite() {
    if (!consumptionId || isUpdatingFavorite) return

    isUpdatingFavorite = true

    try {
      if (isFavorited) {
        // Remove from favorites
        const result = await handleApiCallWithAuthRedirect(async () => {
          return await apiClient.DELETE("/favorites/{consumption_id}", {
            params: { path: { consumption_id: consumptionId! } }
          })
        })

        if (result.error) {
          toast.add(formatErrorForUser(result.error), "error")
          return
        }

        isFavorited = false
        toast.add("Removed from favorites", "success")
      } else {
        // Add to favorites
        const result = await handleApiCallWithAuthRedirect(async () => {
          return await apiClient.POST("/favorites", {
            body: {
              consumption_id: consumptionId!
            }
          })
        })

        if (result.error) {
          toast.add(formatErrorForUser(result.error), "error")
          return
        }

        isFavorited = true
        toast.add("Added to favorites", "success")
      }
    } catch (err) {
      console.error("Failed to toggle favorite:", err)
      toast.add("Failed to update favorites. Please try again.", "error")
    } finally {
      isUpdatingFavorite = false
    }
  }

  async function startRecording() {
    if (isProcessing || isRecording) return
    
    try {
      clearError()
      const stream = await navigator.mediaDevices.getUserMedia({
        audio: {
          echoCancellation: true,
          noiseSuppression: true,
          sampleRate: 44100,
        },
      })

      // Use webm/opus if supported, fallback to available formats
      const options = { mimeType: "audio/webm; codecs=opus" }
      if (!MediaRecorder.isTypeSupported(options.mimeType)) {
        // Fallback options
        const fallbacks = ["audio/webm", "audio/mp4", "audio/wav", ""]
        for (const mimeType of fallbacks) {
          if (!mimeType || MediaRecorder.isTypeSupported(mimeType)) {
            options.mimeType = mimeType
            break
          }
        }
      }

      mediaRecorder = new MediaRecorder(
        stream,
        options.mimeType ? options : undefined,
      )

      const audioChunks: Blob[] = []
      mediaRecorder.ondataavailable = (event) => {
        if (event.data.size > 0) {
          audioChunks.push(event.data)
        }
      }

      mediaRecorder.onstop = () => {
        audioBlob = new Blob(audioChunks, {
          type: options.mimeType || "audio/webm",
        })
        stream.getTracks().forEach((track) => track.stop())
        // Automatically upload when recording stops
        uploadAudio()
      }

      mediaRecorder.start()
      isRecording = true
      status = "Recording... Tap to stop"
    } catch (err) {
      setError("Failed to access microphone. Please ensure you have given permission.")
      console.error("Error accessing microphone:", err)
    }
  }

  function stopRecording() {
    if (mediaRecorder && mediaRecorder.state === "recording") {
      mediaRecorder.stop()
      isRecording = false
      status = "Processing..."
    }
  }

  async function uploadAudio() {
    if (!audioBlob || isProcessing) return

    try {
      isProcessing = true
      status = "⏳ Processing..."
      clearError()
      clearResults()

      const formData = new FormData()
      formData.append("audio", audioBlob, "audio.webm")

      const response = await apiClient.POST("/consumption", {
        body: formData as any, // FormData for multipart/form-data
      })

      if (response.error) {
        throw new Error(`API Error: ${response.error}`)
      }

      const data = response.data
      transcript = data?.transcript || ""
      result = data
      consumptionId = data?.id || null
      lastSubmissionId = consumptionId
      lastSubmissionMode = false // Audio submission
      status = "✅ Complete"
      hasSubmitted = true
      
      // Check favorite status for the new consumption
      if (consumptionId) {
        checkFavoriteStatus()
      }
      
      // Clear audio state after successful submission
      audioBlob = null
    } catch (err) {
      setError(`Error processing audio: ${err}`)
      status = "❌ Error occurred"
      console.error("Upload error:", err)
    } finally {
      isProcessing = false
    }
  }

  // Function to toggle between recording and text modes
  function toggleMode() {
    if (isProcessing) return
    
    isTextMode = !isTextMode
    clearError()
    updateStatus()
    saveModeState()
  }

  // Utility functions for cleaner state management
  function clearError() {
    error = ""
  }

  function setError(message: string) {
    error = message
  }

  function clearResults() {
    transcript = ""
    result = null
    consumptionId = null
    isFavorited = false
  }

  function updateStatus() {
    if (isProcessing) {
      status = "⏳ Processing..."
    } else if (isRecording) {
      status = "Recording... Tap to stop"
    } else if (isTextMode === true) {
      status = "Ready to type"
    } else if (isTextMode === false) {
      status = "Ready to record"
    } else {
      // During initial load, keep status neutral
      status = "Loading..."
    }
  }

  // Function to persist mode state
  function saveModeState() {
    if (!browser) return
    
    const timestamp = Date.now()
    setStorageJSON(DRAFT_STORAGE_KEYS.MODE, {
      isTextMode,
      timestamp
    })
  }

  // Function to restore mode state
  function restoreModeState() {
    if (!browser) return
    
    const now = Date.now()
    
    // Restore mode selection - always set a definitive value to prevent flicker
    const modeDraft = getStorageJSON(DRAFT_STORAGE_KEYS.MODE, { isTextMode: false, timestamp: 0 })
    if ((now - modeDraft.timestamp) <= DRAFT_EXPIRY_MS) {
      isTextMode = modeDraft.isTextMode
    } else {
      // Set default to false if no valid draft exists
      isTextMode = false
    }
  }

  // Function to clear mode state
  function clearModeState() {
    if (!browser) return
    removeStorageItem(DRAFT_STORAGE_KEYS.MODE)
  }

  // Function to reset the page to initial state
  function resetToInitialState() {
    // Clear all state
    isRecording = false
    isProcessing = false
    hasSubmitted = false
    lastSubmissionId = null
    // Don't reset lastSubmissionMode here - we want to preserve it for "Record Another"
    mediaRecorder = null
    audioBlob = null
    
    clearError()
    clearResults()
    
    // Don't set isTextMode here - let restoreModeState handle it
    // This prevents the brief flash of wrong state when navigating to the page
  }

  // Function to set initial status based on current mode
  function setInitialStatus() {
    updateStatus()
  }

  // Function to start a new recording/entry (preserving drafts)
  function startNewEntry() {
    resetToInitialState()
    restoreModeState()
    setInitialStatus()
  }

  // Function to start another entry in the same mode as the last submission
  function startAnotherEntry() {
    resetToInitialState()
    
    // If we have a last submission mode, use that instead of restoring drafts
    if (lastSubmissionMode !== null) {
      isTextMode = lastSubmissionMode
      if (lastSubmissionMode === true && textSubmitComponent) {
        textSubmitComponent.reset() // Clear the text input for new entry
      }
    } else {
      // Fallback to draft restoration if no last submission mode
      restoreModeState()
    }
    
    setInitialStatus()
  }

  // Initialize component
  onMount(() => {
    // First restore mode to get the correct mode, then reset other state, then set status
    // This prevents the brief flash of microphone icon when in text mode
    restoreModeState()
    resetToInitialState()
    setInitialStatus()
  })

  // Cleanup on destroy
  onDestroy(() => {
    // Stop any ongoing recording
    if (mediaRecorder && mediaRecorder.state === "recording") {
      mediaRecorder.stop()
    }
    
    // Reset submission mode when navigating away from the page
    // This ensures that when users navigate back, they get draft restoration
    // instead of being locked into their last submission mode
    lastSubmissionMode = null
  })

  // Event handlers for the text submit component
  function handleTextSubmit(event: CustomEvent<{ transcript: string, result: any, consumptionId: string | null, submissionText: string }>) {
    const { transcript: textTranscript, result: textResult, consumptionId: textConsumptionId } = event.detail
    transcript = textTranscript
    result = textResult
    consumptionId = textConsumptionId
    lastSubmissionId = textConsumptionId
    lastSubmissionMode = true // Text submission
    status = "✅ Complete"
    hasSubmitted = true
    
    // Check favorite status for the new consumption
    if (consumptionId) {
      checkFavoriteStatus()
    }
  }

  function handleTextError(event: CustomEvent<{ message: string }>) {
    setError(event.detail.message)
    status = "❌ Error occurred"
  }

  function handleTextProcessing(event: CustomEvent<{ isProcessing: boolean }>) {
    isProcessing = event.detail.isProcessing
    if (isProcessing) {
      status = "⏳ Processing..."
      clearError()
      clearResults()
    }
  }

  async function handleRedo() {
    if (!consumptionId || isProcessing) return

    const confirmed = confirm(
      "Delete this entry and record again? This action cannot be undone.",
    )
    if (!confirmed) return

    try {
      clearError()
      isProcessing = true
      
      const deleteResponse = await apiClient.DELETE("/consumption/{id}", {
        params: { path: { id: consumptionId } },
      })

      if (deleteResponse.error) {
        throw new Error(`Delete failed: ${deleteResponse.error}`)
      }

      // Reset to fresh state for new recording
      startAnotherEntry()
    } catch (err) {
      setError(`Error deleting consumption: ${err}`)
      console.error("Delete error:", err)
    } finally {
      isProcessing = false
    }
  }

  function handleSave(event: CustomEvent) {
    // Update result with saved consumption data
    result = event.detail.consumption
    status = "✅ Updated"
  }

  // Sound effects
  function playStartSound() {
    // Create a simple start sound effect
    if (
      typeof window !== "undefined" &&
      (window.AudioContext || (window as any).webkitAudioContext)
    ) {
      const AudioContextConstructor =
        window.AudioContext || (window as any).webkitAudioContext
      const audioContext = new AudioContextConstructor()
      const oscillator = audioContext.createOscillator()
      const gainNode = audioContext.createGain()

      oscillator.connect(gainNode)
      gainNode.connect(audioContext.destination)

      oscillator.frequency.setValueAtTime(800, audioContext.currentTime)
      oscillator.frequency.exponentialRampToValueAtTime(
        1000,
        audioContext.currentTime + 0.1,
      )

      gainNode.gain.setValueAtTime(0, audioContext.currentTime)
      gainNode.gain.linearRampToValueAtTime(
        0.1,
        audioContext.currentTime + 0.01,
      )
      gainNode.gain.exponentialRampToValueAtTime(
        0.001,
        audioContext.currentTime + 0.1,
      )

      oscillator.start(audioContext.currentTime)
      oscillator.stop(audioContext.currentTime + 0.1)
    }
  }

  function playStopSound() {
    // Create a simple stop sound effect
    if (
      typeof window !== "undefined" &&
      (window.AudioContext || (window as any).webkitAudioContext)
    ) {
      const AudioContextConstructor =
        window.AudioContext || (window as any).webkitAudioContext
      const audioContext = new AudioContextConstructor()
      const oscillator = audioContext.createOscillator()
      const gainNode = audioContext.createGain()

      oscillator.connect(gainNode)
      gainNode.connect(audioContext.destination)

      oscillator.frequency.setValueAtTime(600, audioContext.currentTime)
      oscillator.frequency.exponentialRampToValueAtTime(
        400,
        audioContext.currentTime + 0.15,
      )

      gainNode.gain.setValueAtTime(0, audioContext.currentTime)
      gainNode.gain.linearRampToValueAtTime(
        0.1,
        audioContext.currentTime + 0.01,
      )
      gainNode.gain.exponentialRampToValueAtTime(
        0.001,
        audioContext.currentTime + 0.15,
      )

      oscillator.start(audioContext.currentTime)
      oscillator.stop(audioContext.currentTime + 0.15)
    }
  }

  // Enhanced toggle function with sound effects
  async function toggleRecordingWithSound() {
    if (isProcessing) return
    
    if (isRecording) {
      playStopSound()
      stopRecording()
    } else {
      playStartSound()
      await startRecording()
    }
  }
</script>

<svelte:head>
  <title>Record - {appName}</title>
</svelte:head>

<div
  class="gradient-bg {result ? 'overflow-y-auto' : 'flex flex-col'} min-h-[calc(100vh-8rem)]"
>
  <!-- Main content area -->

  <!-- Main recording interface - only show when not complete -->
  {#if !result}
    <div class="flex-1 flex items-center justify-center px-4">
      <div class="text-center max-w-4xl w-full space-y-8">
        
        {#if isTextMode === null}
          <!-- Loading state during SSR/initialization -->
          <div class="flex justify-center">
            <div class="loading loading-spinner loading-lg text-primary"></div>
          </div>
        {:else if isTextMode === true}
          <!-- Chat-style text input interface -->
          <ConsumptionTextSubmit
            bind:this={textSubmitComponent}
            disabled={isProcessing}
            on:submit={handleTextSubmit}
            on:error={handleTextError}
            on:processing={handleTextProcessing}
          />
        {:else}
          <!-- Recording button interface -->
          <div class="flex justify-center">
            <button
              class="record-button {isRecording
                ? 'recording'
                : ''} {isProcessing ? 'processing' : ''}"
              on:click={toggleRecordingWithSound}
              disabled={isProcessing}
              aria-label={isRecording ? "Stop recording" : "Start recording"}
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
        {/if}

        <!-- Status message -->
        <div class="status-text mt-6">
          {#if isRecording}
            <p class="text-lg font-medium" style="color: var(--color-dark);">
              Recording... Tap to stop
            </p>
          {:else if error}
            <p class="text-lg text-error font-medium">{error}</p>
          {:else if isTextMode === true}
            <div class="space-y-2">
              <p class="text-sm text-base-content/70 flex items-center justify-center gap-1 flex-wrap">
                <span>Press</span>
                <kbd class="kbd kbd-sm">Enter</kbd>
                <span>to submit</span>
              </p>
            </div>
          {:else if isTextMode === false}
            <div class="space-y-2">
              <p class="text-xl font-semibold text-base-content">
                Tap to record your meal
              </p>
              <p class="text-sm text-base-content/70">
                Speak naturally about what you ate
              </p>
            </div>
          {:else}
            <!-- Loading state -->
            <div class="space-y-2">
              <p class="text-lg text-base-content/70">
                Loading...
              </p>
            </div>
          {/if}
        </div>
      </div>
    </div>
  {/if}

  <!-- Results section (only shown when there are results) -->
  {#if transcript || result}
    <div class="bg-base-100 p-6 fade-in">
      <div class="container mx-auto max-w-4xl space-y-6">
        <ConsumptionDisplay
          consumption={result}
          {transcript}
          showRedoButton={true}
          autoShowLabelEdit={true}
          editable={true}
          on:redo={handleRedo}
          on:save={handleSave}
        />

        <!-- Navigation buttons -->
        {#if status === "✅ Complete" || status === "✅ Updated"}
          <div class="flex flex-col sm:flex-row justify-center gap-4 mt-8">
            <!-- Star button for favorites -->
            {#if consumptionId}
              <button
                class="btn btn-ghost min-h-[44px]"
                class:loading={isUpdatingFavorite}
                on:click={toggleFavorite}
                disabled={isUpdatingFavorite}
                title={isFavorited ? "Remove from favorites" : "Add to favorites"}
              >
                {#if isUpdatingFavorite}
                  <span class="loading loading-spinner loading-sm mr-2"></span>
                {:else}
                  <span class="text-xl mr-2">{isFavorited ? "⭐" : "☆"}</span>
                {/if}
                {isFavorited ? "Remove from Favorites" : "Add to Favorites"}
              </button>
            {/if}
            
            <a href="/summary" class="btn btn-outline min-h-[44px]">
              <svg
                class="w-4 h-4 mr-2"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9 19v-6a2 2 0 00-2 2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
                />
              </svg>
              View Summary
            </a>
            <button
              class="btn btn-primary min-h-[44px]"
              on:click={startAnotherEntry}
            >
              <svg
                class="w-4 h-4 mr-2"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 6v6m0 0v6m0-6h6m-6 0H6"
                />
              </svg>
              Record Another
            </button>
          </div>
        {/if}
      </div>
    </div>
  {/if}

  <!-- Single toggle button at bottom -->
  {#if !result && isTextMode !== null}
    <div class="fixed bottom-6 left-1/2 transform -translate-x-1/2 z-10">
      <button
        class="toggle-button"
        on:click={toggleMode}
        disabled={isProcessing}
        aria-label="Switch to {isTextMode ? 'voice' : 'text'} mode"
      >
        {#if isTextMode}
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z"/>
          </svg>
        {:else}
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"/>
          </svg>
        {/if}
      </button>
    </div>
  {/if}
</div>

<style>
  .record-button {
    width: 200px;
    height: 200px;
    border-radius: 50%;
    border: 4px solid;
    cursor: pointer;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    position: relative;
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: bold;
  }

  .record-button:not(.recording):not(.processing) {
    border-color: hsl(var(--p));
    background-color: hsl(var(--p));
    color: hsl(var(--pc));
    transform: scale(1);
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  }

  .record-button:not(.recording):not(.processing):hover {
    transform: scale(1.05);
    box-shadow:
      0 25px 50px -12px rgba(0, 0, 0, 0.25),
      0 0 0 4px hsla(var(--p), 0.25);
  }

  .record-button:not(.recording):not(.processing):active {
    transform: scale(0.95);
  }

  .record-button.recording {
    border-color: var(--color-dark);
    background-color: var(--color-dark);
    color: var(--color-dark-content);
    animation: pulse-recording 1.5s ease-in-out infinite;
  }

  .record-button.processing {
    border-color: hsl(var(--wa));
    background-color: hsl(var(--wa));
    color: hsl(var(--wac));
    animation: spin 2s linear infinite;
  }

  @keyframes pulse-recording {
    0%,
    100% {
      transform: scale(1);
      box-shadow: 0 0 0 0 color-mix(in srgb, var(--color-dark) 70%, transparent);
    }
    50% {
      transform: scale(1.05);
      box-shadow: 0 0 0 20px
        color-mix(in srgb, var(--color-dark) 0%, transparent);
    }
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(-360deg);
    }
  }

  .status-text {
    transition: all 0.3s ease;
  }

  .fade-in {
    animation: fadeIn 0.5s ease-in;
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .gradient-bg {
    background: linear-gradient(135deg, hsl(var(--b1)), hsl(var(--b2)));
  }

  .toggle-button {
    width: 56px;
    height: 56px;
    border-radius: 28px;
    border: none;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    
    /* Glass morphism effect */
    background: rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.2);
    box-shadow: 
      0 8px 32px rgba(0, 0, 0, 0.1),
      inset 0 1px 0 rgba(255, 255, 255, 0.2);
    
    color: hsl(var(--bc));
  }

  .toggle-button:hover:not(:disabled) {
    transform: translateY(-2px);
    background: rgba(255, 255, 255, 0.15);
    box-shadow: 
      0 12px 40px rgba(0, 0, 0, 0.15),
      inset 0 1px 0 rgba(255, 255, 255, 0.3);
  }

  .toggle-button:active:not(:disabled) {
    transform: translateY(0);
    background: rgba(255, 255, 255, 0.08);
  }

  .toggle-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
  }
</style>
