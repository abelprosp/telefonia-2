using Goal.Application.Commands;
using Luxus.Connect.Infra.Crosscutting.Errors;
using OneOf;
using OneOf.Types;

namespace Luxus.Connect.Contracts.PhoneLines.Commands;

public sealed record UnassignPhoneLineCustomerCommand(
    string OrganizationId,
    string PhoneLineId,
    DateOnly EndDate
) : ICommand<OneOf<None, AppError>>;
