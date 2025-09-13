<script lang="ts">
  import { page } from "$app/stores"

  import { topNavItems, getIconComponent } from "$lib/navigation/navigation"

  $: currentPath = $page.url.pathname
</script>

<header class="navbar bg-base-200 shadow-lg">
  <div class="navbar-start">
    <!-- App logo and Beta label -->
    <div class="flex items-center gap-3">
      <a href="/" class="flex items-center">
        <span class="text-2xl noot-logo">NOOT</span>
      </a>
      <div class="badge badge-outline text-xs beta-badge">Beta</div>
    </div>
  </div>

  <div class="navbar-center">
    <!-- Keep center empty for balance -->
  </div>

  <div class="navbar-end">
    <!-- Mobile hamburger menu using DaisyUI's native dropdown -->
    <div class="dropdown dropdown-end lg:hidden">
      <div
        tabindex="0"
        role="button"
        class="btn btn-square btn-ghost"
        aria-label="Open navigation menu"
      >
        <svg
          class="w-7 h-7"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M4 6h16M4 12h16M4 18h16"
          />
        </svg>
      </div>
      <!-- svelte-ignore a11y-no-noninteractive-tabindex -->
      <ul
        tabindex="0"
        class="menu menu-sm dropdown-content mt-3 z-[1] p-2 shadow bg-base-100 rounded-box w-52"
      >
        {#each topNavItems as item}
          <li>
            <a
              href={item.href}
              class="flex items-center gap-3 py-3 px-4 min-h-[44px]"
              class:active={currentPath === item.href}
            >
              <svelte:component this={getIconComponent(item.icon)} className="w-6 h-6" />
              {item.label}
            </a>
          </li>
        {/each}
      </ul>
    </div>

    <!-- Desktop navigation -->
    <div class="hidden lg:flex">
      <ul class="menu menu-horizontal px-1">
        {#each topNavItems as item}
          <li>
            <a
              href={item.href}
              class="flex items-center gap-2"
              class:active={currentPath === item.href}
            >
              <svelte:component this={getIconComponent(item.icon)} className="w-4 h-4" />
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

  .beta-badge {
    border-color: #bb704f;
    color: #bb704f;
  }
</style>
