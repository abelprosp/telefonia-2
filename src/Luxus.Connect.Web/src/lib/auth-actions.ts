export async function performLogout(auth: { logout: () => Promise<void> }) {
  await auth.logout();
}
