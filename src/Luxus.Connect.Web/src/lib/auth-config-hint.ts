import { env } from '@/env';

function isLocalHost(url: string) {
  return /localhost|127\.0\.0\.1/i.test(url);
}

export function getAuthConfigHint(): string {
  const authUrl = env.VITE_AUTH_URL.replace(/\/+$/, '');
  const apiUrl = env.VITE_API_URL;
  const pageIsLocal = isLocalHost(window.location.hostname);
  const authIsLocal = isLocalHost(authUrl);
  const oidc = `${authUrl}/realms/luxus/.well-known/openid-configuration`;
  const sameOriginAuth =
    !pageIsLocal && authUrl.startsWith(window.location.origin);

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
      'O browser pediu o OpenID do Keycloak neste domínio e recebeu HTML do frontend.',
      'Isso acontece quando /auth não é proxied para o Keycloak.',
      '',
      `VITE_AUTH_URL=${authUrl}`,
      `VITE_API_URL=${apiUrl}`,
      `Endpoint OIDC: ${oidc}`,
      '',
      'Na VPS:',
      '1. Keycloak com KC_HTTP_RELATIVE_PATH=/auth (reinicie o container)',
      '2. sudo bash scripts/setup-vps-nginx.sh',
      '3. curl -sI "' + oidc + '"  → Content-Type: application/json',
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
