import { vi, describe, it, expect, beforeEach } from "vitest"

vi.mock("@supabase/supabase-js")
vi.mock("$env/dynamic/private")
vi.mock("resend")

import { createClient, type User } from "@supabase/supabase-js"
import { Resend } from "resend"
import * as mailer from "./mailer"

describe("mailer", () => {
  const mockSend = vi.fn().mockResolvedValue({ id: "mock-email-id" })

  const mockSupabaseClient = {
    auth: {
      admin: {
        getUserById: vi.fn(),
      },
    },
    from: vi.fn().mockReturnThis(),
    select: vi.fn().mockReturnThis(),
    eq: vi.fn().mockReturnThis(),
    single: vi.fn(),
  }

  beforeEach(async () => {
    vi.clearAllMocks()
    const { env } = await import("$env/dynamic/private")
    env.PRIVATE_RESEND_API_KEY = "mock_resend_api_key"
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    ;(createClient as any).mockReturnValue(mockSupabaseClient)

    vi.mocked(Resend).mockImplementation(
      () =>
        ({
          emails: {
            send: mockSend,
          },
        }) as unknown as Resend,
    )
  })

  describe("sendUserEmail", () => {
    const mockUser = { id: "user123", email: "user@example.com" }

    it("logs welcome email as stub", async () => {
      const originalConsoleLog = console.log
      console.log = vi.fn()

      await mailer.sendUserEmail({
        user: mockUser as User,
        subject: "Test",
        _from_email: "test@example.com",
        _template_name: "welcome_email",
        _template_properties: {
          companyName: "Test Company",
          WebsiteBaseUrl: "https://test.com",
        },
      })

      expect(console.log).toHaveBeenCalledWith("User email (stub):", mockUser, "Test")
      expect(mockSend).not.toHaveBeenCalled()
      
      console.log = originalConsoleLog
    })

    it("logs user email as stub", async () => {
      const originalConsoleLog = console.log
      console.log = vi.fn()

      await mailer.sendUserEmail({
        user: mockUser as User,
        subject: "Test",
        _from_email: "test@example.com",
        _template_name: "welcome_email",
        _template_properties: {},
      })

      expect(mockSend).not.toHaveBeenCalled()

      expect(console.log).toHaveBeenCalledWith(
        "User email (stub):",
        mockUser,
        "Test",
      )

      console.log = originalConsoleLog
    })
  })

  describe("sendTemplatedEmail", () => {
    it("logs templated email as stub", async () => {
      const originalConsoleLog = console.log
      console.log = vi.fn()

      await mailer.sendTemplatedEmail({
        subject: "Test subject",
        _from_email: "from@example.com",
        to_emails: ["to@example.com"],
        _template_name: "welcome_email",
        _template_properties: {
          companyName: "Test Company",
          WebsiteBaseUrl: "https://test.com",
        },
      })

      expect(console.log).toHaveBeenCalledWith("Templated email (stub):", "Test subject", ["to@example.com"])
      expect(mockSend).not.toHaveBeenCalled()
      
      console.log = originalConsoleLog
    })
  })
})
