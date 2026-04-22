using Goal.Application.Commands;
using Luxus.Connect.Contracts.PhoneLines.Responses;
using Luxus.Connect.Infra.Crosscutting.Errors;
using OneOf;

namespace Luxus.Connect.Contracts.PhoneLines.Commands;

public sealed record AssignPhoneLineCustomerCommand(
    string OrganizationId,
    string PhoneLineId,
    string CustomerId,
    DateOnly StartDate
) : ICommand<OneOf<PhoneLineCustomerLinkResponse, AppError>>;
