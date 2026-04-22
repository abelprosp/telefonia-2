using Goal.Domain.Aggregates;

namespace Luxus.Connect.Domain.BillingCycles.Aggregates;

public interface IBillingCycleRepository : IRepository<BillingCycle>
{
    Task<BillingCycle?> GetByCodeAsync(string providerId, string code, CancellationToken cancellationToken = default);
    Task<BillingCycle?> GetByIdAsync(string organizationId, string id, CancellationToken cancellationToken = default);
}
