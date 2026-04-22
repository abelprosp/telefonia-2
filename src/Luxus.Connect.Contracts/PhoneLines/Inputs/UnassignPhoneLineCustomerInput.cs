using Luxus.Connect.Contracts.PhoneLines.Commands;

namespace Luxus.Connect.Contracts.PhoneLines.Inputs;

public sealed record UnassignPhoneLineCustomerInput(DateOnly? EndDate)
{
    public UnassignPhoneLineCustomerCommand ToCommand(string organizationId, string phoneLineId)
        => new(organizationId, phoneLineId, EndDate ?? DateOnly.FromDateTime(DateTime.UtcNow));
}
