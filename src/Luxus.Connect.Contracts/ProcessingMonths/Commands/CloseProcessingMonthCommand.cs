using Goal.Application.Commands;
using Luxus.Connect.Contracts.ProcessingMonths.Responses;
using Luxus.Connect.Infra.Crosscutting.Errors;
using OneOf;

namespace Luxus.Connect.Contracts.ProcessingMonths.Commands;

public sealed record CloseProcessingMonthCommand(
    string OrganizationId,
    string Id
) : ICommand<OneOf<GetProcessingMonthResponse, AppError>>;
