using Goal.Infra.Crosscutting.Collections;
using Goal.Infra.Crosscutting.Extensions;
using Luxus.Connect.Contracts.PhoneLines.Responses;
using Luxus.Connect.Domain.Customers.Aggregates;
using Luxus.Connect.Domain.PhoneLines.Aggregates;
using Microsoft.EntityFrameworkCore;

namespace Luxus.Connect.Infra.Data.Query.Repositories.PhoneLines;

internal sealed class PhoneLineQueryRepository(AppDbContext context)
    : AppQueryRepository(context), IPhoneLineQueryRepository
{
    public async Task<IEnumerable<PhoneLineCustomerLinkResponse>> ListCustomerLinksAsync(
        string organizationId,
        string phoneLineId,
        CancellationToken cancellationToken = default)
    {
        var links = await context
            .Set<PhoneLineCustomerLink>()
            .AsNoTracking()
            .Where(l =>
                l.PhoneLineId == phoneLineId
                && l.PhoneLine.ProviderAccount.ContractingCompany.Provider.OrganizationId == organizationId)
            .OrderByDescending(l => l.StartDate)
            .Select(l => new
            {
                l.PhoneLineId,
                l.CustomerId,
                CustomerName = l.Customer.Name,
                CustomerDocument = l.Customer.Documents
                    .Where(d =>
                        d.DocumentType == CustomerDocumentType.CPF
                        || d.DocumentType == CustomerDocumentType.CNPJ)
                    .Select(d => d.Number)
                    .FirstOrDefault(),
                l.StartDate,
                l.EndDate
            })
            .ToListAsync(cancellationToken);

        return links.Select(l => new PhoneLineCustomerLinkResponse(
            l.PhoneLineId,
            l.CustomerId,
            l.CustomerName,
            l.CustomerDocument,
            l.StartDate,
            l.EndDate,
            l.EndDate is null));
    }

    public async Task<GetPhoneLineResponse?> LoadAsync(string organizationId, string id, CancellationToken cancellationToken = default)
    {
        var entity = await context
            .Set<PhoneLine>()
            .AsNoTracking()
            .Select(l => new
            {
                l.Id,
                l.ProviderPlanId,
                ProviderPlanName = l.ProviderPlan.Name,
                l.ProviderAccountId,
                ProviderAccountNumber = l.ProviderAccount.AccountNumber,
                l.CostCenterId,
                CostCenterName = l.CostCenter != null ? l.CostCenter.Name : null,
                l.LastInvoiceId,
                LastInvoiceNumber = l.LastInvoice != null ? l.LastInvoice.Number : null,
                l.TitularLineId,
                TitularLineNumber = l.TitularLine != null ? l.TitularLine.Number : null,
                l.Number,
                l.LineClassification,
                l.Status,
                l.TransitionSubStatus,
                l.TransitionStartedAt,
                l.ActivationDate,
                l.CancellationDate,
                l.BaseCost,
                l.CostWithConsumption,
                l.ProviderAccount.ContractingCompany.Provider.OrganizationId,
                Children = l.ChildrenLines.Select(c => new
                {
                    c.Id,
                    c.ProviderPlanId,
                    c.ProviderAccountId,
                    c.CostCenterId,
                    c.LastInvoiceId,
                    c.TitularLineId,
                    c.Number,
                    c.LineClassification,
                    c.Status,
                    c.TransitionSubStatus,
                    c.TransitionStartedAt,
                    c.ActivationDate,
                    c.CancellationDate,
                    Plan = new
                    {
                        c.ProviderPlan.Id,
                        c.ProviderPlan.ProviderId,
                        c.ProviderPlan.Name,
                        c.ProviderPlan.Code
                    },
                    Services = c.PhoneLineServices.Select(s => new
                    {
                        s.Id,
                        s.PhoneLineId,
                        s.ProviderPlanServiceId,
                        s.Name,
                        s.Code,
                        s.Recurring,
                        s.Price,
                        s.Active
                    })
                }),
                Services = l.PhoneLineServices.Select(s => new
                {
                    s.Id,
                    s.PhoneLineId,
                    s.ProviderPlanServiceId,
                    s.Name,
                    s.Code,
                    s.Recurring,
                    s.Price,
                    s.Active
                })
            })
            .SingleOrDefaultAsync(
                l => l.OrganizationId == organizationId && l.Id == id,
                cancellationToken);

        return entity is null
            ? null
            : new GetPhoneLineResponse(
                entity.Id,
                entity.ProviderPlanId,
                entity.ProviderPlanName,
                entity.ProviderAccountId,
                entity.ProviderAccountNumber,
                entity.CostCenterId,
                entity.CostCenterName,
                entity.LastInvoiceId,
                entity.LastInvoiceNumber,
                entity.TitularLineId,
                entity.TitularLineNumber,
                entity.Number,
                entity.LineClassification.GetDescription(),
                entity.Status.GetDescription(),
                entity.TransitionSubStatus?.GetDescription(),
                entity.TransitionStartedAt,
                entity.ActivationDate,
                entity.CancellationDate,
                entity.BaseCost,
                entity.CostWithConsumption,
                [.. entity.Children.Select(c => new GetChildPhoneLineResponse(
                    c.Id,
                    c.ProviderPlanId,
                    c.ProviderAccountId,
                    c.CostCenterId,
                    c.LastInvoiceId,
                    c.TitularLineId,
                    c.Number,
                    c.LineClassification.GetDescription(),
                    c.Status.GetDescription(),
                    c.TransitionSubStatus?.GetDescription(),
                    c.TransitionStartedAt,
                    c.ActivationDate,
                    c.CancellationDate,
                    new GetProviderPlanResponse(c.Plan.Id, c.Plan.ProviderId, c.Plan.Name, c.Plan.Code),
                    c.Services.Select(s => new GetPhoneLineServiceResponse(
                        s.Id, s.PhoneLineId, s.ProviderPlanServiceId, s.Name, s.Code, s.Recurring, s.Price, s.Active))
                ))],
                [.. entity.Services.Select(s => new GetPhoneLineServiceResponse(
                    s.Id, s.PhoneLineId, s.ProviderPlanServiceId, s.Name, s.Code, s.Recurring, s.Price, s.Active))]
            );
    }

    public async Task<IPagedList<ListPhoneLineResponse>> QueryAsync(string organizationId, PageSearch pageSearch, CancellationToken cancellationToken = default)
    {
        IQueryable<PhoneLine> query = context
            .Set<PhoneLine>()
            .AsNoTracking()
            .Where(l => l.ProviderAccount.ContractingCompany.Provider.OrganizationId == organizationId);

        return await ProjectToPagedListAsync(query, pageSearch, cancellationToken);
    }

    public async Task<IPagedList<ListPhoneLineResponse>> QueryByStatusAsync(string organizationId, string? status, PageSearch pageSearch, CancellationToken cancellationToken = default)
    {
        IQueryable<PhoneLine> query = context
            .Set<PhoneLine>()
            .AsNoTracking()
            .Where(l => l.ProviderAccount.ContractingCompany.Provider.OrganizationId == organizationId);

        if (!string.IsNullOrWhiteSpace(status)
            && Enum.TryParse(status, ignoreCase: true, out PhoneLineStatus statusEnum))
        {
            query = query.Where(l => l.Status == statusEnum);
        }

        return await ProjectToPagedListAsync(query, pageSearch, cancellationToken);
    }

    private static async Task<IPagedList<ListPhoneLineResponse>> ProjectToPagedListAsync(
        IQueryable<PhoneLine> query,
        PageSearch pageSearch,
        CancellationToken cancellationToken)
    {
        int totalCount = await query.CountAsync(cancellationToken);

        var items = await query
            .Select(l => new
            {
                l.Id,
                l.ProviderPlanId,
                ProviderPlanName = l.ProviderPlan.Name,
                l.ProviderAccountId,
                ProviderAccountNumber = l.ProviderAccount.AccountNumber,
                l.CostCenterId,
                CostCenterName = l.CostCenter != null ? l.CostCenter.Name : null,
                l.LastInvoiceId,
                LastInvoiceNumber = l.LastInvoice != null ? l.LastInvoice.Number : null,
                l.TitularLineId,
                TitularLineNumber = l.TitularLine != null ? l.TitularLine.Number : null,
                l.Number,
                l.LineClassification,
                l.Status,
                l.TransitionSubStatus,
                l.TransitionStartedAt,
                l.ActivationDate,
                l.CancellationDate,
                l.BaseCost,
                l.CostWithConsumption
            })
            .Paginate(pageSearch.PageIndex, pageSearch.PageSize)
            .ToListAsync(cancellationToken);

        return new PagedList<ListPhoneLineResponse>(
            items.Select(l => new ListPhoneLineResponse(
                l.Id,
                l.ProviderPlanId,
                l.ProviderPlanName,
                l.ProviderAccountId,
                l.ProviderAccountNumber,
                l.CostCenterId,
                l.CostCenterName,
                l.LastInvoiceId,
                l.LastInvoiceNumber,
                l.TitularLineId,
                l.TitularLineNumber,
                l.Number,
                l.LineClassification.GetDescription(),
                l.Status.GetDescription(),
                l.TransitionSubStatus?.GetDescription(),
                l.TransitionStartedAt,
                l.ActivationDate,
                l.CancellationDate,
                l.BaseCost,
                l.CostWithConsumption)),
            totalCount);
    }
}