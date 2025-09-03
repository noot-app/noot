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
      status = "Ready to record"
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
        <!-- Main recording button -->
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
    <div class="bg-base-100 p-6 fade-in">
      <div class="container mx-auto max-w-4xl space-y-6">
        <!-- Action buttons (Edit/Redo) - only show if we have a consumption ID -->
        {#if consumptionId && status === "✅ Complete"}
          <div class="flex justify-center gap-4 mb-6">
            {#if !isEditing}
              <button
                class="btn btn-outline btn-primary"
                on:click={startEdit}
                disabled={isSubmitting}
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
                    d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                  />
                </svg>
                ✏️ Edit
              </button>
              <button
                class="btn btn-outline btn-error"
                on:click={redoRecording}
                disabled={isSubmitting}
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
                    d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                  />
                </svg>
                🔄 Redo
              </button>
            {:else}
              <button
                class="btn btn-primary"
                on:click={saveEdit}
                disabled={isSubmitting}
              >
                {#if isSubmitting}
                  <span class="loading loading-spinner loading-sm mr-2"></span>
                {:else}
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
                      d="M5 13l4 4L19 7"
                    />
                  </svg>
                {/if}
                Save Changes
              </button>
              <button
                class="btn btn-outline btn-ghost"
                on:click={cancelEdit}
                disabled={isSubmitting}
              >
                Cancel
              </button>
            {/if}
          </div>
        {/if}

        <!-- Transcript -->
        {#if transcript}
          <Card title="What you said:" compact>
            <p class="text-lg italic">"{transcript}"</p>
          </Card>
        {/if}

        <!-- Nutrition Summary -->
        {#if result?.summary}
          <Card title="Nutrition Summary" variant="primary">
            <NutritionStats
              calories={result.summary.totals.calories}
              protein={result.summary.totals.protein_g}
              carbs={result.summary.totals.total_carbs_g}
              fat={result.summary.totals.total_fat_g}
              size="compact"
              className="bg-transparent shadow-none"
            />
          </Card>
        {/if}

        <!-- Labels Section -->
        {#if consumptionId}
          <div class="card bg-base-200 shadow-lg">
            <div class="card-body">
              <div class="flex justify-between items-center mb-4">
                <h3 class="card-title text-sm flex items-center gap-2">
                  <TagIcon className="w-4 h-4" />
                  Labels
                </h3>

                {#if !isEditingLabels}
                  <button
                    class="btn btn-outline btn-sm"
                    on:click={startLabelEditing}
                    disabled={isAddingLabel}
                  >
                    <svg
                      class="w-4 h-4 mr-1"
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
                    Edit Labels
                  </button>
                {/if}
              </div>

              <!-- Applied Labels Display (when not editing) -->
              {#if !isEditingLabels}
                {#if currentLabels.size > 0}
                  <div class="flex flex-wrap gap-2">
                    {#each Array.from(currentLabels) as labelName}
                      {@const labelData = availableLabels.find(
                        (l) => l.name === labelName,
                      )}
                      {#if labelData}
                        <Label name={labelData.name} color={labelData.color} />
                      {/if}
                    {/each}
                  </div>
                {:else}
                  <div class="flex items-center gap-2">
                    <p class="text-sm text-base-content/70">
                      No labels applied to this meal
                    </p>
                    {#if availableLabels.length === 0 && !isLoadingLabels}
                      <button class="btn btn-xs" on:click={retryLoadLabels}
                        >Load labels</button
                      >
                    {/if}
                  </div>
                {/if}
              {/if}

              <!-- Label Selection Interface (GitHub-style) -->
              {#if isEditingLabels}
                <div class="space-y-4">
                  {#if availableLabels.length === 0 && !isLoadingLabels}
                    <div class="flex items-center gap-2">
                      <p class="text-sm text-base-content/70">
                        No labels available for selection.
                      </p>
                      <button class="btn btn-xs" on:click={retryLoadLabels}
                        >Retry</button
                      >
                    </div>
                  {/if}
                  <div class="grid grid-cols-1 gap-2 max-h-64 overflow-y-auto">
                    {#each availableLabels as label}
                      {@const isSelected = pendingLabels.has(label.name)}
                      <button
                        class="flex items-center justify-between p-3 rounded-lg border transition-all hover:bg-base-300 {isSelected
                          ? 'bg-base-300 border-primary'
                          : 'bg-base-100 border-base-300'}"
                        on:click={() => toggleLabelSelection(label.name)}
                        disabled={isAddingLabel}
                      >
                        <div class="flex items-center gap-3">
                          <div class="checkbox-wrapper">
                            <input
                              type="checkbox"
                              class="checkbox checkbox-primary checkbox-sm"
                              checked={isSelected}
                              readonly
                            />
                          </div>
                          <div class="flex items-center gap-2">
                            <div
                              class="w-3 h-3 rounded-full"
                              style="background-color: {label.color?.startsWith(
                                '#',
                              )
                                ? label.color
                                : `#${label.color}`};"
                            ></div>
                            <span class="font-medium">{label.name}</span>
                          </div>
                        </div>
                        {#if label.description}
                          <span
                            class="text-xs text-base-content/60 truncate ml-2"
                            >{label.description}</span
                          >
                        {/if}
                      </button>
                    {/each}
                  </div>

                  <!-- Action Buttons -->
                  <div
                    class="flex justify-end gap-2 pt-2 border-t border-base-300"
                  >
                    <button
                      class="btn btn-ghost btn-sm"
                      on:click={cancelLabelChanges}
                      disabled={isAddingLabel}
                    >
                      Cancel
                    </button>
                    <button
                      class="btn btn-primary btn-sm"
                      on:click={applyLabelChanges}
                      disabled={isAddingLabel}
                    >
                      {#if isAddingLabel}
                        <span class="loading loading-spinner loading-sm mr-1"
                        ></span>
                        Applying...
                      {:else}
                        <svg
                          class="w-4 h-4 mr-1"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M5 13l4 4L19 7"
                          />
                        </svg>
                        Apply Changes
                      {/if}
                    </button>
                  </div>
                </div>
              {/if}

              {#if isLoadingLabels}
                <div
                  class="flex items-center gap-2 text-sm text-base-content/70"
                >
                  <span class="loading loading-spinner loading-sm"></span>
                  Loading labels...
                </div>
              {/if}
            </div>
          </div>
        {/if}

        <!-- Food Items -->
        {#if result?.items && result.items.length > 0}
          <div class="card bg-base-200 shadow-lg">
            <div class="card-body">
              <div class="flex justify-between items-center mb-4">
                <h3 class="card-title text-sm">Food Items</h3>
                {#if isEditing}
                  <span class="badge badge-warning">Editing Mode</span>
                {/if}
              </div>
              <div class="space-y-3">
                {#each result.items as item, index}
                  <div class="card bg-base-100 shadow">
                    <div class="card-body p-4">
                      <h4 class="font-semibold">{item.item.name}</h4>

                      <!-- Quantity controls (editable in edit mode) -->
                      <div class="flex items-center gap-2 mt-2">
                        {#if isEditing}
                          <div class="flex items-center gap-2">
                            <label
                              for="quantity-{index}"
                              class="text-sm font-medium">Quantity:</label
                            >
                            <button
                              class="btn btn-circle btn-sm btn-outline"
                              on:click={() =>
                                updateItemQuantity(
                                  index,
                                  Math.max(
                                    0.1,
                                    (item.item.user_quantity || 1) - 0.5,
                                  ),
                                )}
                            >
                              -
                            </button>
                            <input
                              id="quantity-{index}"
                              type="number"
                              class="input input-sm input-bordered w-20 text-center"
                              value={item.item.user_quantity}
                              on:input={(e) => {
                                const target = e.target as HTMLInputElement
                                updateItemQuantity(
                                  index,
                                  parseFloat(target.value) || 1,
                                )
                              }}
                              min="0.1"
                              step="0.5"
                            />
                            <button
                              class="btn btn-circle btn-sm btn-outline"
                              on:click={() =>
                                updateItemQuantity(
                                  index,
                                  (item.item.user_quantity || 1) + 0.5,
                                )}
                            >
                              +
                            </button>
                            {#if item.item.user_unit}
                              <span class="text-sm text-base-content/70"
                                >{item.item.user_unit}</span
                              >
                            {/if}
                          </div>
                        {:else}
                          <div class="flex items-center gap-2">
                            {#if item.item.user_quantity && item.item.user_unit}
                              <p class="text-sm text-base-content/70">
                                {item.item.user_quantity}
                                {item.item.user_unit}
                                {#if item.item.user_unit !== "g" && item.item.user_unit !== "gram" && item.item.user_unit !== "grams"}
                                  <span class="text-xs text-base-content/50">
                                    ({Math.round(item.item.grams)}g)
                                  </span>
                                {/if}
                              </p>
                            {:else}
                              <p class="text-sm text-base-content/70">
                                {Math.round(item.item.grams)}g
                              </p>
                            {/if}
                          </div>
                        {/if}
                      </div>

                      {#if item.item.nutrients}
                        <div
                          class="grid grid-cols-1 md:grid-cols-2 gap-2 text-sm mt-3"
                        >
                          <div class="bg-base-200 rounded p-2">
                            <span class="font-medium">Calories:</span>
                            {Math.round(item.item.nutrients.calories)}
                          </div>
                          <div class="bg-base-200 rounded p-2">
                            <span class="font-medium">Protein:</span>
                            {item.item.nutrients.protein_g.toFixed(1)}g
                          </div>
                          <div class="bg-base-200 rounded p-2">
                            <span class="font-medium">Carbohydrates:</span>
                            {item.item.nutrients.total_carbs_g.toFixed(1)}g
                          </div>
                          <div class="bg-base-200 rounded p-2">
                            <span class="font-medium">Total Fat:</span>
                            {item.item.nutrients.total_fat_g.toFixed(1)}g
                          </div>
                        </div>

                        <!-- Nutrient Composition Dropdown -->
                        <NutrientComposition nutrients={item.item.nutrients} />
                      {/if}
                    </div>
                  </div>
                {/each}
              </div>
            </div>
          </div>
        {/if}

        <!-- Meal Goals Progress - Show how this meal contributes to daily goals -->
        {#if result?.items && result.items.length > 0}
          <div class="space-y-4">
            <Goals
              currentNutrition={currentMealNutrition}
              showMealContribution={true}
            />
          </div>
        {/if}

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
                  d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
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
                labelsLoaded = false
                isEditing = false
                availableLabels = [] // Reset labels for new recording
                currentLabels = new Set()
                pendingLabels = new Set()
                isEditingLabels = false
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
