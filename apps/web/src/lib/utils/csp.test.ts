import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { 
  createNoncedScript, 
  createNoncedStyle, 
  setSafeInnerHTML, 
  isValidNonce, 
  enableCSPDebugging 
} from "./csp"

// Mock import.meta.env
vi.mock("import.meta.env", () => ({
  MODE: "test"
}))

// Mock SecurityPolicyViolationEvent for JSDOM environment
class MockSecurityPolicyViolationEvent extends Event {
  blockedURI: string
  violatedDirective: string
  originalPolicy: string
  sourceFile: string
  lineNumber: number

  constructor(type: string, init: any) {
    super(type)
    this.blockedURI = init.blockedURI || ""
    this.violatedDirective = init.violatedDirective || ""
    this.originalPolicy = init.originalPolicy || ""
    this.sourceFile = init.sourceFile || ""
    this.lineNumber = init.lineNumber || 0
  }
}

// Add to global for tests
;(globalThis as any).SecurityPolicyViolationEvent = MockSecurityPolicyViolationEvent

describe("CSP Utilities", () => {
  beforeEach(() => {
    // Clear any existing event listeners
    document.removeEventListener("securitypolicyviolation", vi.fn())
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  describe("createNoncedScript", () => {
    it("creates script element with nonce attribute", () => {
      const nonce = "abc123def456"
      const scriptContent = "console.log('test')"
      
      const script = createNoncedScript(scriptContent, nonce)
      
      expect(script.tagName).toBe("SCRIPT")
      expect(script.getAttribute("nonce")).toBe(nonce)
      expect(script.textContent).toBe(scriptContent)
    })

    it("handles empty script content", () => {
      const nonce = "abc123def456"
      const script = createNoncedScript("", nonce)
      
      expect(script.getAttribute("nonce")).toBe(nonce)
      expect(script.textContent).toBe("")
    })
  })

  describe("createNoncedStyle", () => {
    it("creates style element with nonce attribute", () => {
      const nonce = "abc123def456"
      const cssContent = "body { color: red; }"
      
      const style = createNoncedStyle(cssContent, nonce)
      
      expect(style.tagName).toBe("STYLE")
      expect(style.getAttribute("nonce")).toBe(nonce)
      expect(style.textContent).toBe(cssContent)
    })

    it("handles empty CSS content", () => {
      const nonce = "abc123def456"
      const style = createNoncedStyle("", nonce)
      
      expect(style.getAttribute("nonce")).toBe(nonce)
      expect(style.textContent).toBe("")
    })
  })

  describe("setSafeInnerHTML", () => {
    it("sets innerHTML and adds nonces to inline scripts", () => {
      const container = document.createElement("div")
      const nonce = "abc123def456"
      const htmlContent = '<script>alert("test")</script><p>Content</p>'
      
      setSafeInnerHTML(container, htmlContent, nonce)
      
      const script = container.querySelector("script")
      expect(script?.getAttribute("nonce")).toBe(nonce)
      expect(script?.textContent).toBe('alert("test")')
      expect(container.querySelector("p")?.textContent).toBe("Content")
    })

    it("sets innerHTML and adds nonces to inline styles", () => {
      const container = document.createElement("div")
      const nonce = "abc123def456"
      const htmlContent = '<style>body { color: blue; }</style><p>Content</p>'
      
      setSafeInnerHTML(container, htmlContent, nonce)
      
      const style = container.querySelector("style")
      expect(style?.getAttribute("nonce")).toBe(nonce)
      expect(style?.textContent).toBe("body { color: blue; }")
    })

    it("does not override existing nonces", () => {
      const container = document.createElement("div")
      const nonce = "abc123def456"
      const existingNonce = "existing123"
      const htmlContent = `<script nonce="${existingNonce}">alert("test")</script>`
      
      setSafeInnerHTML(container, htmlContent, nonce)
      
      const script = container.querySelector("script")
      expect(script?.getAttribute("nonce")).toBe(existingNonce)
    })

    it("handles mixed content with scripts and styles", () => {
      const container = document.createElement("div")
      const nonce = "abc123def456"
      const htmlContent = `
        <script>alert("test1")</script>
        <style>body { color: red; }</style>
        <script nonce="existing">alert("test2")</script>
        <p>Content</p>
      `
      
      setSafeInnerHTML(container, htmlContent, nonce)
      
      const scripts = container.querySelectorAll("script")
      const styles = container.querySelectorAll("style")
      
      expect(scripts[0]?.getAttribute("nonce")).toBe(nonce)
      expect(scripts[1]?.getAttribute("nonce")).toBe("existing")
      expect(styles[0]?.getAttribute("nonce")).toBe(nonce)
    })
  })

  describe("isValidNonce", () => {
    it("validates correct base64 nonces", () => {
      expect(isValidNonce("abc123def456")).toBe(true)
      expect(isValidNonce("ABC123DEF456")).toBe(true)
      expect(isValidNonce("abc+def/123=")).toBe(true)
      expect(isValidNonce("abc+def/123==")).toBe(true)
      expect(isValidNonce("abcdefgh")).toBe(true)
    })

    it("rejects invalid nonces", () => {
      expect(isValidNonce("")).toBe(false)
      expect(isValidNonce("abc")).toBe(false) // Too short
      expect(isValidNonce("abc@def")).toBe(false) // Invalid character
      expect(isValidNonce("abc def")).toBe(false) // Space not allowed
      expect(isValidNonce("abc===")).toBe(false) // Too many padding chars
    })

    it("handles non-string inputs", () => {
      expect(isValidNonce(null as any)).toBe(false)
      expect(isValidNonce(undefined as any)).toBe(false)
      expect(isValidNonce(123 as any)).toBe(false)
      expect(isValidNonce({} as any)).toBe(false)
    })
  })

  describe("enableCSPDebugging", () => {
    let consoleLogSpy: any
    let consoleGroupSpy: any
    let consoleErrorSpy: any
    let consoleWarnSpy: any
    let consoleGroupEndSpy: any
    let addEventListenerSpy: any

    beforeEach(() => {
      consoleLogSpy = vi.spyOn(console, "log").mockImplementation(() => {})
      consoleGroupSpy = vi.spyOn(console, "group").mockImplementation(() => {})
      consoleErrorSpy = vi.spyOn(console, "error").mockImplementation(() => {})
      consoleWarnSpy = vi.spyOn(console, "warn").mockImplementation(() => {})
      consoleGroupEndSpy = vi.spyOn(console, "groupEnd").mockImplementation(() => {})
      addEventListenerSpy = vi.spyOn(document, "addEventListener")
    })

    it("sets up CSP violation event listener in non-production", () => {
      enableCSPDebugging()
      
      expect(addEventListenerSpy).toHaveBeenCalledWith(
        "securitypolicyviolation",
        expect.any(Function)
      )
      expect(consoleLogSpy).toHaveBeenCalledWith(
        expect.stringContaining("CSP debugging enabled")
      )
    })

    it("does not run in production mode", () => {
      vi.mocked(import.meta.env).MODE = "production"
      
      enableCSPDebugging()
      
      expect(addEventListenerSpy).not.toHaveBeenCalled()
      expect(consoleLogSpy).not.toHaveBeenCalled()
    })

    it("logs CSP violations with detailed information", () => {
      // Mock browser environment for this test
      vi.mocked(import.meta.env).MODE = "development"
      
      enableCSPDebugging()
      
      // Verify event listener was added
      expect(addEventListenerSpy).toHaveBeenCalledWith(
        "securitypolicyviolation",
        expect.any(Function)
      )
      
      // Get the actual event handler that was registered
      const eventHandler = addEventListenerSpy.mock.calls[0][1] as EventListener
      
      // Simulate a CSP violation event
      const violationEvent = new MockSecurityPolicyViolationEvent("securitypolicyviolation", {
        blockedURI: "inline",
        violatedDirective: "script-src",
        originalPolicy: "script-src 'self'",
        sourceFile: "test.html",
        lineNumber: 10
      })

      // Call the handler directly
      eventHandler(violationEvent as any)
      
      expect(consoleGroupSpy).toHaveBeenCalledWith("🚨 CSP Violation Detected")
      expect(consoleErrorSpy).toHaveBeenCalledWith("Blocked URI:", "inline")
      expect(consoleErrorSpy).toHaveBeenCalledWith("Violated Directive:", "script-src")
      expect(consoleGroupEndSpy).toHaveBeenCalled()
    })

    it("provides helpful suggestions for script-src violations", () => {
      vi.mocked(import.meta.env).MODE = "development"
      
      enableCSPDebugging()
      
      const eventHandler = addEventListenerSpy.mock.calls[0][1] as EventListener
      const violationEvent = new MockSecurityPolicyViolationEvent("securitypolicyviolation", {
        violatedDirective: "script-src 'self'",
        blockedURI: "inline"
      })

      eventHandler(violationEvent as any)
      
      expect(consoleWarnSpy).toHaveBeenCalledWith(
        expect.stringContaining("Script blocked - ensure inline scripts have nonce")
      )
    })

    it("provides helpful suggestions for style-src violations", () => {
      vi.mocked(import.meta.env).MODE = "development"
      
      enableCSPDebugging()
      
      const eventHandler = addEventListenerSpy.mock.calls[0][1] as EventListener
      const violationEvent = new MockSecurityPolicyViolationEvent("securitypolicyviolation", {
        violatedDirective: "style-src 'self'",
        blockedURI: "inline"
      })

      eventHandler(violationEvent as any)
      
      expect(consoleWarnSpy).toHaveBeenCalledWith(
        expect.stringContaining("Style blocked - ensure inline styles have nonce")
      )
    })
  })
})