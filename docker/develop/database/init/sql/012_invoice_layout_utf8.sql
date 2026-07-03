-- Corrige textos com acentuação no layout padrão de fatura (encoding seguro em qualquer client SQL).
UPDATE "InvoiceLayoutTemplates"
SET "ConfigJson" = '{
  "theme": {
    "primaryColor": "#4a4a4a",
    "accentColor": "#00a0c6",
    "borderColor": "#222222",
    "headerBackground": "#ffffff",
    "titleColor": "#1a1a1a",
    "textColor": "#333333",
    "tableHeaderBackground": "#f7f7f7",
    "borderRadius": 12
  },
  "branding": {
    "logoDataUrl": "",
    "companyName": "LUXUS",
    "tagline": "SOLU\u00c7\u00c3O EM TELEFONIA",
    "documentTitle": "Detalhamento da Fatura"
  },
  "sections": {
    "userData": { "enabled": true, "title": "Dados do Usu\u00e1rio" },
    "accountValue": { "enabled": true, "title": "VALOR DA SUA CONTA" },
    "billingDates": { "enabled": true },
    "accountSummary": { "enabled": true, "title": "Resumo da Conta" },
    "detailedConsumption": { "enabled": true, "title": "Consumo Detalhado" }
  },
  "labels": {
    "name": "Nome:",
    "address": "Endere\u00e7o:",
    "phone": "N\u00famero do telefone:",
    "totalServices": "Total Servi\u00e7os:",
    "discounts": "Descontos:",
    "billingPeriod": "Per\u00edodo de faturamento:",
    "referenceMonth": "M\u00eas de refer\u00eancia:",
    "dueDate": "Data de Vencimento:",
    "description": "Descri\u00e7\u00e3o",
    "quantity": "Quantidade",
    "type": "Tipo",
    "unitPrice": "Pre\u00e7o Unit\u00e1rio",
    "total": "Total",
    "totalLabel": "Total:"
  }
}'::jsonb,
    "UpdatedAt" = NOW()
WHERE "Code" = 'default-invoice-layout';
