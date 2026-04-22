using Goal.Infra.Crosscutting.Extensions;
using Luxus.Connect.Domain.PhoneLines.Aggregates;

namespace Luxus.Connect.Contracts.Providers.Responses;

public sealed record GetProviderPhoneLineResponse(
    string Id,
    string ProviderPlanId,
    string ProviderPlanName,
    string ProviderAccountId,
    string ProviderAccountNumber,
    string? CostCenterId,
    string? CostCenterName,
    string? LastInvoiceId,
    string? LastInvoiceNumber,
    string? TitularLineId,
    string? TitularLineNumber,
    string Number,
    string LineClassification,
    string Status,
    string? TransitionSubStatus,
    DateTimeOffset? TransitionStartedAt,
    DateOnly? ActivationDate,
    DateOnly? CancellationDate)
{
    public static explicit operator GetProviderPhoneLineResponse(PhoneLine entity)
    {
        return new GetProviderPhoneLineResponse(
            entity.Id,
            entity.ProviderPlanId,
            entity.ProviderPlan.Name,
            entity.ProviderAccountId,
            entity.ProviderAccount.AccountNumber,
            entity.CostCenterId,
            entity.CostCenter?.Name,
            entity.LastInvoiceId,
            entity.LastInvoice?.Number,
            entity.TitularLineId,
            entity.TitularLine?.Number,
            entity.Number,
            entity.LineClassification.GetDescription(),
            entity.Status.GetDescription(),
            entity.TransitionSubStatus?.GetDescription(),
            entity.TransitionStartedAt,
            entity.ActivationDate,
            entity.CancellationDate);
    }
}
