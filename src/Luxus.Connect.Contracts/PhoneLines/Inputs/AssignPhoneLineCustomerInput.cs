using Luxus.Connect.Contracts.PhoneLines.Commands;

namespace Luxus.Connect.Contracts.PhoneLines.Inputs;

public sealed record AssignPhoneLineCustomerInput(
    string CustomerId,
    DateOnly? StartDate)
{
    public AssignPhoneLineCustomerCommand ToCommand(string organizationId, string phoneLineId)
        => new(organizationId, phoneLineId, CustomerId, StartDate ?? DateOnly.FromDateTime(DateTime.UtcNow));
}
