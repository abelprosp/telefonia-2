using Goal.Domain.Aggregates;

namespace Luxus.Connect.Domain.Customers.Aggregates;

public interface ICustomerAttachmentRepository : IRepository<CustomerAttachment>
{
    Task<CustomerAttachment?> GetByIdAsync(
        string organizationId,
        string customerId,
        string attachmentId,
        CancellationToken cancellationToken = default);
}
