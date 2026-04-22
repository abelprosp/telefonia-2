using Luxus.Connect.Domain.Customers.Aggregates;

namespace Luxus.Connect.Contracts.Customers.Responses;

public sealed record GetCustomerAddressResponse(
    string Id,
    string CustomerId,
    string Street,
    string Number,
    string Neighborhood,
    string City,
    string State,
    string ZipCode,
    string? Complement,
    string? Country)
{
    public static explicit operator GetCustomerAddressResponse(CustomerAddress entity)
    {
        return new GetCustomerAddressResponse(
            entity.Id,
            entity.CustomerId,
            entity.Street,
            entity.Number,
            entity.Neighborhood,
            entity.City,
            entity.State,
            entity.ZipCode,
            entity.Complement,
            entity.Country);
    }
}
