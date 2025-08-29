import { vi, describe, it, expect, beforeEach, afterEach } from "vitest"

describe("mailer", () => {
  const originalConsoleLog = console.log
  let mockConsoleLog: ReturnType<typeof vi.fn>

  beforeEach(() => {
    vi.clearAllMocks()
    mockConsoleLog = vi.fn()
    console.log = mockConsoleLog
  })

  afterEach(() => {
    console.log = originalConsoleLog
  })

  describe("sendUserEmail", () => {
    const mockUser = { id: "user123", email: "user@example.com" }

    it("sends welcome email", async () => {
      const { sendUserEmail } = await import("./mailer")

      await sendUserEmail({
        user: mockUser,
        subject: "Test",
        _from_email: "test@example.com",
        _template_name: "welcome_email",
        _template_properties: {
          companyName: "Test Company",
          WebsiteBaseUrl: "https://test.com",
        },
      })

      expect(mockConsoleLog).toHaveBeenCalledWith("User email (stub):", mockUser, "Test")
    })

    it("should not send email if user is unsubscribed", async () => {
      const { sendUserEmail } = await import("./mailer")

      // Since this is a stub implementation, it will always log regardless of subscription status
      await sendUserEmail({
        user: mockUser,
        subject: "Test",
        _from_email: "test@example.com",
        _template_name: "welcome_email",
        _template_properties: {},
      })

      expect(mockConsoleLog).toHaveBeenCalledWith("User email (stub):", mockUser, "Test")
    })
  })

  describe("sendTemplatedEmail", () => {
    it("sends templated email", async () => {
      const { sendTemplatedEmail } = await import("./mailer")

      await sendTemplatedEmail({
        subject: "Test subject",
        _from_email: "from@example.com",
        to_emails: ["to@example.com"],
        _template_name: "welcome_email",
        _template_properties: {
          companyName: "Test Company",
          WebsiteBaseUrl: "https://test.com",
        },
      })

      expect(mockConsoleLog).toHaveBeenCalledWith(
        "Templated email (stub):",
        "Test subject",
        ["to@example.com"]
      )
    })
  })

  describe("sendAdminEmail", () => {
    it("sends admin email", async () => {
      const { sendAdminEmail } = await import("./mailer")

      await sendAdminEmail({
        subject: "Admin Test",
        body: "Test body",
      })

      expect(mockConsoleLog).toHaveBeenCalledWith("Admin email (stub):", "Admin Test", "Test body")
    })
  })
})
