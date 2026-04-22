using Luxus.Connect.Contracts.PhoneLines.Commands;

namespace Luxus.Connect.Contracts.PhoneLines.Inputs;

public sealed record TransferPhoneLineCustomerInput(
    string CustomerId,
    DateOnly? TransferDate)
{
    public TransferPhoneLineCustomerCommand ToCommand(string organizationId, string phoneLineId)
        => new(organizationId, phoneLineId, CustomerId, TransferDate ?? DateOnly.FromDateTime(DateTime.UtcNow));
}
