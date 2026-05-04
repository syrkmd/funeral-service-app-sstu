/**
 * Normalize phone number by removing spaces, dashes, and other formatting
 * Example: "+7 (999) 123-45-67" → "+79991234567"
 */
export function normalizePhone(phone: string): string {
  return phone.replace(/[\s\-\(\)]/g, '');
}
