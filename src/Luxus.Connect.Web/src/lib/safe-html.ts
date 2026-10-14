import DOMPurify from 'dompurify';

// Prevent reverse-tabnabbing: any anchor that opens a new tab must carry rel=noopener.
DOMPurify.addHook('afterSanitizeAttributes', (node) => {
  if (node.tagName === 'A' && node.getAttribute('target')) {
    node.setAttribute('rel', 'noopener noreferrer');
  }
});

// Sanitize at display time as well, including contracts stored before the fix.
export function sanitizeContractHtml(html: string): string {
  return DOMPurify.sanitize(html, {
    USE_PROFILES: { html: true },
    FORBID_TAGS: ['style', 'form', 'input', 'button', 'iframe'],
    FORBID_ATTR: ['style']
  });
}
