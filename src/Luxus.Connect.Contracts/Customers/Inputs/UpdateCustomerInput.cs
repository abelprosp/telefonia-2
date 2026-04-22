using Luxus.Connect.Contracts.Customers.Commands;

namespace Luxus.Connect.Contracts.Customers.Inputs;

public sealed record UpdateCustomerInput(
    string Name,
    string? LegalName = null,
    string? StateRegistration = null,
    DateOnly? BirthOrOpeningDate = null,
    string? ResponsibleSalespersonUserId = null
)
{
    public UpdateCustomerCommand ToCommand(string id)
        => new(id, Name, LegalName, StateRegistration, BirthOrOpeningDate, ResponsibleSalespersonUserId);
}
