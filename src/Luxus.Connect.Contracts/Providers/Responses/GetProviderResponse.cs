using Luxus.Connect.Domain.Providers.Aggregates;

namespace Luxus.Connect.Contracts.Providers.Responses;

public sealed record GetProviderResponse(
    string Id,
    string OrganizationId,
    string Name,
    string Slug,
    bool Active,
    IEnumerable<GetProviderPlanResponse> Plans)
{
    public static explicit operator GetProviderResponse(Provider entity)
    {
        return new GetProviderResponse(
            entity.Id,
            entity.OrganizationId,
            entity.Name,
            entity.Slug,
            entity.Active,
            [.. entity.ProviderPlans.Select(p => (GetProviderPlanResponse)p)]);
    }
}
