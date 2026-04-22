using Luxus.Connect.Domain.Providers.Aggregates;

namespace Luxus.Connect.Contracts.Providers.Responses;

public sealed record GetProviderPlanServiceResponse(
    string Id,
    string Name,
    bool Active,
    bool Recurring,
    decimal? Price)
{
    public static explicit operator GetProviderPlanServiceResponse(ProviderPlanService entity)
    {
        return new GetProviderPlanServiceResponse(
            entity.Id,
            entity.Name,
            entity.Active,
            entity.Recurring,
            entity.Price);
    }
}
