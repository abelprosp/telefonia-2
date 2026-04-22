using Goal.Domain.Aggregates;
using Luxus.Connect.Domain.Providers.Aggregates;

namespace Luxus.Connect.Domain.Customers.Aggregates;

public class CustomerProviderLink : Entity
{
    protected CustomerProviderLink()
        : base()
    {
    }

    private CustomerProviderLink(Customer customer, Provider provider, DateOnly startDate)
        : this()
    {
        Customer = customer;
        CustomerId = customer.Id;
        Provider = provider;
        ProviderId = provider.Id;
        StartDate = startDate;
    }

    public string CustomerId { get; private set; } = default!;
    public string ProviderId { get; private set; } = default!;
    public DateOnly StartDate { get; private set; }
    public DateOnly? EndDate { get; private set; }

    public Customer Customer { get; private set; } = default!;
    public Provider Provider { get; private set; } = default!;

    public bool IsActive => EndDate is null;

    public static CustomerProviderLink Create(Customer customer, Provider provider, DateOnly startDate)
    {
        ArgumentNullException.ThrowIfNull(customer);
        ArgumentNullException.ThrowIfNull(provider);

        return new(customer, provider, startDate);
    }

    public void Close(DateOnly endDate)
    {
        if (EndDate is not null)
            throw new InvalidOperationException("Vínculo com operadora já encerrado.");

        if (endDate < StartDate)
            throw new InvalidOperationException("Data final não pode ser anterior à data inicial.");

        EndDate = endDate;
    }
}
