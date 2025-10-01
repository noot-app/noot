<script lang="ts">
  import Modal from "$lib/components/Modal.svelte"
  
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

  // Summary text for ToS and Privacy Policy
  const tosSummary = `By using Noot, you agree to: use the service for personal, non-commercial purposes; maintain accurate account information; not misuse or abuse the service; and understand that AI-generated content may contain inaccuracies. The service is provided "as is" with no warranties, and Noot limits its liability for damages. You also agree to resolve disputes through arbitration.`

  const privacySummary = `We collect your account information, audio recordings, and usage data to provide our nutrition tracking service. Your information is processed to generate nutrition analysis and improve our services. We share data with service providers like Stripe, Supabase, and AI transcription services, but we do not sell your personal information. You have rights to access, correct, or delete your data.`
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
        <h4 class="font-semibold text-base-content mb-1">Terms of Service Summary:</h4>
        <p class="text-base-content/70">{tosSummary}</p>
        <a
          href="/terms-of-service"
          target="_blank"
          rel="noopener noreferrer"
          class="link link-primary text-xs mt-1 inline-block"
        >
          Read full Terms of Service →
        </a>
      </div>
      
      <div class="divider my-2"></div>
      
      <div>
        <h4 class="font-semibold text-base-content mb-1">Privacy Policy Summary:</h4>
        <p class="text-base-content/70">{privacySummary}</p>
        <a
          href="/privacy-policy"
          target="_blank"
          rel="noopener noreferrer"
          class="link link-primary text-xs mt-1 inline-block"
        >
          Read full Privacy Policy →
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
<Modal bind:show={showToSModal} title="Terms of Service" size="lg">
  <div class="prose prose-sm max-w-none overflow-y-auto max-h-[60vh]">
    <p class="text-sm text-base-content/70 mb-4">
      Please review our full Terms of Service document for complete details.
    </p>
    <div class="text-xs space-y-3">
      <h3 class="font-semibold text-base">Key Points:</h3>
      <ul class="list-disc list-inside space-y-2 text-base-content/80">
        <li>Use service for personal, non-commercial purposes only</li>
        <li>Maintain accurate account information</li>
        <li>AI-generated content may contain inaccuracies - verify information</li>
        <li>Service provided "as is" without warranties</li>
        <li>Not medical advice - consult healthcare professionals</li>
        <li>Disputes resolved through arbitration</li>
        <li>Limited liability for damages</li>
      </ul>
    </div>
  </div>
  <div slot="actions" class="flex gap-2 justify-end">
    <a
      href="/terms-of-service"
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
<Modal bind:show={showPrivacyModal} title="Privacy Policy" size="lg">
  <div class="prose prose-sm max-w-none overflow-y-auto max-h-[60vh]">
    <p class="text-sm text-base-content/70 mb-4">
      Please review our full Privacy Policy document for complete details.
    </p>
    <div class="text-xs space-y-3">
      <h3 class="font-semibold text-base">Key Points:</h3>
      <ul class="list-disc list-inside space-y-2 text-base-content/80">
        <li>We collect account info, audio recordings, and usage data</li>
        <li>Data used to provide nutrition tracking and improve services</li>
        <li>Shared with service providers (Stripe, Supabase, AI services)</li>
        <li>We do not sell your personal information</li>
        <li>You can access, correct, or delete your data</li>
        <li>Data encrypted in transit and at rest</li>
        <li>Not HIPAA covered - not medical advice</li>
      </ul>
    </div>
  </div>
  <div slot="actions" class="flex gap-2 justify-end">
    <a
      href="/privacy-policy"
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
