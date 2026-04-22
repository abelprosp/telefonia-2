using Goal.Domain.Aggregates;
using Luxus.Connect.Infra.Crosscutting.Extensions;

namespace Luxus.Connect.Domain.Providers.Aggregates;

public class ContractingCompany : Entity
{
    protected ContractingCompany()
        : base()
    {
    }

    private ContractingCompany(Provider provider, string legalName, string taxId)
        : this()
    {
        Provider = provider;
        ProviderId = provider.Id;

        LegalName = legalName;
        TaxId = taxId.NormalizeDigitsOnly();
    }

    public string ProviderId { get; private set; } = default!;
    public string LegalName { get; private set; } = default!;
    public string TaxId { get; private set; } = default!;
    public Provider Provider { get; private set; } = default!;
    public IEnumerable<ProviderAccount> ProviderAccounts { get; private set; } = Enumerable.Empty<ProviderAccount>().ToList();

    public static ContractingCompany Create(Provider provider, string legalName, string taxId)
    {
        ArgumentNullException.ThrowIfNull(provider);
        ArgumentException.ThrowIfNullOrWhiteSpace(legalName);

        return new(provider, legalName, taxId);
    }
}
