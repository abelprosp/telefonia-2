using Goal.Domain.Aggregates;

namespace Luxus.Connect.Domain.Providers.Aggregates;

public interface IProviderPlanServiceRepository : IRepository<ProviderPlanService>
{
    Task<ProviderPlanService?> GetByProviderAndNameAsync(string providerPlanId, string name, CancellationToken cancellationToken);

    Task<ProviderPlanService?> GetByPlanAndServiceAsync(
        string providerPlanId,
        string id,
        CancellationToken cancellationToken = default);
}
