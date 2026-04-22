using Goal.Infra.Data.Query;
using Luxus.Connect.Contracts.Customers.Responses;

namespace Luxus.Connect.Infra.Data.Query.Repositories.Customers;

public interface ICustomerProcessingMonthBillingReadinessQueryRepository : IQueryRepository
{
    Task<GetCustomerProcessingMonthBillingReadinessResponse?> LoadAsync(
        string organizationId,
        string customerId,
        string processingMonthId,
        CancellationToken cancellationToken = default);
}
