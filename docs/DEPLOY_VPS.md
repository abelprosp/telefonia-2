# Deploy na VPS

A autenticação é própria do Telefonia, com senha, sessões HttpOnly e MFA TOTP. A configuração de contas e recuperação está em [AUTH_LOCAL.md](AUTH_LOCAL.md).

## Configuração

Use Docker Compose v2, DNS apontando para a VPS e nginx no host com TLS. O nginx encaminha `/api/` para `127.0.0.1:8010` e a interface para `127.0.0.1:3005`. O MinIO usa `127.0.0.1:19000` para o proxy de objetos. Não publique bancos ou painéis administrativos na internet.

Configure `.env` a partir de `docker/env.deploy.example`, preservando credenciais e volumes existentes. Defina `PUBLIC_ORIGIN=https://telefonia.redobrai.online`, credenciais dedicadas de Postgres, RabbitMQ e MinIO e imagens de produção fixadas por digest. A chave persistente de autenticação é gerada por `configure-local-auth.py`.

## Atualizar

Antes de atualizar, faça backup dos dados e do `.env`. As migrações SQL são aplicadas pela API durante a inicialização.

```bash
cd ~/telefonia-2
git pull --ff-only
python3 scripts/configure-local-auth.py
bash docker-up.sh prod
bash scripts/setup-vps-nginx.sh
docker compose -f docker-compose.yml -f docker-compose.prod.yml ps
curl -fsS https://telefonia.redobrai.online/api/health
```

`docker-up.sh` recompila as imagens e recria os serviços alterados. Não é necessário executar `down`. Use `DOCKER_PULL_POLICY=never` somente quando as imagens das dependências já estiverem disponíveis localmente.

## Contas e MFA

Siga [o guia de autenticação](AUTH_LOCAL.md) para criar o primeiro administrador, importar contas offline ou recuperar senha/MFA. Não crie uma conta nova para substituir outra existente sem avaliar seus vínculos com clientes e vendas.

## Diagnóstico

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml logs --tail 80 connect-api connect-web
```

Para erro de login, confira a origem HTTPS exata em `PUBLIC_ORIGIN`, o cookie da sessão e a configuração do proxy. Para erro MFA, confira a hora do celular e use um código de recuperação se necessário. Para 502, verifique os serviços e as portas do proxy. Não compartilhe logs com senhas, tokens ou cookies.
