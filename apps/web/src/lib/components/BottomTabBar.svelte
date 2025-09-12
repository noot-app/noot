<script lang="ts">
  import { page } from "$app/stores"
  import { goto } from "$app/navigation"

  import TagIcon from "$lib/components/icons/Tag.svelte"
  import UserIcon from "$lib/components/icons/User.svelte"
  import ChartBarIcon from "$lib/components/icons/ChartBar.svelte"
  import SquaresTwoByTwo from "$lib/components/icons/SquaresTwoByTwo.svelte"
  import MicrophoneIcon from "$lib/components/icons/Microphone.svelte"
  import CalendarIcon from "$lib/components/icons/calendar-days.svelte"
  import TimelineIcon from "$lib/components/icons/Timeline.svelte"
  import CursorArrowRaysIcon from "./icons/CursorArrowRays.svelte"

  // Tab items - using the most important/frequently used pages
  const tabItems = [
    { href: "/record", label: "Record", icon: "microphone" },
    { href: "/summary", label: "Summary", icon: "chart-bar" },
    { href: "/log", label: "Log", icon: "timeline" },
    { href: "/dashboard", label: "More", icon: "dashboard" },
    { href: "/profile", label: "Profile", icon: "user" },
  ]

  $: currentPath = $page.url.pathname

  function handleTabClick(href: string) {
    goto(href)
  }

  function getIconComponent(iconName: string) {
    switch (iconName) {
      case "microphone":
        return MicrophoneIcon
      case "chart-bar":
        return ChartBarIcon
      case "dashboard":
        return SquaresTwoByTwo
      case "timeline":
        return TimelineIcon
      case "calendar":
        return CalendarIcon
      case "tag":
        return TagIcon
      case "user":
        return UserIcon
      case "cursor-arrow-rays":
        return CursorArrowRaysIcon
      default:
        return MicrophoneIcon
    }
  }
</script>

<div class="bottom-tab-bar safe-area-bottom safe-area-x">
  {#each tabItems as item}
    {@const isActive = currentPath === item.href}
    <button
      class="tab-button"
      class:active={isActive}
      on:click={() => handleTabClick(item.href)}
      aria-label={item.label}
    >
      <div class="tab-icon">
        <svelte:component 
          this={getIconComponent(item.icon)} 
          className="w-6 h-6" 
        />
      </div>
      {#if isActive}
        <div class="active-indicator"></div>
      {/if}
    </button>
  {/each}
</div>

<style>
  .bottom-tab-bar {
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    background: var(--color-base-100);
    border-top: 1px solid var(--color-base-300);
    display: flex;
    justify-content: space-around;
    align-items: center;
    /* Use shared variable and cap with safe-area */
    height: calc(var(--bottom-tabs-height) + env(safe-area-inset-bottom, 0px));
    z-index: 50;
    box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.1);
    padding-bottom: env(safe-area-inset-bottom, 0px);
  }

  .tab-button {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    flex: 1;
    height: 100%;
    background: transparent;
    border: none;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .tab-button:hover {
    background: var(--color-base-200);
  }

  .tab-button:active {
    transform: scale(0.95);
  }

  .tab-icon {
    color: var(--color-base-content-lighter);
    transition: color 0.2s ease, transform 0.2s ease;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .tab-button.active .tab-icon {
    color: var(--color-primary);
    transform: scale(1.1);
  }

  .active-indicator {
    position: absolute;
    top: 6px;
    left: 50%;
    transform: translateX(-50%);
    width: 6px;
    height: 6px;
    background: var(--color-primary);
    border-radius: 50%;
  }

  /* Ensure content doesn't get hidden behind the tab bar */
  :global(body.has-bottom-tabs) {
    padding-bottom: var(--bottom-tabs-height);
  }

  /* Remove top padding from main content on mobile since there's no navbar */
  :global(body.has-bottom-tabs main) {
    padding-top: 0;
  }
</style>
