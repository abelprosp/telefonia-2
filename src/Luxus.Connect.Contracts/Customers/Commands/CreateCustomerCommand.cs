using Goal.Application.Commands;
using Luxus.Connect.Contracts.Customers.Responses;
using Luxus.Connect.Infra.Crosscutting.Errors;
using OneOf;

namespace Luxus.Connect.Contracts.Customers.Commands;

public sealed record CreateCustomerAddressCommand(
    string Street,
    string Number,
    string Neighborhood,
    string City,
    string State,
    string ZipCode,
    string? Complement = null,
    string Country = "Brasil"
);

public sealed record CreateCustomerCommand : ICommand<OneOf<CreateCustomerResponse, AppError>>
{
    public required string ProviderId { get; init; }
    public required string Type { get; init; }
    public required string Name { get; init; }
    public string? LegalName { get; init; }
    public required string Document { get; init; }
    public string? StateRegistration { get; init; }
    public DateOnly? BirthOrOpeningDate { get; init; }
    public string? ResponsibleSalespersonUserId { get; init; }
    public IList<CreateCustomerAddressCommand> Addresses { get; init; } = [];
}
