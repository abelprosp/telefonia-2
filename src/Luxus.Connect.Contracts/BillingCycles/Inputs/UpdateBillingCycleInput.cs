using Luxus.Connect.Contracts.BillingCycles.Commands;

namespace Luxus.Connect.Contracts.BillingCycles.Inputs;

public sealed record UpdateBillingCycleInput(
    string ProviderId,
    string Code,
    string Name,
    DateOnly StartDate,
    DateOnly EndDate)
{
    public UpdateBillingCycleCommand ToCommand(string organizationId, string id)
        => new(organizationId, id, ProviderId, Code, Name, StartDate, EndDate);
}
