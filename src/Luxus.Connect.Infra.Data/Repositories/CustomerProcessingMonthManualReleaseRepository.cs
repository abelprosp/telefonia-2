using Goal.Infra.Data;
using Luxus.Connect.Domain.Customers.Aggregates;
using Microsoft.EntityFrameworkCore;

namespace Luxus.Connect.Infra.Data.Repositories;

internal sealed class CustomerProcessingMonthManualReleaseRepository(AppDbContext context)
    : Repository<CustomerProcessingMonthManualRelease>(context), ICustomerProcessingMonthManualReleaseRepository
{
    public Task<CustomerProcessingMonthManualRelease?> GetAsync(
        string organizationId,
        string customerId,
        string processingMonthId,
        CancellationToken cancellationToken = default)
    {
        return Context
            .Set<CustomerProcessingMonthManualRelease>()
            .SingleOrDefaultAsync(
                e => e.OrganizationId == organizationId
                    && e.CustomerId == customerId
                    && e.ProcessingMonthId == processingMonthId,
                cancellationToken);
    }
}
