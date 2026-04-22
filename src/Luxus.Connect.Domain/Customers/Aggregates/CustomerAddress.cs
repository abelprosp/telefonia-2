using Goal.Domain.Aggregates;

namespace Luxus.Connect.Domain.Customers.Aggregates;

public class CustomerAddress : Entity
{
    protected CustomerAddress()
        : base()
    {
    }

    private CustomerAddress(
        Customer customer,
        string street,
        string number,
        string neighborhood,
        string city,
        string state,
        string zipCode,
        string? complement,
        string country)
        : this()
    {
        Customer = customer;
        CustomerId = customer.Id;

        Street = street;
        Number = number;
        Neighborhood = neighborhood;
        City = city;
        State = state;
        ZipCode = zipCode;
        Complement = complement;
        Country = country;
    }

    public string CustomerId { get; private set; } = default!;
    public string Street { get; private set; } = default!;
    public string Number { get; private set; } = default!;
    public string Neighborhood { get; private set; } = default!;
    public string? Complement { get; private set; }
    public string City { get; private set; } = default!;
    public string State { get; private set; } = default!;
    public string ZipCode { get; private set; } = default!;
    public string Country { get; private set; } = default!;
    public Customer Customer { get; private set; } = default!;

    public void Replace(
        string street,
        string number,
        string neighborhood,
        string city,
        string state,
        string zipCode,
        string? complement,
        string country)
    {
        Street = street;
        Number = number;
        Neighborhood = neighborhood;
        City = city;
        State = state;
        ZipCode = zipCode;
        Complement = complement;
        Country = country;
    }

    public static CustomerAddress Create(
        Customer customer,
        string street,
        string number,
        string neighborhood,
        string city,
        string state,
        string zipCode,
        string? complement,
        string country)
    {
        ArgumentNullException.ThrowIfNull(customer);
        ArgumentException.ThrowIfNullOrWhiteSpace(street);
        ArgumentException.ThrowIfNullOrWhiteSpace(number);
        ArgumentException.ThrowIfNullOrWhiteSpace(city);
        ArgumentException.ThrowIfNullOrWhiteSpace(state);
        ArgumentException.ThrowIfNullOrWhiteSpace(zipCode);
        ArgumentException.ThrowIfNullOrWhiteSpace(country);

        return new CustomerAddress(
            customer,
            street,
            number,
            neighborhood,
            city,
            state,
            zipCode,
            complement,
            country);
    }
}
