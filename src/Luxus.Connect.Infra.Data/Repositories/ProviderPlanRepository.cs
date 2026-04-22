using Goal.Infra.Data;
using Luxus.Connect.Domain.Providers.Aggregates;
using Microsoft.EntityFrameworkCore;

namespace Luxus.Connect.Infra.Data.Repositories;

internal sealed class ProviderPlanRepository(AppDbContext context)
    : Repository<ProviderPlan>(context), IProviderPlanRepository
{
    public Task<ProviderPlan?> GetAsync(string providerId, string planId, CancellationToken cancellationToken)
    {
        return Context
            .Set<ProviderPlan>()
            .FirstOrDefaultAsync(
                p => p.ProviderId == providerId && p.Id == planId,
                cancellationToken);
    }

    public Task<ProviderPlan?> GetByProviderAndCode(string providerId, string planCode, CancellationToken cancellationToken)
    {
        return Context
            .Set<ProviderPlan>()
            .FirstOrDefaultAsync(
                p => p.ProviderId == providerId && p.Code == planCode,
                cancellationToken);
    }

    public Task<ProviderPlan?> GetByProviderAndCodeAsync(string providerId, string code, CancellationToken cancellationToken = default)
    {
        return Context
            .Set<ProviderPlan>()
            .FirstOrDefaultAsync(
                p => p.ProviderId == providerId && p.Code == code,
                cancellationToken);
    }
}
