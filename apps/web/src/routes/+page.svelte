<script lang="ts">
  import { PUBLIC_APP_NAME } from "$env/static/public";
</script>

<svelte:head>
  <title>{PUBLIC_APP_NAME} - AI-Powered Nutrition Logger</title>
</svelte:head>

<div class="min-h-screen bg-base-100">
  <div class="container mx-auto px-4 py-8">
    <!-- Header -->
    <header class="text-center mb-12">
      <h1 class="text-4xl font-bold text-primary mb-4">
        Welcome to {PUBLIC_APP_NAME}
      </h1>
      <p class="text-lg text-base-content/70 max-w-2xl mx-auto">
        Log your meals effortlessly with AI-powered voice recognition. 
        Just speak what you ate, and get detailed nutrition information instantly.
      </p>
    </header>

    <!-- Navigation -->
    <nav class="flex justify-center mb-12">
      <div class="btn-group">
        <a href="/record" class="btn btn-primary">
          <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
                  d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
          </svg>
          Start Recording
        </a>
        <a href="/summary" class="btn btn-outline">
          <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
                  d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
          </svg>
          View Summary
        </a>
      </div>
    </nav>

    <!-- Features -->
    <div class="grid md:grid-cols-3 gap-8 max-w-4xl mx-auto">
      <div class="card bg-base-200 shadow-xl">
        <div class="card-body text-center">
          <div class="text-3xl mb-4">🎤</div>
          <h3 class="card-title justify-center">Voice Recognition</h3>
          <p>Simply speak what you ate and let AI transcribe and analyze your meals</p>
        </div>
      </div>
      
      <div class="card bg-base-200 shadow-xl">
        <div class="card-body text-center">
          <div class="text-3xl mb-4">🥗</div>
          <h3 class="card-title justify-center">Smart Analysis</h3>
          <p>Get detailed nutrition information including calories, protein, carbs, and more</p>
        </div>
      </div>
      
      <div class="card bg-base-200 shadow-xl">
        <div class="card-body text-center">
          <div class="text-3xl mb-4">📊</div>
          <h3 class="card-title justify-center">Track Progress</h3>
          <p>View daily and weekly summaries of your nutrition intake</p>
        </div>
      </div>
    </div>
  </div>
</div>