using Goal.Infra.Crosscutting.Extensions;
using Luxus.Connect.Domain.PhoneLines.Aggregates;

namespace Luxus.Connect.Contracts.PhoneLines.Responses;

public sealed record GetPhoneLineResponse(
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
    DateOnly? CancellationDate,
    decimal? BaseCost,
    decimal? CostWithConsumption,
    IEnumerable<GetChildPhoneLineResponse> Children,
    IEnumerable<GetPhoneLineServiceResponse> Services)
{
    public static explicit operator GetPhoneLineResponse(PhoneLine entity)
    {
        return new GetPhoneLineResponse(
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
            entity.CancellationDate,
            entity.BaseCost,
            entity.CostWithConsumption,
            entity.ChildrenLines.Select(s => (GetChildPhoneLineResponse)s),
            entity.PhoneLineServices.Select(s => (GetPhoneLineServiceResponse)s));
    }
}