using Goal.Domain.Aggregates;

namespace Luxus.Connect.Domain.Providers.Aggregates;

public interface IContractingCompanyRepository : IRepository<ContractingCompany>
{
    Task<ContractingCompany?> GetByProviderAndTaxIdAsync(
        string providerId,
        string taxId,
        CancellationToken cancellationToken = default);
}
