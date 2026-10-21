# Segurança e autenticação

O login usa autenticação nativa do Telefonia: Argon2id, sessões revogáveis no PostgreSQL, cookie HttpOnly e MFA TOTP com recuperação de uso único. Configuração e recuperação: [AUTH_LOCAL.md](AUTH_LOCAL.md).

## Operação em produção

- Use HTTPS e uma origem exata em PUBLIC_ORIGIN/CORS_ORIGINS.
- Preserve AUTH_SECRET: trocar essa chave impede ler os autenticadores cifrados existentes.
- Administradores e financeiro devem concluir o cadastro do MFA; guarde os códigos de recuperação fora do aplicativo autenticador.
- Gestão de usuários exige perfil autorizado e isolamento por organização. Alterações de conta invalidam sessões.
- Mantenha imagens de dependências fixadas por digest e escolha versões compatíveis com os dados existentes.
- Restrinja portas de bancos, filas e armazenamento à rede interna ou loopback.
- Faça backup do PostgreSQL, documentos e configurações; teste restauração em ambiente separado.
- Integrações usam chaves próprias e empresa explícita; essas chaves não concedem sessão de usuário.
- Não registre senhas, cookies, tokens ou códigos de recuperação nos logs.

Testes automatizados não substituem a validação do proxy HTTPS, login, MFA, permissões e restauração em homologação antes do deploy.
