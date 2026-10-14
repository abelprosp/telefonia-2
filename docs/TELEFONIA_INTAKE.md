# Recebimento de clientes e vendas

## Endpoints

- `POST /api/v1/intake/customers`: recebe o envelope `client + sale + attachments`, com `X-API-Key`. URL pública: `https://telefonia.redobrai.online/api/v1/intake/customers` (sem barra final).
- `POST /api/v1/customers`: mantém o cadastro plano atual e também aceita o envelope, usando o JWT e as permissões operacionais existentes. A chave de intake não autoriza este endpoint.
- `GET /api/v1/customers/{id}/imported-sales`: últimas 100 vendas importadas, em ordem de recebimento, acessível por JWT operacional e restrito à organização.
- `GET /api/v1/customers/{id}/attachments/{attachmentId}/download`: URL temporária de documento pertencente ao cliente, com a mesma autorização operacional.
- O detalhe normal `GET /api/v1/sales/{id}` passa a incluir `intake` quando a venda veio de outro sistema.

## Configurar na VPS receptora

No `.env` do telefonia-2:

```dotenv
TELEFONIA_INTAKE_API_KEY=<segredo aleatório de pelo menos 32 caracteres>
TELEFONIA_INTAKE_ORG_ID=<ID da organização de destino>
```

Gere uma chave com `openssl rand -hex 32`. Não utilize credenciais do banco, do storage ou a chave do agente financeiro. A integração fica desabilitada quando as duas variáveis estão vazias. Definir apenas uma delas impede a inicialização para evitar configuração parcial.

A organização é fixada pelo servidor; não é aceita do remetente. Para a organização Luxus padrão, o ID é `00000000-0000-0000-0000-000000000001`; para outra organização, use seu ID real. A chave permite somente receber clientes/vendas, sem acesso de administrador ou ao financeiro.

No sistema remetente:

```dotenv
TELEFONIA_INTAKE_URL=https://telefonia.redobrai.online/api/v1/intake/customers
TELEFONIA_INTAKE_API_KEY=<o mesmo segredo configurado no telefonia-2>
TELEFONIA_INTAKE_TIMEOUT_MS=90000
```

O remetente deve enviar `Content-Type: application/json` e `X-API-Key: <segredo>`. Também é aceito `Authorization: Bearer <segredo>` na rota dedicada.

Guarde a chave somente no backend dos dois sistemas. Não use variáveis `VITE_*` nem exponha o segredo em JavaScript de navegador.

## Payload

O formato abaixo mantém os nomes camelCase recebidos do outro sistema. Valores e comissão são **reais**, não percentuais, com até duas casas decimais. O servidor converte para centavos inteiros, sem arredondamento silencioso. CPF, CNPJ, telefone, ICCID, IMEI e EID devem ser strings para preservar zeros iniciais.

```json
{
  "sourceSystem": "sistema-comercial",
  "sale": {
    "id": "cmaq1exemplo",
    "protocol": "VND-ABC123",
    "taskProtocol": "LUX-2026-00001",
    "soldAt": "2026-10-14T14:30:00.000Z",
    "concludedAt": "2026-10-14T18:05:00.000Z",
    "value": 79.9,
    "commissionValue": 10,
    "operatorName": "Vivo",
    "planName": "Controle",
    "lineNumber": "51985370873",
    "simType": "CHIP",
    "isVirginChip": false,
    "iccid": "8955100000000000000",
    "deviceImei": "490154203237518",
    "deviceEid": "89049032000000000000000000000000",
    "devicePlatform": "ANDROID",
    "isPortability": true,
    "portabilityNumber": "51999998888",
    "donorOperator": "TIM",
    "isUpgrade": false,
    "upgradeOfProtocol": "",
    "contractFormat": "ZAPSIGN",
    "partnerName": "META CELL MATRIZ",
    "branchName": "CRUZEIRO",
    "notes": "observação da venda"
  },
  "client": {
    "name": "Cliente de exemplo",
    "document": "02113419084",
    "documentType": "CPF",
    "rg": "1234567890",
    "email": "cliente@example.com",
    "phone": "51999998888",
    "address": "Rua Exemplo",
    "addressNumber": "100",
    "complement": "apto 2",
    "neighborhood": "Centro",
    "city": "Lajeado",
    "state": "RS",
    "zipCode": "95900000"
  },
  "attachments": []
}
```

`sale.id`, `sale.protocol`, `sale.soldAt`, `sale.value`, `sale.commissionValue`, `client.name`, `client.document` e `client.documentType` são obrigatórios. `sourceSystem` é opcional e assume `external`; use um nome estável e distinto quando houver mais de um remetente. `client.id` opcional é preservado no histórico. Para PJ, envie `documentType: "CNPJ"`, `legalName` e omita o RG pessoal. CEP/UF, e-mail e telefones são validados como no cadastro normal.

Portabilidade exige `portabilityNumber` e `donorOperator`. Upgrade exige `upgradeOfProtocol`. Conclusão não pode ser anterior à venda. Valores negativos, mais de duas casas decimais, campos desconhecidos e conteúdo após o JSON são recusados.

## Anexos

Cada entrada contém `id`, `type`, `name`, `mimeType`, `contentBase64`. O Base64 deve conter os bytes reais do arquivo, sem prefixo `data:` e sem reticências de exemplo. Os IDs dos anexos devem ser únicos dentro do envio.

São aceitos PDF, JPEG, PNG e WebP. O conteúdo é conferido contra o MIME informado, e nomes com caminhos são recusados. Limites: 10 anexos, 8 MiB por arquivo, 16 MiB decodificados no total e 24 MiB de JSON. Há limite de quatro requisições simultâneas por rota de criação. As demais rotas conservam o limite de 2 MiB.

Os documentos são enviados ao bucket privado configurado, com chaves criadas pelo servidor e isoladas por organização. A conta do storage precisa apenas da permissão `PutObject` já usada no upload normal. O banco armazena metadados, não o Base64; o conteúdo não entra no log de auditoria. PDFs anexados como `CONTRACT` não são automaticamente considerados assinados.

As gravações no banco são transacionais. Em falhas de storage, os novos registros são revertidos. Se houver falha após gravar um objeto e antes de confirmar o banco, esse objeto pode permanecer sem referência; o reenvio do mesmo payload usa a mesma chave, sem destruir arquivos de uma confirmação incerta. Não há remoção automática de objetos órfãos nesta entrega.

## Reenvios e cadastro existente

A identidade da venda é `organização + sourceSystem + sale.id`. Primeiro recebimento: `201`; reenvio idêntico: `200`, com `replayed: true`, sem repetir venda, cliente ou upload. Mesmo ID com conteúdo diferente: `409 INTAKE_REPLAY_CONFLICT`; nenhum registro é sobrescrito. Datas e valores normalizados são considerados na comparação; mantenha a ordem dos anexos.

Clientes são procurados por CPF/CNPJ na organização receptora. Um cliente ativo existente é reaproveitado, mantendo seu cadastro; os dados recebidos ficam preservados como snapshot na venda. Um cliente inativo ou mais de um cadastro para o documento produz conflito. Cadastros novos recebem contatos e endereço informados. As criações normal e por intake usam o mesmo bloqueio por documento para evitar duplicidade concorrente.

Resposta:

```json
{
  "customer_id": "uuid-interno-do-cliente",
  "sale_id": "uuid-interno-da-venda",
  "intake_id": "uuid-do-recebimento",
  "replayed": false
}
```

A venda entra no módulo existente como confirmada, sem baixa de pagamento, cobrança automática ou disparo de assinatura. Datas exatas e os demais campos recebidos permanecem no histórico de importação. Valor e comissão não geram automaticamente lançamentos financeiros.

Operadora, plano, número, chip, portabilidade, aparelho, parceiro e filial são informações comerciais da origem, exibidas na ficha e no detalhe da venda. Não são convertidas automaticamente em vínculos operacionais: o payload não fornece conta da operadora, IDs internos, marca/modelo do aparelho nem as aprovações necessárias. Assim, não há ativação de linha ou movimentação de estoque fictícia.

## Publicação e verificação

1. Configure a chave e a organização no `.env` da VPS e a mesma chave no remetente.
2. Atualize o nginx da VPS com o arquivo correspondente em `docker/nginx/`. Ele inclui as exceções de 24 MiB nas duas URLs de criação; sem isso, arquivos maiores podem receber `413` antes de chegar à API. Execute `sudo nginx -t` e recarregue somente se passar.
3. Execute `bash docker-up.sh prod` sem `down`. A migração `034_customer_intake.sql` é aplicada pela API ao iniciar; faça backup do banco antes de publicar alterações de esquema.
4. Faça um envio de teste, reenvie o mesmo arquivo e confira `201` seguido de `200`. Abra o cliente, a seção de vendas recebidas e o detalhe da venda. Confira documentos, comissão e valores.

Exemplo de chamada (defina `TELEFONIA_INTAKE_API_KEY` no shell sem publicar seu valor):

```sh
curl --fail-with-body --max-time 90 \
  -H 'Content-Type: application/json' \
  -H "X-API-Key: $TELEFONIA_INTAKE_API_KEY" \
  --data-binary @payload.json \
  https://telefonia.redobrai.online/api/v1/intake/customers
```

Contrato formal: `docs/TELEFONIA_INTAKE.openapi.yaml`. Testes de integração exigem PostgreSQL de teste; nunca execute a suíte contra o banco de produção.

## Validação local desta entrega

Testes Go e `go vet` passaram, incluindo moeda exata, campos condicionais, MIME/Base64, isolamento da chave de intake e upload privado em servidor HTTP de teste. A interface compilou; os componentes novos passaram no ESLint. O contrato OpenAPI e suas referências foram validados, assim como o Compose com chave/organização fictícias.

O teste de integração `TestIntakeTransactionReplayAndTenantIsolation` foi incluído, mas ficou ignorado por indisponibilidade do PostgreSQL local. O Docker estava desligado. A migração real, o fluxo completo com MinIO/PostgreSQL e `nginx -t` ainda precisam de homologação antes da publicação na VPS.
