using Luxus.Connect.Domain.Providers.Aggregates;

namespace Luxus.Connect.Contracts.PhoneLines.Responses;

public sealed record GetProviderPlanResponse(
    string Id,
    string ProviderId,
    string Name,
    string Code)
{
    public static explicit operator GetProviderPlanResponse(ProviderPlan entity)
    {
        return new GetProviderPlanResponse(
            entity.Id,
            entity.ProviderId,
            entity.Name,
            entity.Code);
    }
}