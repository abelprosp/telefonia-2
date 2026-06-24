# Sicredi — configuração em produção

Este guia descreve como gerar cobranças (boleto híbrido + PIX) e registrar pagamentos automaticamente no Luxus Connect.

## Fluxo

1. **Gerar fatura** — manual (cliente) ou em massa; o boleto Sicredi é emitido automaticamente.
2. **Cliente paga** — boleto ou PIX.
3. **Confirmação de pagamento** — via webhook do Sicredi (imediato) ou sincronização automática a cada 15 minutos.
4. **Status na UI** — fatura marcada como **Pago** com data em `sicredi_paid_at`; conta a receber atualizada.

## Variáveis de ambiente (API)

```env
SICREDI_ENABLED=true
SICREDI_SANDBOX=false

SICREDI_API_KEY=<x-api-key do portal Sicredi — produção>
SICREDI_PASSWORD=<código de acesso do Internet Banking>
SICREDI_USERNAME=010480179
SICREDI_COOPERATIVA=0179
SICREDI_POSTO=14
SICREDI_CODIGO_BENEFICIARIO=01048

# URL pública HTTPS da API (Railway, domínio próprio, etc.)
SICREDI_PUBLIC_API_URL=https://sua-api.railway.app

# Token que o Sicredi enviará no header Authorization do webhook
SICREDI_WEBHOOK_TOKEN=<token forte, ex. luxus-connect-sicredi-wh-2026>

# Opcional: registra webhook automaticamente ao subir a API
SICREDI_AUTO_REGISTER_WEBHOOK=true
```

**Importante:** nunca use `SICREDI_SANDBOX=true` com credenciais de produção (retorna 403).

## Migrações

Aplique as migrações no Postgres antes de usar em produção:

```bash
./scripts/railway-apply-migrations.sh
```

Inclui tabelas/colunas Sicredi (`008`, `009`, `013`, `014`).

## Passo a passo

### 1. Deploy da API com URL pública

No Railway (ou outro host), configure `SICREDI_PUBLIC_API_URL` com a URL HTTPS do serviço da API.

O webhook ficará em:

```
https://sua-api.railway.app/v1/webhooks/sicredi
```

### 2. Configurar integração (UI)

Em **Financeiro → Faturas de clientes**:

1. Clique em **Testar conexão** — valida OAuth com o Sicredi.
2. Informe a URL pública da API (se ainda não estiver no `.env`).
3. Clique em **Configurar produção** — valida token, URL, conexão e registra o webhook no Sicredi.

Alternativa via API:

```http
POST /v1/collections/sicredi/setup-production
Authorization: Bearer <token>
Content-Type: application/json

{ "public_api_url": "https://sua-api.railway.app" }
```

### 3. Gerar cobranças

- **Cliente individual:** detalhe do cliente → Gerar fatura.
- **Em massa:** Financeiro → Faturas → Gerar em massa.

A resposta indica `sicredi_boleto_status`: `issued` (sucesso) ou `failed` (com `sicredi_boleto_error`).

Requisitos para o boleto:

- Cliente com CPF/CNPJ cadastrado.
- Sicredi habilitado e conectado.

### 4. Confirmar pagamento

**Automático (recomendado):**

- Webhook `LIQUIDACAO` → marca fatura como paga e registra pagamento na conta a receber.
- Cron interno (15 min) consulta boletos liquidados no Sicredi como fallback.

**Manual:**

- Na fatura: botão **Sincronizar pagamento Sicredi**.
- Na listagem: **Sincronizar pagamentos Sicredi** (lote).

### 5. Idempotência

Pagamentos não duplicam: o sistema verifica `sicredi_paid_at` e referência `Sicredi {nossoNumero}` antes de registrar.

Eventos de webhook duplicados são ignorados (índice único por `nossoNumero` + `tipoEvento`).

## Endpoints úteis

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/v1/collections/sicredi/status` | Status da integração |
| POST | `/v1/collections/sicredi/test-connection` | Testa OAuth |
| POST | `/v1/collections/sicredi/setup-production` | Setup completo |
| POST | `/v1/collections/sicredi/register-webhook` | Só registra webhook |
| POST | `/v1/webhooks/sicredi` | Webhook público (Sicredi) |
| POST | `/v1/collections/sync-sicredi-payments` | Sync em lote |
| POST | `/v1/customer-billing-documents/{id}/sync-sicredi-payment` | Sync por fatura |

## Troubleshooting

| Problema | Solução |
|----------|---------|
| 403 Access denied | `SICREDI_SANDBOX=false` com credenciais de produção |
| Boleto `failed` | Verifique CPF/CNPJ do cliente e logs da API |
| Webhook não chega | URL deve ser HTTPS pública; use **Configurar produção** |
| Token inválido no webhook | `SICREDI_WEBHOOK_TOKEN` deve coincidir com o registrado no Sicredi |
| Pagamento não aparece | Use sync manual; confira se boleto está liquidado no Sicredi |
