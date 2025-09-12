<script lang="ts">
  import { onMount } from "svelte"
  import { apiClient } from "$lib/api/client"
  import { toast } from "$lib/stores/toast"
  import Toast from "$lib/components/Toast.svelte"
  import FormField from "$lib/components/FormField.svelte"
  import ConfirmModal from "$lib/components/ConfirmModal.svelte"
  import ColorPicker from "$lib/components/ColorPicker.svelte"
  import Label from "$lib/components/Label.svelte"
  import TagIcon from "$lib/components/icons/Tag.svelte"
  import { formatErrorForUser, handleApiCallWithAuthRedirect } from "$lib/utils/error-handling"
  import type { paths } from "$lib/api/schema"

  type LabelsResponse =
    paths["/labels"]["get"]["responses"]["200"]["content"]["application/json"]
  type Label = LabelsResponse["labels"][0]
  type CreateLabelRequest =
    paths["/labels"]["post"]["requestBody"]["content"]["application/json"]
  type UpdateLabelRequest =
    paths["/labels/{id}"]["put"]["requestBody"]["content"]["application/json"]

  let labels: Label[] = []
  let loading = true
  let error = ""

  // New/Edit label modal state
  let showModal = false
  let editingLabel: Label | null = null
  let modalTitle = ""
  let labelName = ""
  let labelDescription = ""
  let labelColor = ""

  // Delete confirmation modal
  let showDeleteModal = false
  let labelToDelete: Label | null = null

  // Form validation
  $: isValidName =
    labelName.trim().length > 0 &&
    labelName.length <= 39 &&
    /^[a-zA-Z0-9]([a-zA-Z0-9]|-(?=[a-zA-Z0-9]))*$/.test(labelName.trim())
  $: isValidColor = /^#?[0-9a-f]{6}$/i.test(labelColor)
  $: isValidDescription = labelDescription.length <= 250
  $: canSave = isValidName && isValidColor && isValidDescription

  async function loadLabels() {
    try {
      loading = true
      const result = await handleApiCallWithAuthRedirect(async () => {
        return await apiClient.GET("/labels")
      })

      if (result.error) {
        error = result.error
        return
      }

      labels = result.data?.labels || []
      error = ""
    } catch (err) {
      console.error("Error loading labels:", err)
      error = "Failed to load labels. Please try again."
    } finally {
      loading = false
    }
  }

  function openCreateModal() {
    editingLabel = null
    modalTitle = "Create Label"
    labelName = ""
    labelDescription = ""
    labelColor = "#FFD700" // Default to gold
    showModal = true
  }

  function openEditModal(label: Label) {
    editingLabel = label
    modalTitle = "Edit Label"
    labelName = label.name
    labelDescription = label.description || ""
    labelColor = label.color.startsWith("#") ? label.color : `#${label.color}`
    showModal = true
  }

  function closeModal() {
    showModal = false
    editingLabel = null
    labelName = ""
    labelDescription = ""
    labelColor = ""
  }

  // Color normalization handled inside Label component

  async function saveLabel() {
    if (!canSave) return

    try {
      const colorWithoutHash = labelColor.replace("#", "")

      if (editingLabel) {
        // Update existing label
        const response = await apiClient.PUT("/labels/{id}", {
          params: { path: { id: editingLabel.id } },
          body: {
            name: labelName.trim(),
            description: labelDescription.trim() || null,
            color: colorWithoutHash,
          },
        })

        if (response.error) {
          toast.error(formatErrorForUser(response.error))
          return
        }

        // Update the label in the list
        const index = labels.findIndex((l) => l.id === editingLabel!.id)
        if (index >= 0 && response.data) {
          labels[index] = {
            ...response.data,
            consumption_count: labels[index].consumption_count,
            item_count: labels[index].item_count,
          }
        }

        toast.success("Label updated successfully")
      } else {
        // Create new label
        const response = await apiClient.POST("/labels", {
          body: {
            name: labelName.trim(),
            description: labelDescription.trim() || undefined,
            color: colorWithoutHash,
          },
        })

        if (response.error) {
          toast.error(formatErrorForUser(response.error))
          return
        }

        if (response.data) {
          // Add new label to the list with zero usage counts
          labels = [
            ...labels,
            { ...response.data, consumption_count: 0, item_count: 0 },
          ]
        }

        toast.success("Label created successfully")
      }

      closeModal()
    } catch (err) {
      console.error("Error saving label:", err)
      toast.error("Failed to save label. Please try again.")
    }
  }

  function openDeleteModal(label: Label) {
    labelToDelete = label
    showDeleteModal = true
  }

  function closeDeleteModal() {
    showDeleteModal = false
    labelToDelete = null
  }

  async function deleteLabel() {
    if (!labelToDelete) return

    try {
      const response = await apiClient.DELETE("/labels/{id}", {
        params: { path: { id: labelToDelete.id } },
      })

      if (response.error) {
        toast.error(formatErrorForUser(response.error))
        return
      }

      // Remove label from the list
      labels = labels.filter((l) => l.id !== labelToDelete!.id)
      toast.success("Label deleted successfully")

      closeDeleteModal()
    } catch (err) {
      console.error("Error deleting label:", err)
      toast.error("Failed to delete label. Please try again.")
    }
  }

  onMount(() => {
    loadLabels()
  })
</script>

<svelte:head>
  <title>Labels - Noot</title>
  <meta
    name="description"
    content="Manage your labels for organizing meals and nutrition tracking."
  />
</svelte:head>

<div class="min-h-full bg-base-100">
  <div class="container mx-auto px-4 py-8 max-w-6xl">
    <!-- Header -->
    <div class="mb-8">
      <div
        class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-4"
      >
        <div>
          <h1 class="text-3xl font-bold text-base-content flex items-center gap-3">
            <TagIcon className="w-8 h-8" />
            Labels
          </h1>
          <p class="text-base-content-lighter mt-2">
            Organize your meals and nutrition tracking with custom labels.
          </p>
        </div>
        <button
          class="btn btn-primary min-h-[44px] shrink-0"
          on:click={openCreateModal}
        >
          <svg
            class="w-5 h-5 mr-2"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 6v6m0 0v6m0-6h6m-6 0H6"
            />
          </svg>
          New Label
        </button>
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
        <svg
          class="stroke-current shrink-0 w-6 h-6"
          fill="none"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <span>{error}</span>
        <div>
          <button class="btn btn-sm btn-ghost min-h-[44px]" on:click={loadLabels}>
            Try Again
          </button>
        </div>
      </div>
    {:else}
      <!-- Labels list -->
      <div class="card bg-base-100 shadow-xl">
        <div class="card-body">
          {#if labels.length === 0}
            <div class="text-center py-12">
              <TagIcon
                className="w-16 h-16 mx-auto text-base-content/50 mb-4"
              />
              <h3 class="text-lg font-medium text-base-content/70 mb-2">
                No labels yet
              </h3>
              <p class="text-base-content/50 mb-4">
                Create your first label to start organizing your meals.
              </p>
              <button class="btn btn-primary" on:click={openCreateModal}>
                Create Label
              </button>
            </div>
          {:else}
            <div class="space-y-2">
              {#each labels as label}
                <div
                  class="flex flex-col sm:flex-row sm:items-center sm:justify-between p-4 border border-base-300 rounded-lg hover:bg-base-50 transition-colors gap-3 sm:gap-0"
                >
                  <div class="flex flex-col gap-2">
                    <div class="flex items-center gap-3">
                      <!-- Color badge -->
                      <Label name={label.name} color={label.color} />
                    </div>
                    <div class="flex flex-col gap-1 ml-0 sm:ml-12">
                      {#if label.description}
                        <span class="text-sm text-base-content/70"
                          >{label.description}</span
                        >
                      {/if}
                      <span class="text-xs text-base-content/50">
                        {label.consumption_count}
                        {label.consumption_count === 1
                          ? "consumption"
                          : "consumptions"}
                      </span>
                    </div>
                  </div>

                  <div class="flex gap-2 self-start sm:self-center">
                    <button
                      class="btn btn-ghost btn-sm min-h-[44px] min-w-[44px]"
                      aria-label={`Edit label ${label.name}`}
                      title={`Edit label ${label.name}`}
                      on:click={() => openEditModal(label)}
                    >
                      <svg
                        class="w-4 h-4"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2"
                          d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                        />
                      </svg>
                    </button>
                    <button
                      class="btn btn-ghost btn-sm min-h-[44px] min-w-[44px] text-error hover:bg-error hover:text-error-content"
                      aria-label={`Delete label ${label.name}`}
                      title={`Delete label ${label.name}`}
                      on:click={() => openDeleteModal(label)}
                    >
                      <svg
                        class="w-4 h-4"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2"
                          d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                        />
                      </svg>
                    </button>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>
    {/if}
  </div>
</div>

<!-- Create/Edit Label Modal -->
{#if showModal}
  <div class="modal modal-open">
    <div class="modal-box max-w-md">
      <h3 class="font-bold text-lg">{modalTitle}</h3>

      <form on:submit|preventDefault={saveLabel} class="space-y-4 mt-4">
        <!-- Label Name -->
        <FormField
          id="labelName"
          label="Label Name"
          bind:value={labelName}
          required
          maxlength={39}
          placeholder="e.g. breakfast, healthy"
          error={!isValidName && labelName.length > 0
            ? "Name must be 1-39 characters and contain only letters, numbers, and hyphens"
            : ""}
        />

        <!-- Label Description -->
        <FormField
          id="labelDescription"
          label="Description (Optional)"
          bind:value={labelDescription}
          maxlength={250}
          placeholder="Optional description for this label"
          error={!isValidDescription
            ? "Description must be 250 characters or less"
            : ""}
        />

        <!-- Color Selection -->
        <div class="form-control">
          <label class="label" for="labelColor">
            <span class="label-text">Color</span>
          </label>

          <ColorPicker 
            bind:selectedColor={labelColor} 
            previewName={labelName || "Label preview"} 
          />

          {#if !isValidColor && labelColor.length > 0}
            <div class="label">
              <span class="label-text-alt text-error"
                >Please enter a valid hex color (e.g. #FF0000)</span
              >
            </div>
          {/if}
        </div>
      </form>

      <div class="modal-action">
        <button class="btn btn-ghost" on:click={closeModal}>Cancel</button>
        <button
          class="btn btn-primary"
          disabled={!canSave}
          on:click={saveLabel}
        >
          {editingLabel ? "Update" : "Create"}
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Delete Confirmation Modal -->
<ConfirmModal
  show={showDeleteModal}
  title="Delete Label"
  message="Are you sure you want to delete the label '{labelToDelete?.name}'? This will remove it from all associated consumptions."
  confirmText="Delete"
  confirmVariant="error"
  onConfirm={deleteLabel}
  onCancel={closeDeleteModal}
/>

<Toast />

<style>
</style>
