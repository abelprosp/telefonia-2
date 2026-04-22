using Goal.Application.Commands;
using Luxus.Connect.Contracts.ProcessingMonths.Responses;
using Luxus.Connect.Infra.Crosscutting.Errors;
using OneOf;

namespace Luxus.Connect.Contracts.ProcessingMonths.Commands;

public sealed record CloseProcessingMonthInContingencyCommand(
    string OrganizationId,
    string Id,
    string Justification
) : ICommand<OneOf<GetProcessingMonthResponse, AppError>>;
