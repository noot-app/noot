<script lang="ts">
  import { browser } from '$app/environment';
  import { onMount } from 'svelte';

  // Props for customization
  export let lightLogo = '/images/logo-light.png';
  export let darkLogo = '/images/logo-dark.png';
  export let lightThemeColor = '#be7454';
  export let darkThemeColor = '#3a3b40';

  onMount(() => {
    if (!browser) return;

    // Function to update favicon based on color scheme
    function updateFavicon(isDark: boolean) {
      // Update PNG favicon
      const faviconLink = document.querySelector('link[rel="icon"][type="image/png"]') as HTMLLinkElement;
      if (faviconLink) {
        faviconLink.href = isDark ? darkLogo : lightLogo;
      }

      // Update theme color
      const lightThemeMeta = document.querySelector('meta[name="theme-color"][media*="light"]') as HTMLMetaElement;
      const darkThemeMeta = document.querySelector('meta[name="theme-color"][media*="dark"]') as HTMLMetaElement;
      
      if (lightThemeMeta) lightThemeMeta.content = lightThemeColor;
      if (darkThemeMeta) darkThemeMeta.content = darkThemeColor;
    }

    // Check initial color scheme
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    updateFavicon(mediaQuery.matches);

    // Listen for changes
    const handleChange = (e: MediaQueryListEvent) => updateFavicon(e.matches);
    mediaQuery.addEventListener('change', handleChange);

    // Cleanup
    return () => {
      mediaQuery.removeEventListener('change', handleChange);
    };
  });
</script>

<!-- This component doesn't render anything visible -->
