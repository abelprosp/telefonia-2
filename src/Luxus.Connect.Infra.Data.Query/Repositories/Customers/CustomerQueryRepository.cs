using Goal.Infra.Crosscutting.Collections;
using Goal.Infra.Crosscutting.Extensions;
using Luxus.Connect.Contracts.Customers.Responses;
using Luxus.Connect.Domain.Customers.Aggregates;
using Luxus.Connect.Domain.PhoneLines.Aggregates;
using Microsoft.EntityFrameworkCore;

namespace Luxus.Connect.Infra.Data.Query.Repositories.Customers;

internal sealed class CustomerQueryRepository(AppDbContext context)
    : AppQueryRepository(context), ICustomerQueryRepository
{
    public async Task<IReadOnlyList<CustomerProviderLinkResponse>> ListProviderLinksAsync(
        string organizationId,
        string customerId,
        CancellationToken cancellationToken = default)
    {
        List<CustomerProviderLinkResponse> items = await (
            from link in context.Set<CustomerProviderLink>().AsNoTracking()
            join provider in context.Set<Domain.Providers.Aggregates.Provider>().AsNoTracking()
                on link.ProviderId equals provider.Id
            where link.CustomerId == customerId && link.Customer.OrganizationId == organizationId
            orderby link.StartDate descending
            select new CustomerProviderLinkResponse(
                link.CustomerId,
                link.ProviderId,
                provider.Name,
                link.StartDate,
                link.EndDate,
                link.EndDate == null))
            .ToListAsync(cancellationToken);

        return items;
    }

    public async Task<IReadOnlyList<CustomerAttachmentResponse>?> ListAttachmentsAsync(
        string organizationId,
        string customerId,
        CancellationToken cancellationToken = default)
    {
        bool customerOk = await context
            .Set<Customer>()
            .AsNoTracking()
            .AnyAsync(
                c => c.Id == customerId && c.OrganizationId == organizationId,
                cancellationToken);

        if (!customerOk)
        {
            return null;
        }

        List<CustomerAttachmentResponse> items = await context
            .Set<CustomerAttachment>()
            .AsNoTracking()
            .Where(a => a.CustomerId == customerId && a.OrganizationId == organizationId)
            .OrderBy(a => a.UploadedAtUtc)
            .Select(a => new CustomerAttachmentResponse(
                a.Id,
                a.Title,
                a.OriginalFileName,
                a.StorageBucket,
                a.StorageObjectKey,
                a.ContentType,
                a.SizeBytes,
                a.UploadedAtUtc))
            .ToListAsync(cancellationToken);

        return items;
    }

    public async Task<IPagedList<CustomerPhoneLineLinkResponse>> QueryPhoneLinesAsync(
        string organizationId,
        string customerId,
        PageSearch pageSearch,
        CancellationToken cancellationToken = default)
    {
        IQueryable<PhoneLineCustomerLink> query = context
            .Set<PhoneLineCustomerLink>()
            .AsNoTracking()
            .Where(l =>
                l.CustomerId == customerId
                && l.Customer.OrganizationId == organizationId)
            .OrderByDescending(l => l.StartDate);

        int totalCount = await query.CountAsync(cancellationToken);

        var items = await query
            .Select(l => new
            {
                l.CustomerId,
                l.PhoneLineId,
                PhoneLineNumber = l.PhoneLine.Number,
                PhoneLineStatus = l.PhoneLine.Status,
                l.PhoneLine.LineClassification,
                l.StartDate,
                l.EndDate
            })
            .Paginate(pageSearch.PageIndex, pageSearch.PageSize)
            .ToListAsync(cancellationToken);

        return new PagedList<CustomerPhoneLineLinkResponse>(
            items.Select(i => new CustomerPhoneLineLinkResponse(
                i.CustomerId,
                i.PhoneLineId,
                i.PhoneLineNumber,
                i.PhoneLineStatus.GetDescription(),
                i.LineClassification.GetDescription(),
                i.StartDate,
                i.EndDate,
                i.EndDate is null)),
            totalCount);
    }

    public async Task<ListCustomerResponse?> LoadAsync(string id, CancellationToken cancellationToken = default)
    {
        var entity = await context
            .Set<Customer>()
            .AsNoTracking()
            .Select(c => new
            {
                c.Id,
                c.Active,
                c.Type,
                c.Name,
                c.LegalName,
                c.BirthOrOpeningDate,
                c.ResponsibleSalespersonUserId,
                Documents = c.Documents.Select(d => new
                {
                    d.DocumentType,
                    d.Number
                })
            })
            .SingleOrDefaultAsync(c => c.Id == id, cancellationToken);

        if (entity is null)
            return null;

        var cpfCnpj = entity.Documents.SingleOrDefault(
            d => d.DocumentType is CustomerDocumentType.CPF or CustomerDocumentType.CNPJ);

        var stateRegistration = entity.Documents.FirstOrDefault(
            d => d.DocumentType == CustomerDocumentType.STATE_REGISTRATION);

        return new ListCustomerResponse(
            entity.Id,
            entity.Active,
            entity.Type.GetDescription(),
            entity.Name,
            cpfCnpj?.Number ?? string.Empty,
            stateRegistration?.Number,
            entity.LegalName,
            entity.BirthOrOpeningDate,
            entity.ResponsibleSalespersonUserId);
    }

    public async Task<IPagedList<ListCustomerResponse>> QueryAsync(
        PageSearch pageSearch,
        string? providerId,
        CancellationToken cancellationToken = default)
    {
        IQueryable<Customer> query = context
            .Set<Customer>()
            .AsNoTracking();

        if (!string.IsNullOrWhiteSpace(providerId))
        {
            query = query.Where(c => c.ProviderLinks.Any(l => l.ProviderId == providerId && l.EndDate == null));
        }

        int totalCount = await query.CountAsync(cancellationToken);

        var items = await query
            .Select(c => new
            {
                c.Id,
                c.Active,
                c.Type,
                c.Name,
                c.LegalName,
                c.BirthOrOpeningDate,
                c.ResponsibleSalespersonUserId,
                Documents = c.Documents.Select(d => new
                {
                    d.DocumentType,
                    d.Number
                })
            })
            .Paginate(pageSearch.PageIndex, pageSearch.PageSize)
            .ToListAsync(cancellationToken);

        return new PagedList<ListCustomerResponse>(
            items.Select(c =>
            {
                var cpfCnpj = c.Documents.SingleOrDefault(
                    d => d.DocumentType is CustomerDocumentType.CPF or CustomerDocumentType.CNPJ);

                var stateRegistration = c.Documents.FirstOrDefault(
                    d => d.DocumentType == CustomerDocumentType.STATE_REGISTRATION);

                return new ListCustomerResponse(
                    c.Id,
                    c.Active,
                    c.Type.GetDescription(),
                    c.Name,
                    cpfCnpj?.Number ?? string.Empty,
                    stateRegistration?.Number,
                    c.LegalName,
                    c.BirthOrOpeningDate,
                    c.ResponsibleSalespersonUserId);
            }),
            totalCount);
    }
}