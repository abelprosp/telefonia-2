using Goal.Infra.Crosscutting.Collections;
using Goal.Infra.Data.Query;
using Luxus.Connect.Contracts.Providers.Responses;

namespace Luxus.Connect.Infra.Data.Query.Repositories.Providers;

public interface IProviderQueryRepository : IQueryRepository
{
    Task<GetProviderResponse?> GetWithDetailsAsync(string organizationId, string id, CancellationToken cancellationToken = default);
    Task<IPagedList<ListProvidersResponse>> QueryAsync(string organizationId, PageSearch pageSearch, CancellationToken cancellationToken = default);
}
