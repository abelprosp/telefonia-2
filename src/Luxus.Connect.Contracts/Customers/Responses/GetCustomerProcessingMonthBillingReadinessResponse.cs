namespace Luxus.Connect.Contracts.Customers.Responses;

/// <summary>§11.2 — Pendente vs Liberado para faturamento no mês.</summary>
public sealed record GetCustomerProcessingMonthBillingReadinessResponse(
    string CustomerId,
    string ProcessingMonthId,
    /// <summary>Ex.: "Pendente" ou "Liberado para faturamento" (linguagem da spec).</summary>
    string StatusDisplayName,
    bool IsReleasedForBilling,
    bool IsAutomaticallyComplete,
    bool IsManuallyReleased,
    /// <summary>
    /// Quando true, a completude automática compara contas das empresas contratantes cujo CNPJ = documento CNPJ do cliente.
    /// Clientes apenas CPF não entram nesta regra até existir vínculo linha–cliente (ver backlog).
    /// </summary>
    bool AutomaticEvaluationUsesCnpjContractingCompanies,
    int AccountsExpectedForAutomaticRule,
    int AccountsWithInvoiceInProcessingMonth,
    GetCustomerProcessingMonthBillingReadinessManualReleaseDto? ManualRelease);

public sealed record GetCustomerProcessingMonthBillingReadinessManualReleaseDto(
    string Justification,
    DateTimeOffset ReleasedAt,
    string ReleasedByUserId);
