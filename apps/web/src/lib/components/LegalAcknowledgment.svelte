<script lang="ts">
  import Modal from "$lib/components/Modal.svelte"
  import { LEGAL_SUMMARIES } from "$lib/constants/legal"
  
  export let showCheckbox: boolean = false
  export let checked: boolean = false
  export let disabled: boolean = false

  let showToSModal = false
  let showPrivacyModal = false
  let expanded = false

  function toggleExpanded() {
    expanded = !expanded
  }

  function openToS() {
    showToSModal = true
  }

  function openPrivacy() {
    showPrivacyModal = true
  }

  const { termsOfService, privacyPolicy } = LEGAL_SUMMARIES
</script>

<div class="space-y-3">
  <!-- Main acknowledgment text -->
  <div class="text-xs text-base-content/70 text-center">
    By signing up, you agree to the
    <button
      type="button"
      class="link link-primary font-medium"
      on:click={openToS}
      aria-label="View Terms of Service"
    >
      Terms of Service
    </button>
    and
    <button
      type="button"
      class="link link-primary font-medium"
      on:click={openPrivacy}
      aria-label="View Privacy Policy"
    >
      Privacy Policy
    </button>.
    <button
      type="button"
      class="link link-primary text-xs ml-1"
      on:click={toggleExpanded}
      aria-expanded={expanded}
      aria-label={expanded ? "Hide summary" : "Show summary"}
    >
      {expanded ? "Hide" : "Read more"}
    </button>
  </div>

  <!-- Expandable summary section -->
  {#if expanded}
    <div 
      class="text-xs bg-base-200 rounded-lg p-4 space-y-3 border border-base-300"
      role="region"
      aria-label="Legal terms summary"
    >
      <div>
        <h4 class="font-semibold text-base-content mb-1">{termsOfService.title} Summary:</h4>
        <p class="text-base-content/70">{termsOfService.summary}</p>
        <a
          href={termsOfService.url}
          target="_blank"
          rel="noopener noreferrer"
          class="link link-primary text-xs mt-1 inline-block"
        >
          Read full {termsOfService.title} →
        </a>
      </div>
      
      <div class="divider my-2"></div>
      
      <div>
        <h4 class="font-semibold text-base-content mb-1">{privacyPolicy.title} Summary:</h4>
        <p class="text-base-content/70">{privacyPolicy.summary}</p>
        <a
          href={privacyPolicy.url}
          target="_blank"
          rel="noopener noreferrer"
          class="link link-primary text-xs mt-1 inline-block"
        >
          Read full {privacyPolicy.title} →
        </a>
      </div>
    </div>
  {/if}

  <!-- Optional checkbox for explicit consent -->
  {#if showCheckbox}
    <div class="flex items-start gap-2 justify-center">
      <input
        type="checkbox"
        id="legal-consent"
        class="checkbox checkbox-sm mt-0.5"
        bind:checked
        {disabled}
        aria-required="true"
      />
      <label
        for="legal-consent"
        class="text-xs text-base-content/70 cursor-pointer"
      >
        I have read and agree to the Terms of Service and Privacy Policy
      </label>
    </div>
  {/if}
</div>

<!-- Terms of Service Modal -->
<Modal bind:show={showToSModal} title={termsOfService.title} size="lg">
  <div class="prose prose-sm max-w-none overflow-y-auto max-h-[60vh]">
    <p class="text-sm text-base-content/70 mb-4">
      Please review our full {termsOfService.title} document for complete details.
    </p>
    <div class="text-xs space-y-3">
      <h3 class="font-semibold text-base">Key Points:</h3>
      <ul class="list-disc list-inside space-y-2 text-base-content/80">
        {#each termsOfService.keyPoints as point}
          <li>{point}</li>
        {/each}
      </ul>
    </div>
  </div>
  <div slot="actions" class="flex gap-2 justify-end">
    <a
      href={termsOfService.url}
      target="_blank"
      rel="noopener noreferrer"
      class="btn btn-outline btn-sm"
    >
      View Full Document
    </a>
    <button class="btn btn-primary btn-sm" on:click={() => showToSModal = false}>
      Close
    </button>
  </div>
</Modal>

<!-- Privacy Policy Modal -->
<Modal bind:show={showPrivacyModal} title={privacyPolicy.title} size="lg">
  <div class="prose prose-sm max-w-none overflow-y-auto max-h-[60vh]">
    <p class="text-sm text-base-content/70 mb-4">
      Please review our full {privacyPolicy.title} document for complete details.
    </p>
    <div class="text-xs space-y-3">
      <h3 class="font-semibold text-base">Key Points:</h3>
      <ul class="list-disc list-inside space-y-2 text-base-content/80">
        {#each privacyPolicy.keyPoints as point}
          <li>{point}</li>
        {/each}
      </ul>
    </div>
  </div>
  <div slot="actions" class="flex gap-2 justify-end">
    <a
      href={privacyPolicy.url}
      target="_blank"
      rel="noopener noreferrer"
      class="btn btn-outline btn-sm"
    >
      View Full Document
    </a>
    <button class="btn btn-primary btn-sm" on:click={() => showPrivacyModal = false}>
      Close
    </button>
  </div>
</Modal>
