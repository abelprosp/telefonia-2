using Goal.Infra.Data;
using Luxus.Connect.Domain.Providers.Aggregates;
using Microsoft.EntityFrameworkCore;

namespace Luxus.Connect.Infra.Data.Repositories;

internal sealed class ProviderPlanServiceRepository(AppDbContext context)
    : Repository<ProviderPlanService>(context), IProviderPlanServiceRepository
{
    public Task<ProviderPlanService?> GetByProviderAndNameAsync(string providerPlanId, string name, CancellationToken cancellationToken)
    {
        return Context
            .Set<ProviderPlanService>()
            .FirstOrDefaultAsync(
                x => x.ProviderPlanId == providerPlanId && x.Name == name,
                cancellationToken);
    }

    public Task<ProviderPlanService?> GetByPlanAndServiceAsync(
        string providerPlanId,
        string id,
        CancellationToken cancellationToken = default)
    {
        return Context
            .Set<ProviderPlanService>()
            .FirstOrDefaultAsync(
                x => x.ProviderPlanId == providerPlanId && x.Id == id,
                cancellationToken);
    }
}
