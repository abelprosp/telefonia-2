using Goal.Infra.Crosscutting.Collections;
using Goal.Infra.Data.Query;
using Luxus.Connect.Contracts.BillingCycles.Responses;

namespace Luxus.Connect.Infra.Data.Query.Repositories.BillingCycles;

public interface IBillingCycleQueryRepository : IQueryRepository
{
    Task<GetBillingCycleResponse?> LoadAsync(string organizationId, string id, CancellationToken cancellationToken = default);
    Task<IPagedList<ListBillingCycleResponse>> QueryAsync(string organizationId, PageSearch pageSearch, CancellationToken cancellationToken = default);
}
