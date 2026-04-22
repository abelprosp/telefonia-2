using Goal.Domain.Aggregates;
using Luxus.Connect.Domain.BillingCycles.Aggregates;
using Luxus.Connect.Domain.PhoneLines.Aggregates;

namespace Luxus.Connect.Domain.Providers.Aggregates;

public class ProviderAccount : Entity
{
    protected ProviderAccount()
        : base()
    {
    }

    private ProviderAccount(ContractingCompany contractingCompany, string accountNumber)
        : this()
    {
        ContractingCompany = contractingCompany;
        ContractingCompanyId = contractingCompany.Id;

        AccountNumber = accountNumber;
    }

    public string ContractingCompanyId { get; private set; } = default!;
    public string AccountNumber { get; private set; } = default!;
    public ContractingCompany ContractingCompany { get; private set; } = default!;
    public IEnumerable<BillingCycle> BillingCycles { get; private set; } = Enumerable.Empty<BillingCycle>().ToList();
    public IEnumerable<ProviderInvoice> Invoices { get; private set; } = Enumerable.Empty<ProviderInvoice>().ToList();
    public IEnumerable<PhoneLine> PhoneLines { get; private set; } = Enumerable.Empty<PhoneLine>().ToList();

    public static ProviderAccount Create(ContractingCompany contractingCompany, string accountNumber)
    {
        ArgumentNullException.ThrowIfNull(contractingCompany);
        ArgumentException.ThrowIfNullOrWhiteSpace(accountNumber);

        return new ProviderAccount(contractingCompany, accountNumber);
    }
}
