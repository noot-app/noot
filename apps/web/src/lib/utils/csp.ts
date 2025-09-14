/**
 * CSP (Content Security Policy) utilities for secure content handling
 */

/**
 * Safely creates an inline script with proper nonce attribute
 * Use this when you need to create dynamic scripts in Svelte components
 * 
 * @param scriptContent - The JavaScript code to execute
 * @param nonce - The CSP nonce value from layout data
 * @returns HTMLElement that can be safely inserted into the DOM
 */
export function createNoncedScript(scriptContent: string, nonce: string): HTMLElement {
  const script = document.createElement('script')
  script.setAttribute('nonce', nonce)
  script.textContent = scriptContent
  return script
}

/**
 * Safely creates an inline style with proper nonce attribute
 * Use this when you need to create dynamic styles in Svelte components
 * 
 * @param cssContent - The CSS rules to apply
 * @param nonce - The CSP nonce value from layout data
 * @returns HTMLElement that can be safely inserted into the DOM
 */
export function createNoncedStyle(cssContent: string, nonce: string): HTMLElement {
  const style = document.createElement('style')
  style.setAttribute('nonce', nonce)
  style.textContent = cssContent
  return style
}

/**
 * Safely sets innerHTML on an element, ensuring any inline scripts/styles have nonces
 * Use this instead of directly setting innerHTML when content might contain scripts/styles
 * 
 * @param element - The target element
 * @param htmlContent - The HTML content to set
 * @param nonce - The CSP nonce value from layout data
 */
export function setSafeInnerHTML(element: HTMLElement, htmlContent: string, nonce: string): void {
  // First set the content
  element.innerHTML = htmlContent
  
  // Then add nonces to any inline scripts or styles
  const scripts = element.querySelectorAll('script:not([nonce])')
  const styles = element.querySelectorAll('style:not([nonce])')
  
  scripts.forEach(script => script.setAttribute('nonce', nonce))
  styles.forEach(style => style.setAttribute('nonce', nonce))
}

/**
 * Validates that a nonce is properly formatted (base64)
 * Useful for debugging CSP issues
 * 
 * @param nonce - The nonce to validate
 * @returns boolean indicating if nonce is valid
 */
export function isValidNonce(nonce: string): boolean {
  if (!nonce || typeof nonce !== 'string') return false
  
  // Base64 pattern: letters, numbers, +, /, with optional = padding
  const base64Pattern = /^[A-Za-z0-9+/]*={0,2}$/
  return base64Pattern.test(nonce) && nonce.length >= 8
}

/**
 * Development helper for CSP debugging
 * Logs CSP violations and provides debugging information
 * Only active in development mode
 */
export function enableCSPDebugging(): void {
  if (typeof window === 'undefined' || import.meta.env.MODE === 'production') return
  
  document.addEventListener('securitypolicyviolation', (event) => {
    console.group('🚨 CSP Violation Detected')
    console.error('Blocked URI:', event.blockedURI)
    console.error('Violated Directive:', event.violatedDirective)
    console.error('Original Policy:', event.originalPolicy)
    console.error('Source File:', event.sourceFile)
    console.error('Line Number:', event.lineNumber)
    console.groupEnd()
    
    // Provide helpful suggestions
    if (event.violatedDirective.includes('script-src')) {
      console.warn('💡 Script blocked - ensure inline scripts have nonce attribute or use createNoncedScript()')
    }
    if (event.violatedDirective.includes('style-src')) {
      console.warn('💡 Style blocked - ensure inline styles have nonce attribute or use createNoncedStyle()')
    }
  })
  
  console.log('🛡️ CSP debugging enabled - violations will be logged to console')
}