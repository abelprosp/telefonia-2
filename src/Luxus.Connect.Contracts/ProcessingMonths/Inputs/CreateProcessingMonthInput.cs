using Luxus.Connect.Contracts.ProcessingMonths.Commands;

namespace Luxus.Connect.Contracts.ProcessingMonths.Inputs;

public sealed record CreateProcessingMonthInput(
    string ProviderId,
    int Year,
    int Month,
    string DisplayName)
{
    public CreateProcessingMonthCommand ToCommand(string organizationId)
        => new(organizationId, ProviderId, Year, Month, DisplayName);
}
