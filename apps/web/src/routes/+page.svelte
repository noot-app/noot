<script lang="ts">
  import { PUBLIC_APP_NAME } from "$env/static/public";
  import { onMount } from "svelte";
  
  let phoneAnimated = false;
  
  onMount(() => {
    // Trigger phone animation after a brief delay
    setTimeout(() => {
      phoneAnimated = true;
    }, 500);
  });
</script>

<svelte:head>
  <title>{PUBLIC_APP_NAME}</title>
  <meta name="description" content="The easiest nutrition tracker ever. Press record, speak naturally, get instant nutrition insights with AI-powered voice recognition." />
</svelte:head>

<style>
  /* Ensure navbar stays above background effects */
  :global(nav, .navbar, header, .navbar-start, .navbar-center, .navbar-end) {
    position: relative;
    z-index: 1000 !important;
  }
  
  :global(html, body) {
    overflow-x: hidden;
  }
  
  .gradient-hero {
    background: 
      /* Fill layer that extends sunset colors upward */
      linear-gradient(to top, #d35400 0%, #e67e22 15%, #f39c12 35%, #f5c28b 50%, hsl(var(--b1)) 70%),
      linear-gradient(135deg, hsl(var(--b1)), hsl(var(--b2)));
    background-size: 100% 120%, 100% 100%;
    position: relative;
    overflow: visible;
  }
  
  .gradient-hero::after {
    content: "";
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 20vh;
    background: linear-gradient(to bottom, 
      transparent 0%, 
      rgba(255, 255, 255, 0.3) 50%, 
      rgba(255, 255, 255, 0.8) 90%, 
      hsl(var(--b1)) 100%);
    z-index: 3;
    pointer-events: none;
  }
  
  .hero-content {
    position: relative;
    z-index: 2;
  }
  
  /* Rising sun from horizon effect */
  .rising-sun {
    position: absolute;
    bottom: -10vh;
    left: 0;
    right: 0;
    height: 150vh;
    border-radius: 0%;
    background: 
      radial-gradient(ellipse 60vw 50vh at center bottom, #ff3000 0%, #ff4500 10%, #ff6b35 20%, #ff8c42 35%, transparent 70%),
      radial-gradient(ellipse 50vw 45vh at center bottom, #d2691e 0%, #e67e22 5%, #f39c12 10%, transparent 60%),
      radial-gradient(ellipse 40vw 40vh at center bottom, #b22222 0%, #dc143c 3%, #e74c3c 5%, transparent 50%),
      radial-gradient(ellipse 80vw 65vh at center bottom, 
        #ff3000 0%,
        #ff4500 10%, 
        #ff6b35 15%, 
        #f39c12 25%, 
        #e67e22 40%, 
        #d35400 60%, 
        transparent 90%);
    background-blend-mode: screen, multiply, overlay, normal;
    opacity: 0.6;
    z-index: -10;
    mask: linear-gradient(to bottom, 
      black 0%, 
      black 70%, 
      transparent 95%);
    -webkit-mask: linear-gradient(to bottom, 
      black 0%, 
      black 70%, 
      transparent 95%);
    pointer-events: none;
  }
  
  .rising-sun::after {
    content: "";
    position: absolute;
    inset: 0;
    border-radius: inherit;
    background-image: 
      url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='80' height='80'><filter id='coarseGrain'><feTurbulence type='fractalNoise' baseFrequency='0.8' numOctaves='2' stitchTiles='stitch'/></filter><rect width='100%' height='100%' filter='url(%23coarseGrain)' opacity='1'/></svg>"),
      url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='120' height='120'><filter id='medGrain'><feTurbulence type='turbulence' baseFrequency='1.2' numOctaves='3' seed='7'/></filter><rect width='100%' height='100%' filter='url(%23medGrain)' opacity='0.9'/></svg>"),
      url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='60' height='60'><filter id='heavyGrain'><feTurbulence type='fractalNoise' baseFrequency='0.5' numOctaves='1' seed='3'/></filter><rect width='100%' height='100%' filter='url(%23heavyGrain)' opacity='1'/></svg>");
    background-size: 80px 80px, 120px 120px, 60px 60px;
    opacity: 0.95;
    mix-blend-mode: hard-light;
    pointer-events: none;
  }

  
  @keyframes grain {
    0%, 100% { transform: translate(0, 0); }
    25% { transform: translate(-3px, -3px); }
    50% { transform: translate(3px, -3px); }
    75% { transform: translate(-3px, 3px); }
  }
  
  .phone-float {
    animation: float 3s ease-in-out infinite;
  }
  
  @keyframes float {
    0%, 100% { transform: translateY(0px) rotate(0deg); }
    50% { transform: translateY(-10px) rotate(1deg); }
  }
  
  .pulse-record {
    animation: pulse-gentle 2s ease-in-out infinite;
  }
  
  @keyframes pulse-gentle {
    0%, 100% { transform: scale(1); opacity: 1; }
    50% { transform: scale(1.05); opacity: 0.9; }
  }
  
  .fade-in-up {
    animation: fadeInUp 0.8s ease-out;
  }
  
  @keyframes fadeInUp {
    from {
      opacity: 0;
      transform: translateY(30px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
  
  .phone-border {
    border-color: oklch(32.428% 0.0141 285.558);
  }
  
  .learn-how-btn {
    transition: all 0.3s ease;
  }
  
  .learn-how-btn:hover {
    background-color: hsl(var(--p) / 0.1);
    border-color: hsl(var(--p) / 0.6);
    color: hsl(var(--p));
    transform: translateY(-1px);
  }
  
  .gradient-text {
    background: linear-gradient(135deg, #e89852 0%, #cc902f 20%, #d38c46 40%, #cf6c26 60%, #c57553 80%, #b16645 100%);
    background-clip: text;
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }
</style>

<!-- Hero Section -->
<section class="min-h-screen gradient-hero flex items-center py-12">
  <!-- Rising Sun Background -->
  <div class="rising-sun"></div>
  
  <div class="container mx-auto px-4 hero-content">
    <div class="grid lg:grid-cols-2 gap-12 items-center max-w-6xl mx-auto">
      <!-- Hero Content -->
      <div class="text-center lg:text-left space-y-6 fade-in-up">
        <h1 class="text-6xl lg:text-8xl font-bold leading-tight gradient-text">
          Meet Noot
        </h1>
        <p class="text-xl lg:text-2xl text-base-content/80 max-w-lg mx-auto lg:mx-0">
          The easiest nutrition tracker ever. Press record, speak naturally, get instant nutrition insights.
        </p>
        <div class="flex flex-col sm:flex-row gap-4 justify-center lg:justify-start">
          <a href="/record" class="btn btn-primary btn-lg text-lg px-8">
            <svg class="w-6 h-6 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
                    d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
            </svg>
            Try {PUBLIC_APP_NAME} Now
          </a>
          <a href="#how-it-works" class="btn btn-outline btn-lg text-lg px-8 learn-how-btn">
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
            <div class="flex flex-col items-center justify-center h-full space-y-6 p-8">
              <!-- Large Record Button -->
              <div class="flex justify-center">
                <div class="w-32 h-32 rounded-full bg-primary flex items-center justify-center pulse-record shadow-2xl">
                  <svg class="w-16 h-16 text-primary-content pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" 
                          d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
                  </svg>
                </div>
              </div>
              
              <!-- Instruction Text -->
              <div class="text-center space-y-2">
                <p class="text-lg font-semibold text-base-content">Tap to record your meal</p>
                <p class="text-sm text-base-content/70">Speak naturally about what you ate</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</section>

<!-- Key Features Section -->
<section class="py-20 bg-base-100">
  <div class="container mx-auto px-4 max-w-6xl">
    <div class="text-center mb-16">
      <h2 class="text-4xl font-bold text-primary mb-4">Why {PUBLIC_APP_NAME}?</h2>
      <p class="text-lg text-base-content/70 max-w-2xl mx-auto">
        Traditional nutrition apps are clunky and time-consuming. {PUBLIC_APP_NAME} makes tracking effortless.
      </p>
    </div>
    
    <div class="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
      <!-- Voice-First Interface -->
      <div class="card bg-base-200 shadow-xl hover:shadow-2xl transition-shadow">
        <div class="card-body text-center">
          <div class="text-4xl mb-4">🎙️</div>
          <h3 class="card-title justify-center text-lg">Voice-First Interface</h3>
          <p class="text-sm">No typing, no searching databases. Just speak naturally about your meals.</p>
        </div>
      </div>
      
      <!-- AI-Powered Analysis -->
      <div class="card bg-base-200 shadow-xl hover:shadow-2xl transition-shadow">
        <div class="card-body text-center">
          <div class="text-4xl mb-4">🤖</div>
          <h3 class="card-title justify-center text-lg">AI-Powered Analysis</h3>
          <p class="text-sm">Advanced AI understands brands, portions, and preparation methods to give you accurate nutrition data.</p>
        </div>
      </div>
      
      <!-- Complete Nutrition Tracking -->
      <div class="card bg-base-200 shadow-xl hover:shadow-2xl transition-shadow">
        <div class="card-body text-center">
          <div class="text-4xl mb-4">📊</div>
          <h3 class="card-title justify-center text-lg">Complete Nutrition Tracking</h3>
          <p class="text-sm">Track 40+ nutrients including vitamins, minerals, and macros. See daily goals and trends.</p>
        </div>
      </div>
      
      <!-- Lightning Fast -->
      <div class="card bg-base-200 shadow-xl hover:shadow-2xl transition-shadow">
        <div class="card-body text-center">
          <div class="text-4xl mb-4">⚡</div>
          <h3 class="card-title justify-center text-lg">Lightning Fast</h3>
          <p class="text-sm">Get results in seconds. From voice to full nutrition breakdown instantly.</p>
        </div>
      </div>
    </div>
  </div>
</section>

<!-- How It Works Section -->
<section id="how-it-works" class="py-20 bg-base-200">
  <div class="container mx-auto px-4 max-w-4xl">
    <div class="text-center mb-16">
      <h2 class="text-4xl font-bold text-primary mb-4">How It Works</h2>
      <p class="text-lg text-base-content/70">
        Three simple steps to track your nutrition like never before
      </p>
    </div>
    
    <div class="grid md:grid-cols-3 gap-8">
      <!-- Step 1 -->
      <div class="text-center space-y-4">
        <div class="w-16 h-16 bg-primary text-primary-content rounded-full flex items-center justify-center text-2xl font-bold mx-auto">
          1
        </div>
        <h3 class="text-xl font-semibold">Tap the microphone button</h3>
        <p class="text-base-content/70">Simple one-tap interface to start recording your meal</p>
      </div>
      
      <!-- Step 2 -->
      <div class="text-center space-y-4">
        <div class="w-16 h-16 bg-secondary text-secondary-content rounded-full flex items-center justify-center text-2xl font-bold mx-auto">
          2
        </div>
        <h3 class="text-xl font-semibold">Say what you ate</h3>
        <p class="text-base-content/70">Speak naturally: "I had a latte with oat milk and a banana"</p>
      </div>
      
      <!-- Step 3 -->
      <div class="text-center space-y-4">
        <div class="w-16 h-16 bg-accent text-accent-content rounded-full flex items-center justify-center text-2xl font-bold mx-auto">
          3
        </div>
        <h3 class="text-xl font-semibold">Get instant nutrition analysis</h3>
        <p class="text-base-content/70">Complete breakdown with calories, protein, vitamins, and more</p>
      </div>
    </div>
  </div>
</section>

<!-- Pricing/CTA Section -->
<section id="pricing" class="py-20 bg-primary/5">
  <div class="container mx-auto px-4 max-w-4xl text-center">
    <div class="space-y-8">
      <div>
        <h2 class="text-4xl font-bold text-primary mb-4">Start Tracking Smarter Today</h2>
        <p class="text-xl text-base-content/80 max-w-2xl mx-auto">
          Join thousands of users who have simplified their nutrition tracking with voice-first AI technology.
        </p>
      </div>
      
      <div class="card bg-base-100 shadow-xl max-w-md mx-auto">
        <div class="card-body text-center">
          <h3 class="text-2xl font-bold text-primary">Free to Start</h3>
          <div class="divider"></div>
          <ul class="text-left space-y-2">
            <li class="flex items-center">
              <svg class="w-5 h-5 text-success mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
              </svg>
              Unlimited voice recordings
            </li>
            <li class="flex items-center">
              <svg class="w-5 h-5 text-success mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
              </svg>
              Full 40+ nutrient analysis
            </li>
            <li class="flex items-center">
              <svg class="w-5 h-5 text-success mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
              </svg>
              Daily nutrition insights
            </li>
            <li class="flex items-center">
              <svg class="w-5 h-5 text-success mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
              </svg>
              No credit card required
            </li>
          </ul>
        </div>
      </div>
      
      <a href="/record" class="btn btn-primary btn-lg text-lg px-12">
        <svg class="w-6 h-6 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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
  <div class="container mx-auto px-4 py-12 max-w-6xl">
    <!-- Logo -->
    <div class="flex justify-center mb-10">
      <img src="/images/noot.svg" alt="{PUBLIC_APP_NAME} Logo" class="w-32 h-32" />
    </div>
    
    <!-- Navigation Links -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-8 mb-8">
      <div class="text-center space-y-3">
        <h6 class="font-semibold text-base-content mb-3">Product</h6>
        <div class="space-y-2">
          <a href="#how-it-works" class="block text-base-content/70 hover:text-base-content transition-colors">How it Works</a>
          <a href="/record" class="block text-base-content/70 hover:text-base-content transition-colors">Features</a>
          <a href="#pricing" class="block text-base-content/70 hover:text-base-content transition-colors">Pricing</a>
        </div>
      </div>
      
      <div class="text-center space-y-3">
        <h6 class="font-semibold text-base-content mb-3">Company</h6>
        <div class="space-y-2">
          <span class="block text-base-content/40 cursor-not-allowed">About</span>
          <span class="block text-base-content/40 cursor-not-allowed">Blog</span>
          <span class="block text-base-content/40 cursor-not-allowed">Contact</span>
        </div>
      </div>
      
      <div class="text-center space-y-3">
        <h6 class="font-semibold text-base-content mb-3">Legal</h6>
        <div class="space-y-2">
          <span class="block text-base-content/40 cursor-not-allowed">Privacy Policy</span>
          <span class="block text-base-content/40 cursor-not-allowed">Terms of Service</span>
          <span class="block text-base-content/40 cursor-not-allowed">Cookie Policy</span>
        </div>
      </div>
      
      <div class="text-center space-y-3">
        <h6 class="font-semibold text-base-content mb-3">Social</h6>
        <div class="space-y-2">
          <a href="https://github.com/GrantBirki/noot" class="block text-base-content/70 hover:text-base-content transition-colors" target="_blank" rel="noopener noreferrer">GitHub</a>
          <span class="block text-base-content/40 cursor-not-allowed">Twitter</span>
          <span class="block text-base-content/40 cursor-not-allowed">Discord</span>
        </div>
      </div>
    </div>
    
    <!-- Divider -->
    <div class="border-t border-base-300 pt-6">
      <!-- Copyright -->
      <div class="text-center mb-2">
        <p class="text-sm text-base-content/70">Copyright © {new Date().getFullYear()} - All rights reserved by {PUBLIC_APP_NAME}</p>
      </div>
      
      <!-- Made with love -->
      <div class="text-center">
        <p class="text-xs text-base-content/60">Made with ❤️ in San Francisco, CA</p>
      </div>
    </div>
  </div>
</footer>
