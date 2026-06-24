import { env } from '@/env';

function isLocalHost(url: string) {
  return /localhost|127\.0\.0\.1/i.test(url);
}

export function getAuthConfigHint(): string {
  const authUrl = env.VITE_AUTH_URL;
  const apiUrl = env.VITE_API_URL;
  const pageIsLocal = isLocalHost(window.location.hostname);
  const authIsLocal = isLocalHost(authUrl);

  if (authIsLocal && !pageIsLocal) {
    return [
      'O frontend em produção foi compilado com URLs de desenvolvimento.',
      `VITE_AUTH_URL atual: ${authUrl}`,
      `VITE_API_URL atual: ${apiUrl}`,
      '',
      'No Railway (serviço connect-web):',
      '1. Variables → marque VITE_AUTH_URL, VITE_API_URL, VITE_CLIENT_ID e VITE_CLIENT_SECRET como disponíveis no build',
      '2. VITE_AUTH_URL = URL pública do Keycloak (ex.: https://seu-keycloak.up.railway.app)',
      '3. VITE_API_URL = URL pública da API (ex.: https://sua-api.up.railway.app)',
      '4. Faça redeploy do connect-web após guardar',
      '',
      `No Keycloak, adicione redirect URI: ${window.location.origin}/*`
    ].join('\n');
  }

  if (pageIsLocal) {
    return [
      'Confirme no ficheiro src/Luxus.Connect.Web/.env:',
      'VITE_AUTH_URL=http://localhost:8081',
      'VITE_API_URL=http://localhost:8002',
      'Keycloak deve estar a correr (docker compose).'
    ].join('\n');
  }

  return [
    'Verifique se o Keycloak está acessível a partir do browser:',
    `VITE_AUTH_URL=${authUrl}`,
    `VITE_API_URL=${apiUrl}`,
    `Endpoint OIDC: ${authUrl}/realms/luxus/.well-known/openid-configuration`,
    '',
    `Redirect URI no Keycloak: ${window.location.origin}/*`,
    'Após alterar VITE_* no Railway, é obrigatório redeploy do connect-web.'
  ].join('\n');
}
