using Goal.Application.Commands;
using Luxus.Connect.Contracts.BillingCycles.Responses;
using Luxus.Connect.Infra.Crosscutting.Errors;
using OneOf;

namespace Luxus.Connect.Contracts.BillingCycles.Commands;

public sealed record CreateBillingCycleCommand(
    string OrganizationId,
    string ProviderId,
    string Code,
    string Name,
    DateOnly StartDate,
    DateOnly EndDate
) : ICommand<OneOf<CreateBillingCycleResponse, AppError>>;
