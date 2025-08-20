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
        Just speak what you ate, and get detailed nutrition information for 39 essential nutrients.
      </p>
      <p class="text-sm text-accent font-semibold mt-2">
        ✨ Complete DRI Coverage - Track 100% of nutrients with established dietary guidelines
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
          <p>Get detailed nutrition information for 39 nutrients including vitamins, minerals, and macronutrients</p>
        </div>
      </div>
      
      <div class="card bg-base-200 shadow-xl">
        <div class="card-body text-center">
          <div class="text-3xl mb-4">📊</div>
          <h3 class="card-title justify-center">Comprehensive Tracking</h3>
          <p>Track all 39 essential nutrients with established dietary guidelines for complete nutrition insights</p>
        </div>
      </div>
    </div>

    <!-- Comprehensive Nutrition Tracking -->
    <div class="max-w-6xl mx-auto mt-16 mb-12">
      <div class="text-center mb-8">
        <h2 class="text-3xl font-bold text-primary mb-4">Complete Nutrition Tracking</h2>
        <p class="text-lg text-base-content/70">
          Track every nutrient that matters with our comprehensive 39-nutrient coverage
        </p>
      </div>

      <div class="grid lg:grid-cols-3 md:grid-cols-2 gap-6">
        <!-- Macronutrients & Energy -->
        <div class="card bg-gradient-to-br from-primary/10 to-primary/5 border border-primary/20">
          <div class="card-body">
            <h3 class="card-title text-primary flex items-center">
              <span class="text-xl mr-2">⚡</span>
              Macronutrients & Energy
            </h3>
            <ul class="text-sm space-y-1 text-base-content/80">
              <li>• Calories</li>
              <li>• Protein</li>
              <li>• Total Fat & Saturated Fat</li>
              <li>• Trans Fat & Cholesterol</li>
              <li>• Total & Added Sugars</li>
              <li>• Carbohydrates & Fiber</li>
            </ul>
          </div>
        </div>

        <!-- Vitamins -->
        <div class="card bg-gradient-to-br from-secondary/10 to-secondary/5 border border-secondary/20">
          <div class="card-body">
            <h3 class="card-title text-secondary flex items-center">
              <span class="text-xl mr-2">🍊</span>
              Essential Vitamins
            </h3>
            <ul class="text-sm space-y-1 text-base-content/80">
              <li>• Vitamins A, C, D, E, K</li>
              <li>• B-Complex: B1, B2, B3, B6, B12</li>
              <li>• Folate & Biotin</li>
              <li>• Pantothenic Acid (B5)</li>
              <li>• Choline</li>
            </ul>
          </div>
        </div>

        <!-- Minerals & Electrolytes -->
        <div class="card bg-gradient-to-br from-accent/10 to-accent/5 border border-accent/20">
          <div class="card-body">
            <h3 class="card-title text-accent flex items-center">
              <span class="text-xl mr-2">🧂</span>
              Minerals & Electrolytes
            </h3>
            <ul class="text-sm space-y-1 text-base-content/80">
              <li>• Sodium, Potassium, Chloride</li>
              <li>• Calcium, Magnesium, Phosphorus</li>
              <li>• Iron, Zinc, Copper</li>
              <li>• Selenium, Manganese</li>
              <li>• Iodine, Chromium, Fluoride</li>
              <li>• Molybdenum</li>
            </ul>
          </div>
        </div>
      </div>

      <!-- DRI Achievement Badge -->
      <div class="text-center mt-8">
        <div class="inline-flex items-center bg-success/10 border border-success/30 rounded-full px-6 py-3">
          <svg class="w-5 h-5 text-success mr-2" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
          </svg>
          <span class="text-success font-semibold">100% DRI Coverage</span>
          <span class="text-success/70 ml-2">• All nutrients with established dietary guidelines</span>
        </div>
      </div>
    </div>
  </div>
</div>