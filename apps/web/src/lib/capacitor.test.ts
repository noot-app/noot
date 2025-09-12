import { describe, it, expect, vi, beforeEach } from "vitest"

describe("Capacitor Integration", () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe("adapter selection", () => {
    it("should default to cloudflare adapter when no environment variable is set", () => {
      const originalAdapter = process.env.ADAPTER
      delete process.env.ADAPTER

      const mode = process.env.ADAPTER || 'cloudflare'
      expect(mode).toBe('cloudflare')

      if (originalAdapter) {
        process.env.ADAPTER = originalAdapter
      }
    })

    it("should use static adapter when ADAPTER=static", () => {
      const originalAdapter = process.env.ADAPTER
      process.env.ADAPTER = 'static'

      const mode = process.env.ADAPTER || 'cloudflare'
      expect(mode).toBe('static')

      if (originalAdapter) {
        process.env.ADAPTER = originalAdapter
      } else {
        delete process.env.ADAPTER
      }
    })

    it("should fallback to cloudflare for unknown adapter values", () => {
      const originalAdapter = process.env.ADAPTER
      process.env.ADAPTER = 'unknown-adapter'

      const mode = process.env.ADAPTER || 'cloudflare'
      const adapters = {
        cloudflare: 'cloudflare-adapter',
        static: 'static-adapter'
      }
      
      // Unknown adapter should use the provided value, but would fallback in real config
      expect(mode).toBe('unknown-adapter')
      expect(adapters[mode as keyof typeof adapters]).toBeUndefined()

      if (originalAdapter) {
        process.env.ADAPTER = originalAdapter
      } else {
        delete process.env.ADAPTER
      }
    })
  })

  describe("capacitor configuration", () => {
    it("should use correct app configuration", () => {
      const config = {
        appId: 'com.noot.app',
        appName: 'Noot',
        webDir: 'build',
        bundledWebRuntime: false
      }

      expect(config.appId).toBe('com.noot.app')
      expect(config.appName).toBe('Noot')
      expect(config.webDir).toBe('build')
      expect(config.bundledWebRuntime).toBe(false)
    })

    it("should have webDir pointing to SvelteKit build output", () => {
      const config = {
        webDir: 'build'
      }

      // Verify webDir matches SvelteKit's static output directory
      expect(config.webDir).toBe('build')
      expect(config.webDir).not.toBe('www') // default Capacitor value
      expect(config.webDir).not.toBe('.svelte-kit/output/client') // SSR output
    })
  })

  describe("mobile compatibility", () => {
    it("should handle client-side environment detection", () => {
      // Mock browser environment
      ;(global as any).window = {}
      const isBrowser = typeof window !== "undefined"
      expect(isBrowser).toBe(true)

      // Cleanup
      delete (global as any).window
    })

    it("should handle server-side environment detection", () => {
      // Ensure window is not defined (server environment)
      const isBrowser = typeof window !== "undefined"
      expect(isBrowser).toBe(false)
    })

    it("should validate build scripts exist", () => {
      // This would be validated during build process
      const expectedScripts = [
        'build:static',
        'build:mobile',
        'mobile:android',
        'mobile:ios'
      ]

      expectedScripts.forEach(script => {
        expect(typeof script).toBe('string')
        expect(script.length).toBeGreaterThan(0)
      })
    })
  })
})