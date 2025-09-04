<script lang="ts">
  import { onMount } from "svelte"
  import { page } from "$app/stores"
  import { goto } from "$app/navigation"
  import { apiClient } from "$lib/api/client"
  import { toast } from "$lib/stores/toast"
  import Toast from "$lib/components/Toast.svelte"
  import FormField from "$lib/components/FormField.svelte"
  import FormSelect from "$lib/components/FormSelect.svelte"
  import ConfirmModal from "$lib/components/ConfirmModal.svelte"
  import KeyIcon from "$lib/components/icons/Key.svelte"
  import EyeIcon from "$lib/components/icons/Eye.svelte"
  import CubeIcon from "$lib/components/icons/Cube.svelte"
  import TrashIcon from "$lib/components/icons/Trash.svelte"
  import ArrowPathIcon from "$lib/components/icons/ArrowPath.svelte"
  import ClockIcon from "$lib/components/icons/Clock.svelte"
  import { formatErrorForUser, handleApiCallWithAuthRedirect } from "$lib/utils/error-handling"
  import type { paths } from "$lib/api/schema"

  type APIKeysResponse =
    paths["/api-keys"]["get"]["responses"]["200"]["content"]["application/json"]
  type APIKey = APIKeysResponse["api_keys"][0]
  type CreateAPIKeyRequest =
    paths["/api-keys"]["post"]["requestBody"]["content"]["application/json"]
  type CreateAPIKeyResponse =
    paths["/api-keys"]["post"]["responses"]["201"]["content"]["application/json"]

  let apiKeys: APIKey[] = []
  let loading = true
  let error = ""
  let isProUser = false

  // New API key modal state
  let showCreateModal = false
  let keyName = ""
  let keyScope: "read" | "read_write" = "read"
  let keyExpiresAt = ""
  let isCreating = false

  // API key created modal state (shows secret once)
  let showSecretModal = false
  let newApiKeySecret = ""
  let newApiKeyName = ""

  // Delete confirmation modal
  let showDeleteModal = false
  let keyToDelete: APIKey | null = null

  // Rotate confirmation modal
  let showRotateModal = false
  let keyToRotate: APIKey | null = null

  // Copy functionality
  let copyTimeoutId: number | null = null

  // Scope options
  const scopeOptions = [
    { 
      value: "read", 
      label: "Read Only", 
      description: "Can only view data (GET requests)" 
    },
    { 
      value: "read_write", 
      label: "Full Access", 
      description: "Can view and modify data (all HTTP methods)" 
    }
  ]

  // Form validation
  $: isValidName = keyName.trim().length > 0 && keyName.length <= 100
  $: isValidExpiration = !keyExpiresAt || new Date(keyExpiresAt) > new Date()
  $: canCreate = isValidName && isValidExpiration && !isCreating

  onMount(async () => {
    // Check if user has pro access by trying to load API keys
    await loadApiKeys()
  })

  async function loadApiKeys() {
    try {
      loading = true
      error = ""
      
      const result = await handleApiCallWithAuthRedirect(async () => {
        return await apiClient.GET("/api-keys")
      })

      if (result.error) {
        // If 403, user is not a pro user
        if (result.error.includes("Forbidden") || result.error.includes("permission")) {
          isProUser = false
          error = "API keys are available for Pro users only. Upgrade your subscription to access this feature."
          return
        }
        
        error = result.error
        return
      }

      isProUser = true
      apiKeys = result.data?.api_keys || []
    } catch (err) {
      console.error("Error loading API keys:", err)
      error = "Failed to load API keys. Please try again."
    } finally {
      loading = false
    }
  }

  function openCreateModal() {
    if (!isProUser) return
    
    keyName = ""
    keyScope = "read"
    keyExpiresAt = ""
    showCreateModal = true
  }

  function closeCreateModal() {
    showCreateModal = false
    keyName = ""
    keyScope = "read"
    keyExpiresAt = ""
  }

  async function createApiKey() {
    if (!canCreate || !isProUser) return

    try {
      isCreating = true
      
      const keyData: CreateAPIKeyRequest = {
        name: keyName.trim(),
        scope: keyScope,
        expires_at: keyExpiresAt ? new Date(keyExpiresAt).toISOString() : null
      }

      const response = await apiClient.POST("/api-keys", {
        body: keyData,
      })

      if (response.error) {
        toast.error(formatErrorForUser(response.error))
        return
      }

      if (response.data) {
        // Store the secret and key name for the secret modal
        newApiKeySecret = response.data.secret
        newApiKeyName = response.data.api_key.name
        
        // Add the new API key to the list (without secret)
        apiKeys = [...apiKeys, response.data.api_key]
        
        // Close create modal and show secret modal
        closeCreateModal()
        showSecretModal = true
        
        toast.success("API key created successfully")
      }
    } catch (err) {
      console.error("Error creating API key:", err)
      toast.error("Failed to create API key. Please try again.")
    } finally {
      isCreating = false
    }
  }

  function openDeleteModal(apiKey: APIKey) {
    keyToDelete = apiKey
    showDeleteModal = true
  }

  function closeDeleteModal() {
    showDeleteModal = false
    keyToDelete = null
  }

  async function deleteApiKey() {
    if (!keyToDelete || !isProUser) return

    try {
      const response = await apiClient.DELETE("/api-keys/{id}", {
        params: { path: { id: keyToDelete.id } },
      })

      if (response.error) {
        toast.error(formatErrorForUser(response.error))
        return
      }

      // Remove the key from the list
      apiKeys = apiKeys.filter((k) => k.id !== keyToDelete!.id)
      toast.success("API key deleted successfully")
      
      closeDeleteModal()
    } catch (err) {
      console.error("Error deleting API key:", err)
      toast.error("Failed to delete API key. Please try again.")
    }
  }

  function openRotateModal(apiKey: APIKey) {
    keyToRotate = apiKey
    showRotateModal = true
  }

  function closeRotateModal() {
    showRotateModal = false
    keyToRotate = null
  }

  async function rotateApiKey() {
    if (!keyToRotate || !isProUser) return

    try {
      const response = await apiClient.POST("/api-keys/{id}/rotate", {
        params: { path: { id: keyToRotate.id } },
      })

      if (response.error) {
        toast.error(formatErrorForUser(response.error))
        return
      }

      if (response.data) {
        // Store the new secret and key name for the secret modal
        newApiKeySecret = response.data.secret
        newApiKeyName = response.data.api_key.name
        
        // Update the API key in the list
        const keyIndex = apiKeys.findIndex((k) => k.id === keyToRotate!.id)
        if (keyIndex >= 0) {
          apiKeys[keyIndex] = response.data.api_key
          apiKeys = [...apiKeys] // Trigger reactivity
        }
        
        // Close rotate modal and show secret modal
        closeRotateModal()
        showSecretModal = true
        
        toast.success("API key rotated successfully")
      }
    } catch (err) {
      console.error("Error rotating API key:", err)
      toast.error("Failed to rotate API key. Please try again.")
    }
  }

  async function copySecret() {
    try {
      await navigator.clipboard.writeText(newApiKeySecret)
      toast.success("API key copied to clipboard")
      
      // Clear any existing timeout
      if (copyTimeoutId) {
        clearTimeout(copyTimeoutId)
      }
      
      // Auto-close the modal after a delay
      copyTimeoutId = window.setTimeout(() => {
        showSecretModal = false
        newApiKeySecret = ""
        newApiKeyName = ""
      }, 3000)
    } catch (err) {
      console.error("Failed to copy:", err)
      toast.error("Failed to copy to clipboard")
    }
  }

  function closeSecretModal() {
    showSecretModal = false
    newApiKeySecret = ""
    newApiKeyName = ""
    
    if (copyTimeoutId) {
      clearTimeout(copyTimeoutId)
      copyTimeoutId = null
    }
  }

  function formatDate(dateStr: string) {
    return new Date(dateStr).toLocaleDateString()
  }

  function formatDateTime(dateStr: string) {
    return new Date(dateStr).toLocaleString()
  }

  function getStatusBadge(apiKey: APIKey) {
    const now = new Date()
    
    if (apiKey.revoked_at) {
      return { text: "Revoked", class: "badge-error" }
    }
    
    if (apiKey.expires_at && new Date(apiKey.expires_at) <= now) {
      return { text: "Expired", class: "badge-warning" }
    }
    
    return { text: "Active", class: "badge-success" }
  }

  function getScopeDisplay(scope: "read" | "read_write") {
    return scope === "read" ? "Read Only" : "Full Access"
  }

  // Clean up timeout on component destroy
  onMount(() => {
    return () => {
      if (copyTimeoutId) {
        clearTimeout(copyTimeoutId)
      }
    }
  })
</script>

<svelte:head>
  <title>API Keys - Noot</title>
  <meta
    name="description"
    content="Manage your API keys for programmatic access to your nutrition data."
  />
</svelte:head>

<div class="min-h-screen bg-base-100">
  <div class="container mx-auto px-4 py-8 max-w-6xl">
    <!-- Header -->
    <div class="mb-8">
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-4">
        <div>
          <h1 class="text-3xl font-bold text-base-content">API Keys</h1>
          <p class="text-base-content/70 mt-2">
            Create and manage API keys for programmatic access to your data.
          </p>
        </div>
        {#if isProUser}
          <button
            class="btn btn-primary min-h-[44px] shrink-0"
            on:click={openCreateModal}
          >
            <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
            </svg>
            New API Key
          </button>
        {/if}
      </div>
    </div>

    <!-- Loading state -->
    {#if loading}
      <div class="flex justify-center items-center py-12">
        <span class="loading loading-spinner loading-lg"></span>
      </div>
    {:else if error}
      <!-- Error state -->
      <div class="alert alert-error">
        <svg class="stroke-current shrink-0 w-6 h-6" fill="none" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span>{error}</span>
        {#if !isProUser}
          <div>
            <a href="/profile" class="btn btn-sm btn-ghost">
              Go to Profile
            </a>
          </div>
        {:else}
          <div>
            <button class="btn btn-sm btn-ghost" on:click={loadApiKeys}>
              Try Again
            </button>
          </div>
        {/if}
      </div>
    {:else if !isProUser}
      <!-- Pro upgrade prompt -->
      <div class="card bg-base-100 shadow-xl">
        <div class="card-body text-center py-12">
          <div class="mx-auto w-16 h-16 bg-warning/10 rounded-full flex items-center justify-center mb-4">
            <KeyIcon class_="w-8 h-8 text-warning" />
          </div>
          <h2 class="text-2xl font-bold mb-2">Pro Feature</h2>
          <p class="text-base-content/70 max-w-md mx-auto mb-6">
            API keys are exclusive to Pro users. Upgrade your subscription to create secure API keys for programmatic access to your nutrition data.
          </p>
          <div class="flex gap-3 justify-center">
            <a href="/profile" class="btn btn-outline">
              View Profile
            </a>
            <button class="btn btn-primary">
              Upgrade to Pro
            </button>
          </div>
        </div>
      </div>
    {:else}
      <!-- API Keys list -->
      <div class="card bg-base-100 shadow-xl">
        <div class="card-body">
          {#if apiKeys.length === 0}
            <!-- Empty state -->
            <div class="text-center py-12">
              <div class="mx-auto w-16 h-16 bg-base-200 rounded-full flex items-center justify-center mb-4">
                <KeyIcon class_="w-8 h-8 text-base-content/40" />
              </div>
              <h3 class="text-lg font-semibold mb-2">No API Keys</h3>
              <p class="text-base-content/60 mb-4">
                Create your first API key to start accessing your data programmatically.
              </p>
              <button
                class="btn btn-primary"
                on:click={openCreateModal}
              >
                <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                </svg>
                Create API Key
              </button>
            </div>
          {:else}
            <!-- API Keys grid -->
            <div class="grid gap-4 lg:grid-cols-2">
              {#each apiKeys as apiKey}
                {@const status = getStatusBadge(apiKey)}
                {@const scope = apiKey.scope as "read" | "read_write"}
                <div class="card bg-base-200 border border-base-300">
                  <div class="card-body p-6">
                    <div class="flex items-start justify-between mb-4">
                      <div class="flex-1 min-w-0">
                        <h3 class="font-semibold text-lg truncate mb-1">
                          {apiKey.name}
                        </h3>
                        <div class="flex items-center gap-2 text-sm text-base-content/60">
                          <span class="font-mono bg-base-100 px-2 py-1 rounded text-xs">
                            {apiKey.prefix}***
                          </span>
                        </div>
                      </div>
                      <div class="badge {status.class}">{status.text}</div>
                    </div>

                    <!-- API Key details -->
                    <div class="space-y-2 mb-4">
                      <div class="flex items-center gap-2 text-sm">
                        {#if scope === 'read'}
                          <EyeIcon class_="w-4 h-4" />
                        {:else}
                          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                          </svg>
                        {/if}
                        <span class="text-base-content/70">Scope:</span>
                        <span class="font-medium">{getScopeDisplay(scope)}</span>
                      </div>
                      
                      <div class="flex items-center gap-2 text-sm">
                        <CubeIcon class_="w-4 h-4" />
                        <span class="text-base-content/70">Created:</span>
                        <span>{formatDate(apiKey.created_at)}</span>
                      </div>
                      
                      {#if apiKey.last_used_at}
                        <div class="flex items-center gap-2 text-sm">
                          <ClockIcon class_="w-4 h-4" />
                          <span class="text-base-content/70">Last used:</span>
                          <span>{formatDateTime(apiKey.last_used_at)}</span>
                        </div>
                      {:else}
                        <div class="flex items-center gap-2 text-sm">
                          <ClockIcon class_="w-4 h-4" />
                          <span class="text-base-content/70">Last used:</span>
                          <span class="text-warning">Never</span>
                        </div>
                      {/if}
                      
                      {#if apiKey.expires_at}
                        <div class="flex items-center gap-2 text-sm">
                          <ClockIcon class_="w-4 h-4" />
                          <span class="text-base-content/70">Expires:</span>
                          <span class:text-warning={new Date(apiKey.expires_at) <= new Date()}>
                            {formatDate(apiKey.expires_at)}
                          </span>
                        </div>
                      {/if}
                    </div>

                    <!-- Actions -->
                    {#if !apiKey.revoked_at}
                      <div class="flex gap-2">
                        <button
                          class="btn btn-outline btn-sm flex-1"
                          on:click={() => openRotateModal(apiKey)}
                        >
                          <ArrowPathIcon class_="w-4 h-4" />
                          Rotate
                        </button>
                        <button
                          class="btn btn-error btn-sm flex-1"
                          on:click={() => openDeleteModal(apiKey)}
                        >
                          <TrashIcon class_="w-4 h-4" />
                          Delete
                        </button>
                      </div>
                    {/if}
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>
    {/if}

    <!-- Usage Documentation -->
    {#if isProUser}
      <div class="card bg-base-100 shadow-xl mt-8">
        <div class="card-body">
          <h2 class="card-title">Using Your API Keys</h2>
          <div class="prose max-w-none">
            <p class="text-base-content/70">
              Include your API key in the <code>X-API-Key</code> header when making requests to the API.
            </p>
            
            <div class="mockup-code text-sm mt-4">
              <pre><code>curl -H "X-API-Key: noot_abc123_your-secret-here" \
  https://api.nootapp.io/api/v1/consumptions</code></pre>
            </div>

            <div class="grid md:grid-cols-2 gap-6 mt-6">
              <div>
                <h3 class="text-lg font-semibold mb-2">API Scopes</h3>
                <ul class="space-y-2">
                  <li class="flex items-start gap-2">
                    <div class="badge badge-info badge-sm mt-0.5">READ</div>
                    <div class="text-sm">
                      <strong>Read Only:</strong> Access to GET endpoints only. Perfect for monitoring and analytics.
                    </div>
                  </li>
                  <li class="flex items-start gap-2">
                    <div class="badge badge-warning badge-sm mt-0.5">WRITE</div>
                    <div class="text-sm">
                      <strong>Full Access:</strong> Complete access to all endpoints. Can create, update, and delete data.
                    </div>
                  </li>
                </ul>
              </div>

              <div>
                <h3 class="text-lg font-semibold mb-2">Security Best Practices</h3>
                <ul class="text-sm space-y-1 text-base-content/70">
                  <li>Store API keys securely and never commit them to version control</li>
                  <li>Use environment variables in your applications</li>
                  <li>Rotate keys regularly or when they may be compromised</li>
                  <li>Use read-only keys when write access isn't needed</li>
                  <li>Set expiration dates for temporary integrations</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      </div>
    {/if}
  </div>
</div>

<!-- Create API Key Modal -->
{#if showCreateModal}
  <div class="modal modal-open">
    <div class="modal-box max-w-md">
      <h3 class="font-bold text-lg">Create API Key</h3>

      <form on:submit|preventDefault={createApiKey} class="space-y-4 mt-4">
        <!-- Key Name -->
        <FormField
          id="keyName"
          label="Name"
          bind:value={keyName}
          required
          maxlength={100}
          placeholder="e.g. Production Server, Mobile App"
          error={!isValidName && keyName.length > 0 ? "Name is required (max 100 characters)" : ""}
        />

        <!-- Scope Selection -->
        <FormSelect
          id="keyScope"
          label="Scope"
          bind:value={keyScope}
          options={scopeOptions.map(opt => ({ value: opt.value, label: opt.label }))}
          required
        />
        
        <!-- Scope Description -->
        {#each scopeOptions as option}
          {#if option.value === keyScope}
            <div class="text-sm text-base-content/60 -mt-2">
              {option.description}
            </div>
          {/if}
        {/each}

        <!-- Expiration Date (Optional) -->
        <FormField
          id="keyExpiresAt"
          label="Expiration Date (Optional)"
          type="datetime-local"
          bind:value={keyExpiresAt}
          min={new Date().toISOString().slice(0, 16)}
          placeholder="Leave empty for no expiration"
          error={!isValidExpiration ? "Expiration must be in the future" : ""}
        />
        
        {#if !keyExpiresAt}
          <div class="text-sm text-base-content/60 -mt-2">
            This key will never expire unless manually revoked
          </div>
        {/if}
      </form>

      <div class="modal-action">
        <button class="btn btn-ghost" on:click={closeCreateModal}>Cancel</button>
        <button
          class="btn btn-primary"
          disabled={!canCreate}
          on:click={createApiKey}
        >
          {#if isCreating}
            <span class="loading loading-spinner loading-sm"></span>
            Creating...
          {:else}
            Create Key
          {/if}
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- API Key Secret Modal (shown once) -->
{#if showSecretModal}
  <div class="modal modal-open">
    <div class="modal-box max-w-lg">
      <h3 class="font-bold text-lg text-success">API Key Created</h3>
      
      <div class="alert alert-warning mt-4">
        <svg class="w-6 h-6 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16c-.77.833.192 2.5 1.732 2.5z" />
        </svg>
        <div>
          <h3 class="font-bold">Important!</h3>
          <div class="text-sm">This is the only time you'll see the full API key. Copy it now and store it securely.</div>
        </div>
      </div>

      <div class="mt-6">
        <label class="label" for="api-key-secret">
          <span class="label-text font-medium">API Key: {newApiKeyName}</span>
        </label>
        <div class="flex gap-2">
          <input
            id="api-key-secret"
            type="text"
            class="input input-bordered flex-1 font-mono text-sm"
            value={newApiKeySecret}
            readonly
          />
          <button
            class="btn btn-outline"
            on:click={copySecret}
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
            Copy
          </button>
        </div>
      </div>

      <div class="modal-action">
        <button class="btn btn-primary" on:click={closeSecretModal}>
          I've Saved It
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Delete Confirmation Modal -->
<ConfirmModal
  show={showDeleteModal}
  title="Delete API Key"
  message="Are you sure you want to delete the API key '{keyToDelete?.name}'? This action will permanently remove and invalidate the key."
  confirmText="Delete Key"
  confirmVariant="error"
  onConfirm={deleteApiKey}
  onCancel={closeDeleteModal}
/>

<!-- Rotate Confirmation Modal -->
<ConfirmModal
  show={showRotateModal}
  title="Rotate API Key"
  message="This will generate a new secret for '{keyToRotate?.name}' and immediately invalidate the old one. Any applications using the current key will need to be updated."
  confirmText="Rotate Key"
  confirmVariant="warning"
  onConfirm={rotateApiKey}
  onCancel={closeRotateModal}
/>

<Toast />
