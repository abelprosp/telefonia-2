/** A rejected OTP or current password must not discard a valid session. */
export function isExpiredSession(status: number, code?: string): boolean {
  return status === 401 && code === 'UNAUTHORIZED';
}

export function authNetworkMessage(error: unknown): string {
  if (
    error instanceof Error &&
    ['TimeoutError', 'AbortError'].includes(error.name)
  ) {
    return 'O servidor demorou para responder. Tente novamente.';
  }
  return 'Sem conexão com o servidor. Verifique sua rede e tente novamente.';
}
