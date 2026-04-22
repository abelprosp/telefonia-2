using Goal.Domain.Aggregates;
using Luxus.Connect.Domain.PhoneLines.Aggregates;
using Luxus.Connect.Domain.Providers.Aggregates;

namespace Luxus.Connect.Domain.Controllership.Aggregates;

public class CostCenter : Entity
{
    protected CostCenter()
        : base()
    {
    }

    private CostCenter(string origanizationId, string name, string description)
        : this()
    {
        OrganizationId = origanizationId;
        Name = name;
        Description = description;
    }

    public string OrganizationId { get; private set; } = default!;
    public string Name { get; private set; } = default!;
    public string Description { get; private set; } = default!;
    public IEnumerable<PhoneLine> PhoneLines { get; private set; } = Enumerable.Empty<PhoneLine>().ToList();
    public IEnumerable<ProviderInvoice> ProviderInvoices { get; private set; } = Enumerable.Empty<ProviderInvoice>().ToList();

    public static CostCenter Create(string origanizationId, string name, string description)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(origanizationId);
        ArgumentException.ThrowIfNullOrWhiteSpace(name);
        ArgumentException.ThrowIfNullOrWhiteSpace(description);

        return new CostCenter(origanizationId, name, description);
    }
}
