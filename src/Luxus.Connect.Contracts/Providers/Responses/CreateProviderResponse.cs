using Luxus.Connect.Domain.Providers.Aggregates;

namespace Luxus.Connect.Contracts.Providers.Responses;

public sealed record CreateProviderResponse(string Id, string Name, string Slug, bool Active)
{
    public static explicit operator CreateProviderResponse(Provider entity)
    {
        return new CreateProviderResponse(
            entity.Id,
            entity.Name,
            entity.Slug,
            entity.Active);
    }
}
