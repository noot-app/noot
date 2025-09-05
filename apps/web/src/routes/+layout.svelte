<script lang="ts">
  import "../app.css"
  import { navigating } from "$app/stores"
  import { expoOut } from "svelte/easing"
  import { slide } from "svelte/transition"
  import { onMount } from "svelte"
  import { units } from "$lib/stores/units"
  import { initAuth } from "$lib/auth/store"
  import Navbar from "$lib/components/Navbar.svelte"
  import DevBanner from "$lib/components/DevBanner.svelte"
  import AdaptiveFavicon from "$lib/components/AdaptiveFavicon.svelte"

  interface Props {
    children?: import("svelte").Snippet
    data?: any
  }

  let { children, data }: Props = $props()

  // Navigation timeout handling
  let navigationTimeout: NodeJS.Timeout | null = null
  let isNavigationStuck = $state(false)

  // Watch for navigation state changes using $effect
  $effect(() => {
    if ($navigating) {
      // Clear any existing timeout
      if (navigationTimeout) {
        clearTimeout(navigationTimeout)
      }
      
      // Set a timeout to detect stuck navigation
      navigationTimeout = setTimeout(() => {
        if ($navigating) {
          console.warn("Navigation appears to be stuck, this may indicate a routing issue")
          isNavigationStuck = true
          
          // Provide user feedback after 15 seconds
          setTimeout(() => {
            if ($navigating && isNavigationStuck) {
              console.error("Navigation timeout - please try refreshing the page")
            }
          }, 15000)
        }
      }, 10000) // 10 second timeout
    } else {
      // Navigation completed, clear timeout and reset stuck state
      if (navigationTimeout) {
        clearTimeout(navigationTimeout)
        navigationTimeout = null
      }
      isNavigationStuck = false
    }
  })

  onMount(() => {
    units.init()

    // Initialize auth with SSR session data
    initAuth(data?.session)

    // Cleanup timeout on unmount
    return () => {
      if (navigationTimeout) {
        clearTimeout(navigationTimeout)
      }
    }
  })
</script>

<!-- Adaptive favicon management -->
<AdaptiveFavicon />

<DevBanner />

{#if $navigating}
  <!-- 
    Loading animation for next page since svelte doesn't show any indicator. 
     - delay 100ms because most page loads are instant, and we don't want to flash 
     - long 12s duration because we don't actually know how long it will take
     - exponential easing so fast loads (>100ms and <1s) still see enough progress,
       while slow networks see it moving for a full 12 seconds
  -->
  <div
    class="fixed w-full top-0 right-0 left-0 h-1 z-50 bg-primary"
    class:bg-warning={isNavigationStuck}
    in:slide={{ delay: 100, duration: 12000, axis: "x", easing: expoOut }}
  ></div>
  
  {#if isNavigationStuck}
    <!-- Show warning message for stuck navigation -->
    <div class="fixed top-4 left-1/2 transform -translate-x-1/2 z-50">
      <div class="alert alert-warning shadow-lg max-w-md">
        <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-3.08-9.4a9 9 0 1118.16 0" />
        </svg>
        <div>
          <h3 class="font-bold">Navigation taking longer than expected</h3>
          <div class="text-xs">Please wait or try refreshing the page</div>
        </div>
      </div>
    </div>
  {/if}
{/if}

<Navbar />
<main>
  {@render children?.()}
</main>
