# SISTEMA LUXUS GESTÃO

Documento de Especificação Funcional — V1 e V2
Versão do documento: 2.0 · Fevereiro 2026

## Sobre este documento

Este documento especifica funcionalmente o Sistema Luxus Gestão, cobrindo as versões 1 (V1) e 2 (V2). O objetivo é dar aos desenvolvedores uma referência precisa, sem ambiguidades, sobre o que o sistema deve fazer — e o que está fora de escopo em cada versão.

| Versão 1 — MVP                                                                                                                                                               | Versão 2 — Expansão                                                                                                                   |
| ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| Produto mínimo viável, simples e fiel à operação real da Luxus. Cobre: controle de linhas, estoque, clientes, composição financeira básica e ciclo de faturamento principal. | Iniciada após estabilização e validação da V1 em operação real. Funcionalidades V2 são marcadas ao longo do documento com o selo [V2] |

#### ℹ Sobre o sistema financeiro externo

O Sistema Luxus Gestão calcula e consolida valores, mas não emite boletos nem notas fiscais. A exportação dos valores para o sistema financeiro responsável pela emissão é abordada na Seção 13.

### 1. Glossário Operacional

Os termos abaixo são usados de forma padronizada em todo o documento e devem ser refletidos nos nomes de entidades, campos e variáveis do sistema.

| Termo                   | Definição                                                                                                                                                                                            |
| ----------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Linha                   | A entidade central do sistema. Um número de celular ativo, com todos os seus metadados, serviços e vínculos. Tudo no sistema gira em torno da linha.                                                 |
| Conta (Nuvem)           | Agrupador de linhas junto à operadora. Cada Conta pertence a um CNPJ do Grupo Luxus e pode conter centenas ou milhares de linhas.                                                                    |
| Fatura Origem           | Documento mensal emitido pela operadora à Luxus para uma Conta específica. É a autoridade primária para liberação de cobrança. Cada CNPJ pode ter mais de uma conta.                                 |
| Ciclo                   | Período de 30 dias considerado por uma Fatura Origem para gerar cobranças. Cada Conta tem seu próprio ciclo (datas de início e fim variáveis). Não confundir com "mês de processamento".             |
| Mês de Processamento    | Conceito operacional da Luxus. Agrupa todas as Faturas Origem que a equipe designa como pertencentes a um mesmo mês de cobrança ao cliente — independentemente dos ciclos individuais de cada Conta. |
| Portfólio de Serviços   | Catálogo pré-cadastrado de serviços comercializáveis. Funciona como cardápio controlado para composição das linhas.                                                                                  |
| Composição da Linha     | Conjunto de serviços, descontos, aparelhos e excedentes que define o valor a ser cobrado por uma linha em um ciclo.                                                                                  |
| CNPJ Luxus (Contratada) | CNPJ do Grupo Luxus sob o qual o contrato com um cliente foi celebrado. É um atributo do cadastro do cliente, não da linha.                                                                          |

### 2. Módulo de Entrada de Dados — Faturas da Operadora

O sistema é alimentado pelas Faturas Origem das operadoras. A fatura é a autoridade primária: nenhuma linha é cobrada de cliente sem que tenha aparecido em fatura, salvo por comando manual de administrador (ver Seção 13).

#### 2.1 Importação de Faturas

• Formatos suportados: TXT (preferencial) e PDF (alternativo).
• Cada fatura importada é identificada pela combinação: Operadora + CNPJ Luxus + Conta + Vencimento.
• O sistema deve controlar as importações de forma a rejeitar a importação de uma fatura com combinação idêntica já existente (mesmo mês, mesma conta), assim como de meses diferentes, exibindo erro claro ao operador.
• O operador deve designar explicitamente a qual "mês de processamento" cada lote de faturas pertence no momento da importação.

#### 2.2 Validação Automática por Linha

Ao processar cada linha contida na fatura, o sistema responde automaticamente às seguintes perguntas:

| Condição detectada                                     | Ação automática do sistema                                                                                                              |
| ------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------- |
| Linha inexistente no sistema                           | Cria a linha e classifica como Em Estoque. Gera log: "Linha criada em estoque automaticamente a partir da fatura [Conta / Vencimento]." |
| Linha existe e está vinculada a cliente ativo no ciclo | Marca linha como processada para o ciclo. Libera para cálculo.                                                                          |
| Linha existe mas estava "Em Transição"                 | Atualiza status para Ativa com a data de início constante na fatura. Libera para cálculo. Gera log de conciliação.                      |
| Linha existe mas está no Estoque                       | Mantém no estoque. Nenhuma cobrança gerada.                                                                                             |
| Linha que estava ativa some da fatura                  | Não consolida cobrança. Mantém o vínculo com o cliente, mas com status "Aguardando Fatura". Gera alerta ao operador.                    |
| Linha que estava em Estoque some da fatura             | A linha fica como “inativa” no estoque com indicação da última fatura em que ela apareceu (ou a primeira em que NÃO apareceu).          |

```⚠ Escopo da validação

A validação é estritamente operacional: o sistema verifica existência e vínculo das linhas. Não analisa chamadas, minutos, consumo detalhado nem conformidade tarifária da fatura.
```

#### 2.3 Relatórios de Movimentação

A cada mês de processamento concluído, o sistema deve gerar automaticamente:
• Relatório de Entradas: linhas que apareceram em fatura pela primeira vez ou que retornaram após inatividade.
• Relatório de Saídas: linhas que constavam no mês anterior e não aparecem no corrente.
• Relatório de Pendências de Ativação: linhas em status "Em Transição" que não apareceram em fatura, com indicação de há quantos ciclos estão pendentes.

#### 2.4 Processamento em Modo de Contingência

Se uma ou mais Faturas Origem não estiverem disponíveis no prazo esperado, o sistema deve permitir que o mês seja processado com base nos dados já registrados (clientes, linhas, serviços ativos), sem aguardar indefinidamente pela operadora. Essa operação exige comando explícito de administrador e gera log de justificativa.

### 3. Controle de Estoque de Linhas

O estoque é formado e mantido de forma automática a partir do processamento das Faturas Origem.

#### 3.1 Definição e Critério de Inclusão

São consideradas Em Estoque todas as linhas que constam nas Faturas Origem importadas e não estão vinculadas a nenhum cliente no período correspondente.

```
Regra estrutural
Toda linha que consta em fatura deve estar em uma de duas situações: vinculada a um cliente ativo, ou em estoque. Não existe linha faturada sem destino.
```

#### 3.2 Dados de Cada Linha em Estoque

Todos os campos abaixo são preenchidos automaticamente a partir da fatura, sem entrada manual:

| Campo           | Descrição                                                                                                              |
| --------------- | ---------------------------------------------------------------------------------------------------------------------- |
| Número da linha | Identificador único. Não pode haver duplicatas no estoque.                                                             |
| Status          | Em Estoque ou Inativa.                                                                                                 |
| Operadora       | Originada da fatura.                                                                                                   |
| CNPJ Luxus      | CNPJ do Grupo Luxus ao qual a Conta pertence na operadora.                                                             |
| Conta (Nuvem)   | Identificador da Conta na operadora.                                                                                   |
| Custo base      | Valor total de serviços da linha, excluindo parcelamentos de aparelho e multas.                                        |
| Com consumo     | Sim/Não. Indica se houve consumo registrado na fatura.                                                                 |
| Última fatura   | Referência da última Fatura Origem em que a linha apareceu. Preenchida automaticamente quando a linha passa a Inativa. |

#### 3.3 Regras de Atualização e Inativação

• O estoque é atualizado a cada processamento de fatura. Os dados da linha (operadora, CNPJ, conta) refletem sempre a última fatura processada.
• Se uma linha apareceu em meses anteriores e não aparece no mês corrente, passa automaticamente para status Inativa, com registro da última fatura em que esteve presente.
• Uma linha Inativa que reaparecer em fatura retorna automaticamente ao status Em Estoque.
• Não pode existir a mesma linha duas vezes no estoque — unicidade garantida pelo número da linha.
• Quando uma linha muda de Conta, operadora ou de CNPJ Luxus dentro da operadora, os dados são atualizados conforme a fatura mais recente. Histórico de conta anterior não é necessário.

### 4. Módulo de Clientes

Este módulo define os clientes da Luxus e os vínculos entre clientes e linhas. É a partir daqui que se determina quem paga o quê.

#### 4.1 Cadastro de Cliente

Todo cliente possui os seguintes atributos de cadastro:

| Campo                   | Observação                                                                                                                             |
| ----------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| Tipo                    | PF (Pessoa Física) ou PJ (Pessoa Jurídica).                                                                                            |
| CPF / CNPJ              | Documento principal do cliente.                                                                                                        |
| Nome / Razão Social     | Nome completo (PF) ou razão social (PJ).                                                                                               |
| Data de ativação        | Data de início do relacionamento comercial.                                                                                            |
| CNPJ Luxus (Contratada) | CNPJ do Grupo Luxus sob o qual o contrato foi celebrado. Atributo do cliente, não da linha.                                            |
| Vendedor responsável    | Quem abriu o cliente. Qualquer vendedor pode ser associado a qualquer cliente ou linha, independentemente de quem realizou a abertura. |
| Documentação            | Arquivos vinculados ao cadastro (contratos, anexos, etc.).                                                                             |
| Flag: PJ Revendedor     | [V2] Indica se o cliente PJ revende linhas a usuários finais. Ativa o Processamento Duplo (ver Seção 8).                               |

#### 4.2 Tipos de Cliente

Pessoa Física (PF): contratação direta de linhas para uso próprio ou grupo menor.
Pessoa Jurídica (PJ): pode contratar linhas para uso interno ou para revenda. A flag "PJ Revendedor" ativa o módulo de Processamento Duplo [V2].

#### 4.3 Geração Automática de Contratos

Cada venda — abertura de cliente, vínculo de nova linha ou adição de serviço — deve gerar automaticamente um contrato da Luxus com os respectivos anexos, preenchido com as TAGs do cadastro do cliente.

```
📄 Documento complementar
O mapeamento completo das TAGs disponíveis, os templates de contrato e seus respectivos anexos serão detalhados em documento separado, a ser fornecido antes do início da implementação desta funcionalidade.
```

### 5. Hierarquia de Cobrança — Titular, Dependente e Normal

Cada linha vinculada a um cliente recebe uma classificação hierárquica que determina como sua cobrança é gerada. Essa classificação define diretamente quantos boletos (unidades de cobrança) um cliente terá no ciclo.

#### 5.1 Classificações

| Classificação | Comportamento de cobrança                                                                   | Observação                                                    |
| ------------- | ------------------------------------------------------------------------------------------- | ------------------------------------------------------------- |
| Normal        | Gera cobrança individual. Um boleto por linha Normal.                                       | Classificação padrão. Toda linha começa como Normal.          |
| Titular       | Gera um único boleto que consolida o valor desta linha mais o de todas as suas Dependentes. | Deve ter ao menos uma Dependente para manter a classificação. |
| Dependente    | Não gera cobrança própria. Seu valor é somado ao boleto do Titular.                         | Deve estar obrigatoriamente vinculada a um Titular.           |

#### 5.2 Regras Estruturais da Hierarquia

• Toda linha Dependente deve estar vinculada a um Titular. Se o vínculo se romper por qualquer motivo, a linha é reclassificada automaticamente para Normal.
• Se um Titular perder todas as suas Dependentes, é reclassificado automaticamente para Normal.
• Reclassificações automáticas são registradas em log com data, motivo e linhas afetadas.
• Todas as linhas, independentemente da posição hierárquica, devem ter um cadastro próprio e unitário para conter os devidos dados cadastrais.

#### 5.3 Cenários de Agrupamento

Os cenários abaixo ilustram como a hierarquia se traduz em boletos:

| Cenário             | Descrição                                                                                | Boletos gerados             |
| ------------------- | ---------------------------------------------------------------------------------------- | --------------------------- |
| Linha única         | Cliente com uma única linha (classificação Normal).                                      | 1 boleto                    |
| Múltiplas Normais   | Cliente com duas ou mais linhas, todas Normais.                                          | 1 boleto por linha Normal   |
| Grupo com Titular   | Um Titular com N Dependentes. Os clientes das Dependentes podem ser cadastros distintos. | 1 boleto no nome do Titular |
| Múltiplos Titulares | Cliente com duas linhas Titulares, cada uma com suas Dependentes.                        | 2 boletos (um por Titular)  |

### 6. Composição Financeira da Linha

Cada linha vinculada a um cliente possui uma ficha de composição que define como seu valor é calculado no ciclo. A composição é montada pelo operador a partir do Portfólio de Serviços pré-cadastrado.

#### 6.1 Portfólio de Serviços

O Portfólio é o catálogo de serviços disponíveis para uso nas linhas. Funciona como um cardápio controlado, garantindo padronização e governança de preços. Cada serviço cadastrado possui:

| Atributo                 | Descrição                                                                                                                                         |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| Nome em sistema          | Identificação interna para busca e operação.                                                                                                      |
| Nome em fatura           | Nome que aparece na composição da cobrança enviada ao cliente.                                                                                    |
| Tipo                     | Assinatura, Dados, SMS, Roaming ou Outros.                                                                                                        |
| Operadora                | Operadora à qual o serviço pertence.                                                                                                              |
| Valor padrão             | Valor base. Pode ser customizado na linha sem alterar o cadastro do portfólio.                                                                    |
| Definição de aplicação   | Define se o serviço é cobrado pela Luxus ao cliente, pelo cliente ao usuário final, ou nos dois casos. Relevante para o Processamento Duplo [V2]. |
| Regra de disponibilidade | Global (disponível para todos os clientes) ou Exclusivo (disponível apenas para um cliente específico).                                           |

#### 6.2 Componentes da Composição de uma Linha

Ao vincular uma linha a um cliente, o operador monta sua composição selecionando itens do Portfólio. Os componentes possíveis são:

##### 6.2.1 Serviços Recorrentes

Serviços do Portfólio aplicados à linha. Ao adicionar um serviço, o operador define:
• Data de início da vigência.
• Data de fim (quando aplicável).
• Valor customizado para esta linha (opcional — se não informado, usa o valor padrão do Portfólio).
• Vendedor responsável pela linha (pode ser diferente do vendedor que abriu o cliente).

```
Regra de unicidade por tipo
Não pode existir mais de um serviço do mesmo tipo ativo simultaneamente em uma linha. Exemplo: duas assinaturas ativas na mesma linha são bloqueadas pelo sistema.
```

##### 6.2.2 Descontos

Podem ser aplicados no nível da linha individualmente. Atributos:
• Nome do desconto.
• Valor (fixo em R$).
• Data de início e data de fim (opcional).
• Aceita proporcionalidade (S ou N). Sobre proporcionalidade, item 10 deste doc.

Adicionalmente, é possível aplicar desconto no nível do cliente, que incide sobre o total consolidado do boleto (Titular ou Normal). Atributos idênticos ao desconto de linha: nome, valor e vigência.

##### 6.2.3 Aparelhos

Parcelamento de aparelho associado à linha. Atributos:
• Valor total do aparelho.
• Número de parcelas (meses).
• Datas de início e fim calculadas automaticamente.
• Valor da parcela mensal calculado automaticamente (valor total ÷ número de parcelas).

##### 6.2.4 Cobrança Avulsa

Para renegociações, ajustes e valores que não se enquadram nos outros componentes. Atributos:
• Nome descritivo.
• Valor.
• Data de início e data de fim (opcional).
• Aceita proporcionalidade (S ou N). Sobre proporcionalidade, item 10 deste doc.

##### 6.2.5 Excedentes

V1 => Entrada manual: o operador adiciona manualmente os valores de excedente à linha com base na fatura recebida. Sem automação de reconhecimento de termos. |
V2 => Detecção automática: o sistema mantém uma lista de termos configurável (ex.: "Roaming Internacional"). Quando detectados na fatura, os eventos podem gerar cobrança automaticamente conforme as regras da linha. Regras por linha: flag "Cobrar excedentes" (padrão: Sim). Se Não, o excedente é ignorado mesmo quando detectado.
Opção A — Espelhado: o cliente paga exatamente o valor cobrado pela operadora.
Opção B — Tabelado: o cliente paga um valor fixo pré-definido, independente do custo real da operadora.

### 7. Linhas em Transição — Venda Antecipada

O sistema permite cadastrar e configurar linhas que já foram negociadas com um cliente, mas que ainda dependem de processos externos (trâmites na operadora) para aparecerem na fatura da Luxus.

#### 7.1 O que o sistema permite no cadastro antecipado

1.  Vincular a linha a um cliente existente fora dos CNPJs processados pelo sistema— pois a linha pode estar vindo de fora do parque Luxus (sendo os tipos de trâmite para tal: portabilidade, TT, pré-pago).
2.  Definir o tipo de migração (obrigatório para rastreabilidade):
    – Portabilidade — linha vinda de outra operadora.
    – TT (Transferência de Titularidade) — linha existente migrando de outro titular para a Luxus.
    – PP (Migração Pré-Pago) — linha pré-paga sendo migrada para plano Luxus.
3.  Pré-configurar a composição completa da linha (serviços, valores, aparelhos).

```
Regra obrigatória
Mesmo em transição, a linha deve ter ao menos um serviço pré-configurado no cadastro antecipado. O cadastro sem composição não é permitido — regra estrutural da Seção 14.
```

#### 7.2 Status e Conciliação Automática

A linha em transição recebe o status: "Em Transição — Aguardando [Tipo]" (ex.: Aguardando Portabilidade).

Quando a linha aparece na importação de fatura da operadora:
• O sistema cruza o número da linha na fatura com o cadastrado em transição.
• O status é atualizado automaticamente para Ativa, com a data de início informada pela operadora na fatura.
• A composição pré-configurada entra em vigor.
• Gera log: "Linha [número] conciliada automaticamente. Status: Ativa desde [data]."

#### 7.3 Gestão de Divergências

Se a linha não aparecer na fatura conforme o previsto, ela permanece em transição e é listada no Relatório de Pendências de Ativação (Seção 2.3) com indicação de há quantos ciclos está pendente, para que a equipe cobre a operadora ou investigue falhas no processo de migração.

### 8. Relação Linha × Cliente ao Longo do Tempo

Linhas e clientes têm relacionamentos que evoluem. O sistema trata toda mudança como encerramento de um vínculo e abertura de um novo, preservando o histórico completo.

#### 8.1 Encerramento de Vínculo

• O operador define a data de fim do vínculo atual da linha com o cliente.
• Nenhum dado histórico é apagado — o vínculo encerrado permanece visível com status inativo e suas datas de início e fim.
• Após o encerramento, a linha pode ser: vinculada a outro cliente diretamente (transferência direta), devolvida ao estoque ou cancelada definitivamente (quando a linha sai das Contas da Luxus).

#### 8.2 Transferência Direta de Linha entre Clientes

O operador pode transferir uma linha diretamente do cliente A para o cliente B em um único ato operacional, sem que a linha passe pelo estoque. O sistema:
• Registra a data de transição informada pelo operador.
• Encerra o vínculo com o cliente A na data informada.
• Abre o novo vínculo com o cliente B na mesma data.
• Preserva o histórico do vínculo anterior com o cliente A.
• A composição financeira do novo vínculo (cliente B) deve ser configurada no ato da transferência.

#### 8.3 Cancelamento de Cliente

O cancelamento de um cliente acontece automaticamente quando todas as linhas não estiverem mais ativas com a Luxus (ou não houver alguma em cadastro antecipado). Sistema cancela um cliente quando todas as suas linhas já foram desvinculadas (encerradas ou transferidas), exibindo mensagem clara ao operador.

Regra estrutural
Cliente com linhas ativas sempre fica como Ativo. Cliente outrora Cancelado que ativa linha volta automaticamente para Ativo. Resumo: 1 ou mais linhas ativas = cliente Ativo; 0 linhas ativas = cliente Cancelado.

### 9. Ciclo de Faturamento e Mês de Processamento

O sistema não utiliza mês calendário como unidade de cobrança. Trabalha com o conceito de Mês de Processamento — um agrupador operacional definido pela equipe da Luxus.

#### 9.1 Mês de Processamento

Cada Conta (Nuvem) possui seu próprio ciclo de faturamento junto à operadora — com datas de início, fim e vencimento próprios, que variam de conta para conta. Não existe um ciclo único para toda a operação.

O "Mês de Processamento" é o conceito que unifica isso: a equipe da Luxus designa, no momento da importação, quais faturas pertencem ao mesmo mês de cobrança ao cliente. Todas as faturas designadas a um mesmo mês são processadas juntas e geram uma cobrança consolidada por cliente.

| Operadora / CNPJ Luxus    | Conta    | Vencimento | Mês de Processamento |
| ------------------------- | -------- | ---------- | -------------------- |
| VIVO / 00.000.000/0000-00 | 12345678 | 01/07/26   | Julho 2026           |
| VIVO / 00.000.000/0000-00 | 98756423 | 01/07/26   | Julho 2026           |
| TIM / 00.111.222/0000-01  | 6549843  | 01/07/26   | Julho 2026           |
| TIM / 00.111.222/0000-01  | 1111111  | 16/07/26   | Julho 2026           |
| VIVO / 00.000.000/0000-00 | 2222222  | 20/07/26   | Julho 2026           |

Neste exemplo, cinco faturas com vencimentos distintos (1/7, 16/7 e 20/7) são todas designadas ao Mês de Processamento "Julho 2026" e compõem juntas a cobrança daquele mês para os clientes Luxus.

#### 9.2 Blocos de Processamento

Ao longo de cada mês, a Luxus executa aproximadamente três blocos de processamento (número variável), conforme as faturas ficam disponíveis pelas operadoras. Alguns blocos trazem lotes maiores de faturas; outros trazem faturas pontuais. Todos os blocos designados ao mesmo mês compõem o "processamento do mês".

Enquanto o mês estiver aberto (não consolidado), o processamento pode ser refeito e ajustes são permitidos.

### 10. Proporcionalidade

Sempre que um evento começa ou termina no meio do ciclo, o valor é calculado proporcionalmente por dia. A regra se aplica a todos os serviços e a outros componentes de composição de cada linha que estejam com a proporcionalidade ativada (vide composição abordada no ponto 6.2.

| Situação                           | Exemplo                                                        | Cálculo                            |
| ---------------------------------- | -------------------------------------------------------------- | ---------------------------------- |
| Serviço iniciado no meio do ciclo  | Serviço de R$ 30,00 ativado no 16º dia de um ciclo de 30 dias. | R$ 30,00 ÷ 30 × 15 dias = R$ 15,00 |
| Serviço encerrado no meio do ciclo | Serviço de R$ 30,00 cancelado no 10º dia do ciclo.             | R$ 30,00 ÷ 30 × 10 dias = R$ 10,00 |
| Linha cancelada no meio do ciclo   | Linha com composição total de R$ 60,00 cancelada no 20º dia.   | R$ 60,00 ÷ 30 × 20 dias = R$ 40,00 |

```
Referência de dias
Todo ciclo de fatura tem duração de 30 dias para fins de cálculo proporcional, independente do mês calendário.
```

### 11. Processamento Mensal e Fechamento do Ciclo

#### 11.1 Fluxo de Processamento

O processamento mensal segue sempre a mesma sequência:

1.  Importação das Faturas Origem da operadora e designação ao Mês de Processamento.
2.  Identificação das linhas: o sistema cruza as linhas da fatura com o cadastro existente.
3.  Separação entre Estoque e Clientes: linhas sem vínculo vão para estoque; linhas vinculadas seguem para cálculo.
4.  Aplicação de vigências: para cada linha, o sistema verifica quais componentes estavam ativos no ciclo.
5.  Cálculo proporcional: componentes com início ou fim no meio do ciclo são calculados por dia.
6.  Consolidação por boleto: valores de Dependentes são somados ao Titular; Normais geram valor individual.
7.  Verificação de completude: o sistema checa se todas as linhas de cada cliente foram processadas (ver 11.2).
8.  Geração dos valores finais por cliente e exportação para o sistema financeiro externo.

#### 11.2 Trava de Completude por Cliente

Um cliente pode ter linhas distribuídas em múltiplas Contas com vencimentos diferentes. O sistema monitora o status de processamento de cada cliente:

| Status do cliente         | Condição                                                                                           |
| ------------------------- | -------------------------------------------------------------------------------------------------- |
| Pendente                  | Pelo menos uma linha do cliente pertence a uma Conta ainda não processada no mês.                  |
| Liberado para faturamento | 100% das linhas do cliente foram processadas (independente de virem de 1 ou 10 Contas diferentes). |

Quando todas as linhas do cliente estão processadas, o sistema soma tudo e libera a geração do valor final. O operador pode liberar manualmente (por cliente ou em lote) nos casos em que:
• Uma fatura não foi disponibilizada pela operadora no prazo esperado.
• Uma linha não apareceu em nenhuma fatura do mês (e a equipe confirma que a cobrança deve ocorrer assim mesmo).
Liberação manual exige comando explícito de administrador e gera log com justificativa.

#### 11.3 Fechamento do Ciclo

Após a validação interna e a geração dos valores finais, o operador executa o fechamento do ciclo. A partir desse momento:
• Os valores são consolidados e deixam de ser alteráveis.
• Qualquer ajuste retroativo é bloqueado pelo sistema.
• Correções de períodos fechados devem ser feitas preferencialmente via lançamentos compensatórios em ciclos futuros.
• Exceções servem para alterações pontuais que possibilitam o reprocessamento de clientes específicos para emissão correta. Não há reprocessamento em lote: se a alteração for feita, o novo valor a ser calculado pelo sistema é de cliente em cliente.

```
Trava de alterações retroativas
Uma vez fechado o ciclo, o sistema impede qualquer modificação de vigências, vínculos ou composições que impactem aquele período, exceto se feito por administradores.
```

#### 11.4 Exportação para Sistema Financeiro

Ao fechar o ciclo, o sistema disponibiliza para exportação os valores consolidados por cliente (e por boleto, respeitando a hierarquia Titular/Normal). O formato e o protocolo de exportação devem ser definidos em conjunto com a equipe responsável pelo sistema financeiro externo.

### 12. Vigência Contratual e Fidelidade

[V2] Esta seção é exclusiva da Versão 2.
Cada linha vinculada a um cliente possui um módulo de Gestão de Vigência Contratual com os seguintes atributos:

| Campo                        | Descrição                                                       |
| ---------------------------- | --------------------------------------------------------------- |
| Data de início da fidelidade | Data em que a fidelidade contratual da linha se inicia.         |
| Prazo inicial (meses)        | Duração do período de fidelidade inicial, em meses.             |
| Data final prevista          | Calculada automaticamente.                                      |
| Flag: renovação automática   | Sim/Não.                                                        |
| Período de renovação (meses) | Usado quando flag = Sim.                                        |
| Histórico de renovações      | Data + tipo (Automática / Por Alteração) + usuário responsável. |
| Status                       | Com contrato ativo ou Contrato expirado.                        |

#### 12.1 Renovação Automática

Se a flag estiver marcada como Sim, ao atingir a data de término, o sistema renova automaticamente pelo período configurado e registra o evento no histórico. A renovação automática adiciona períodos consecutivos sem alterar o prazo inicial do contrato original.

#### 12.2 Renovação por Alteração Contratual

Sempre que uma alteração estrutural da linha ocorrer (upgrade, downgrade, aquisição de aparelho, inclusão de novo serviço ou qualquer modificação que altere a composição estrutural), o sistema exibe ao operador:

```
"Esta alteração pode gerar renovação contratual. Deseja renovar a fidelidade pelo prazo inicial?"
Opções: Sim ou Não. Se Sim, o evento é registrado como "Renovação por Alteração". Se Não, a vigência permanece inalterada e a decisão é registrada no histórico.
```

A lista de eventos que disparam essa pergunta deve ser configurável pelo administrador. Por padrão, inclui: upgrade/downgrade de dados, aquisição de aparelho, alteração de plano e inclusão de novo serviço.

### 13. Processamento Duplo — PJ Revendedor

[V2] Esta seção é exclusiva da Versão 2.
Aplica-se exclusivamente a clientes PJ com a flag "PJ Revendedor" marcada. Nesses casos, a mesma linha passa a ter dois conjuntos paralelos de cálculo, representando perspectivas distintas:

|                            | Perspectiva A — Luxus → Cliente                            | Perspectiva B — Cliente → Usuário Final                                                             |
| -------------------------- | ---------------------------------------------------------- | --------------------------------------------------------------------------------------------------- |
| O que representa           | O que a Luxus cobra do cliente PJ.                         | O que o cliente PJ cobra do seu usuário final.                                                      |
| Composição                 | Serviços, descontos e regras normais da linha.             | Composição financeira independente, gerida pelo cliente.                                            |
| Rótulo (label)             | Nome do cliente                                            | Campo livre (ex.: nome do colaborador, setor, unidade). Não afeta a identificação técnica da linha. |
| Classificadores adicionais | UA, Setor, Centro de Custo ou outros campos configuráveis. | Idem — configuráveis conforme estrutura interna do cliente.                                         |

Flag de espelhamento: quando ativada, a Perspectiva B replica automaticamente todos os itens possíveis da Perspectiva A, atendendo casos em que o cliente apenas quer identificar o usuário final sem alterar valores ou regras.
Botão de cópia: ao ser ativada, a Perspectiva B replica automaticamente todos os itens possíveis da Perspectiva A, mas podem ser editados. O botão serve para facilitar e acelerar os preenchimentos.

### 14. Regras Estruturais do Sistema

As regras abaixo são invioláveis. Nenhum perfil de usuário, nem mesmo administradores, pode contorná-las — salvo os casos onde o próprio documento especifica "comando humano explícito".

| Regra                                                                                          | Justificativa / Exceção                                                                                                                                       |
| ---------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Cobrança só ocorre se a linha constar na Fatura Origem.                                        | Exceção: comando explícito de administrador com log de justificativa.                                                                                         |
| Venda e cadastro de linha independem da fatura.                                                | O vínculo com o cliente pode ser criado antes da linha aparecer em fatura (Seção 7).                                                                          |
| Uma linha não pode ser cadastrada para um cliente sem ao menos um serviço ativo na composição. | Aplica-se também a linhas em transição (venda antecipada).                                                                                                    |
| Nenhum histórico é apagado.                                                                    | Serviços, faturas, vínculos e clientes cancelados permanecem visíveis. O que muda são os status. Tudo possui data de início e, quando aplicável, data de fim. |
| Enquanto o ciclo não estiver fechado, ajustes são permitidos.                                  | Após o fechamento, alterações retroativas só são permitidas para administradores.                                                                             |
| Toda ação automática do sistema gera log visível ao operador.                                  | Reclassificações, conciliações, renovações automáticas — nada ocorre silenciosamente.                                                                         |
| Toda ação humana gera registro de usuário, data e descrição da ação.                           | Auditoria completa de operações.                                                                                                                              |

### 15. Tabela de Eventos, Regras e Ações Automáticas

Esta tabela consolida os principais eventos do sistema, as regras de negócio associadas, as ações automáticas esperadas e os logs gerados. Não é exaustiva — outras amarrações podem e devem ser implementadas conforme necessário.

#### Eventos de Hierarquia

| Evento                                       | Regra de negócio                         | Ação automática                                | Log gerado                                                |
| -------------------------------------------- | ---------------------------------------- | ---------------------------------------------- | --------------------------------------------------------- |
| Titular é cancelado / colocado como “Normal” | Dependente não pode existir sem Titular. | Reclassifica todas as Dependentes para Normal. | "Dependentes reclassificadas por ausência de Titular."    |
| Titular fica sem Dependentes                 | Titular sem Dependentes não faz sentido. | Reclassifica o Titular para Normal.            | "Titular [X] reclassificado por ausência de Dependentes." |

#### Eventos de Composição Financeira

| Evento                                                            | Regra de negócio                    | Ação automática                                            | Log gerado                                                                    |
| ----------------------------------------------------------------- | ----------------------------------- | ---------------------------------------------------------- | ----------------------------------------------------------------------------- |
| Serviço inicia no meio do ciclo                                   | Proporcionalidade obrigatória.      | Calcula valor proporcional por dia automaticamente.        | —                                                                             |
| Serviço encerra no meio do ciclo                                  | Cobrança apenas pelo período ativo. | Ajusta valor proporcional.                                 | —                                                                             |
| Tentativa de cadastrar dois serviços do mesmo tipo na mesma linha | Duplicidade não permitida.          | Bloqueia o cadastro.                                       | "Erro: já existe serviço do tipo [X] ativo nesta linha no período informado." |
| Linha cancelada no meio do ciclo                                  | Proporcionalidade obrigatória.      | Calcula valor proporcional de todos os componentes ativos. |                                                                               |

#### Transferência entre clientes

| Evento                                       | Regra de negócio               | Ação automática                                                       | Log gerado                                         |
| -------------------------------------------- | ------------------------------ | --------------------------------------------------------------------- | -------------------------------------------------- |
| Transferência direta de linha entre clientes | Histórico deve ser preservado. | Encerra vínculo com cliente A e abre com cliente B na data informada. | "Linha [X] transferida de [A] para [B] em [data]." |

## Conceitos Gerais

| Módulo/Funcionalidade              | Descrição da Regra                                                                                          | "Disponibilidade (V1 - MVP)" | "Disponibilidade (V2 - Expansão)" | Ações Automáticas e Logs                                                                                                 | Impacto Financeiro/Cálculo                                                                                                                                                                                                          |
| ---------------------------------- | ----------------------------------------------------------------------------------------------------------- | ---------------------------- | --------------------------------- | ------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Faturamento (Entrada de Dados)     | Importação de faturas da operadora (TXT/PDF) vinculadas a um Mês de Processamento designado pelo operador.  | Disponível                   | Disponível                        | Rejeita duplicatas; cria linhas no estoque automaticamente. Log: 'Linha criada em estoque automaticamente...'            | Autoridade primária para cobrança; sem fatura não há cobrança (salvo comando manual).                                                                                                                                               |
| Estoque de Linhas                  | Linhas que constam em fatura mas não estão vinculadas a clientes. Unicidade garantida pelo número da linha. | Disponível                   | Disponível                        | Atualização automática de status (Ativa/Inativa/Estoque) conforme presença na fatura mais recente.                       | Custo base e plano (Ex.: Plano Smart...) extraído da fatura (serviços exceto aparelhos/multas).                                                                                                                                     |
| Linhas em Transição                | Cadastro antecipado de linhas (Portabilidade, TT, PP) antes de aparecerem na fatura.                        | Disponível                   | Disponível                        | Conciliação automática quando o número aparece na fatura. Log: 'Linha X conciliada automaticamente'.                     | Pré-configuração da composição financeira para vigorar na ativação.                                                                                                                                                                 |
| Hierarquia de Cobrança             | Classificação das linhas em Normal, Titular ou Dependente para agrupamento de boletos de clientes.          | Disponível                   | Disponível                        | Reclassificação automática para 'Normal' se vínculo Titular/Dependente quebrar. Log: 'Dependentes reclassificadas...'    | Define o número de boletos: Titular consolida dependentes; Normal gera boleto individual.                                                                                                                                           |
| Proporcionalidade                  | Cálculo pro-rata para serviços ou linhas que iniciam ou terminam no meio do ciclo de 30 dias.               | Disponível                   | Disponível                        | Cálculo automático baseado em: ((Valor / 30) X Dias Ativos)                                                              | Ajusta o valor final da linha proporcionalmente ao tempo de uso no ciclo.                                                                                                                                                           |
| Gestão de Clientes (PJ Revendedor) | Flag 'PJ Revendedor' que habilita funcionalidades de revenda de linhas para usuários finais.                | Não disponível               | Exclusivo V2                      | Ativa o módulo de Processamento Duplo.                                                                                   | Permite definir preços distintos para o cliente PJ e para o usuário final que o respectivo cliente repassa a linha. Gera um Cálculo paralelo de duas perspectivas financeiras para a mesma linha (Luxus Cliente e Cliente Usuário). |
| Excedentes                         | Cobrança de valores que extrapolam o plano contratado (ex: Roaming, chamadas extras).                       | Entrada Manual               | Detecção Automática               | V2: Detecção por termos configuráveis e aplicação de regras de 'Espelhado' ou 'Tabelado'.                                | Adiciona valores variáveis à fatura do cliente conforme uso detectado.                                                                                                                                                              |
| Gestão de Fidelidade               | Controle de vigência contratual, prazos de carência e renovações.                                           | Não disponível               | Exclusivo V2                      | Renovação automática de linha conforme flag; gatilho de renovação em upgrades/aparelhos/renegociações. Log de histórico. | Monitoramento de status de contrato (Ativo/Expirado). Este é da fidelidade contratual, e não se o cliente "ainda é cliente ou não".                                                                                                 |
