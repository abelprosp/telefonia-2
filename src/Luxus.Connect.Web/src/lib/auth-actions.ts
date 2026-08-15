import type { AuthContextProps } from 'react-oidc-context';

/**
 * Realiza o encerramento seguro e completo de sessão do usuário.
 * Tenta signoutRedirect primeiro para limpar a sessão no Identity Provider (Keycloak),
 * com fallback para limpeza local de credenciais e redirecionamento.
 */
export async function performLogout(auth: Pick<AuthContextProps, 'signoutRedirect' | 'removeUser' | 'signoutSilent'>) {
  try {
    sessionStorage.clear();
    localStorage.removeItem('luxus_last_auth_redirect');

    if (typeof auth.signoutRedirect === 'function') {
      await auth.signoutRedirect({
        post_logout_redirect_uri: window.location.origin
      });
      return;
    }
  } catch (err) {
    console.warn('signoutRedirect failed, executing fallback removal:', err);
  }

  try {
    if (typeof auth.removeUser === 'function') {
      await auth.removeUser();
    }
  } catch (err) {
    console.warn('removeUser error:', err);
  }

  // Redireciona para a raiz
  window.location.href = '/';
}
