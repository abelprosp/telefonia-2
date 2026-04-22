using Goal.Infra.Data;
using Luxus.Connect.Domain.PhoneLines.Aggregates;
using Microsoft.EntityFrameworkCore;

namespace Luxus.Connect.Infra.Data.Repositories;

internal sealed class PhoneLineRepository(AppDbContext context)
    : Repository<PhoneLine>(context), IPhoneLineRepository
{
    public async Task<PhoneLine?> GetByIdAsync(string organizationId, string id, CancellationToken cancellationToken = default)
    {
        return await Context
            .Set<PhoneLine>()
            .Include(l => l.ProviderAccount)
            .ThenInclude(a => a.ContractingCompany)
            .ThenInclude(c => c.Provider)
            .Include(l => l.CustomerLinks)
            .ThenInclude(link => link.Customer)
            .SingleOrDefaultAsync(
                l => l.Id == id && l.ProviderAccount.ContractingCompany.Provider.OrganizationId == organizationId,
                cancellationToken);
    }

    public async Task<PhoneLine?> GetByNumberAsync(string number, CancellationToken cancellationToken = default)
    {
        return await Context
            .Set<PhoneLine>()
            .Include(l => l.CustomerLinks)
            .FirstOrDefaultAsync(
                l => l.Number == number,
                cancellationToken);
    }

    public async Task<PhoneLine?> GetByAccountAndNumberAsync(string providerAccountId, string number, CancellationToken cancellationToken = default)
    {
        return await Context
            .Set<PhoneLine>()
            .Include(l => l.CustomerLinks)
            .FirstOrDefaultAsync(
                l => l.ProviderAccountId == providerAccountId && l.Number == number,
                cancellationToken);
    }

    public async Task<IEnumerable<PhoneLine>> ListByAccountIdAsync(string providerAccountId, CancellationToken cancellationToken = default)
    {
        return await Context
            .Set<PhoneLine>()
            .Include(l => l.CustomerLinks)
            .Where(l => l.ProviderAccountId == providerAccountId)
            .ToListAsync(cancellationToken);
    }
}
