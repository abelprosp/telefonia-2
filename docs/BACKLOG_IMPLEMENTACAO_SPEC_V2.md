# Backlog de implementação — aderência à especificação funcional v2

Documento operacional para fechar a lacuna entre o [Documento de Especificação Funcional-v2.md](./Documento%20de%20Especificação%20Funcional-v2.md) e o código em `src/`.

**Como usar:** cada épico pode virar milestone/épico no gestor de trabalho; as tarefas são incrementos entregáveis (backend, frontend ou ambos).

---

## Prioridade e rótulos

| Tag    | Significado                                                                                                           |
| ------ | --------------------------------------------------------------------------------------------------------------------- |
| **P0** | Bloqueia operação fiel ao doc (mês de processamento, matriz §2.2, vínculo linha–cliente, fecho/export se obrigatório) |
| **P1** | Alta — relatórios, contingência, composição mínima utilizável                                                         |
| **P2** | Média — qualidade, auditoria, PDF, refinamentos                                                                       |
| **P3** | Baixa / polish / UX                                                                                                   |
| **V2** | Marcado como V2 na spec ou depende de decisão de escopo                                                               |
| **B**  | Eixo B (cobrança company → cliente final) — projeto separado                                                          |

---

## Épico 1 — Mês de processamento e fecho (§2.1, §9, §11)

| ID  | Tarefa                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | Prioridade                                                           |
| --- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------- | ----- |
| 1.1 | Modelar **Mês de Processamento** (agregado + EF + repositório): operadora, período civil ou etiqueta Luxus, estado Aberto/Fechado, organização — **feito:** `ProcessingMonth`, `ProcessingMonthStatus`, `IProcessingMonthRepository`, tabela `ProcessingMonths`, migração `AddProcessingMonths`, `IAppUnitOfWork.ProcessingMonths`                                                                                                                                                                                                                                      | P0 ✅                                                                |
| 1.2 | Exigir `processing_month_id` (ou equivalente) em `ProviderInvoiceImportRequest` / fluxo de importação; associar `ProviderInvoice` ao mês — **feito:** input/command/validator/handler, colunas `ProcessingMonthId` em `ProviderInvoiceImportRequest` e `ProviderInvoice`, FKs e migração `AddProcessingMonthToInvoiceImportAndInvoice`                                                                                                                                                                                                                                  | P0 ✅                                                                |
| 1.3 | Regras: bloquear novas importações / operações mutáveis quando mês **fechado** — **feito (escopo atual):** bloqueio em `RequestInvoiceImportCommandHandler` (solicitação) e `ProcessImportInvoiceCommandHandler` (processamento), validando `ProcessingMonthStatus.OPEN`                                                                                                                                                                                                                                                                                                | P0 ✅                                                                |
| 1.4 | Comandos + API: criar mês, listar, fechar (`CloseProcessingMonth`), fecho em contingência com justificativa e utilizador (§2.4 mínimo administrativo) — **feito:** `ProcessingMonthsController` (GET/list, GET/id, POST/create, POST/close, POST/close-contingency), handlers e validações (`CreateProcessingMonth`, `CloseProcessingMonth`, `CloseProcessingMonthInContingency`), query repository e auditoria de fecho com `ClosedBy`/`ClosedInContingency`/`ContingencyJustification`                                                                                | P0 ✅                                                                |
| 1.5 | Atualizar chave de duplicidade de fatura se o doc exigir unicidade por **mês de processamento** explícito (além de conta/vencimento/contratante) — **feito:** `FindDuplicateByBusinessKeyAsync` exige `ProcessingMonthId` na chave; índice único em `ProviderInvoices` alterado de `(ProviderAccountId, ContractingCompanyId, BillingCycleId, DueDate)` para `(ProviderAccountId, ContractingCompanyId, ProcessingMonthId, DueDate)`; enum `ProviderInvoiceDuplication` simplificado (`Duplicate` / `None`); migração `ProviderInvoiceUniqueKeyIncludesProcessingMonth` | P0 ✅                                                                |
| 1.6 | **§11.2** — estado por cliente no mês: Pendente vs Liberado; comando de liberação admin — **feito:** entidade `CustomerProcessingMonthManualRelease` (justificativa + utilizador + instante), repositório/UoW, leitura `GetCustomerProcessingMonthBillingReadinessResponse` (automático por CNPJ = `ContractingCompany.TaxId` e faturas no mês; PF/sem regra → Pendente até liberação manual), `ManuallyReleaseCustomerForProcessingMonthCommand`, API `GET/POST .../customers/{id}/processing-months/{processingMonthId}/billing-readiness                             | manual-release`, migração `AddCustomerProcessingMonthManualReleases` | P1 ✅ |
| 1.7 | **§11.3** — trava retroativa: após fecho, impedir alteração de vigências/composições do período (extensão além de só `BillingCycle`) — **feito:** `ProcessingMonthDateRange` + `IProcessingMonthRepository.ExistsClosedIntersectingDateRangeAsync` (interseção intervalo ↔ mês civil de `ProcessingMonth` fechado); bloqueio em `CreateBillingCycle`, `UpdateBillingCycle` e criação de `BillingCycle` na importação quando o intervalo do ciclo intersecta competência de mês já fechado; notificação `PROCESSING_MONTH_RETROACTIVE_CHANGE_BLOCKED`                    | P1 ✅                                                                |
| 1.8 | Frontend: ecrãs de gestão de mês, fecho, contingência, filtro de faturas por mês — **feito:** `/processing-months`, filtro `processing_month_id` em `/invoices`, API de listagem de faturas com filtro opcional                                                                                                                                                                                                                                                                                                                                                         | P0 ✅                                                                |

---

## Épico 2 — Matriz automática §2.2 (importação)

| ID  | Tarefa                                                                                                                                                                                                                                                                                                                                                                                                                                                            | Prioridade |
| --- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------- |
| 2.1 | Na importação, após 110D: chamar `PhoneLine.ApplyImportedLinePresence` com data de emissão da fatura (transição → ativa) — **feito:** `ProcessImportInvoiceCommandHandler` agora usa `phoneLine.ApplyImportedLinePresence(invoice, invoice.IssueDate)` no processamento de linhas 110D                                                                                                                                                                            | P0 ✅      |
| 2.2 | Novo número na fatura: criar linha em **Em estoque** (`IN_STOCK`) quando sem cliente; estado inicial coerente com spec (não deixar só `INACTIVE` se o doc exige estoque) — **feito:** no `ProcessImportInvoiceCommandHandler`, linha nova identificada no 110D é marcada como `IN_STOCK` após aplicar presença importada                                                                                                                                          | P0 ✅      |
| 2.3 | Pós-processamento: linhas da conta ausentes no 110D → `AWAITING_INVOICE` / inativa em estoque conforme tabela §2.2 (incl. variação com/sem cliente) — **feito (escopo atual sem vínculo linha-cliente):** ausentes no 110D são pós-processadas por conta; `IN_STOCK` → `INACTIVE` (estoque inativo), e linhas operacionais (`ACTIVE`/`IN_TRANSITION`/`AWAITING_INVOICE`) → `AWAITING_INVOICE`                                                                     | P0 ✅      |
| 2.4 | Validador estrutural §3.1: cliente da linha vs cliente da fatura / estados coerentes (reintroduzir se necessário após vínculo linha–cliente) — **feito (escopo atual):** validação estrutural no import para rejeitar linha presente na fatura em estado incoerente (`INACTIVE`/`CANCELLED`/`SUSPENDED` ou `IN_TRANSITION` sem subtipo); checagem de cliente-da-linha vs cliente-da-fatura permanece dependente do vínculo explícito PhoneLine↔Customer (épico 3) | P0 ✅      |
| 2.5 | Logs estruturados + **alertas ao operador** (UI ou fila de notificações) para eventos da matriz — **feito:** resumo estruturado no `ProcessImportInvoiceCommandHandler` com contadores de transição (`new in stock`, `transition→active`, `absent→awaiting`, `absent→inactive stock`) + publicação em fila via evento `InvoiceImportMatrixAlertEvent` (topic `providers.invoice_import.matrix_alert`)                                                             | P1 ✅      |
| 2.6 | Garantir unicidade global do número de linha no estoque (§3) — **feito:** índice único global em `PhoneLines.Number` + migração `PhoneLineNumberGlobalUnique`; importação passa a resolver linha por número global e bloquear conflito entre contas                                                                                                                                                                                                               | P1 ✅      |

---

## Épico 3 — Vínculo linha ↔ cliente final (§3, §4, §8)

| ID  | Tarefa                                                                                                                                                                                                                                                                                                                                                                                                                                     | Prioridade |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------- |
| 3.1 | Modelar vínculo **PhoneLine ↔ Customer** (temporal se necessário: vigência início/fim) + migração — **feito:** entidade temporal `PhoneLineCustomerLink` (início/fim), navegações em `PhoneLine`/`Customer`, regra de 1 vínculo ativo por linha (índice único filtrado), migração `AddPhoneLineCustomerLinks`                                                                                                                              | P0 ✅      |
| 3.2 | Regras: linha em fatura sem cliente → estoque; com cliente ativo → cobrança conforme classificação — **feito (escopo atual):** importação usa vínculo ativo `PhoneLineCustomerLink`; linha presente no 110D sem cliente ativo permanece/entra em `IN_STOCK`; com cliente ativo aplica presença importada (incl. transição→ativa); ausente no 110D: sem cliente ativo → inativa em estoque, com cliente ativo → `AWAITING_INVOICE`          | P0 ✅      |
| 3.3 | API + Application: associar/desassociar cliente, listar linhas por cliente / cliente por linha — **feito:** comandos/handlers de associação e desassociação de vínculo ativo da linha; API `POST/DELETE /phone-lines/{id}/customer-links` + consultas `GET /phone-lines/{id}/customer-links` e `GET /customers/{id}/phone-lines` (paginada)                                                                                                | P0 ✅      |
| 3.4 | Importação: resolver ou sugerir cliente (regras com 011D / cadastro) — **feito (escopo atual):** resolução automática por documento 011D quando existir 1 cliente ativo compatível na operadora; em casos ambíguos, mantém sem vínculo automático e gera sugestão via log estruturado com IDs candidatos                                                                                                                                   | P0 ✅      |
| 3.5 | **§8.1–8.2** — transferir linha entre clientes com histórico imutável — **feito:** comando transacional `TransferPhoneLineCustomer` (fecha vínculo ativo e abre novo na mesma data, sem apagar histórico), com validações de cliente-alvo distinto, vínculo ativo obrigatório e compatibilidade de operadora; endpoint `POST /phone-lines/{id}/customer-links/transfer`                                                                    | P1 ✅      |
| 3.6 | **§8.3** — inativar cliente automaticamente quando zero linhas ativas (+ reativação se aplicável) — **feito:** nos fluxos de vínculo (`assign`, `transfer`, `unassign`) o cliente origem é inativado quando fica sem linhas ativas; cliente destino é reativado ao receber linha. Reativação também aplicada na associação automática via importação (011D)                                                                                | P1 ✅      |
| 3.7 | Alinhar **CNPJ Luxus (contratada)** no `Customer` com `ContractingCompany` e validações de importação — **feito:** importação valida documento 011D (CPF/CNPJ); para CNPJ mantém resolução/criação da `ContractingCompany` por CNPJ; para CPF exige cliente único existente na operadora e deriva CNPJ da contratada a partir do cadastro do cliente; bloqueia inconsistência `Customer`↔`ContractingCompany` com erro de regra de negócio | P1 ✅      |
| 3.8 | Frontend: vínculos na ficha da linha e do cliente — **feito:** ficha da linha com histórico de vínculos + ações de vincular/transferir/desvincular cliente ativo; ficha do cliente com listagem paginada de linhas vinculadas (status, classificação, vigência) e atalho para abrir detalhe da linha                                                                                                                                       | P0 ✅      |

---

## Épico 4 — Relatórios §2.3

| ID  | Tarefa                                                                                                                  | Prioridade |
| --- | ----------------------------------------------------------------------------------------------------------------------- | ---------- |
| 4.1 | Repositório de leitura: **Entradas** (primeira aparição ou retorno após ausência) filtrado por **mês de processamento** | P1         |
| 4.2 | **Saídas** (presente no mês anterior, ausente no corrente)                                                              | P1         |
| 4.3 | **Pendências de ativação** (em transição sem fatura no mês; contagem em ciclos ou meses — alinhar ao doc)               | P1         |
| 4.4 | API + autorização admin                                                                                                 | P1         |
| 4.5 | Export CSV/Excel/PDF (opcional por fase)                                                                                | P2         |
| 4.6 | Frontend: substituir placeholder `/reports/transition-pending` por dados reais + navegação para os três relatórios      | P1         |

---

## Épico 5 — Contingência §2.4 (negócio completo)

| ID  | Tarefa                                                                                                     | Prioridade |
| --- | ---------------------------------------------------------------------------------------------------------- | ---------- |
| 5.1 | Definir com negócio: o que é “processar o mês” sem fatura (snapshot de cobrança? só fecho administrativo?) | P1         |
| 5.2 | Implementar comandos + persistência de justificativa + auditoria                                           | P1         |
| 5.3 | Frontend: fluxo admin                                                                                      | P2         |

---

## Épico 6 — Estoque §3.2–3.3

| ID  | Tarefa                                                                                                                                                                                                                                                                                                                                                                    | Prioridade |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------- |
| 6.1 | Campos **custo base** e **com consumo** na linha (ou agregado derivado da fatura) preenchidos na importação — **feito:** adicionados campos `BaseCost` e `CostWithConsumption` em `PhoneLine` (persistidos e expostos nas responses); importação preenche snapshot financeiro por linha com base no total do registro 110D                                                | P1 ✅      |
| 6.2 | Atualização de **última fatura** ao inativar / ausência (já parcialmente com `LastInvoiceId`) — **feito:** no pós-processamento de ausência (linhas não presentes no 110D), a linha passa a registrar explicitamente a fatura processada em `LastInvoiceId` antes da transição para `INACTIVE` (estoque) ou `AWAITING_INVOICE`                                            | P1 ✅      |
| 6.3 | Documentar no doc de produto se histórico de mudança de conta/CNPJ é obrigatório (spec §3.3 menciona que histórico não é necessário — apenas refletir última fatura) — **feito:** `docs/PRODUTO_E_ROADMAP.md` atualizado na seção de Estoque (§3) explicitando que histórico de troca conta/CNPJ não é obrigatório no escopo atual, prevalecendo reflexo da última fatura | P3 ✅      |

---

## Épico 7 — Clientes §4

| ID  | Tarefa                                                                                                                                                                                                                                                                        | Prioridade |
| --- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------- |
| 7.1 | **Vendedor responsável** no modelo + API + UI — **feito:** `Customer.ResponsibleSalespersonUserId` (nullable, 256); `POST /customers` e `PATCH /customers/{id}`; listagem/detalhe; migração EF; OpenAPI + UI (cadastro, edição, coluna na lista)                              | P2 ✅      |
| 7.2 | Documentação/arquivos no cadastro (upload, metadados) — **feito:** entidade `CustomerAttachment`; `GET/POST/DELETE …/customers/{id}/attachments`; registro pós-upload (presigned URL + metadados); OpenAPI + Kubb + UI na ficha do cliente; migração `AddCustomerAttachments` | P2 ✅      |
| 7.3 | **PJ Revendedor** [V2] + flag em `Customer`                                                                                                                                                                                                                                   | V2         |
| 7.4 | **§4.3** — geração automática de contratos + TAGs (depende de documento de templates)                                                                                                                                                                                         | V2 / P3    |

---

## Épico 8 — Hierarquia §5

| ID  | Tarefa                                                                                    | Prioridade |
| --- | ----------------------------------------------------------------------------------------- | ---------- |
| 8.1 | API para definir classificação e titular em `PhoneLine`                                   | P1         |
| 8.2 | Logs de reclassificação (auditoria de negócio)                                            | P2         |
| 8.3 | Regras de consolidação de cobrança / boletos (depende de Eixo B ou integração financeira) | P1 / B     |

---

## Épico 9 — Composição financeira e portfólio §6

| ID  | Tarefa                                                                                            | Prioridade |
| --- | ------------------------------------------------------------------------------------------------- | ---------- |
| 9.1 | API CRUD para `PhoneLineService` (adicionar/remover serviço do plano na linha, preço, recorrente) | P1         |
| 9.2 | Unicidade **um serviço ativo por tipo** (`ServiceType` em `ProviderPlanService` se necessário)    | P1         |
| 9.3 | Descontos em linha e em cliente; parcelas aparelho; cobrança avulsa (modelo + API)                | P2         |
| 9.4 | Excedentes: manual V1; regras automáticas [V2]                                                    | P2 / V2    |
| 9.5 | Frontend: UI de composição na ficha da linha                                                      | P1         |

---

## Épico 10 — Proporcionalidade §10

| ID   | Tarefa                                                                                                   | Prioridade |
| ---- | -------------------------------------------------------------------------------------------------------- | ---------- |
| 10.1 | Expor API de pré-visualização usando `ProportionalBillingCalculator` (datas + valores)                   | P1         |
| 10.2 | Persistir opção por serviço/linha (ex.: flag) e integrar no fecho ou na composição — definir com negócio | P1         |
| 10.3 | Opcional: aplicar pro-rata na importação (só se produto exigir reescrita vs valores operadora)           | P2         |

---

## Épico 11 — Exportação sistema financeiro §11.4 / doc §13

| ID   | Tarefa                                                               | Prioridade |
| ---- | -------------------------------------------------------------------- | ---------- |
| 11.1 | Definir formato (CSV, JSON, API dedicada, SFTP, etc.) com financeiro | P0         |
| 11.2 | Implementar geração pós-fecho ou sob demanda + permissões admin      | P0         |
| 11.3 | Frontend: download / agendamento (se aplicável)                      | P1         |

---

## Épico 12 — PDF §2.1

| ID   | Tarefa                                                      | Prioridade |
| ---- | ----------------------------------------------------------- | ---------- |
| 12.1 | Pipeline de extração/normalização por operadora             | P2         |
| 12.2 | Mesmas regras de negócio que TXT (matriz, duplicidade, mês) | P2         |

---

## Épico 13 — Auditoria §14–15

| ID   | Tarefa                                                                                           | Prioridade |
| ---- | ------------------------------------------------------------------------------------------------ | ---------- |
| 13.1 | Tabela ou store append-only de eventos de negócio (importação, hierarquia, transferência, fecho) | P2         |
| 13.2 | UI de consulta para admin                                                                        | P3         |

---

## Épico 14 — V2 (§12–13)

| ID   | Tarefa                                                  | Prioridade |
| ---- | ------------------------------------------------------- | ---------- |
| 14.1 | Fidelidade: vigência, renovação, gatilhos configuráveis | V2         |
| 14.2 | Processamento duplo / espelhamento para PJ Revendedor   | V2         |
| 14.3 | Excedentes automáticos com termos configuráveis         | V2         |

---

## Épico 15 — Eixo B (cobrança gerada pela company)

| ID   | Tarefa                                                                     | Prioridade |
| ---- | -------------------------------------------------------------------------- | ---------- |
| 15.1 | Event storming / bounded context: agregados distintos de `ProviderInvoice` | B          |
| 15.2 | Modelo de fatura interna, boleto consolidado, integração com §5 e §11      | B          |

---

## Épico 0 — Base técnica e qualidade

**Priorização:** executar **por último**, depois dos épicos 1–15 (higiene, testes de integração, documentação OpenAPI — não bloqueia entrega funcional da spec).

| ID  | Tarefa                                                                                                                                                                  | Prioridade |
| --- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------- |
| 0.1 | Corrigir compilação de `Luxus.Connect.IntegrationTests` (alinhamento a `ProviderInvoice`, `PhoneLine`, remoção de `ProcessingMonth` / namespaces antigos)               | P1         |
| 0.2 | Cobrir com testes a importação TXT: duplicidade, criação de `PhoneLine`, persistência de `ProviderInvoice`                                                              | P1         |
| 0.3 | Revisar notificações/mensagens em `Notifications` que ainda falam em “processing month” sem entidade correspondente — alinhar texto ao modelo real ou à futura entidade | P2         |
| 0.4 | Documentar no OpenAPI/Scalar os fluxos principais (importação, storage)                                                                                                 | P3         |

---

## Ordem sugerida de execução

1. **Épico 1** + **2** + **3** em paralelo controlado (mês de processamento, matriz §2.2, vínculo cliente são interdependentes).
2. **Épico 4** e **6** assim que mês e linhas estiverem estáveis.
3. **Épico 9** e **8** em paralelo com **10** conforme prioridade financeira.
4. **Épico 11** quando fecho mensal for real.
5. **5**, **7**, **12**, **13**, **14**, **15** conforme negócio.
6. **Épico 0** por último (integração tests, smoke automatizado, notificações/OpenAPI).

---

## Critério de “done” global (aderência v2)

- Operador consegue: criar **mês de processamento**, associar **lotes** de importação, ver **matriz §2.2** aplicada com **alertas**, **fechar** o mês (e contingência quando definido), extrair **relatórios §2.3**, gerir **vínculo linha–cliente** e **composição**, e **exportar** para o financeiro — tudo com **auditoria** mínima.
- Itens **V2** e **Eixo B** podem permanecer fora do “done” de V1 se o produto assim decidir;
