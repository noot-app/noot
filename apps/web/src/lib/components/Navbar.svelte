<script lang="ts">
  import { page } from "$app/stores";
  import { PUBLIC_APP_NAME } from "$env/static/public";
  
  let isMenuOpen = false;

  function toggleMenu() {
    isMenuOpen = !isMenuOpen;
  }

  function closeMenu() {
    isMenuOpen = false;
  }

  // Navigation items
  const navItems = [
    { href: "/record", label: "Record", icon: "M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" },
    { href: "/summary", label: "Summary", icon: "M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" },
    { href: "/profile", label: "Profile", icon: "M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" }
  ];

  $: currentPath = $page.url.pathname;
</script>

<header class="navbar bg-base-200 shadow-lg">
  <div class="navbar-start">
    <!-- App logo and Beta label -->
    <div class="flex items-center gap-3">
      <a href="/" class="flex items-center">
        <img src="/images/noot.svg" alt="{PUBLIC_APP_NAME} Logo" class="w-8 h-8" />
      </a>
      <div class="badge badge-outline text-xs" style="border-color: #bb704f; color: #bb704f;">
        Beta
      </div>
    </div>
  </div>

  <div class="navbar-center">
    <!-- Keep center empty for balance -->
  </div>

  <div class="navbar-end">
    <!-- Mobile hamburger menu -->
    <div class="dropdown lg:hidden">
      <button 
        class="btn btn-square btn-ghost"
        on:click={toggleMenu}
        aria-label="Toggle navigation menu"
      >
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          {#if isMenuOpen}
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          {:else}
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
          {/if}
        </svg>
      </button>
      
      {#if isMenuOpen}
        <ul class="menu menu-sm dropdown-content mt-3 z-[1] p-2 shadow bg-base-100 rounded-box w-52">
          {#each navItems as item}
            <li>
              <a 
                href={item.href} 
                class="flex items-center gap-3"
                class:active={currentPath === item.href}
                on:click={closeMenu}
              >
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={item.icon} />
                </svg>
                {item.label}
              </a>
            </li>
          {/each}
        </ul>
      {/if}
    </div>

    <!-- Desktop navigation -->
    <div class="hidden lg:flex">
      <ul class="menu menu-horizontal px-1">
        {#each navItems as item}
          <li>
            <a 
              href={item.href}
              class="flex items-center gap-2"
              class:active={currentPath === item.href}
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={item.icon} />
              </svg>
              {item.label}
            </a>
          </li>
        {/each}
      </ul>
    </div>
  </div>
</header>

<style>
  .menu li > a.active {
    background-color: oklch(var(--p));
    color: oklch(var(--pc));
  }
  
  /* Ensure mobile menu doesn't get cut off */
  .dropdown:focus-within .dropdown-content {
    display: block;
  }
</style>
