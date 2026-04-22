using Luxus.Connect.Domain.Providers.Aggregates;

namespace Luxus.Connect.Contracts.Providers.Responses;

public sealed record ListProvidersResponse(string Id, string Name, string Slug, bool Active)
{
    public static explicit operator ListProvidersResponse(Provider entity)
    {
        return new ListProvidersResponse(
            entity.Id,
            entity.Name,
            entity.Slug,
            entity.Active);
    }
}
