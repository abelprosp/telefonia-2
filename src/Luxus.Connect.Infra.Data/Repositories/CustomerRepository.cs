using Goal.Infra.Data;
using Luxus.Connect.Domain.Customers.Aggregates;
using Luxus.Connect.Domain.PhoneLines.Aggregates;
using Luxus.Connect.Infra.Crosscutting.Extensions;
using Microsoft.EntityFrameworkCore;

namespace Luxus.Connect.Infra.Data.Repositories;

internal sealed class CustomerRepository(AppDbContext context)
    : Repository<Customer>(context), ICustomerRepository
{
    public async Task<bool> HasActivePhoneLinesAsync(
        string organizationId,
        string customerId,
        string? excludingPhoneLineId = null,
        CancellationToken cancellationToken = default)
    {
        IQueryable<PhoneLineCustomerLink> query = Context
            .Set<PhoneLineCustomerLink>()
            .Where(l =>
                l.Customer.OrganizationId == organizationId
                && l.CustomerId == customerId
                && l.EndDate == null);

        if (!string.IsNullOrWhiteSpace(excludingPhoneLineId))
        {
            query = query.Where(l => l.PhoneLineId != excludingPhoneLineId);
        }

        return await query.AnyAsync(cancellationToken);
    }

    public Task<Customer?> GetByIdAsync(string organizationId, string id, CancellationToken cancellationToken = default)
    {
        return Context
            .Set<Customer>()
            .Include(c => c.Addresses)
            .Include(c => c.Documents)
            .Include(c => c.ProviderLinks)
            .Where(c => c.OrganizationId == organizationId && c.Id == id)
            .SingleOrDefaultAsync(cancellationToken);
    }

    public async Task<IEnumerable<Customer>> ListByDocumentAsync(string organizationId, string document, CancellationToken cancellationToken = default)
    {
        string normalizedDocument = document.NormalizeDigitsOnly();

        return await Context
            .Set<Customer>()
            .Include(c => c.Addresses)
            .Include(c => c.Documents)
            .Include(c => c.ProviderLinks)
            .Where(c => c.OrganizationId == organizationId
                && c.Documents.Any(d => (d.DocumentType == CustomerDocumentType.CNPJ || d.DocumentType == CustomerDocumentType.CPF)
                && d.Number == normalizedDocument))
            .ToListAsync(cancellationToken);
    }
}
