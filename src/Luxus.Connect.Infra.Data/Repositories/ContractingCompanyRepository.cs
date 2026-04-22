using Goal.Infra.Data;
using Luxus.Connect.Domain.Providers.Aggregates;
using Microsoft.EntityFrameworkCore;

namespace Luxus.Connect.Infra.Data.Repositories;

internal sealed class ContractingCompanyRepository(AppDbContext context)
    : Repository<ContractingCompany>(context), IContractingCompanyRepository
{
    public Task<ContractingCompany?> GetByProviderAndTaxIdAsync(
        string providerId,
        string taxIdDigits,
        CancellationToken cancellationToken = default)
    {
        string digits = new(taxIdDigits.Where(char.IsDigit).ToArray());
        if (digits.Length != 14)
            return Task.FromResult<ContractingCompany?>(null);

        return Context
            .Set<ContractingCompany>()
            .FirstOrDefaultAsync(
                c => c.ProviderId == providerId && c.TaxId == digits,
                cancellationToken);
    }
}
