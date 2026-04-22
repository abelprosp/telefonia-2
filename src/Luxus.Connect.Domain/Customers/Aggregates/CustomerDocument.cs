using Goal.Domain.Aggregates;

namespace Luxus.Connect.Domain.Customers.Aggregates;

public class CustomerDocument : Entity
{
    protected CustomerDocument()
        : base()
    {
    }

    private CustomerDocument(Customer customer, CustomerDocumentType documentType, string number)
        : this()
    {
        Customer = customer;
        CustomerId = customer.Id;

        DocumentType = documentType;
        Number = number;
    }

    public string CustomerId { get; private set; } = default!;
    public CustomerDocumentType DocumentType { get; private set; }
    public string Number { get; private set; } = default!;
    public Customer Customer { get; private set; } = default!;

    public void UpdateNumber(string number)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(number);
        Number = number;
    }

    public static CustomerDocument Create(Customer customer, CustomerDocumentType documentType, string number)
    {
        ArgumentNullException.ThrowIfNull(customer);
        ArgumentException.ThrowIfNullOrWhiteSpace(number);

        return new CustomerDocument(customer, documentType, number);
    }
}
