using Goal.Application.Commands;
using Luxus.Connect.Contracts.Customers.Responses;
using Luxus.Connect.Infra.Crosscutting.Errors;
using OneOf;

namespace Luxus.Connect.Contracts.Customers.Commands;

public sealed record ManuallyReleaseCustomerForProcessingMonthCommand(
    string OrganizationId,
    string CustomerId,
    string ProcessingMonthId,
    string Justification
) : ICommand<OneOf<GetCustomerProcessingMonthBillingReadinessResponse, AppError>>;
