using Goal.Infra.Crosscutting.Extensions;
using Luxus.Connect.Domain.PhoneLines.Aggregates;

namespace Luxus.Connect.Contracts.PhoneLines.Responses;

public sealed record GetChildPhoneLineResponse(
    string Id,
    string ProviderPlanId,
    string ProviderAccountId,
    string? CostCenterId,
    string? LastInvoiceId,
    string? TitularLineId,
    string Number,
    string LineClassification,
    string Status,
    string? TransitionSubStatus,
    DateTimeOffset? TransitionStartedAt,
    DateOnly? ActivationDate,
    DateOnly? CancellationDate,
    GetProviderPlanResponse Plan,
    IEnumerable<GetPhoneLineServiceResponse> Services)
{
    public static explicit operator GetChildPhoneLineResponse(PhoneLine entity)
    {
        return new GetChildPhoneLineResponse(
            entity.Id,
            entity.ProviderPlanId,
            entity.ProviderAccountId,
            entity.CostCenterId,
            entity.LastInvoiceId,
            entity.TitularLineId,
            entity.Number,
            entity.LineClassification.GetDescription(),
            entity.Status.GetDescription(),
            entity.TransitionSubStatus?.GetDescription(),
            entity.TransitionStartedAt,
            entity.ActivationDate,
            entity.CancellationDate,
            (GetProviderPlanResponse)entity.ProviderPlan,
            entity.PhoneLineServices.Select(s => (GetPhoneLineServiceResponse)s));
    }
}
