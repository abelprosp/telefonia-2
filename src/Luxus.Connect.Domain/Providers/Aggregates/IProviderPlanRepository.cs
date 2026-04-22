using Goal.Domain.Aggregates;

namespace Luxus.Connect.Domain.Providers.Aggregates;

public interface IProviderPlanRepository : IRepository<ProviderPlan>
{
    Task<ProviderPlan?> GetAsync(string providerId, string planId, CancellationToken cancellationToken);

    Task<ProviderPlan?> GetByProviderAndCode(string providerId, string planCode, CancellationToken cancellationToken);

    Task<ProviderPlan?> GetByProviderAndCodeAsync(string providerId, string code, CancellationToken cancellationToken = default);
}
