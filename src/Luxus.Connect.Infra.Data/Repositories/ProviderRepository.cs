using Goal.Infra.Data;
using Luxus.Connect.Domain.Providers.Aggregates;
using Microsoft.EntityFrameworkCore;

namespace Luxus.Connect.Infra.Data.Repositories;

internal sealed class ProviderRepository(AppDbContext context)
    : Repository<Provider>(context), IProviderRepository
{
    public async Task<Provider?> GetBySlugAsync(string organizationId, string slug, CancellationToken cancellationToken = default)
    {
        return await Context
            .Set<Provider>()
            .FirstOrDefaultAsync(
                o => o.OrganizationId == organizationId && o.Slug == slug,
                cancellationToken);
    }

    public async Task<Provider?> GetByIdAsync(string organizationId, string id, CancellationToken cancellationToken = default)
    {
        return await Context
            .Set<Provider>()
            .FirstOrDefaultAsync(
                o => o.OrganizationId == organizationId && o.Id == id,
                cancellationToken);
    }
}
