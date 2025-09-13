<script lang="ts">
  import { page } from "$app/stores"
  import { goto } from "$app/navigation"

  import { bottomTabItems, bottomMenuTrigger, drawerNavItems, getIconComponent, type NavItemBase } from "$lib/navigation/navigation"

  const tabItems: NavItemBase[] = [...bottomTabItems, bottomMenuTrigger]

  $: currentPath = $page.url.pathname

  // Optimistic highlight to remove perceived delay before route store updates
  let pendingHref: string | null = null

  function handleTabClick(href: string) {
    if (href !== currentPath) {
      pendingHref = href
    }
    goto(href)
  }

  let showDrawer = false

  function toggleDrawer() { showDrawer = !showDrawer }
  function closeDrawer() { showDrawer = false }

  // Clear pending once navigation reflects new path
  $: if (pendingHref && currentPath === pendingHref) {
    pendingHref = null
  }

  // Icon component resolver now centralized in navigation.ts
</script>

<div class="bottom-tab-bar safe-area-bottom safe-area-x">
  {#each tabItems as item}
    {#if item.type === "route"}
      {@const isActive = currentPath === item.href}
      <button
        class="tab-button"
        class:active={isActive}
        class:pending-active={pendingHref === item.href}
        on:click={() => item.href && handleTabClick(item.href)}
        aria-label={item.label}
      >
        <div class="tab-icon">
          <svelte:component this={getIconComponent(item.icon)} className="w-6 h-6" />
        </div>
      </button>
    {:else}
      <button
        class="tab-button"
        on:click={toggleDrawer}
        aria-label="Open navigation menu"
        aria-haspopup="dialog"
        aria-expanded={showDrawer}
      >
        <div class="tab-icon">
          <svelte:component this={getIconComponent(item.icon)} className="w-6 h-6" />
        </div>
      </button>
    {/if}
  {/each}
</div>

{#if showDrawer}
  <div class="drawer-overlay" role="presentation" on:click={closeDrawer}></div>
  <div class="drawer-panel safe-area-bottom" role="dialog" aria-label="Navigation Menu">
    <div class="drawer-handle"></div>
    <nav class="drawer-nav" aria-label="Main navigation">
      <ul>
        {#each drawerNavItems as item}
          <li>
            <a href={item.href} on:click={closeDrawer} class:active={currentPath === item.href}>
              <span class="nav-icon">
                <svelte:component this={getIconComponent(item.icon)} className="w-5 h-5" />
              </span>
              <span class="nav-text">{item.label}</span>
            </a>
          </li>
        {/each}
      </ul>
    </nav>
    <button class="close-button" on:click={closeDrawer} aria-label="Close menu">Close</button>
  </div>
{/if}

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

  /* Removed hover background & active press scaling per design simplification */

  .tab-icon {
    color: var(--color-base-content-lighter);
    /* Remove transition for instantaneous feedback */
    transition: none;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .tab-button.active .tab-icon,
  .tab-button.pending-active .tab-icon {
    color: var(--color-primary);
  }

  /* Drawer styles */
  .drawer-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.4); backdrop-filter: blur(2px); z-index: 60; }
  .drawer-panel { position: fixed; left:0; right:0; bottom:0; background: var(--color-base-100); border-top-left-radius:18px; border-top-right-radius:18px; padding:0.75rem 1rem 1.25rem; box-shadow:0 -4px 16px rgba(0,0,0,0.15); z-index:70; animation: slide-up 160ms ease-out; }
  .drawer-handle { width:48px; height:5px; background:var(--color-base-300); border-radius:4px; margin:0 auto 0.75rem; }
  .drawer-nav ul { list-style:none; padding:0; margin:0 0 0.75rem; }
  .drawer-nav li + li { margin-top:0.5rem; }
  .drawer-nav a { display:flex; align-items:center; gap:0.65rem; padding:0.65rem 0.75rem; border-radius:10px; font-size:0.95rem; font-weight:500; color:var(--color-base-content); background:var(--color-base-200); position:relative; }
  .drawer-nav a:active { background:var(--color-base-300); }
  .drawer-nav a.active { background:var(--color-primary); color:var(--color-primary-content,#fff); }
  .drawer-nav a.active .nav-icon { color:var(--color-primary-content,#fff); }
  .nav-icon { display:flex; align-items:center; justify-content:center; color:var(--color-base-content-lighter); }
  .drawer-nav a.active .nav-icon { color:inherit; }
  .close-button { width:100%; background:var(--color-base-200); border:none; padding:0.7rem 0.75rem; border-radius:10px; font-size:0.9rem; font-weight:500; color:var(--color-base-content-lighter); }
  .close-button:active { background:var(--color-base-300); }
  @keyframes slide-up { from { transform:translateY(16px); opacity:0; } to { transform:translateY(0); opacity:1; } }

  /* Ensure content doesn't get hidden behind the tab bar */
  :global(body.has-bottom-tabs) {
    padding-bottom: var(--bottom-tabs-height);
  }

  /* Remove top padding from main content on mobile since there's no navbar */
  :global(body.has-bottom-tabs main) {
    padding-top: 0;
  }
</style>
