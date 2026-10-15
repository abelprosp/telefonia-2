# Autenticação própria do Telefonia

O login e a gestão de usuários usam a API Go e o PostgreSQL existentes.
O Keycloak não participa do arranque, login, permissões, MFA ou recuperação.
Não é necessário configurar realm, client, redirect URI, SMTP ou outro serviço.

## Atualizar a VPS existente

Faça backup do banco antes de aplicar a migração `035_local_auth.sql`.
Os clientes, vendas e demais registros existentes são mantidos.
O login antigo será encerrado: o token do Keycloak deixa de autorizar a API.
Guarde também o banco antigo `luxus_kc_dev` e o `.env` para recuperação.

```bash
cd ~/telefonia-2
git pull --ff-only
python3 scripts/configure-local-auth.py
docker compose -f docker-compose.yml -f docker-compose.prod.yml build connect-api connect-web
```

A configuração gera `AUTH_SECRET` somente quando ausente, preservando o resto
do `.env`. Guarde esse arquivo em backup protegido: a chave cifra os segredos
do autenticador. Não troque a chave a cada deploy.

### Preservar contas antigas

Antes de parar o contêiner antigo do Keycloak, execute:

```bash
python3 scripts/migrate-keycloak-users.py
```

O comando consulta o provedor antigo somente durante essa importação.
Preserva IDs, empresa, perfis, status e documento de cada usuário para manter
os vínculos de vendedores, parceiros, portal e auditoria.
É transacional, paginado e pode ser repetido: contas locais existentes não são
sobrescritas. Contas sem empresa válida interrompem a importação; corrija esses
vínculos na origem antes de tentar novamente. Usuários sem papel operacional
ficam com acesso de portal, sem receber privilégios por padrão.

Para uma conta sem organização, é possível escolher sua empresa somente durante
a migração, sem editar o Keycloak. Exemplo, se `arturabel` pertence à Luxus padrão:

```bash
python3 scripts/migrate-keycloak-users.py \
  --assign-org arturabel=00000000-0000-0000-0000-000000000001
```

Repita `--assign-org USUARIO=UUID` para outras contas sem empresa, escolhendo
individualmente o destino correto. Vínculos existentes de outra empresa não
são sobrescritos. Um nome não encontrado interrompe sem importação parcial.

Se todas as contas sem empresa pertencerem à mesma organização, escolha
explicitamente esse destino para o lote inteiro. Exemplo para a Luxus padrão:

```bash
python3 scripts/migrate-keycloak-users.py \
  --assign-unassigned-to 00000000-0000-0000-0000-000000000001
```

Isso resolve todas as contas sem empresa de uma vez. Contas que já possuem
organização continuam na empresa original. `--assign-org` tem prioridade para
exceções individuais. Sem essas opções, a migração continua recusando contas
sem vínculo, em vez de escolher uma empresa automaticamente.

Senhas e autenticadores antigos não são copiados. Defina uma nova senha para
cada conta importada, começando pelo seu administrador (substitua `SEU_USUARIO`):

```bash
bash scripts/local-auth-admin.sh reset-password SEU_USUARIO
```

O script lê e confirma a senha sem exibi-la ou colocá-la no histórico do shell.
Se a senha administrativa do Keycloak foi alterada desde o bootstrap, a importação
pode falhar; informe a senha atual via `read -rs KEYCLOAK_ADMIN_PASSWORD`,
`export KEYCLOAK_ADMIN_PASSWORD`, repita e depois execute
`unset KEYCLOAK_ADMIN_PASSWORD`. Essa credencial não é necessária depois.

### Instalação sem contas antigas

Se não houver contas para importar, crie o primeiro administrador:

```bash
bash scripts/local-auth-admin.sh init-admin admin admin@SUA-EMPRESA.com.br
```

Não existe usuário ou senha universal. `init-admin` recusa criar outro
administrador quando já existe um; para recuperar acesso use `reset-password`.

### Aplicar e verificar

```bash
DOCKER_PULL_POLICY=never bash docker-up.sh prod
bash scripts/setup-vps-nginx.sh
curl -fsS https://telefonia.redobrai.online/api/health
```

O modo `never` pressupõe que as imagens das dependências já estejam no cache.
O deploy não executa `down` e preserva contêineres antigos para migração/rollback.
Abra o site e entre com a nova senha. Contas Master e Financeiro precisam
cadastrar o autenticador pelo QR code na primeira entrada de produção.
Guarde os oito códigos de recuperação exibidos: cada um só funciona uma vez.
Após confirmar o acesso e as permissões, o Keycloak antigo pode ser parado:

```bash
docker stop keycloak.connect.luxus
```

Não apague o banco/volume antigo antes de validar a migração e seu backup.
Reverter o código exige restaurar também a configuração anterior; as novas
tabelas de autenticação não removem os dados antigos do Keycloak.

## Uso diário e recuperação

- Crie, desative ou redefina senhas de outras contas em **Usuários** (Master).
- Troque sua senha em **Configurações → Segurança da conta**, informando a senha atual.
- Alterar senha, perfil, empresa, status ou dados da conta encerra suas sessões.
- O administrador não pode remover seu próprio acesso pela tela de usuários.
- Para perda do autenticador, use um código de recuperação na tela de login.
- Se todos os códigos se perderem, o operador autorizado da VPS pode executar:

```bash
bash scripts/local-auth-admin.sh reset-password SEU_USUARIO
bash scripts/local-auth-admin.sh reset-mfa SEU_USUARIO
```

O reset do MFA encerra as sessões e exige novo cadastro do autenticador.
Não há endpoint público de reset, cadastro público ou envio de senha por e-mail.

## Proteções implementadas

- Argon2id com salt aleatório individual (19 MiB, duas iterações, uma lane).
- Senha de 12 caracteres no mínimo, até 256 bytes; sem truncamento silencioso.
- Sessão opaca de 256 bits; somente o hash SHA-256 fica no banco.
- Cookie `__Host-telefonia_session`, `Secure`, `HttpOnly`, `SameSite=Strict`, sem Domain.
- Expiração absoluta de 12 horas e por inatividade de 30 minutos.
- Sessão sem MFA pendente expira em dez minutos e não autoriza operações.
- Token rotacionado após MFA; TOTP e códigos de recuperação resistem a reenvio.
- TOTP cifrado com AES-GCM; códigos de recuperação armazenados somente como hash.
- Limites persistidos de tentativas por conta, sessão e peer; quatro operações
  simultâneas nas rotas de autenticação. Respostas de senha incorreta não revelam
  se o usuário existe ou está desativado.
- Escritas por cookie exigem `Origin` exato em `CORS_ORIGINS`; não use curingas.
- Papéis e empresa são lidos do banco em cada requisição; alterações de conta
  invalidam sessões na mesma transação, com proteção contra alterações concorrentes.
- Eventos de login, MFA e recuperação registrados sem senhas, tokens ou códigos.

`CORS_ORIGINS` já usa `PUBLIC_ORIGIN` no Compose de produção. Frontend e API
devem estar no mesmo site; a configuração padrão usa `/api` no mesmo domínio.
Integrações de intake e agente continuam com suas chaves e organização fixada;
essas chaves não autorizam o login nem a gestão de usuários.

Referências: [OWASP: senhas](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html),
[sessões](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html),
[CSRF](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html).

## Testes locais

Use um PostgreSQL descartável com as migrações de `db/migrations` aplicadas:

```bash
AUTH_TEST_DATABASE_URL=postgres://... go test ./internal/localauth -count=1
```

A suíte cobre login, QR de cadastro, confirmação, repetição de TOTP, uso único
de recuperação, CSRF, rotação do cookie, logout, expiração e desativação da conta.
Nunca aponte os testes para o banco de produção.
