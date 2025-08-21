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
  <title>{PUBLIC_APP_NAME} - Just Say What You Ate!</title>
  <meta name="description" content="The easiest nutrition tracker ever. Press record, speak naturally, get instant nutrition insights with AI-powered voice recognition." />
</svelte:head>

<style>
  .gradient-hero {
    background: linear-gradient(135deg, hsl(var(--b1)), hsl(var(--b2)));
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
</style>

<!-- Hero Section -->
<section class="min-h-screen gradient-hero flex items-center py-12">
  <div class="container mx-auto px-4">
    <div class="grid lg:grid-cols-2 gap-12 items-center max-w-6xl mx-auto">
      <!-- Hero Content -->
      <div class="text-center lg:text-left space-y-6 fade-in-up">
        <h1 class="text-5xl lg:text-6xl font-bold text-primary leading-tight">
          Just say what you ate!
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
          <a href="#how-it-works" class="btn btn-outline btn-lg text-lg px-8">
            Learn How
          </a>
        </div>
      </div>
      
      <!-- Phone Mockup -->
      <div class="flex justify-center lg:justify-end">
        <div class="mockup-phone {phoneAnimated ? 'phone-float' : ''}">
          <div class="camera"></div>
          <div class="display">
            <div class="artboard artboard-demo phone-1 bg-gradient-to-br from-base-100 to-base-200">
              <!-- Simplified Record Screen Interface -->
              <div class="flex flex-col items-center justify-center h-full space-y-6 p-8">
                <!-- Large Record Button -->
                <div class="flex justify-center">
                  <div class="w-32 h-32 rounded-full bg-primary flex items-center justify-center cursor-pointer pulse-record shadow-2xl">
                    <svg class="w-16 h-16 text-primary-content" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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

<!-- Social Proof Section -->
<section class="py-20 bg-base-100">
  <div class="container mx-auto px-4 max-w-6xl">
    <div class="text-center mb-16">
      <h2 class="text-4xl font-bold text-primary mb-4">Trusted by Health-Conscious Users</h2>
    </div>
    
    <div class="grid md:grid-cols-2 lg:grid-cols-3 gap-8 mb-12">
      <!-- Stats -->
      <div class="stat bg-base-200 rounded-box text-center">
        <div class="stat-figure text-primary">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="inline-block w-8 h-8 stroke-current">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
          </svg>
        </div>
        <div class="stat-title">Meals Logged</div>
        <div class="stat-value text-primary">10,000+</div>
      </div>
      
      <div class="stat bg-base-200 rounded-box text-center">
        <div class="stat-figure text-secondary">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="inline-block w-8 h-8 stroke-current">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 100 4m0-4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 100 4m0-4v2m0-6V4"></path>
          </svg>
        </div>
        <div class="stat-title">Time Saved Per Meal</div>
        <div class="stat-value text-secondary">5 min</div>
      </div>
      
      <div class="stat bg-base-200 rounded-box text-center">
        <div class="stat-figure text-accent">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="inline-block w-8 h-8 stroke-current">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"></path>
          </svg>
        </div>
        <div class="stat-title">User Satisfaction</div>
        <div class="stat-value text-accent">98%</div>
      </div>
    </div>
    
    <!-- Testimonials -->
    <div class="grid md:grid-cols-2 gap-8">
      <div class="card bg-base-200 shadow-lg">
        <div class="card-body">
          <p class="italic">"Finally, a nutrition tracker that doesn't feel like work. I just speak what I ate and get incredible insights instantly!"</p>
          <div class="flex items-center mt-4">
            <div class="avatar placeholder">
              <div class="bg-primary text-primary-content rounded-full w-10">
                <span class="text-sm">SJ</span>
              </div>
            </div>
            <div class="ml-3">
              <div class="font-semibold">Sarah J.</div>
              <div class="text-sm text-base-content/70">Health Coach</div>
            </div>
          </div>
        </div>
      </div>
      
      <div class="card bg-base-200 shadow-lg">
        <div class="card-body">
          <p class="italic">"The AI accuracy is incredible. It understands portion sizes and even cooking methods. This is the future of nutrition tracking."</p>
          <div class="flex items-center mt-4">
            <div class="avatar placeholder">
              <div class="bg-secondary text-secondary-content rounded-full w-10">
                <span class="text-sm">MR</span>
              </div>
            </div>
            <div class="ml-3">
              <div class="font-semibold">Mike R.</div>
              <div class="text-sm text-base-content/70">Fitness Enthusiast</div>
            </div>
          </div>
        </div>
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
<footer class="footer footer-center bg-base-200 text-base-content p-10">
  <aside class="grid-flow-col items-center">
    <img src="/images/noot.svg" alt="{PUBLIC_APP_NAME} Logo" class="w-12 h-12" />
    <div>
      <p class="font-bold text-lg">{PUBLIC_APP_NAME}</p>
      <p class="text-sm">AI-powered nutrition tracking made simple</p>
    </div>
  </aside>
  <nav class="grid-flow-col gap-4">
    <div class="grid grid-flow-row gap-2">
      <h6 class="footer-title">Product</h6>
      <a href="#how-it-works" class="link link-hover">How it Works</a>
      <a href="/record" class="link link-hover">Features</a>
      <a href="#pricing" class="link link-hover">Pricing</a>
    </div>
    <div class="grid grid-flow-row gap-2">
      <h6 class="footer-title">Company</h6>
      <span class="link link-hover opacity-50 cursor-not-allowed">About</span>
      <span class="link link-hover opacity-50 cursor-not-allowed">Blog</span>
      <span class="link link-hover opacity-50 cursor-not-allowed">Contact</span>
    </div>
    <div class="grid grid-flow-row gap-2">
      <h6 class="footer-title">Legal</h6>
      <span class="link link-hover opacity-50 cursor-not-allowed">Privacy Policy</span>
      <span class="link link-hover opacity-50 cursor-not-allowed">Terms of Service</span>
    </div>
    <div class="grid grid-flow-row gap-2">
      <h6 class="footer-title">Social</h6>
      <a href="https://github.com/GrantBirki/noot" class="link link-hover" target="_blank" rel="noopener noreferrer">GitHub</a>
      <span class="link link-hover opacity-50 cursor-not-allowed">Twitter</span>
    </div>
  </nav>
  <aside>
    <p>Copyright © {new Date().getFullYear()} - All rights reserved by {PUBLIC_APP_NAME}</p>
  </aside>
</footer>