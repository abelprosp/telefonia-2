export function getAuthConfigHint(message: string): string {
  return message
    ? 'Verifique a conexão com a API do Telefonia e tente entrar novamente.'
    : '';
}
