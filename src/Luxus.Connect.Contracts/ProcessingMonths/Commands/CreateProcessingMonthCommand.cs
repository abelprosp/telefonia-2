using Goal.Application.Commands;
using Luxus.Connect.Contracts.ProcessingMonths.Responses;
using Luxus.Connect.Infra.Crosscutting.Errors;
using OneOf;

namespace Luxus.Connect.Contracts.ProcessingMonths.Commands;

public sealed record CreateProcessingMonthCommand(
    string OrganizationId,
    string ProviderId,
    int Year,
    int Month,
    string DisplayName
) : ICommand<OneOf<GetProcessingMonthResponse, AppError>>;
