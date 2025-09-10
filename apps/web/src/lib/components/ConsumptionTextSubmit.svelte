<script lang="ts">
  import { apiClient } from "$lib/api/client"
  import ArrowUpCircle from "$lib/components/icons/ArrowUpCircle.svelte"
  import { browser } from "$app/environment"
  import { getStorageJSON, setStorageJSON, removeStorageItem } from "$lib/utils/secure-storage"
  import { createEventDispatcher, onMount, onDestroy } from "svelte"

  // Event dispatcher for parent communication
  const dispatch = createEventDispatcher<{
    submit: { 
      transcript: string
      result: any
      consumptionId: string | null
      submissionText: string
    }
    error: { message: string }
    processing: { isProcessing: boolean }
  }>()

  // Props
  export let disabled = false
  export let placeholder = "Type naturally about what you ate"
  export let storageKey = 'consumption-text-draft'

  // Storage keys for draft persistence
  const DRAFT_STORAGE_KEY = storageKey
  const DRAFT_EXPIRY_MS = 24 * 60 * 60 * 1000 // 24 hours

  // Component state
  let textInput = ""
  let isProcessing = false
  let draftSaveTimeout: ReturnType<typeof setTimeout> | null = null
  let hasSubmitted = false
  let lastSubmissionText = ""

  // Auto-resize textarea to fit content
  function autoResizeTextarea() {
    const textarea = document.querySelector('.chat-textarea') as HTMLTextAreaElement
    if (textarea) {
      // Reset height to auto to get the correct scrollHeight
      textarea.style.height = 'auto'
      // Set height to scrollHeight to fit content
      textarea.style.height = `${Math.min(textarea.scrollHeight, 200)}px`
    }
  }

  // Debounced draft saving to prevent excessive storage writes
  function debouncedSaveDraft() {
    if (!browser) return
    
    if (draftSaveTimeout) {
      clearTimeout(draftSaveTimeout)
    }
    
    draftSaveTimeout = setTimeout(() => {
      saveDraftState()
    }, 500) // 500ms debounce
  }

  // Function to persist draft state
  function saveDraftState() {
    if (!browser) return
    
    const timestamp = Date.now()
    
    // Only save non-empty text input
    if (textInput.trim().length > 0) {
      setStorageJSON(DRAFT_STORAGE_KEY, {
        text: textInput,
        timestamp
      })
    } else {
      removeStorageItem(DRAFT_STORAGE_KEY)
    }
  }

  // Function to restore draft state
  function restoreDraftState() {
    if (!browser) return
    
    const now = Date.now()
    
    // Restore text input
    const textDraft = getStorageJSON(DRAFT_STORAGE_KEY, { text: '', timestamp: 0 })
    if (textDraft.text.trim().length > 0 && (now - textDraft.timestamp) <= DRAFT_EXPIRY_MS) {
      textInput = textDraft.text
      // Trigger auto-resize after restoration
      setTimeout(autoResizeTextarea, 0)
    }
  }

  // Function to clear draft state (called after successful submission)
  function clearDraftState() {
    if (!browser) return
    removeStorageItem(DRAFT_STORAGE_KEY)
  }

  // Handle text input changes with debounced saving
  function handleTextInput() {
    debouncedSaveDraft()
    autoResizeTextarea()
  }

  // Submit text function
  async function submitText() {
    if (!textInput.trim() || isProcessing || disabled) {
      if (!textInput.trim()) {
        dispatch('error', { message: "Please enter a description of your meal" })
      }
      return
    }

    // Prevent duplicate submissions
    if (hasSubmitted && textInput === lastSubmissionText) {
      dispatch('error', { message: "This meal has already been submitted" })
      return
    }

    try {
      isProcessing = true
      dispatch('processing', { isProcessing: true })

      const submissionText = textInput.trim()
      const response = await apiClient.POST("/consumption", {
        body: {
          text: submissionText
        },
      })

      if (response.error) {
        throw new Error(`API Error: ${response.error}`)
      }

      const data = response.data
      const transcript = data?.transcript || submissionText
      const result = data
      const consumptionId = data?.id || null
      
      lastSubmissionText = submissionText
      hasSubmitted = true

      // Clear text input and draft state after successful submission
      textInput = ""
      clearDraftState()

      // Dispatch success event
      dispatch('submit', {
        transcript,
        result,
        consumptionId,
        submissionText
      })
    } catch (err) {
      dispatch('error', { message: `Error processing text: ${err}` })
      console.error("Submit error:", err)
    } finally {
      isProcessing = false
      dispatch('processing', { isProcessing: false })
    }
  }

  // Public methods that can be called by parent
  export function reset() {
    textInput = ""
    hasSubmitted = false
    lastSubmissionText = ""
    isProcessing = false
    clearDraftState()
  }

  export function getValue() {
    return textInput
  }

  export function setValue(value: string) {
    textInput = value
    setTimeout(autoResizeTextarea, 0)
  }

  // Initialize component
  onMount(() => {
    restoreDraftState()
  })

  // Cleanup on destroy
  onDestroy(() => {
    // Clear timeouts
    if (draftSaveTimeout) {
      clearTimeout(draftSaveTimeout)
    }
    
    // Save current draft state before leaving
    if (browser && textInput.trim()) {
      saveDraftState()
    }
  })
</script>

<!-- Chat-style text input interface -->
<div class="chat-input-container">
  <div class="chat-input-wrapper">
    <textarea
      bind:value={textInput}
      on:input={handleTextInput}
      {placeholder}
      class="chat-textarea"
      disabled={isProcessing || disabled}
      rows="1"
      on:keydown={(e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
          e.preventDefault()
          submitText()
        }
      }}
    ></textarea>
    <button
      class="chat-submit-button"
      on:click={submitText}
      disabled={isProcessing || disabled || !textInput.trim()}
      aria-label="Submit meal description"
    >
      {#if isProcessing}
        <span class="loading loading-spinner text-primary"></span>
      {:else}
        <ArrowUpCircle className="w-8 h-8" variant="solid" />
      {/if}
    </button>
  </div>
</div>

<style>
  /* Chat-style input interface */
  .chat-input-container {
    width: 100%;
    max-width: 768px;
    margin: 0 auto;
    padding: 0 1rem;
  }

  .chat-input-wrapper {
    position: relative;
    background: var(--color-text-box-content);
    border: 1px solid var(--color-base-300);
    border-radius: 16px;
    padding: 14px 60px 14px 20px;
    display: flex;
    align-items: flex-end;
    min-height: 48px;
    cursor: text;
    transition: all 200ms;
    box-shadow: 0 4px 20px hsl(0 0% 0% / 8%);
  }

  .chat-input-wrapper:hover {
    border-color: var(--color-base-content-lighter);
    box-shadow: 0 6px 25px hsl(0 0% 0% / 12%);
  }

  .chat-input-wrapper:focus-within {
    border-color: var(--color-base-content-lighter);
    box-shadow: 0 6px 25px hsl(0 0% 0% / 12%);
  }

  .chat-input-wrapper:hover:focus-within {
    border-color: var(--color-base-content-lighter);
    box-shadow: 0 6px 25px hsl(0 0% 0% / 12%);
  }

  .chat-textarea {
    flex: 1;
    border: none;
    outline: none;
    resize: none;
    background: transparent;
    color: var(--color-base-content);
    font-size: 16px;
    line-height: 1.5;
    padding: 0;
    margin: 0;
    max-height: 200px;
    overflow-y: auto;
    font-family: inherit;
    min-height: 24px;
  }

  .chat-textarea::placeholder {
    color: var(--color-placeholder);
  }

  .chat-textarea:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .chat-submit-button {
    position: absolute;
    right: 6px;
    bottom: 6px;
    width: 36px;
    height: 36px;
    border-radius: 8px;
    border: none;
    background: transparent;
    color: var(--color-primary);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 100ms;
    flex-shrink: 0;
  }

  .chat-submit-button:hover:not(:disabled) {
    background: color-mix(in srgb, var(--color-primary) 10%, transparent);
    transform: scale(0.95);
  }

  .chat-submit-button:active:not(:disabled) {
    transform: scale(0.9);
  }

  .chat-submit-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
  }
</style>
