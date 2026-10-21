# Deploy no Railway

O Telefonia usa API Go, PostgreSQL, RabbitMQ, armazenamento S3 compatível e interface Vite/React. O login e as passkeys WebAuthn são implementados pela própria API.

Configure a API com DATABASE_URL, AUTH_SECRET persistente, RABBITMQ_URL, CORS_ORIGINS com a origem HTTPS exata do site e as variáveis OBJECT_STORAGE_* de credenciais dedicadas. Consulte docker/env.deploy.example para as demais integrações opcionais.

Construa a API com api/Dockerfile. Para o web, use a configuração Docker do frontend e VITE_API_URL=/api. Encaminhe /api para a API no mesmo domínio da interface; cookies de produção são Secure e SameSite=Strict. Apenas apontar a interface para outro domínio de API não reproduz essa configuração de sessão.

Preserve AUTH_SECRET em todos os deploys e backups. Crie/recupere o administrador pelo comando do binário `api auth`, usando entrada padrão para senha. Consulte [AUTH_LOCAL.md](AUTH_LOCAL.md) para importação offline, passkeys e recuperação. Defina `WEBAUTHN_RP_ID` com o domínio público do site (sem esquema ou porta) e mantenha `CORS_ORIGINS` na mesma origem. A API aplica suas migrações na inicialização; valide primeiro em homologação.

Não use credenciais de demonstração em produção. Não publique bancos e painéis de administração. Verifique o health check /health e os logs de inicialização antes de liberar tráfego.
