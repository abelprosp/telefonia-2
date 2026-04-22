using Luxus.Connect.Contracts.BillingCycles.Commands;

namespace Luxus.Connect.Contracts.BillingCycles.Inputs;

public sealed record CreateBillingCycleInput(
    string ProviderId,
    string Code,
    string Name,
    DateOnly StartDate,
    DateOnly EndDate)
{
    public CreateBillingCycleCommand ToCommand(string organizationId)
        => new(organizationId, ProviderId, Code, Name, StartDate, EndDate);
}
