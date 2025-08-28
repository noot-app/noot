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
    data: {
      session: unknown | null
      supabaseEnabled: boolean
      isDevMode: boolean
    }
  }

  let { children, data }: Props = $props()

  // Initialize auth immediately with server session to prevent hydration mismatch
  if (data.session && data.supabaseEnabled) {
    // If we have a server session, initialize the auth store immediately
    // This prevents the flash of login screen on protected routes
    initAuth(data.session)
  } else {
    // No server session, but still initialize auth to check client-side storage
    initAuth()
  }

  onMount(() => {
    units.init();
    
    // Re-initialize auth on mount to ensure client-side state is synchronized
    // This handles cases where client storage might have newer tokens
    if (data.supabaseEnabled && !data.session) {
      initAuth()
    }
  });
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
    in:slide={{ delay: 100, duration: 12000, axis: "x", easing: expoOut }}
  ></div>
{/if}

<Navbar />
<main>
  {@render children?.()}
</main>
