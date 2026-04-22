using Goal.Infra.Crosscutting.Extensions;
using Luxus.Connect.Domain.Providers.Aggregates;

namespace Luxus.Connect.Contracts.Providers.Responses;

public sealed record ListProviderInvoiceResponse(
    string Id,
    string ProviderAccountId,
    string ProviderAccountNumber,
    string ContractingCompanyId,
    string ContractingCompanyName,
    string ProviderId,
    string ProviderName,
    string BillingCycleId,
    string BillingCycleName,
    string? ProcessingMonthId,
    string? CostCenterId,
    string? ParentInvoiceId,
    DateOnly IssueDate,
    DateOnly DueDate,
    decimal TotalAmount,
    string Status,
    decimal SubtotalServices,
    decimal SubtotalUsage,
    decimal SubtotalTaxes,
    decimal SubtotalDiscounts,
    decimal SubtotalInstallments)
{
    public static explicit operator ListProviderInvoiceResponse(ProviderInvoice entity)
    {
        return new ListProviderInvoiceResponse(
            entity.Id,
            entity.ProviderAccountId,
            entity.ProviderAccount.AccountNumber,
            entity.ContractingCompanyId,
            entity.ContractingCompany.LegalName,
            entity.ContractingCompany.ProviderId,
            entity.ContractingCompany.Provider.Name,
            entity.BillingCycleId,
            entity.BillingCycle.Name,
            entity.ProcessingMonthId,
            entity.CostCenterId,
            entity.ParentInvoiceId,
            entity.IssueDate,
            entity.DueDate,
            entity.TotalAmount,
            entity.Status.GetDescription(),
            entity.SubtotalServices,
            entity.SubtotalUsage,
            entity.SubtotalTaxes,
            entity.SubtotalDiscounts,
            entity.SubtotalInstallments);
    }
}
