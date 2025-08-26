// Stub mailer for noot - not needed for PoC
// This is a simplified version that doesn't use Supabase

export const sendAdminEmail = async ({
  subject,
  body,
}: {
  subject: string
  body: string
}) => {
  console.log("Admin email (stub):", subject, body)
  // No-op for PoC
}

export const sendUserEmail = async ({
  user,
  subject,
  _from_email,
  _template_name,
  _template_properties,
}: {
  user: any
  subject: string
  _from_email: string
  _template_name: string
  _template_properties: Record<string, string>
}) => {
  console.log("User email (stub):", user, subject)
  // No-op for PoC
}

export const sendTemplatedEmail = async ({
  subject,
  to_emails,
  _from_email,
  _template_name,
  _template_properties,
}: {
  subject: string
  to_emails: string[]
  _from_email: string
  _template_name: string
  _template_properties: Record<string, string>
}) => {
  console.log("Templated email (stub):", subject, to_emails)
  // No-op for PoC
}