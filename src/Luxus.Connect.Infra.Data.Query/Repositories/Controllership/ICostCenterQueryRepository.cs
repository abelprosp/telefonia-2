using Goal.Infra.Crosscutting.Collections;
using Goal.Infra.Data.Query;
using Luxus.Connect.Contracts.Controllership.Responses;

namespace Luxus.Connect.Infra.Data.Query.Repositories.Controllership;

public interface ICostCenterQueryRepository : IQueryRepository
{
    Task<ListCostCenterResponse?> LoadAsync(string organizationId, string id, CancellationToken cancellationToken = default);
    Task<IPagedList<ListCostCenterResponse>> QueryAsync(string organizationId, PageSearch pageSearch, CancellationToken cancellationToken = default);
}
