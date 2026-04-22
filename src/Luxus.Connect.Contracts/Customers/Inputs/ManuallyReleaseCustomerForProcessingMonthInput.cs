using Luxus.Connect.Contracts.Customers.Commands;

namespace Luxus.Connect.Contracts.Customers.Inputs;

public sealed record ManuallyReleaseCustomerForProcessingMonthInput(string Justification)
{
    public ManuallyReleaseCustomerForProcessingMonthCommand ToCommand(
        string organizationId,
        string customerId,
        string processingMonthId)
        => new(organizationId, customerId, processingMonthId, Justification);
}
