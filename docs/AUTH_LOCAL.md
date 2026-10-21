# Login e passkeys do Telefonia

O Telefonia autentica usuários na API Go e no PostgreSQL existentes. Não exige realm, client, servidor de identidade, SMTP ou serviço pago. O segundo fator é uma passkey (WebAuthn): digital, Face ID, Windows Hello ou chave de segurança física.

## Login diário

1. Entre com usuário e senha no próprio Telefonia.
2. Confirme o acesso com a passkey cadastrada no seu dispositivo. Não há código de seis dígitos para digitar.
3. No primeiro acesso em produção, cadastre uma passkey logo após a senha e guarde os códigos de recuperação em local protegido.

A passkey fica no dispositivo ou no gerenciador de senhas (iCloud Keychain, Google Password Manager, 1Password, Windows Hello). Cadastre mais de um dispositivo quando possível. Se perder o aparelho, use um código de recuperação; cada código funciona uma única vez.

O primeiro cadastro de uma conta sem passkey exige apenas a senha. Conclua-o imediatamente após criar ou importar a conta, antes de compartilhar a senha. Depois que uma passkey existe, uma sessão só-senha não consegue cadastrar outra.

### Requisitos

- HTTPS em produção. WebAuthn não funciona em HTTP, exceto `localhost` durante o desenvolvimento.
- O domínio da interface deve ser o mesmo configurado em `PUBLIC_ORIGIN`/`CORS_ORIGINS`.
- Navegador atualizado com suporte a WebAuthn (Chrome, Edge, Firefox, Safari).
- `WEBAUTHN_RP_ID` é o domínio sem esquema, porta ou caminho. Quando ausente, a API usa o host da primeira origem de `CORS_ORIGINS`.

## Configurar a VPS

Faça backup do PostgreSQL e do `.env` antes de atualizar. As tabelas de clientes e vendas e os usuários locais existentes são preservados.

```bash
cd ~/telefonia-2
git pull --ff-only
python3 scripts/configure-local-auth.py
docker compose -f docker-compose.yml -f docker-compose.prod.yml build connect-api connect-web
```

A configuração cria `AUTH_SECRET` somente se estiver ausente e preserva as demais variáveis. Guarde essa chave junto do backup protegido e não a regenere a cada deploy. `PUBLIC_ORIGIN` deve ser a origem HTTPS do site, sem caminho; o Compose a usa para `CORS_ORIGINS`.

Opcionalmente, defina explicitamente o domínio WebAuthn:

```bash
WEBAUTHN_RP_ID=telefonia.redobrai.online
WEBAUTHN_RP_NAME=Telefonia
```

`WEBAUTHN_RP_ID` precisa ser o domínio da interface ou um sufixo registrável dele. Não inclua `https://`, porta ou barra. Depois de mudar esse valor, as passkeys antigas deixam de valer e os usuários precisam cadastrar novamente.

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

Se a passkey e todos os códigos de recuperação se perderem:

```bash
bash scripts/local-auth-admin.sh reset-mfa SEU_USUARIO
```

`reset-passkeys` é um alias do mesmo comando. O reset remove as passkeys, os códigos de recuperação e as sessões, e exige novo cadastro no próximo login. Não há endpoint público para esse reset.

### Importação offline de contas antigas

Se as contas já foram importadas, não repita a migração: use a mesma conta local. Para contas ainda ausentes, utilize um arquivo JSON de exportação de usuários, sem conectar o Telefonia ao provedor antigo:

```bash
python3 scripts/import-local-users.py /root/backup/users.json
```

A importação preserva UUIDs, empresas e vínculos e não sobrescreve contas locais existentes. Senhas e passkeys da origem não são copiadas. Defina uma senha local depois da importação. Para contas sem empresa, escolha explicitamente o vínculo correto:

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

Se precisar reutilizar imagens de dependências já disponíveis localmente, use `DOCKER_PULL_POLICY=never bash docker-up.sh prod`. O deploy recompila API e interface, sem executar `down` ou apagar volumes. Abra o site e faça login novamente.

### Remover o Keycloak legado

O código atual não usa Keycloak. Depois de confirmar login e permissões na versão nova, remova o contêiner e os volumes antigos da VPS com o script idempotente:

```bash
bash scripts/remove-legacy-keycloak.sh
```

O script apenas lista o que encontrou e pede confirmação. Use `--yes` para execução não interativa e `--purge-volumes` para remover também os volumes `keycloak*`. Faça backup e preserve os volumes antigos até validar a migração. Não remova volumes durante o deploy.

## Proteções

- Senhas com Argon2id e salt individual; mínimo de 12 caracteres e máximo de 256 bytes.
- Sessões opacas revogáveis, com hash no banco e cookie HttpOnly, Secure e SameSite=Strict em produção.
- Expiração absoluta de 12 horas; expiração por inatividade de 30 minutos. Sessão pendente de MFA vale dez minutos.
- WebAuthn com verificação de usuário obrigatória, desafio de uso único armazenado no servidor e validação de origem e RP ID pela biblioteca.
- Códigos de recuperação de uso único, gerados apenas no primeiro cadastro de passkey.
- MFA obrigatório para todas as contas em produção; quem já tem passkey continua exigido nos logins seguintes.
- Uma sessão só-senha não pode adicionar uma segunda passkey: para isso é necessário confirmar uma passkey existente.
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
