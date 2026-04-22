using Goal.Domain.Aggregates;

namespace Luxus.Connect.Domain.Customers.Aggregates;

public interface ICustomerProcessingMonthManualReleaseRepository : IRepository<CustomerProcessingMonthManualRelease>
{
    Task<CustomerProcessingMonthManualRelease?> GetAsync(
        string organizationId,
        string customerId,
        string processingMonthId,
        CancellationToken cancellationToken = default);
}
