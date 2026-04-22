using Goal.Infra.Data;
using Luxus.Connect.Domain.Customers.Aggregates;
using Microsoft.EntityFrameworkCore;

namespace Luxus.Connect.Infra.Data.Repositories;

internal sealed class CustomerAttachmentRepository(AppDbContext context)
    : Repository<CustomerAttachment>(context), ICustomerAttachmentRepository
{
    public async Task<CustomerAttachment?> GetByIdAsync(
        string organizationId,
        string customerId,
        string attachmentId,
        CancellationToken cancellationToken = default)
    {
        return await Context
            .Set<CustomerAttachment>()
            .SingleOrDefaultAsync(
                a =>
                    a.Id == attachmentId
                    && a.CustomerId == customerId
                    && a.OrganizationId == organizationId,
                cancellationToken);
    }
}
