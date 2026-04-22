using Goal.Domain.Aggregates;

namespace Luxus.Connect.Domain.Providers.Aggregates;

public interface IProviderRepository : IRepository<Provider>
{
    Task<Provider?> GetBySlugAsync(string organizationId, string slug, CancellationToken cancellationToken = default);
    Task<Provider?> GetByIdAsync(string organizationId, string id, CancellationToken cancellationToken = default);
}
