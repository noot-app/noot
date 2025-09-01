<script lang="ts">
  import { onMount } from 'svelte'
  
  let versionInfo: {
    commit: string
    commitShort: string
    buildTime: string
    tag: string
  } | null = null
  
  let loading = true
  let error: string | null = null
  
  onMount(async () => {
    try {
      const response = await fetch('/version')
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`)
      }
      versionInfo = await response.json()
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load version info'
    } finally {
      loading = false
    }
  })
</script>

<svelte:head>
  <title>Version Info - Noot</title>
  <meta name="description" content="Version and build information for Noot" />
</svelte:head>

<div class="max-w-2xl mx-auto p-6">
  <h1 class="text-3xl font-bold mb-6">Version Information</h1>
  
  {#if loading}
    <div class="loading loading-spinner loading-md"></div>
    <p class="ml-4">Loading version information...</p>
  {:else if error}
    <div class="alert alert-error">
      <span>Error loading version information: {error}</span>
    </div>
  {:else if versionInfo}
    <div class="card bg-base-200 shadow-xl">
      <div class="card-body">
        <h2 class="card-title">Build Information</h2>
        
        <div class="space-y-4">
          <div class="flex justify-between items-center">
            <span class="font-semibold">Tag:</span>
            <span class="font-mono bg-base-300 px-2 py-1 rounded">{versionInfo.tag}</span>
          </div>
          
          <div class="flex justify-between items-center">
            <span class="font-semibold">Commit SHA:</span>
            <span class="font-mono bg-base-300 px-2 py-1 rounded" title="Full commit: {versionInfo.commit}">
              {versionInfo.commitShort}
            </span>
          </div>
          
          <div class="flex justify-between items-center">
            <span class="font-semibold">Build Time:</span>
            <span class="font-mono bg-base-300 px-2 py-1 rounded">{versionInfo.buildTime}</span>
          </div>
          
          <div class="flex justify-between items-center">
            <span class="font-semibold">Full Commit SHA:</span>
            <span class="font-mono bg-base-300 px-2 py-1 rounded text-xs break-all">
              {versionInfo.commit}
            </span>
          </div>
        </div>
      </div>
    </div>
  {/if}
</div>