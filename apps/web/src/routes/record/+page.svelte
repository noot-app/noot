<script lang="ts">
  import { apiClient } from "$lib/api/client";
  import { PUBLIC_APP_NAME } from "$env/static/public";
  import { onMount } from "svelte";
  import NutritionStats from "$lib/components/NutritionStats.svelte";

  let isRecording = false;
  let mediaRecorder: MediaRecorder | null = null;
  let audioBlob: Blob | null = null;
  let status = "Ready to record";
  let transcript = "";
  let result: any = null;
  let error: string = "";

  async function startRecording() {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ 
        audio: { 
          echoCancellation: true,
          noiseSuppression: true,
          sampleRate: 44100
        } 
      });
      
      // Use webm/opus if supported, fallback to available formats
      const options = { mimeType: 'audio/webm; codecs=opus' };
      if (!MediaRecorder.isTypeSupported(options.mimeType)) {
        // Fallback options
        const fallbacks = [
          'audio/webm',
          'audio/mp4',
          'audio/wav',
          ''
        ];
        for (const mimeType of fallbacks) {
          if (!mimeType || MediaRecorder.isTypeSupported(mimeType)) {
            options.mimeType = mimeType;
            break;
          }
        }
      }

      mediaRecorder = new MediaRecorder(stream, options.mimeType ? options : undefined);
      
      const audioChunks: Blob[] = [];
      mediaRecorder.ondataavailable = (event) => {
        if (event.data.size > 0) {
          audioChunks.push(event.data);
        }
      };

      mediaRecorder.onstop = () => {
        audioBlob = new Blob(audioChunks, { 
          type: options.mimeType || 'audio/webm' 
        });
        stream.getTracks().forEach(track => track.stop());
      };

      mediaRecorder.start();
      isRecording = true;
      status = "Recording... Tap to stop";
      error = "";
    } catch (err) {
      error = "Failed to access microphone. Please ensure you have given permission.";
      console.error("Error accessing microphone:", err);
    }
  }

  function stopRecording() {
    if (mediaRecorder && mediaRecorder.state === "recording") {
      mediaRecorder.stop();
      isRecording = false;
      status = "Processing...";
    }
  }

  async function toggleRecording() {
    if (isRecording) {
      stopRecording();
    } else {
      await startRecording();
    }
  }

  async function uploadAudio() {
    if (!audioBlob) return;

    try {
      status = "⏳ Processing...";
      error = "";
      transcript = "";
      result = null;

      const formData = new FormData();
      formData.append('audio', audioBlob, 'audio.webm');

      const response = await apiClient.POST('/consumption', {
        body: formData as any, // FormData for multipart/form-data
      });

      if (response.error) {
        throw new Error(`API Error: ${response.error}`);
      }

      const data = response.data;
      transcript = data?.transcript || "";
      result = data;
      status = "✅ Complete";
    } catch (err) {
      error = `Error processing audio: ${err}`;
      status = "❌ Error occurred";
      console.error("Upload error:", err);
    }
  }

  // Auto-upload when recording stops
  $: if (audioBlob && status === "Processing...") {
    uploadAudio();
  }
</script>

<svelte:head>
  <title>Record - {PUBLIC_APP_NAME}</title>
</svelte:head>

<div class="min-h-screen bg-base-100">
  <div class="container mx-auto px-4 py-8 max-w-4xl">
    <!-- Header -->
    <div class="text-center mb-8">
      <h1 class="text-3xl font-bold text-primary mb-4">Record Your Meal</h1>
      <p class="text-base-content/70">Tap the microphone and describe what you ate</p>
    </div>

    <!-- Navigation -->
    <div class="flex justify-center mb-8">
      <div class="btn-group">
        <a href="/" class="btn btn-ghost">
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
          Home
        </a>
        <a href="/summary" class="btn btn-outline">Summary</a>
      </div>
    </div>

    <!-- Recording Controls -->
    <div class="text-center mb-8">
      <button 
        class="btn btn-circle btn-lg {isRecording ? 'btn-error' : 'btn-primary'} mb-4"
        on:click={toggleRecording}
        disabled={status === "Processing..."}
      >
        {#if isRecording}
          <svg class="w-8 h-8" fill="currentColor" viewBox="0 0 24 24">
            <rect x="6" y="6" width="12" height="12" rx="2" />
          </svg>
        {:else}
          <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
                  d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
          </svg>
        {/if}
      </button>
      <div class="text-lg font-medium">{status}</div>
    </div>

    <!-- Error Display -->
    {#if error}
      <div class="alert alert-error mb-6">
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span>{error}</span>
      </div>
    {/if}

    <!-- Results -->
    {#if transcript || result}
      <div class="space-y-6">
        <!-- Transcript -->
        {#if transcript}
          <div class="card bg-base-200 shadow-xl">
            <div class="card-body">
              <h2 class="card-title">What you said:</h2>
              <p class="text-lg italic">"{transcript}"</p>
            </div>
          </div>
        {/if}

        <!-- Summary -->
        {#if result?.summary}
          <div class="card bg-primary/10 shadow-xl">
            <div class="card-body">
              <h2 class="card-title text-primary mb-4">Nutrition Summary</h2>
              <NutritionStats 
                calories={result.summary.totals.calories}
                protein={result.summary.totals.protein_g}
                carbs={result.summary.totals.total_carbs_g}
                fat={result.summary.totals.total_fat_g}
                size="compact"
                className="bg-transparent shadow-none"
              />
            </div>
          </div>
        {/if}

        <!-- Items -->
        {#if result?.items && result.items.length > 0}
          <div class="card bg-base-200 shadow-xl">
            <div class="card-body">
              <h2 class="card-title">Food Items</h2>
              <div class="space-y-4">
                {#each result.items as item}
                  <div class="card bg-base-100 shadow">
                    <div class="card-body p-4">
                      <h3 class="text-lg font-semibold">{item.item.name}</h3>
                      {#if item.item.quantity && item.item.unit}
                        <p class="text-sm text-base-content/70">
                          {item.item.quantity} {item.item.unit}
                        </p>
                      {/if}
                      {#if item.item.nutrients}
                        <div class="grid grid-cols-2 md:grid-cols-4 gap-2 text-sm mt-2">
                          <span>Cal: {item.item.nutrients.calories}</span>
                          <span>Pro: {item.item.nutrients.protein_g.toFixed(1)}g</span>
                          <span>Carb: {item.item.nutrients.total_carbs_g.toFixed(1)}g</span>
                          <span>Fat: {item.item.nutrients.total_fat_g.toFixed(1)}g</span>
                        </div>
                      {/if}
                    </div>
                  </div>
                {/each}
              </div>
            </div>
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>
