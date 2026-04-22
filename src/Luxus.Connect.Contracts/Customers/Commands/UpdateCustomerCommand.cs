using Goal.Application.Commands;
using Luxus.Connect.Infra.Crosscutting.Errors;
using OneOf;
using OneOf.Types;

namespace Luxus.Connect.Contracts.Customers.Commands;

public sealed record UpdateCustomerCommand(
    string Id,
    string Name,
    string? LegalName = null,
    string? StateRegistration = null,
    DateOnly? BirthOrOpeningDate = null,
    string? ResponsibleSalespersonUserId = null
) : ICommand<OneOf<None, AppError>>;
