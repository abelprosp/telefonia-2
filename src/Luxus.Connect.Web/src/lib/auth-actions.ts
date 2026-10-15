export async function performLogout(auth: {
  signoutRedirect: () => Promise<void>;
}) {
  await auth.signoutRedirect();
}
