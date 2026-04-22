using Goal.Domain.Aggregates;
using Luxus.Connect.Domain.Customers.Aggregates;

namespace Luxus.Connect.Domain.PhoneLines.Aggregates;

public class PhoneLineCustomerLink : Entity
{
    protected PhoneLineCustomerLink()
        : base()
    {
    }

    private PhoneLineCustomerLink(
        PhoneLine phoneLine,
        Customer customer,
        DateOnly startDate)
        : this()
    {
        PhoneLine = phoneLine;
        PhoneLineId = phoneLine.Id;
        Customer = customer;
        CustomerId = customer.Id;
        StartDate = startDate;
    }

    public string PhoneLineId { get; private set; } = default!;
    public string CustomerId { get; private set; } = default!;
    public DateOnly StartDate { get; private set; }
    public DateOnly? EndDate { get; private set; }
    public PhoneLine PhoneLine { get; private set; } = default!;
    public Customer Customer { get; private set; } = default!;
    public bool IsActive => EndDate is null;

    public void Close(DateOnly endDate)
    {
        if (EndDate is not null)
        {
            return;
        }

        if (endDate < StartDate)
        {
            throw new ArgumentException(
                "A data de término do vínculo não pode ser anterior à data de início.",
                nameof(endDate));
        }

        EndDate = endDate;
    }

    public static PhoneLineCustomerLink Create(
        PhoneLine phoneLine,
        Customer customer,
        DateOnly startDate)
    {
        ArgumentNullException.ThrowIfNull(phoneLine);
        ArgumentNullException.ThrowIfNull(customer);

        return new PhoneLineCustomerLink(phoneLine, customer, startDate);
    }
}
