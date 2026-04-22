using Goal.Infra.Crosscutting.Extensions;
using Luxus.Connect.Domain.Providers.Aggregates;

namespace Luxus.Connect.Contracts.Providers.Responses;

public sealed record GetProviderInvoiceResponse(
    string Id,
    string Number,
    string ProviderAccountId,
    string ProviderAccountNumber,
    string ContractingCompanyId,
    string ContractingCompanyName,
    string ProviderId,
    string ProviderName,
    string BillingCycleId,
    string BillingCycleName,
    string? ProcessingMonthId,
    string? ProcessingMonthName,
    string? CostCenterId,
    string? CostCenterName,
    string? ParentInvoiceId,
    DateOnly IssueDate,
    DateOnly DueDate,
    decimal TotalAmount,
    string Status,
    decimal SubtotalServices,
    decimal SubtotalUsage,
    decimal SubtotalTaxes,
    decimal SubtotalDiscounts,
    decimal SubtotalInstallments,
    IEnumerable<GetProviderPhoneLineResponse> PhoneLines,
    IEnumerable<GetProviderInvoiceItemResponse> ProviderInvoiceItems,
    IEnumerable<GetProviderInvoiceServiceResponse> ProviderInvoiceServices,
    IEnumerable<GetProviderInvoiceQuotaSharingResponse> ProviderInvoiceQuotaSharing)
{
    public static explicit operator GetProviderInvoiceResponse(ProviderInvoice entity)
    {
        return new GetProviderInvoiceResponse(
            entity.Id,
            entity.Number,
            entity.ProviderAccountId,
            entity.ProviderAccount.AccountNumber,
            entity.ContractingCompany.ProviderId,
            entity.ContractingCompany.Provider.Name,
            entity.ContractingCompanyId,
            entity.ContractingCompany.LegalName,
            entity.BillingCycleId,
            entity.BillingCycle.Name,
            entity.ProcessingMonthId,
            entity.ProcessingMonth.DisplayName,
            entity.CostCenterId,
            entity.CostCenter?.Name,
            entity.ParentInvoiceId,
            entity.IssueDate,
            entity.DueDate,
            entity.TotalAmount,
            entity.Status.GetDescription(),
            entity.SubtotalServices,
            entity.SubtotalUsage,
            entity.SubtotalTaxes,
            entity.SubtotalDiscounts,
            entity.SubtotalInstallments,
            [.. entity.PhoneLines.Select(l => (GetProviderPhoneLineResponse)l)],
            [.. entity.ProviderInvoiceItems.Select(i => (GetProviderInvoiceItemResponse)i)],
            [.. entity.ProviderInvoiceServices.Select(s => (GetProviderInvoiceServiceResponse)s)],
            [.. entity.ProviderInvoiceQuotaSharing.Select(q => (GetProviderInvoiceQuotaSharingResponse)q)]);
    }
}
