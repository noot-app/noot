<script lang="ts">
  import { PUBLIC_APP_NAME } from "$env/static/public";
  import { onMount } from "svelte";
  
  // Import component styles
  import "$lib/styles/gradient-hero.css";
  import "$lib/styles/phone-mockup.css";
  import "$lib/styles/hero-components.css";
  
  let phoneAnimated = false;
  
  onMount(() => {
    // Trigger phone animation after a brief delay
    setTimeout(() => {
      phoneAnimated = true;
    }, 25); // ms delay from dom rendering
  });
</script>

<svelte:head>
  <title>{PUBLIC_APP_NAME}</title>
  <meta name="description" content="The easiest nutrition tracker ever. Press record, speak naturally, get instant nutrition insights with AI-powered voice recognition." />
</svelte:head>

<style>
  /* Page-specific styles only - main component styles moved to separate CSS files */
</style>

<!-- Hero Section -->
<section class="min-h-screen gradient-hero flex items-center py-12">
  <!-- Rising Sun Background -->
  <div class="rising-sun"></div>
  
  <div class="container mx-auto px-4 hero-content">
    <div class="grid lg:grid-cols-2 gap-8 lg:gap-12 items-center max-w-6xl mx-auto">
      <!-- Hero Content -->
      <div class="text-center lg:text-left space-y-4 lg:space-y-6 fade-in-up">
        <h1 class="hero-title text-4xl sm:text-5xl lg:text-6xl xl:text-8xl font-bold leading-tight">
          <span class="gradient-text">Meet <span class="noot-logo">NOOT</span></span>
        </h1>
        <p class="hero-subtitle text-base sm:text-lg lg:text-xl xl:text-2xl text-base-content/80 max-w-lg mx-auto lg:mx-0">
          The easiest nutrition tracker ever. Press record, speak naturally, get instant nutrition insights.
        </p>
        <div class="flex flex-col sm:flex-row gap-3 lg:gap-4 justify-center lg:justify-start">
          <a href="/record" class="btn btn-primary btn-lg text-base lg:text-lg px-6 lg:px-8 btn-mobile">
            <svg class="w-5 h-5 lg:w-6 lg:h-6 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
                    d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
            </svg>
            Try {PUBLIC_APP_NAME} Now
          </a>
          <a href="#how-it-works" class="btn btn-outline btn-lg text-base lg:text-lg px-6 lg:px-8 learn-how-btn btn-mobile">
            Learn How
          </a>
        </div>
      </div>
      
      <!-- Phone Mockup -->
      <div class="flex justify-center lg:justify-end">
        <div class="mockup-phone phone-border {phoneAnimated ? 'phone-float' : ''}">
          <div class="mockup-phone-camera"></div>
          <div class="mockup-phone-display bg-gradient-to-br from-base-100 to-base-200">
            <!-- Simplified Record Screen Interface -->
            <div class="phone-display-content flex flex-col items-center justify-center h-full p-6 lg:p-8" style="gap: 1.5rem;">
              <!-- Large Record Button -->
              <div class="flex justify-center">
                <div class="phone-record-button w-24 h-24 lg:w-32 lg:h-32 rounded-full bg-primary flex items-center justify-center pulse-record shadow-2xl">
                  <svg class="phone-record-icon w-12 h-12 lg:w-16 lg:h-16 text-primary-content pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" 
                          d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
                  </svg>
                </div>
              </div>
              
              <!-- Instruction Text -->
              <div class="phone-instruction-text text-center space-y-1 lg:space-y-2">
                <p class="phone-instruction-title text-base lg:text-lg font-semibold text-base-content">Tap to record your meal</p>
                <p class="phone-instruction-subtitle text-xs lg:text-sm text-base-content/70">Speak naturally about what you ate</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</section>

<!-- Key Features Section -->
<section class="py-12 lg:py-20 bg-base-100">
  <div class="container mx-auto px-4 max-w-6xl">
    <div class="text-center mb-12 lg:mb-16">
      <h2 class="text-2xl sm:text-3xl lg:text-4xl font-bold text-primary mb-4">Why {PUBLIC_APP_NAME}?</h2>
      <p class="text-base lg:text-lg text-base-content/70 max-w-2xl mx-auto">
        Traditional nutrition apps are clunky and time-consuming. {PUBLIC_APP_NAME} makes tracking effortless.
      </p>
    </div>
    
    <div class="grid sm:grid-cols-2 lg:grid-cols-4 gap-6 lg:gap-8">
      <!-- Voice-First Interface -->
      <div class="card bg-base-200 shadow-xl hover:shadow-2xl transition-shadow">
        <div class="card-body text-center p-4 lg:p-6">
          <div class="flex justify-center mb-3 lg:mb-4">
            <svg class="w-12 h-12 lg:w-16 lg:h-16 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
            </svg>
          </div>
          <h3 class="card-title justify-center text-base lg:text-lg">Voice-First Interface</h3>
          <p class="text-sm">No typing, no searching databases. Just speak naturally about your meals.</p>
        </div>
      </div>
      
      <!-- AI-Powered Analysis -->
      <div class="card bg-base-200 shadow-xl hover:shadow-2xl transition-shadow">
        <div class="card-body text-center p-4 lg:p-6">
          <div class="flex justify-center mb-3 lg:mb-4">
            <svg class="w-12 h-12 lg:w-16 lg:h-16 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 00-2.456 2.456zM16.894 20.567L16.5 21.75l-.394-1.183a2.25 2.25 0 00-1.423-1.423L13.5 18.75l1.183-.394a2.25 2.25 0 001.423-1.423L16.5 15.75l.394 1.183a2.25 2.25 0 001.423 1.423l1.183.394-1.183.394a2.25 2.25 0 00-1.423 1.423z" />
            </svg>
          </div>
          <h3 class="card-title justify-center text-base lg:text-lg">AI-Powered Analysis</h3>
          <p class="text-sm">Advanced AI understands brands, portions, and preparation methods to give you accurate nutrition data.</p>
        </div>
      </div>
      
      <!-- Complete Nutrition Tracking -->
      <div class="card bg-base-200 shadow-xl hover:shadow-2xl transition-shadow">
        <div class="card-body text-center p-4 lg:p-6">
          <div class="flex justify-center mb-3 lg:mb-4">
            <svg class="w-12 h-12 lg:w-16 lg:h-16 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z" />
            </svg>
          </div>
          <h3 class="card-title justify-center text-base lg:text-lg">Complete Nutrition Tracking</h3>
          <p class="text-sm">Track 40+ nutrients including vitamins, minerals, and macros. See daily goals and trends.</p>
        </div>
      </div>
      
      <!-- Lightning Fast -->
      <div class="card bg-base-200 shadow-xl hover:shadow-2xl transition-shadow">
        <div class="card-body text-center p-4 lg:p-6">
          <div class="flex justify-center mb-3 lg:mb-4">
            <svg class="w-12 h-12 lg:w-16 lg:h-16 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z" />
            </svg>
          </div>
          <h3 class="card-title justify-center text-base lg:text-lg">Lightning Fast</h3>
          <p class="text-sm">Get results in seconds. From voice to full nutrition breakdown instantly.</p>
        </div>
      </div>
    </div>
  </div>
</section>

<!-- How It Works Section -->
<section id="how-it-works" class="py-12 lg:py-20 bg-base-200">
  <div class="container mx-auto px-4 max-w-4xl">
    <div class="text-center mb-12 lg:mb-16">
      <h2 class="text-2xl sm:text-3xl lg:text-4xl font-bold text-primary mb-4">How It Works</h2>
      <p class="text-base lg:text-lg text-base-content/70">
        Three simple steps to track your nutrition like never before
      </p>
    </div>
    
    <div class="grid sm:grid-cols-3 gap-8 lg:gap-8">
      <!-- Step 1 -->
      <div class="text-center space-y-3 lg:space-y-4">
        <div class="w-12 h-12 lg:w-16 lg:h-16 bg-primary text-primary-content rounded-full flex items-center justify-center text-xl lg:text-2xl font-bold mx-auto">
          1
        </div>
        <h3 class="text-lg lg:text-xl font-semibold">Tap the microphone button</h3>
        <p class="text-sm lg:text-base text-base-content/70">Simple one-tap interface to start recording your meal</p>
      </div>
      
      <!-- Step 2 -->
      <div class="text-center space-y-3 lg:space-y-4">
        <div class="w-12 h-12 lg:w-16 lg:h-16 bg-secondary text-secondary-content rounded-full flex items-center justify-center text-xl lg:text-2xl font-bold mx-auto">
          2
        </div>
        <h3 class="text-lg lg:text-xl font-semibold">Say what you ate</h3>
        <p class="text-sm lg:text-base text-base-content/70">Speak naturally: "I had a latte with oat milk and a banana"</p>
      </div>
      
      <!-- Step 3 -->
      <div class="text-center space-y-3 lg:space-y-4">
        <div class="w-12 h-12 lg:w-16 lg:h-16 bg-accent text-accent-content rounded-full flex items-center justify-center text-xl lg:text-2xl font-bold mx-auto">
          3
        </div>
        <h3 class="text-lg lg:text-xl font-semibold">Get instant nutrition analysis</h3>
        <p class="text-sm lg:text-base text-base-content/70">Complete breakdown with calories, protein, vitamins, and more</p>
      </div>
    </div>
  </div>
</section>

<!-- Pricing/CTA Section -->
<section id="pricing" class="py-12 lg:py-20 bg-primary/5">
  <div class="container mx-auto px-4 max-w-4xl text-center">
    <div class="space-y-6 lg:space-y-8">
      <div>
        <h2 class="text-2xl sm:text-3xl lg:text-4xl font-bold text-primary mb-4">Start Tracking Smarter Today</h2>
        <p class="text-base lg:text-xl text-base-content/80 max-w-2xl mx-auto">
          Join a community of people that care about quality nutrition.
        </p>
      </div>
      
      <div class="card bg-base-100 shadow-xl max-w-md mx-auto">
        <div class="card-body text-center p-4 lg:p-6">
          <h3 class="text-xl lg:text-2xl font-bold text-primary">Free to Start</h3>
          <div class="divider"></div>
          <ul class="text-left space-y-2">
            <li class="flex items-center">
              <svg class="w-4 h-4 lg:w-5 lg:h-5 text-success mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
              </svg>
              <span class="text-sm lg:text-base">Unlimited voice recordings</span>
            </li>
            <li class="flex items-center">
              <svg class="w-4 h-4 lg:w-5 lg:h-5 text-success mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
              </svg>
              <span class="text-sm lg:text-base">Analyzes 48 nutrients</span>
            </li>
            <li class="flex items-center">
              <svg class="w-4 h-4 lg:w-5 lg:h-5 text-success mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
              </svg>
              <span class="text-sm lg:text-base">Daily nutrition insights</span>
            </li>
            <li class="flex items-center">
              <svg class="w-4 h-4 lg:w-5 lg:h-5 text-success mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
              </svg>
              <span class="text-sm lg:text-base">No credit card required</span>
            </li>
          </ul>
        </div>
      </div>
      
      <a href="/record" class="btn btn-primary btn-lg text-base lg:text-lg px-8 lg:px-12">
        <svg class="w-5 h-5 lg:w-6 lg:h-6 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
                d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
        </svg>
        Get Started Now
      </a>
    </div>
  </div>
</section>

<!-- Footer Section -->
<footer class="bg-base-200 text-base-content">
  <div class="container mx-auto px-4 py-8 lg:py-12 max-w-6xl">
    <!-- Logo -->
    <div class="flex justify-center mb-8 lg:mb-10">
      <span class="text-4xl lg:text-6xl noot-logo" style="font-weight: 800;">NOOT</span>
    </div>
    
    <!-- Navigation Links -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-6 lg:gap-8 mb-6 lg:mb-8">
      <div class="text-center space-y-2 lg:space-y-3">
        <h6 class="font-semibold text-sm lg:text-base text-base-content mb-2 lg:mb-3">Product</h6>
        <div class="space-y-1 lg:space-y-2">
          <a href="#how-it-works" class="block text-xs lg:text-sm text-base-content/70 hover:text-base-content transition-colors">How it Works</a>
          <a href="/record" class="block text-xs lg:text-sm text-base-content/70 hover:text-base-content transition-colors">Features</a>
          <a href="#pricing" class="block text-xs lg:text-sm text-base-content/70 hover:text-base-content transition-colors">Pricing</a>
        </div>
      </div>
      
      <div class="text-center space-y-2 lg:space-y-3">
        <h6 class="font-semibold text-sm lg:text-base text-base-content mb-2 lg:mb-3">Company</h6>
        <div class="space-y-1 lg:space-y-2">
          <span class="block text-xs lg:text-sm text-base-content/40 cursor-not-allowed">About</span>
          <span class="block text-xs lg:text-sm text-base-content/40 cursor-not-allowed">Blog</span>
          <span class="block text-xs lg:text-sm text-base-content/40 cursor-not-allowed">Contact</span>
        </div>
      </div>
      
      <div class="text-center space-y-2 lg:space-y-3">
        <h6 class="font-semibold text-sm lg:text-base text-base-content mb-2 lg:mb-3">Legal</h6>
        <div class="space-y-1 lg:space-y-2">
          <span class="block text-xs lg:text-sm text-base-content/40 cursor-not-allowed">Privacy Policy</span>
          <span class="block text-xs lg:text-sm text-base-content/40 cursor-not-allowed">Terms of Service</span>
          <span class="block text-xs lg:text-sm text-base-content/40 cursor-not-allowed">Cookie Policy</span>
        </div>
      </div>
      
      <div class="text-center space-y-2 lg:space-y-3">
        <h6 class="font-semibold text-sm lg:text-base text-base-content mb-2 lg:mb-3">Social</h6>
        <div class="space-y-1 lg:space-y-2">
          <a href="https://github.com/GrantBirki/noot" class="block text-xs lg:text-sm text-base-content/70 hover:text-base-content transition-colors" target="_blank" rel="noopener noreferrer">GitHub</a>
          <span class="block text-xs lg:text-sm text-base-content/40 cursor-not-allowed">Twitter</span>
          <span class="block text-xs lg:text-sm text-base-content/40 cursor-not-allowed">Discord</span>
        </div>
      </div>
    </div>
    
    <!-- Divider -->
    <div class="border-t border-base-300 pt-4 lg:pt-6">
      <!-- Copyright -->
      <div class="text-center mb-2">
        <p class="text-xs lg:text-sm text-base-content/70">Copyright © {new Date().getFullYear()} - All rights reserved by {PUBLIC_APP_NAME}</p>
      </div>
      
      <!-- Made with love -->
      <div class="text-center">
        <p class="text-xs text-base-content/60 flex items-center justify-center gap-1">
          Made with 
          <svg class="w-3 h-3 text-red-500" fill="currentColor" viewBox="0 0 24 24">
            <path d="M11.645 20.91l-.007-.003-.022-.012a15.247 15.247 0 01-.383-.218 25.18 25.18 0 01-4.244-3.17C4.688 15.36 2.25 12.174 2.25 8.25 2.25 5.322 4.714 3 7.688 3A5.5 5.5 0 0112 5.052 5.5 5.5 0 0116.313 3c2.973 0 5.437 2.322 5.437 5.25 0 3.925-2.438 7.111-4.739 9.256a25.175 25.175 0 01-4.244 3.17 15.247 15.247 0 01-.383.219l-.022.012-.007.004-.003.001a.752.752 0 01-.704 0l-.003-.001z" />
          </svg>
          in San Francisco, CA
        </p>
      </div>
    </div>
  </div>
</footer>
