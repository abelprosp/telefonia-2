using Goal.Domain.Aggregates;

namespace Luxus.Connect.Domain.Providers.Aggregates;

public interface IProviderAccountRepository : IRepository<ProviderAccount>
{
    Task<ProviderAccount?> GetByContractingCompanyAndAccountNumber(string contractingCompanyId, string accountNumber, CancellationToken cancellationToken);
}
