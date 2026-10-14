# Correções de segurança e interface

## Alterações

- Contratos: escape dos dados pessoais, sanitização HTML no servidor e no navegador, inclusive contratos antigos. Tags de formulário, scripts e estilos arbitrários não são permitidas.
- Senhas: `/v1/me` não redefine mais a senha via acesso administrativo. A interface abre a segurança da conta Keycloak, cuja ação de atualização exige reautenticação em produção.
- Arquivos: acesso genérico restrito a master, funcionários, operadores e financeiro; clientes e parceiros usam endpoints que verificam o documento ou ticket. Uploads genéricos recebem uma chave nova, retornada em `object_key`, e ficam limitados ao bucket configurado. URLs expiram em até 15 minutos.
- Comercial: mantém cadastro/edição de clientes e vendas; não altera operadoras, estoque, templates, fechamentos ou baixas de pagamento. O backend aplica as permissões independentemente da interface.
- JWT: papéis de `resource_access` são lidos somente para o cliente configurado. Master e financeiro precisam de MFA comprovado no token em produção; `/v1/me` continua disponível para permitir a configuração da conta.
- Marca: leitura pública somente em `/v1/whitelabel-settings`; dados internos ficam nos endpoints administrativos. A cor primária é aplicada com texto escolhido por contraste e o logo escuro acompanha o tema.
- API: limite JSON de 2 MiB, controle de concorrência, limites de tempo HTTP e rejeição de conteúdo após o primeiro JSON. Métricas exigem master e o health público não expõe erros de componentes.
- nginx: limites de requisições, limite de 2 MiB na API e no login, HSTS no proxy TLS e CSP da SPA sem scripts inline. Uploads de faturas continuam diretamente no storage, com o limite maior nessa rota.
- Deploy: sem credenciais padrão de storage; conta de aplicação separada do root, política limitada ao bucket, CORS do storage restrito à origem pública e falha explícita para URLs inválidas.
- Interface: atalhos respeitam o perfil, relatórios financeiros usam autorização financeira, links de resumo e rentabilidade levam à seção correspondente, tabelas informam que a ordenação é da página, e a inativação de clientes é descrita corretamente. A lista exibe o nome do vendedor quando disponível no Keycloak da mesma organização.

## Aplicar em produção existente

Estas alterações no repositório não modificam automaticamente um servidor publicado.

1. Preencha `docker/env.deploy.example` no ambiente de produção. `MINIO_ROOT_USER` e `MINIO_ROOT_PASSWORD` devem diferir de `OBJECT_STORAGE_ACCESS_KEY_ID` e `OBJECT_STORAGE_SECRET_ACCESS_KEY`. Use segredos novos e mantenha o arquivo fora do Git.
2. Defina `PUBLIC_ORIGIN` com a origem HTTPS sem barra final. O arquivo nginx da VPS deve usar o mesmo domínio e os certificados correspondentes. O bucket padrão é `luxus-connect`; um bucket diferente exige ajustar também o caminho do proxy de storage.
3. Defina `MINIO_IMAGE`, `MINIO_MC_IMAGE` e `KEYCLOAK_IMAGE` com imagens verificadas e corrigidas, fixadas por digest `@sha256:...`. O script de deploy rejeita referências flutuantes para essas imagens. Preserve a versão principal atual dos serviços ao escolher as imagens para evitar migrações involuntárias do banco.
4. O `minio-init` cria a conta de aplicação e sua política antes de iniciar a API. Ele não ignora falhas de configuração. O CORS é configurado no servidor por `MINIO_API_CORS_ALLOW_ORIGIN=PUBLIC_ORIGIN`; não usa a API de CORS por bucket, indisponível no MinIO Community instalado. Se a chave anteriormente usada pela aplicação era root, rotacione o root e configure a nova chave de aplicação; não revogue uma chave ainda usada por outra integração sem migrá-la.
5. Para um realm **novo**, o Compose usa `docker/keycloak/luxus-realm.prod.json`. Ele não cria usuários de demonstração, exige OTP no browser, configura o claim `amr`, desabilita Direct Access Grants do SPA, exige PKCE S256 e limita redirects e origens ao domínio público.
6. Para um realm **existente**, `--import-realm` não sobrescreve sua configuração. No Admin Console, aplique os mesmos ajustes, preservando usuários, organizações, mappers existentes e vínculos:
   - Realm roles → `user` → Associated roles: adicione os papéis do client `account`, `manage-account` e `view-profile`. Esses papéis permitem administrar a própria conta. Confirme que o usuário do aplicativo possui `user`, que `account`/`account-console` estão habilitados e que o mapper Audience Resolve do `account-console` não foi removido. Sem audience `account`, o portal da conta pode retornar 401 mesmo depois do login. Os arquivos de realm agora incluem os dois papéis como composição de `user`.
   - Realm settings: habilite Brute Force Detection, com limite de 5 falhas e espera crescente; use política de senha de no mínimo 12 caracteres.
   - Client `connect-cli`: desabilite Direct Access Grants; exija PKCE S256; remova redirects/origins locais de produção. Adicione o mapper **Authentication Method Reference (AMR)** aos access tokens.
   - Authentication: copie o fluxo browser para um fluxo próprio. Na subetapa de formulários, mantenha senha e OTP como **Required**. Configure o Authenticator Reference da senha como `pwd` e do OTP como `otp`, com máximo de 36000 segundos, alinhado à duração máxima de sessão. O arquivo de produção contém a estrutura completa do fluxo `connect-browser`.
   - Required actions: habilite Configure OTP como ação padrão. A variável `KC_SPI_REQUIRED_ACTION__UPDATE_PASSWORD__MAX_AUTH_AGE=0` exige reautenticação na troca de senha; mantenha a imagem Keycloak compatível com essa configuração.
   - Entre novamente e confirme que o access token contém `amr` com `otp`. Não marque MFA por meio de um claim fixo: ele deve vir da execução real do autenticador.
7. Atualize a configuração nginx e execute `nginx -t` antes de recarregar. Depois execute `./docker-up.sh prod`.

Os endpoints genéricos de storage recusam clientes, parceiros e vendedores. Integrações que os utilizavam devem usar o endpoint do recurso autorizado. Envio de anexos comerciais deve ser feito por funcionário/operador autorizado; o portal mantém envio de anexos de tickets.

## Verificação

```sh
cd api
go test ./...
go vet ./...
cd ../src/Luxus.Connect.Web
npm ci
npm run test:security
npm run build
npm audit
```

Testes de integração com PostgreSQL exigem uma base de teste; sem ela, os testes existentes de integração são ignorados. Não configure `DATABASE_URL` de produção para executar a suíte: ela aplica migrações e cria registros de teste.

O CSS dos componentes Base UI fica local em `src/base-ui-variants.css`. O aplicativo não depende mais do CLI shadcn para compilar; isso removeu sua cadeia de dependências vulneráveis. A configuração `components.json` continua disponível para manutenção dos componentes.

### Resultado local em 08/10/2026

- `go test ./...` e `go vet ./...`: aprovados. Os testes de integração dependentes do PostgreSQL foram ignorados porque a base local não estava disponível.
- `npm run test:security`: 3 testes aprovados; `npm run build`: aprovado.
- Auditoria npm: zero vulnerabilidades reportadas após atualização das dependências.
- ESLint dos arquivos alterados: zero erros; quatro avisos existentes sobre atualização de estado em efeitos React.
- Compose de produção validado com valores fictícios: imagens por digest, realm correto, credenciais separadas, dependência de inicialização e todas as portas limitadas ao loopback. Scripts shell passaram na verificação de sintaxe.
- O Docker local estava desligado. Containers, fluxo completo do Keycloak, configuração nginx em execução e aparência das telas autenticadas ainda precisam ser conferidos no ambiente de homologação antes de publicar.

Referências de implementação: [Keycloak AMR](https://www.keycloak.org/admin-api/protocol-mappers), [configuração de reautenticação](https://www.keycloak.org/server/all-provider-config), [sanitização HTML OWASP](https://cheatsheetseries.owasp.org/cheatsheets/Cross_Site_Scripting_Prevention_Cheat_Sheet.html).
