/**
 * Legal text summaries and configurations
 * Update these when Terms of Service or Privacy Policy change
 */

export const LEGAL_SUMMARIES = {
  termsOfService: {
    summary: `By using Noot, you agree to: use the service for personal, non-commercial purposes; maintain accurate account information; not misuse or abuse the service; and understand that AI-generated content may contain inaccuracies. The service is provided "as is" with no warranties, and Noot limits its liability for damages. You also agree to resolve disputes through arbitration.`,
    keyPoints: [
      'Use service for personal, non-commercial purposes only',
      'Maintain accurate account information',
      'AI-generated content may contain inaccuracies - verify information',
      'Service provided "as is" without warranties',
      'Not medical advice - consult healthcare professionals',
      'Disputes resolved through arbitration',
      'Limited liability for damages'
    ],
    url: '/terms-of-service',
    title: 'Terms of Service'
  },
  privacyPolicy: {
    summary: `We collect your account information, audio recordings, and usage data to provide our nutrition tracking service. Your information is processed to generate nutrition analysis and improve our services. We share data with service providers like Stripe, Supabase, and AI transcription services, but we do not sell your personal information. You have rights to access, correct, or delete your data.`,
    keyPoints: [
      'We collect account info, audio recordings, and usage data',
      'Data used to provide nutrition tracking and improve services',
      'Shared with service providers (Stripe, Supabase, AI services)',
      'We do not sell your personal information',
      'You can access, correct, or delete your data',
      'Data encrypted in transit and at rest',
      'Not HIPAA covered - not medical advice'
    ],
    url: '/privacy-policy',
    title: 'Privacy Policy'
  }
} as const
