using Goal.Domain.Aggregates;
using Luxus.Connect.Domain.Customers.Aggregates;
using Luxus.Connect.Domain.ProcessingMonths.Aggregates;

namespace Luxus.Connect.Domain.Providers.Aggregates;

public class Provider : Entity
{
    protected Provider()
        : base()
    {
    }

    private Provider(string organizationId, string name, string slug)
        : this()
    {
        OrganizationId = organizationId;
        Name = name;
        Slug = slug;
    }

    public string OrganizationId { get; private set; } = default!;
    public string Name { get; private set; } = default!;
    public string Slug { get; private set; } = default!;
    public bool Active { get; private set; } = true;
    public IEnumerable<ProviderPlan> ProviderPlans { get; private set; } = Enumerable.Empty<ProviderPlan>().ToList();
    public IEnumerable<ContractingCompany> ContractingCompanies { get; private set; } = Enumerable.Empty<ContractingCompany>().ToList();
    public IEnumerable<ProviderInvoiceImportRequest> InvoiceImportRequests { get; private set; } = Enumerable.Empty<ProviderInvoiceImportRequest>().ToList();
    public IEnumerable<ProcessingMonth> ProcessingMonths { get; private set; } = Enumerable.Empty<ProcessingMonth>().ToList();
    public IEnumerable<CustomerProviderLink> ProviderLinks { get; private set; } = Enumerable.Empty<CustomerProviderLink>().ToList();

    public void Update(string name, string slug)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(name);
        ArgumentException.ThrowIfNullOrWhiteSpace(slug);

        Name = name;
        Slug = slug;
    }

    public void Inactivate()
        => Active = false;

    public static Provider Create(string organizationId, string name, string slug)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(organizationId);
        ArgumentException.ThrowIfNullOrWhiteSpace(name);
        ArgumentException.ThrowIfNullOrWhiteSpace(slug);

        return new(organizationId, name, slug);
    }
}
