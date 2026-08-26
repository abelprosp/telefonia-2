# Agente financeiro n8n (OpenAI + WhatsApp)

Agente de atendimento financeiro no WhatsApp. Consulta o Luxus Connect, envia fatura/boleto, confere inadimplência no sistema e no Sicredi, calcula multa/juros e valida comprovantes.

## O que o agente faz

- Localiza cliente por **nome**, **CPF/CNPJ**, **número da linha** ou **número da fatura**
- Informa **vencimento**, valor, PIX e linha digitável
- Envia a **fatura/boleto Sicredi** pelo WhatsApp
- Verifica **inadimplência** no sistema e no **Sicredi**
- Calcula **multa e juros** com as taxas cadastradas em Configurações
- Confere **comprovante** (foto no WhatsApp) com Vision da OpenAI + regras do sistema
- Só **baixa pagamento sozinho** quando o Sicredi confirma liquidação **ou** o PIX da fatura bate com o comprovante

## 1. API Luxus

No ambiente da API:

```
FINANCIAL_AGENT_API_KEY=<chave longa e aleatória>
FINANCIAL_AGENT_ORG_ID=<UUID da organização no Keycloak>
FINANCIAL_AGENT_PUBLIC_URL=https://sua-api.exemplo.com
```

`FINANCIAL_AGENT_ORG_ID` é o mesmo `organization` do token Keycloak (UUID da org). Sem isso o agente não enxerga clientes nem faturas.

## 2. Variáveis no n8n

Em **Settings → Variables** (ou variáveis de ambiente da instância):

| Variável | Exemplo |
|---|---|
| `LUXUS_API_URL` | `https://sua-api.exemplo.com` |
| `LUXUS_AGENT_API_KEY` | a mesma de `FINANCIAL_AGENT_API_KEY` |
| `EVOLUTION_API_URL` | `https://evolution.exemplo.com` |
| `EVOLUTION_API_KEY` | chave da Evolution API |
| `EVOLUTION_INSTANCE` | nome da instância WhatsApp |
| `OPENAI_API_KEY` | chave da OpenAI (ou credencial nativa do n8n) |

## 3. Importar

1. n8n → **Workflows → Import from File**
2. Importe `agente-financeiro.json`
3. Crie/vincule a credencial **OpenAI** no nó `OpenAI Chat Model`
4. No WhatsApp (Evolution), configure o webhook da instância para o URL do nó **Webhook WhatsApp** (produção), evento `MESSAGES_UPSERT`
5. Ative o workflow

## 4. Evolution API

Evento: `messages.upsert`

URL do webhook n8n (após ativar):

```
https://seu-n8n.exemplo.com/webhook/agente-financeiro-whatsapp
```

O fluxo ignora mensagens `fromMe` para não entrar em loop.

## 5. Ferramentas HTTP (Luxus)

Todas autenticadas com header `X-Agent-Key`.

| Ferramenta | Endpoint |
|---|---|
| Consultar cliente | `POST /v1/agent/financial/lookup-customer` |
| Listar faturas | `POST /v1/agent/financial/list-invoices` |
| Pacote da fatura (texto + PIX) | `POST /v1/agent/financial/invoice-package` |
| Inadimplência + Sicredi | `POST /v1/agent/financial/check-delinquency` |
| Calcular juros | `POST /v1/agent/financial/calculate-interest` |
| Sincronizar Sicredi | `POST /v1/agent/financial/sync-sicredi` |
| Validar comprovante | `POST /v1/agent/financial/verify-receipt` |
| Confirmar pagamento | `POST /v1/agent/financial/confirm-payment` |
| PDF do boleto | `GET /v1/agent/financial/invoices/{id}/boleto-pdf` |

## 6. Comprovantes

Quando o cliente manda **foto/PDF**, o n8n:

1. Baixa a mídia na Evolution
2. Extrai dados com **GPT-4o Vision** (valor, data, PIX, banco, pagador)
3. Chama `verify-receipt` no Luxus (casa com fatura + Sicredi)

Baixa automática só com confiança alta **e** (Sicredi liquidado **ou** TxID PIX igual ao da fatura). Caso contrário o agente pede revisão humana e **não** registra o pagamento.

## 7. Teste rápido

No WhatsApp da instância:

1. `Oi, meu nome é ... quero a fatura`
2. `Estou em atraso? Quanto fica com juros?`
3. Envie um comprovante de pagamento

Health check da API:

```
curl -H "X-Agent-Key: $FINANCIAL_AGENT_API_KEY" \
  "$LUXUS_API_URL/v1/agent/financial/health"
```
