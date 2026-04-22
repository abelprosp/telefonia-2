using Goal.Infra.Data;
using Luxus.Connect.Domain.BillingCycles.Aggregates;
using Microsoft.EntityFrameworkCore;

namespace Luxus.Connect.Infra.Data.Repositories;

internal sealed class BillingCycleRepository(AppDbContext context)
    : Repository<BillingCycle>(context), IBillingCycleRepository
{
    public Task<BillingCycle?> GetByCodeAsync(string providerId, string code, CancellationToken cancellationToken = default)
    {
        return Context
            .Set<BillingCycle>()
            .SingleOrDefaultAsync(
                b => b.ProviderId == providerId
                && b.Code == code, cancellationToken);
    }

    public Task<BillingCycle?> GetByIdAsync(string organizationId, string id, CancellationToken cancellationToken = default)
    {
        return Context
            .Set<BillingCycle>()
            .SingleOrDefaultAsync(
                b => b.OrganizationId == organizationId
                && b.Id == id, cancellationToken);
    }
}
