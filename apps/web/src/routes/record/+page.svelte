<script lang="ts">
  import { apiClient } from "$lib/api/client"
  import ConsumptionDisplay from "$lib/components/ConsumptionDisplay.svelte"
  import { getAppName } from "$lib/utils/app-info"

  // Get app name from runtime environment
  $: appName = getAppName()

  let isRecording = false
  let mediaRecorder: MediaRecorder | null = null
  let audioBlob: Blob | null = null
  let status = "Ready to record"
  let transcript = ""
  let result: any = null
  let error: string = ""
  let consumptionId: string | null = null
  
  // Text input mode
  let isTextMode = false
  let textInput = ""

  async function startRecording() {
    try {
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
      }

      mediaRecorder.start()
      isRecording = true
      status = "Recording... Tap to stop"
      error = ""
    } catch (err) {
      error =
        "Failed to access microphone. Please ensure you have given permission."
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
    if (!audioBlob) return

    try {
      status = "⏳ Processing..."
      error = ""
      transcript = ""
      result = null
      consumptionId = null

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
      status = "✅ Complete"
    } catch (err) {
      error = `Error processing audio: ${err}`
      status = "❌ Error occurred"
      console.error("Upload error:", err)
    }
  }

  async function submitText() {
    if (!textInput.trim()) {
      error = "Please enter a description of your meal"
      return
    }

    try {
      status = "⏳ Processing..."
      error = ""
      transcript = ""
      result = null
      consumptionId = null

      const response = await apiClient.POST("/consumption", {
        body: {
          text: textInput.trim()
        },
      })

      if (response.error) {
        throw new Error(`API Error: ${response.error}`)
      }

      const data = response.data
      transcript = data?.transcript || textInput.trim()
      result = data
      consumptionId = data?.id || null
      status = "✅ Complete"

      // Clear text input once processing is complete
      textInput = ""
    } catch (err) {
      error = `Error processing text: ${err}`
      status = "❌ Error occurred"
      console.error("Submit error:", err)
    }
  }

  // Function to toggle between recording and text modes
  function toggleMode() {
    isTextMode = !isTextMode
    // Reset states when switching modes
    error = ""
    status = isTextMode ? "Ready to type" : "Ready to record"
    textInput = ""
    audioBlob = null
  }

  // Auto-upload when recording stops
  $: if (audioBlob && status === "Processing...") {
    uploadAudio()
  }

  async function handleRedo() {
    if (!consumptionId) return

    const confirmed = confirm(
      "Delete this entry and record again? This action cannot be undone.",
    )
    if (!confirmed) return

    try {
      error = ""
      const deleteResponse = await apiClient.DELETE("/consumption/{id}", {
        params: { path: { id: consumptionId } },
      })

      if (deleteResponse.error) {
        throw new Error(`Delete failed: ${deleteResponse.error}`)
      }

      // Reset the interface to recording state
      result = null
      transcript = ""
      consumptionId = null
      status = isTextMode ? "Ready to type" : "Ready to record"
    } catch (err) {
      error = `Error deleting consumption: ${err}`
      console.error("Delete error:", err)
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
      <div class="text-center max-w-md w-full space-y-8">
        
        <!-- Mode toggle -->
        <div class="flex justify-center mb-6">
          <div class="btn-group">
            <button 
              class="btn btn-sm {!isTextMode ? 'btn-primary' : 'btn-outline'}"
              on:click={() => !isTextMode || toggleMode()}
              disabled={status.includes("Processing")}
            >
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z"/>
              </svg>
              Voice
            </button>
            <button 
              class="btn btn-sm {isTextMode ? 'btn-primary' : 'btn-outline'}"
              on:click={() => isTextMode || toggleMode()}
              disabled={status.includes("Processing")}
            >
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"/>
              </svg>
              Type
            </button>
          </div>
        </div>

        {#if isTextMode}
          <!-- Text input interface -->
          <div class="space-y-4">
            <textarea
              bind:value={textInput}
              placeholder="Describe what you ate... (e.g., 'I had a chicken caesar salad with croutons and parmesan cheese')"
              class="textarea textarea-primary w-full h-32 resize-none"
              disabled={status.includes("Processing")}
              on:keydown={(e) => {
                if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
                  e.preventDefault()
                  submitText()
                }
              }}
            ></textarea>
            <button
              class="btn btn-primary btn-lg w-full"
              on:click={submitText}
              disabled={status.includes("Processing") || !textInput.trim()}
            >
              {#if status.includes("Processing")}
                <span class="loading loading-spinner loading-sm"></span>
                Processing...
              {:else}
                Analyze Meal
              {/if}
            </button>
          </div>
        {:else}
          <!-- Recording button interface -->
          <div class="flex justify-center">
            <button
              class="record-button {isRecording
                ? 'recording'
                : ''} {status.includes('Processing') ? 'processing' : ''}"
              on:click={toggleRecordingWithSound}
              disabled={status.includes("Processing")}
              aria-label={isRecording ? "Stop recording" : "Start recording"}
            >
              {#if status.includes("Processing")}
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
        <div class="status-text">
          {#if status.includes("Processing")}
            <p class="text-lg text-warning font-medium">
              Processing your meal...
            </p>
          {:else if isRecording}
            <p class="text-lg font-medium" style="color: var(--color-dark);">
              Recording... Tap to stop
            </p>
          {:else if error}
            <p class="text-lg text-error font-medium">{error}</p>
          {:else if isTextMode}
            <div class="space-y-2">
              <p class="text-xl font-semibold text-base-content">
                Describe what you ate
              </p>
              <p class="text-sm text-base-content/70">
                Type your meal description and click "Analyze Meal" or press Ctrl/Cmd+Enter
              </p>
            </div>
          {:else}
            <div class="space-y-2">
              <p class="text-xl font-semibold text-base-content">
                Tap to record your meal
              </p>
              <p class="text-sm text-base-content/70">
                Speak naturally about what you ate
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
              on:click={() => {
                result = null
                transcript = ""
                consumptionId = null
                audioBlob = null
                status = "Ready to record"
                error = ""
              }}
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
</style>
