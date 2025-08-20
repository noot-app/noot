<script lang="ts">
  import { dev } from '$app/environment';
  import { page } from '$app/stores';
  import { navigating } from '$app/stores';
  import { onMount } from 'svelte';
  
  let pageLoadTime = 0;
  let currentUser = 'Not logged in'; // TODO: Get from auth when implemented
  let memoryUsage = '';
  let networkType = '';
  let isHidden = false;
  let navigationStartTime = 0;
  let errorCount = 0;
  let warningCount = 0;
  
  // Original console methods
  let originalError: typeof console.error;
  let originalWarn: typeof console.warn;
  let errorHandler: ((event: ErrorEvent) => void) | undefined;
  let rejectionHandler: ((event: PromiseRejectionEvent) => void) | undefined;
  
  function copyDebugInfo() {
    const debugInfo = `
Dev Info:
- Route: ${$page.url.pathname}
- Load Time: ${pageLoadTime}ms
- Memory: ${memoryUsage}
- Network: ${networkType}
- Errors: ${errorCount}
- Warnings: ${warningCount}
- User: ${currentUser}
    `.trim();
    
    navigator.clipboard.writeText(debugInfo).then(() => {
      console.log('Debug info copied to clipboard');
    });
  }
  
  function setupConsoleMonitoring() {
    // Store original methods
    originalError = console.error;
    originalWarn = console.warn;
    
    // Override console.error
    console.error = (...args: any[]) => {
      errorCount++;
      originalError.apply(console, args);
    };
    
    // Override console.warn
    console.warn = (...args: any[]) => {
      warningCount++;
      originalWarn.apply(console, args);
    };
    
    // Listen for unhandled errors
    errorHandler = () => {
      errorCount++;
    };
    window.addEventListener('error', errorHandler);
    
    // Listen for unhandled promise rejections
    rejectionHandler = () => {
      errorCount++;
    };
    window.addEventListener('unhandledrejection', rejectionHandler);
  }
  
  function restoreConsoleMonitoring() {
    if (originalError && originalWarn) {
      console.error = originalError;
      console.warn = originalWarn;
    }
    
    if (errorHandler) {
      window.removeEventListener('error', errorHandler);
    }
    
    if (rejectionHandler) {
      window.removeEventListener('unhandledrejection', rejectionHandler);
    }
  }
  
  function clearCounts() {
    errorCount = 0;
    warningCount = 0;
  }
  
  function handleKeydown(event: KeyboardEvent) {
    if (event.key === '`' && !event.ctrlKey && !event.metaKey && !event.altKey) {
      // Only toggle if not in an input field
      const target = event.target as HTMLElement;
      if (target.tagName !== 'INPUT' && target.tagName !== 'TEXTAREA' && !target.isContentEditable) {
        event.preventDefault();
        isHidden = !isHidden;
      }
    }
  }
  
  onMount(() => {
    // Initial page load time
    pageLoadTime = Math.round(performance.now());
    
    // Setup console monitoring
    setupConsoleMonitoring();
    
    // Get memory usage if available
    if ('memory' in performance) {
      const mem = (performance as any).memory;
      memoryUsage = `${Math.round(mem.usedJSHeapSize / 1024 / 1024)}MB`;
    }
    
    // Get network information if available
    if ('connection' in navigator) {
      const conn = (navigator as any).connection;
      networkType = conn?.effectiveType || 'unknown';
    }
    
    // Add global keydown listener
    window.addEventListener('keydown', handleKeydown);
    
    // TODO: Get actual user when auth is implemented
    // currentUser = $authStore?.user?.email || 'Not logged in';
    
    return () => {
      window.removeEventListener('keydown', handleKeydown);
      restoreConsoleMonitoring();
    };
  });
  
  // Update body class based on banner visibility
  $: if (typeof document !== 'undefined') {
    if (dev && !isHidden) {
      document.body.classList.add('dev-mode');
    } else {
      document.body.classList.remove('dev-mode');
    }
  }
  
  // Track navigation start time
  $: if ($navigating) {
    navigationStartTime = performance.now();
  }
  
  // Update page load time when navigation completes
  $: if (!$navigating && navigationStartTime > 0) {
    pageLoadTime = Math.round(performance.now() - navigationStartTime);
    navigationStartTime = 0;
    
    // Update memory usage on navigation
    if ('memory' in performance) {
      const mem = (performance as any).memory;
      memoryUsage = `${Math.round(mem.usedJSHeapSize / 1024 / 1024)}MB`;
    }
  }
</script>

{#if dev && !isHidden}
  <div class="dev-banner">
    <div class="dev-banner-content">
      <span class="dev-label">🛠️ DEV</span>
      <span class="dev-item">Route: <strong>{$page.url.pathname}</strong></span>
      <span class="dev-separator">•</span>
      <span class="dev-item">Load: <strong>{pageLoadTime}ms</strong></span>
      {#if memoryUsage}
        <span class="dev-separator">•</span>
        <span class="dev-item">Memory: <strong>{memoryUsage}</strong></span>
      {/if}
      {#if networkType}
        <span class="dev-separator">•</span>
        <span class="dev-item">Network: <strong>{networkType}</strong></span>
      {/if}
      <span class="dev-separator">•</span>
      <span class="dev-item">
        Errors: <strong class="error-count" class:has-errors={errorCount > 0}>{errorCount}</strong>
      </span>
      <span class="dev-separator">•</span>
      <span class="dev-item">
        Warnings: <strong class="warning-count" class:has-warnings={warningCount > 0}>{warningCount}</strong>
      </span>
      <span class="dev-separator">•</span>
      <span class="dev-item">User: <strong>{currentUser}</strong></span>
      <span class="dev-separator">•</span>
      <button 
        class="dev-button" 
        on:click={copyDebugInfo}
        title="Copy debug info to clipboard"
      >
        📋 Copy
      </button>
      {#if errorCount > 0 || warningCount > 0}
        <span class="dev-separator">•</span>
        <button 
          class="dev-button clear-button" 
          on:click={clearCounts}
          title="Clear error and warning counts"
        >
          🧹 Clear
        </button>
      {/if}
      <span class="dev-separator">•</span>
      <span class="dev-item dev-hint">Press ` to hide</span>
    </div>
  </div>
{/if}

<style>
  .dev-banner {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    z-index: 9999;
    background: linear-gradient(90deg, #660000, #800020, #660000);
    color: #f5f5f5;
    font-size: 11px;
    font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Roboto Mono', monospace;
    border-bottom: 1px solid rgba(255, 255, 255, 0.15);
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.4);
    height: 20px;
    overflow: hidden;
  }
  
  .dev-banner-content {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 0 12px;
    height: 100%;
  }
  
  .dev-label {
    background: rgba(255, 255, 255, 0.15);
    border: 1px solid rgba(255, 255, 255, 0.2);
    color: #f5f5f5;
    padding: 1px 6px;
    border-radius: 3px;
    font-weight: bold;
    flex-shrink: 0;
  }
  
  .dev-item {
    white-space: nowrap;
    flex-shrink: 0;
    color: #e0e0e0;
  }
  
  .dev-item strong {
    color: #c0c0c0;
    font-weight: normal;
  }
  
  .dev-hint {
    opacity: 0.6;
    font-size: 10px;
  }
  
  .error-count {
    color: #c0c0c0;
  }
  
  .error-count.has-errors {
    color: #ff6b6b;
    background: rgba(255, 107, 107, 0.1);
    padding: 0 3px;
    border-radius: 2px;
    border: 1px solid rgba(255, 107, 107, 0.3);
  }
  
  .warning-count {
    color: #c0c0c0;
  }
  
  .warning-count.has-warnings {
    color: #ffd93d;
    background: rgba(255, 217, 61, 0.1);
    padding: 0 3px;
    border-radius: 2px;
    border: 1px solid rgba(255, 217, 61, 0.3);
  }
  
  .clear-button {
    background: rgba(255, 255, 255, 0.15);
    border-color: rgba(255, 255, 255, 0.25);
  }
  
  .clear-button:hover {
    background: rgba(255, 255, 255, 0.25);
  }
  
  .dev-separator {
    opacity: 0.4;
    flex-shrink: 0;
    color: #c0c0c0;
  }
  
  .dev-button {
    background: rgba(255, 255, 255, 0.1);
    border: 1px solid rgba(255, 255, 255, 0.2);
    color: #f5f5f5;
    padding: 1px 4px;
    border-radius: 3px;
    cursor: pointer;
    transition: background-color 0.2s ease;
    font-size: 10px;
  }
  
  .dev-button:hover {
    background: rgba(255, 255, 255, 0.2);
  }
  
  /* Ensure content below banner doesn't get hidden */
  :global(body.dev-mode) {
    padding-top: 20px;
  }
</style>
