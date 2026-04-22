using Goal.Application.Commands;
using Luxus.Connect.Infra.Crosscutting.Errors;
using OneOf;
using OneOf.Types;

namespace Luxus.Connect.Contracts.BillingCycles.Commands;

public sealed record UpdateBillingCycleCommand(
    string OrganizationId,
    string Id,
    string ProviderId,
    string Code,
    string Name,
    DateOnly StartDate,
    DateOnly EndDate
) : ICommand<OneOf<None, AppError>>;
