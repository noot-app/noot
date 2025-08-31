import { describe, it, expect, vi, beforeEach } from "vitest"
import { getValidatedSession } from "./utils"

const mockSupabase: any = {
  auth: {
    getSession: vi.fn(),
    getClaims: vi.fn(),
  },
}

describe("getValidatedSession", () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it("returns null when there is no session", async () => {
    mockSupabase.auth.getSession.mockResolvedValue({ data: { session: null } })
    const res = await getValidatedSession(mockSupabase)
    expect(res).toBeNull()
    expect(mockSupabase.auth.getClaims).not.toHaveBeenCalled()
  })

  it("returns a normalized session when claims are valid", async () => {
    const rawSession = { access_token: "token", refresh_token: "r" }
    mockSupabase.auth.getSession.mockResolvedValue({
      data: { session: rawSession },
    })
    mockSupabase.auth.getClaims.mockResolvedValue({
      data: {
        claims: {
          exp: Math.round(Date.now() / 1000) + 3600,
          sub: "uid",
          email: "e@example.com",
          phone: null,
          user_metadata: {},
          app_metadata: {},
          is_anonymous: false,
        },
      },
      error: null,
    })

    const res = await getValidatedSession(mockSupabase)
    expect(res).toBeTruthy()
    expect(res?.access_token).toBe("token")
    expect(res?.refresh_token).toBe("r")
    expect(res?.user.id).toBe("uid")
    expect(res?.user.email).toBe("e@example.com")
  })

  it("returns null when getClaims returns error", async () => {
    mockSupabase.auth.getSession.mockResolvedValue({
      data: { session: { access_token: "token", refresh_token: "r" } },
    })
    mockSupabase.auth.getClaims.mockResolvedValue({
      data: null,
      error: { message: "bad" },
    })
    const res = await getValidatedSession(mockSupabase)
    expect(res).toBeNull()
  })

  it("returns null when getClaims throws", async () => {
    mockSupabase.auth.getSession.mockResolvedValue({
      data: { session: { access_token: "token", refresh_token: "r" } },
    })
    mockSupabase.auth.getClaims.mockRejectedValue(new Error("boom"))
    const res = await getValidatedSession(mockSupabase)
    expect(res).toBeNull()
  })
})
