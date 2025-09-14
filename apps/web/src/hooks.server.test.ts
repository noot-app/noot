import { describe, it, expect, vi, beforeEach } from "vitest"

// Mock env so Supabase SSR path runs without warnings
vi.mock("$env/dynamic/public", () => ({
  env: {
    PUBLIC_SUPABASE_URL: "http://localhost:54321",
    PUBLIC_SUPABASE_ANON_KEY: "anon",
  },
}))

// Minimal mock for createServerClient; we mostly assert locals wiring
const { mockCreateServerClient } = vi.hoisted(() => ({
  mockCreateServerClient: vi.fn().mockReturnValue({ auth: {} }),
}))
vi.mock("@supabase/ssr", () => ({
  createServerClient: mockCreateServerClient,
}))

// Mock getValidatedSession to control session presence
const { mockGetValidatedSession } = vi.hoisted(() => ({
  mockGetValidatedSession: vi.fn(),
}))
vi.mock("$lib/utils.js", () => ({
  getValidatedSession: mockGetValidatedSession,
}))

// Import after mocks
import { handle } from "./hooks.server"

function makeEvent(pathname: string) {
  const cookies: any = {
    getAll: vi.fn().mockReturnValue([]),
    set: vi.fn(),
  }
  const event: any = {
    url: new URL(`http://example.com${pathname}`),
    cookies,
    locals: {},
    setHeaders: vi.fn(),
  }
  return event
}

describe("hooks.server handle", () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it("allows public routes without session", async () => {
    mockGetValidatedSession.mockResolvedValueOnce(null)
    const event = makeEvent("/")
    const resolve = vi.fn().mockResolvedValue({ ok: true })

    const res = await handle({ event, resolve } as any)

    expect(mockCreateServerClient).toHaveBeenCalled()
    expect(typeof event.locals.getSession).toBe("function")
    expect(typeof event.locals.cspNonce).toBe("string")
    expect(event.locals.cspNonce).toMatch(/^[A-Za-z0-9+/]+=*$/) // Base64 pattern
    expect(resolve).toHaveBeenCalledWith(
      event,
      expect.objectContaining({
        filterSerializedResponseHeaders: expect.any(Function),
        transformPageChunk: expect.any(Function),
      }),
    )
    expect(res).toEqual({ ok: true })
  })

  it("generates unique CSP nonces for each request", async () => {
    mockGetValidatedSession.mockResolvedValue(null)
    
    const event1 = makeEvent("/")
    const event2 = makeEvent("/")
    const resolve = vi.fn().mockResolvedValue({ ok: true })

    await handle({ event: event1, resolve } as any)
    await handle({ event: event2, resolve } as any)

    expect(event1.locals.cspNonce).toBeTruthy()
    expect(event2.locals.cspNonce).toBeTruthy()
    expect(event1.locals.cspNonce).not.toBe(event2.locals.cspNonce)
  })

  it("redirects unauthenticated users from protected paths", async () => {
    mockGetValidatedSession.mockResolvedValueOnce(null)
    const event = makeEvent("/profile")
    const resolve = vi.fn()

    try {
      await handle({ event, resolve } as any)
      throw new Error("Expected redirect to be thrown")
    } catch (e: any) {
      expect(e.status).toBe(303)
      expect(decodeURIComponent(e.location)).toBe("/login?redirect=/profile")
    }
  })

  it("allows authenticated users on protected paths", async () => {
    mockGetValidatedSession.mockResolvedValueOnce({
      access_token: "t",
      refresh_token: "r",
      user: { id: "u" },
    } as any)
    const event = makeEvent("/record/anything")
    const resolve = vi.fn().mockResolvedValue({ ok: true })

    const res = await handle({ event, resolve } as any)

    expect(resolve).toHaveBeenCalled()
    expect(res).toEqual({ ok: true })
  })

  it("exposes locals.getSession that returns the validated session", async () => {
    const fakeSession = {
      access_token: "t",
      refresh_token: "r",
      user: { id: "u" },
    } as any
    // handle() calls getSession once internally; mock for all calls
    mockGetValidatedSession.mockResolvedValue(fakeSession)
    const event = makeEvent("/")
    const resolve = vi.fn().mockResolvedValue({ ok: true })

    await handle({ event, resolve } as any)
    const session = await event.locals.getSession()
    expect(session).toEqual(fakeSession)
  })

  it("filterSerializedResponseHeaders only allows specific headers", async () => {
    mockGetValidatedSession.mockResolvedValueOnce(null)
    const event = makeEvent("/")
    let optionsArg: any
    const resolve = vi.fn().mockImplementation((_event: any, options?: any) => {
      optionsArg = options
      return { ok: true }
    })

    await handle({ event, resolve } as any)

    expect(typeof optionsArg.filterSerializedResponseHeaders).toBe("function")
    const filter = optionsArg.filterSerializedResponseHeaders as (
      name: string,
    ) => boolean
    expect(filter("content-range")).toBe(true)
    expect(filter("x-supabase-api-version")).toBe(true)
    expect(filter("set-cookie")).toBe(false)
  })

  it("injects nonce into scripts and styles and sets CSP headers in production", async () => {
    // Mock production environment
    vi.doMock("$env/dynamic/public", () => ({
      env: {
        PUBLIC_SUPABASE_URL: "http://localhost:54321",
        PUBLIC_SUPABASE_ANON_KEY: "anon",
        PUBLIC_NODE_ENV: "production",
      },
    }))

    const { handle: productionHandle } = await import("./hooks.server")

    mockGetValidatedSession.mockResolvedValueOnce(null)
    const event = makeEvent("/")
    event.setHeaders = vi.fn()
    
    let transformPageChunk: any
    const resolve = vi.fn().mockImplementation((_event: any, options?: any) => {
      transformPageChunk = options?.transformPageChunk
      return { ok: true }
    })

    await productionHandle({ event, resolve } as any)

    // Test HTML transformation with nonce injection
    const testHTML = '<script>console.log("test")</script><style>body{}</style>'
    const transformedHTML = transformPageChunk({ html: testHTML, done: true })
    
    expect(transformedHTML).toContain(`nonce="${event.locals.cspNonce}"`)
    expect(transformedHTML).toMatch(/<script[^>]*nonce="[^"]+">/)
    expect(transformedHTML).toMatch(/<style[^>]*nonce="[^"]+">/)

    // Verify CSP headers are set with nonce
    expect(event.setHeaders).toHaveBeenCalledWith(
      expect.objectContaining({
        'Content-Security-Policy': expect.stringContaining(`'nonce-${event.locals.cspNonce}'`)
      })
    )

    // Verify 'unsafe-inline' is removed from CSP
    const cspCall = (event.setHeaders as any).mock.calls[0][0]['Content-Security-Policy']
    expect(cspCall).not.toContain("'unsafe-inline'")
    expect(cspCall).toContain(`'nonce-${event.locals.cspNonce}'`)
  })
})
