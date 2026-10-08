import DOMPurify from 'dompurify';

// Sanitize at display time as well, including contracts stored before the fix.
export function sanitizeContractHtml(html: string): string {
  return DOMPurify.sanitize(html, {
    USE_PROFILES: { html: true },
    FORBID_TAGS: ['style', 'form', 'input', 'button', 'iframe'],
    FORBID_ATTR: ['style']
  });
}
