using Luxus.Connect.Domain.Providers.Aggregates;

namespace Luxus.Connect.Contracts.Providers.Responses;

public sealed record GetProviderPlanResponse(
    string Id,
    string Name,
    string Code,
    IEnumerable<GetProviderPlanServiceResponse> Services)
{
    public static explicit operator GetProviderPlanResponse(ProviderPlan entity)
    {
        return new GetProviderPlanResponse(
            entity.Id,
            entity.Name,
            entity.Code,
            [.. entity.ProviderPlanServices.Select(p => (GetProviderPlanServiceResponse)p)]);
    }
}
