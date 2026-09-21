import { env } from '@/env';

function isLocalHost(url: string) {
  return /localhost|127\.0\.0\.1/i.test(url);
}

function readOidcQueryError(): { error?: string; description?: string } {
  try {
    const params = new URLSearchParams(window.location.search);
    const error = params.get('error') ?? undefined;
    const description = params.get('error_description') ?? undefined;
    return { error, description };
  } catch {
    return {};
  }
}

export function getAuthConfigHint(authErrorMessage?: string): string {
  const authUrl = env.VITE_AUTH_URL.replace(/\/+$/, '');
  const apiUrl = env.VITE_API_URL;
  const pageIsLocal = isLocalHost(window.location.hostname);
  const authIsLocal = isLocalHost(authUrl);
  const oidc = `${authUrl}/realms/luxus/.well-known/openid-configuration`;
  const sameOriginAuth =
    !pageIsLocal && authUrl.startsWith(window.location.origin);
  const query = readOidcQueryError();
  const combined = `${authErrorMessage ?? ''} ${query.error ?? ''} ${query.description ?? ''}`.toLowerCase();

  if (combined.includes('invalid_scope') || combined.includes('invalid scopes')) {
    return [
      'O Keycloak rejeitou o scope pedido pelo frontend (ex.: tenant-organization).',
      'O proxy /auth está OK — o realm/client em produção está desatualizado em relação ao luxus-realm.json.',
      '',
      'Na VPS, rode:',
      '  bash docker/keycloak/ensure-org-id-mapper.sh',
      'Depois faça logout/login (ou rebuild do connect-web se o scope no bundle mudou).',
      '',
      `Erro Keycloak: ${query.description ?? authErrorMessage ?? query.error ?? 'invalid_scope'}`,
      `Redirect URI: ${window.location.origin}/*`
    ].join('\n');
  }

  if (
    combined.includes('no matching state') ||
    combined.includes('state does not match') ||
    combined.includes('stale state')
  ) {
    return [
      'O state OIDC da sessão anterior expirou ou foi limpo (aba fechada, storage limpo, ou login duplicado).',
      'Use "Fazer Login Novamente" — isso limpa a URL e o sessionStorage e reinicia o fluxo.',
      '',
      `Redirect URI no Keycloak: ${window.location.origin}/*`
    ].join('\n');
  }

  if (authIsLocal && !pageIsLocal) {
    return [
      'O frontend em produção foi compilado com URLs de desenvolvimento.',
      `VITE_AUTH_URL atual: ${authUrl}`,
      `VITE_API_URL atual: ${apiUrl}`,
      '',
      'Recompile o connect-web com:',
      'VITE_AUTH_URL=https://telefonia.redobrai.online/auth',
      'VITE_API_URL=https://telefonia.redobrai.online/api',
      '',
      `No Keycloak, redirect URI: ${window.location.origin}/*`
    ].join('\n');
  }

  if (pageIsLocal) {
    return [
      'Confirme no ficheiro src/Luxus.Connect.Web/.env:',
      'VITE_AUTH_URL=http://localhost:8081/auth',
      'VITE_API_URL=http://localhost:8002',
      'Keycloak deve estar a correr (docker compose) com KC_HTTP_RELATIVE_PATH=/auth.'
    ].join('\n');
  }

  if (sameOriginAuth) {
    return [
      'Se o endpoint OIDC devolver HTML em vez de JSON, o /auth não está proxied ao Keycloak.',
      'Confirme primeiro:',
      `  curl -sI "${oidc}"  → Content-Type: application/json`,
      '',
      `VITE_AUTH_URL=${authUrl}`,
      `VITE_API_URL=${apiUrl}`,
      '',
      'Na VPS:',
      '1. Keycloak com KC_HTTP_RELATIVE_PATH=/auth (reinicie o container)',
      '2. sudo bash scripts/setup-vps-nginx.sh',
      '3. bash docker/keycloak/ensure-org-id-mapper.sh',
      '4. Rebuild do connect-web se VITE_* mudou',
      '',
      `Redirect URI no Keycloak: ${window.location.origin}/*`
    ].join('\n');
  }

  return [
    'Verifique se o Keycloak está acessível a partir do browser:',
    `VITE_AUTH_URL=${authUrl}`,
    `VITE_API_URL=${apiUrl}`,
    `Endpoint OIDC: ${oidc}`,
    '',
    `Redirect URI no Keycloak: ${window.location.origin}/*`,
    'Após alterar VITE_*, é obrigatório rebuild/redeploy do connect-web.'
  ].join('\n');
}
