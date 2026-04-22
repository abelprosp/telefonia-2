using Luxus.Connect.Contracts.ProcessingMonths.Commands;

namespace Luxus.Connect.Contracts.ProcessingMonths.Inputs;

public sealed record CloseProcessingMonthInContingencyInput(string Justification)
{
    public CloseProcessingMonthInContingencyCommand ToCommand(string organizationId, string id)
        => new(organizationId, id, Justification);
}
