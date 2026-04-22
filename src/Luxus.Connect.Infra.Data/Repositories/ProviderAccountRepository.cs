using Goal.Infra.Data;
using Luxus.Connect.Domain.Providers.Aggregates;
using Microsoft.EntityFrameworkCore;

namespace Luxus.Connect.Infra.Data.Repositories;

internal sealed class ProviderAccountRepository(AppDbContext context)
    : Repository<ProviderAccount>(context), IProviderAccountRepository
{
    public async Task<ProviderAccount?> GetByContractingCompanyAndAccountNumber(string contractingCompanyId, string accountNumber, CancellationToken cancellationToken)
    {
        return await Context
            .Set<ProviderAccount>()
            .Where(p => p.ContractingCompanyId == contractingCompanyId && p.AccountNumber == accountNumber)
            .SingleOrDefaultAsync();
    }
}