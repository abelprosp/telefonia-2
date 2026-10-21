/* global DOMException, globalThis */
import assert from 'node:assert/strict';
import { after, test } from 'node:test';
import { JSDOM } from 'jsdom';
import {
  isExpiredSession,
  authNetworkMessage
} from '../src/lib/auth-response.ts';

test('expired sessions return to login while invalid recovery codes and passwords remain retryable', () => {
  assert.equal(isExpiredSession(401, 'UNAUTHORIZED'), true);
  assert.equal(isExpiredSession(401, 'INVALID_RECOVERY_CODE'), false);
  assert.equal(isExpiredSession(401, 'INVALID_CREDENTIALS'), false);
  assert.equal(isExpiredSession(429, 'MFA_LIMIT'), false);
  assert.equal(isExpiredSession(503, 'AUTH_UNAVAILABLE'), false);
});

test('authentication connection failures provide a useful Portuguese message', () => {
  assert.match(
    authNetworkMessage(new DOMException('', 'TimeoutError')),
    /demorou/
  );
  assert.match(
    authNetworkMessage(new TypeError('Failed to fetch')),
    /Sem conexão/
  );
});

const dom = new JSDOM('<!doctype html><html><body></body></html>');
globalThis.window = dom.window;
const { sanitizeContractHtml } = await import('../src/lib/safe-html.ts');
after(() => dom.window.close());

test('stored contracts cannot execute scripts, handlers, SVG or javascript links', () => {
  const result = sanitizeContractHtml(
    '<script>alert(1)</script><img src=x onerror="alert(2)"><svg onload="alert(3)"></svg><a href="javascript:alert(4)">link</a>'
  );
  const doc = new JSDOM(result).window.document;
  assert.equal(
    doc.querySelector('script,svg,[onerror],[onload],a[href^="javascript:"]'),
    null
  );
});

test('contract forms, iframes and style injection are removed', () => {
  const result = sanitizeContractHtml(
    '<style>body{display:none}</style><form action="https://evil.test"><input></form><iframe src="https://evil.test"></iframe><p style="position:fixed">Contrato</p>'
  );
  assert.doesNotMatch(result, /<(style|form|input|iframe)\b|\bstyle=/i);
  assert.match(result, /Contrato/);
});

test('safe contract text and billing tables remain readable', () => {
  const result = sanitizeContractHtml(
    '<h2>Contrato</h2><p>João &amp; Maria</p><table><tr><td>Plano</td><td>R$ 10,00</td></tr></table>'
  );
  assert.match(result, /<h2>Contrato<\/h2>/);
  assert.match(result, /João &amp; Maria/);
  assert.match(result, /<td>R\$ 10,00<\/td>/);
});
