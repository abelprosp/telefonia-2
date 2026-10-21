# Login e MFA do Telefonia

O Telefonia autentica usuários na API Go e no PostgreSQL existentes. Não exige realm, client, servidor de identidade, SMTP ou serviço pago.

## Login diário

1. Entre com usuário e senha no próprio Telefonia.
2. Se a conta usa MFA, informe o código de seis dígitos do seu aplicativo autenticador.
3. No primeiro acesso de uma conta privilegiada em produção, escaneie o QR code no Google Authenticator, Microsoft Authenticator, 1Password ou outro aplicativo TOTP. Confirme o primeiro código e guarde os códigos de recuperação em local protegido.

O aplicativo autenticador funciona sem internet. Os códigos mudam a cada 30 segundos; mantenha a hora do celular automática. Se perder o aparelho, escolha a opção de código de recuperação. Cada código só pode ser utilizado uma vez.

## Configurar a VPS

Faça backup do PostgreSQL e do `.env` antes de atualizar. As tabelas de clientes e vendas e os usuários locais existentes são preservados.

```bash
cd ~/telefonia-2
git pull --ff-only
python3 scripts/configure-local-auth.py
docker compose -f docker-compose.yml -f docker-compose.prod.yml build connect-api connect-web
```

A configuração cria `AUTH_SECRET` somente se estiver ausente e preserva as demais variáveis. Guarde essa chave junto do backup protegido: ela cifra os segredos MFA. Não a regenere a cada deploy. `PUBLIC_ORIGIN` deve ser a origem HTTPS do site, sem caminho; o Compose a usa para `CORS_ORIGINS`.

### Primeiro administrador ou recuperação de acesso

Se ainda não há administrador local:

```bash
bash scripts/local-auth-admin.sh init-admin admin admin@SUA-EMPRESA.com.br
```

Para uma conta já existente:

```bash
bash scripts/local-auth-admin.sh reset-password SEU_USUARIO
```

A senha é solicitada sem exibição e não entra no histórico do shell. Use pelo menos 12 caracteres. O primeiro administrador só é criado em uma instalação sem administrador; não existe senha universal.

Se o autenticador e todos os códigos de recuperação se perderem:

```bash
bash scripts/local-auth-admin.sh reset-mfa SEU_USUARIO
```

O reset encerra sessões e exige novo cadastro do autenticador. Não há endpoint público para esse reset.

### Importação offline de contas antigas

Se as contas já foram importadas, não repita a migração: use a mesma conta local. Para contas ainda ausentes, utilize um arquivo JSON de exportação de usuários, sem conectar o Telefonia ao provedor antigo:

```bash
python3 scripts/import-local-users.py /root/backup/users.json
```

A importação preserva UUIDs, empresas e vínculos e não sobrescreve contas locais existentes. Senhas e autenticadores da origem não são copiados. Defina uma senha local depois da importação. Para contas sem empresa, escolha explicitamente o vínculo correto:

```bash
python3 scripts/import-local-users.py /root/backup/users.json \
  --assign-org SEU_USUARIO=00000000-0000-0000-0000-000000000001
```

`--assign-unassigned-to UUID` atribui um destino explícito às contas sem empresa do lote. Contas com empresa já definida mantêm esse vínculo; exceções individuais usam `--assign-org`. Proteja o arquivo exportado e remova a cópia temporária após validar a importação.

### Aplicar a atualização

```bash
bash docker-up.sh prod
bash scripts/setup-vps-nginx.sh
curl -fsS https://telefonia.redobrai.online/api/health
```

Se precisar reutilizar imagens de dependências já disponíveis localmente, use `DOCKER_PULL_POLICY=never bash docker-up.sh prod`. O deploy recompila API e interface, sem executar `down` ou apagar volumes. Abra o site e faça login novamente. A atualização não para nem apaga automaticamente contêineres ou bancos antigos: valide primeiro o acesso local e seus backups.

Depois de confirmar login, MFA e permissões na versão nova, pare o contêiner legado, caso ele ainda exista na VPS:

```bash
docker stop keycloak.connect.luxus
```

O serviço não faz parte do Compose atual. Preserve os backups antigos até validar a migração; não remova volumes durante o deploy.

## Proteções

- Senhas com Argon2id e salt individual; mínimo de 12 caracteres e máximo de 256 bytes.
- Sessões opacas revogáveis, com hash no banco e cookie HttpOnly, Secure e SameSite=Strict em produção.
- Expiração absoluta de 12 horas; expiração por inatividade de 30 minutos. Sessão pendente de MFA vale dez minutos.
- TOTP cifrado com AES-GCM; proteção contra repetição de códigos e códigos de recuperação de uso único.
- MFA obrigatório para administradores e financeiro em produção; MFA configurado continua exigido nos logins seguintes.
- Escritas exigem Origin exato; sem curingas em CORS_ORIGINS.
- Limites persistidos de tentativas; troca de senha/perfil/status invalida sessões.
- Recuperação administrativa pela VPS, sem envio de senha ou código por e-mail.

Frontend e API devem usar a mesma origem, com API em `/api`. As chaves de integrações não autorizam login nem gestão de usuários.

## Testes

```bash
cd api
go test ./...
```

Para a suíte de autenticação com banco, configure `AUTH_TEST_DATABASE_URL` com um PostgreSQL descartável com as migrações aplicadas. Nunca use o banco de produção nos testes.
