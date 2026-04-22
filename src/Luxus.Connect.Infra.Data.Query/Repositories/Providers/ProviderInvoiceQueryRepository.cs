using Goal.Infra.Crosscutting.Collections;
using Goal.Infra.Crosscutting.Extensions;
using Luxus.Connect.Contracts.Providers.Responses;
using Luxus.Connect.Domain.Providers.Aggregates;
using Microsoft.EntityFrameworkCore;

namespace Luxus.Connect.Infra.Data.Query.Repositories.Providers;

internal sealed class ProviderInvoiceQueryRepository(AppDbContext context)
    : AppQueryRepository(context), IProviderInvoiceQueryRepository
{
    public async Task<GetProviderInvoiceResponse?> LoadAsync(string organizationId, string id, CancellationToken cancellationToken = default)
    {
        var entity = await context
            .Set<ProviderInvoice>()
            .AsNoTracking()
            .Select(i => new
            {
                i.Id,
                i.Number,
                i.ProviderAccountId,
                ProviderAccountNumber = i.ProviderAccount.AccountNumber,
                ContractingCompanyId = i.ContractingCompany.ProviderId,
                ContractingCompanyName = i.ContractingCompany.LegalName,
                ProviderId = i.ContractingCompanyId,
                ProviderName = i.ContractingCompany.Provider.Name,
                i.ContractingCompany.Provider.OrganizationId,
                i.BillingCycleId,
                BillingCycleName = i.BillingCycle.Name,
                i.ProcessingMonthId,
                ProcessingMonthName = i.ProcessingMonth.DisplayName,
                i.CostCenterId,
                CostCenterName = i.CostCenter != null ? i.CostCenter.Name : null,
                i.ParentInvoiceId,
                i.IssueDate,
                i.DueDate,
                i.TotalAmount,
                i.Status,
                i.SubtotalServices,
                i.SubtotalUsage,
                i.SubtotalTaxes,
                i.SubtotalDiscounts,
                i.SubtotalInstallments,
                PhoneLines = i.PhoneLines.Select(l => new
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
                    l.CancellationDate
                }),
                ProviderInvoiceItems = i.ProviderInvoiceItems
                    .Where(it => it.ParentId == null)
                    .Select(it => new
                    {
                        it.Id,
                        it.InvoiceId,
                        it.ParentId,
                        it.Description,
                        it.Quantity,
                        it.TotalPrice,
                        it.ItemType,
                        it.QuotaAmount,
                        it.ConsumedAmount,
                        it.Unit,
                        Children = it.Children.Select(c => new
                        {
                            c.Id,
                            c.InvoiceId,
                            c.ParentId,
                            c.Description,
                            c.Quantity,
                            c.TotalPrice,
                            c.ItemType,
                            c.QuotaAmount,
                            c.ConsumedAmount,
                            c.Unit
                        })
                    }),
                ProviderInvoiceServices = i.ProviderInvoiceServices.Select(s => new
                {
                    s.Id,
                    s.InvoiceId,
                    s.PlanId,
                    PlanName = s.Plan.Name,
                    s.Description,
                    s.Quantity,
                    s.TotalPrice,
                    s.QuotaAmount,
                    s.ConsumedAmount,
                    s.Unit
                }),
                ProviderInvoiceQuotaSharing = i.ProviderInvoiceQuotaSharing.Select(q => new
                {
                    q.Id,
                    q.InvoiceId,
                    q.PhoneLineId,
                    q.Description,
                    q.ConsumedAmount
                })
            })
            .SingleOrDefaultAsync(
                i => i.OrganizationId == organizationId && i.Id == id,
                cancellationToken);

        if (entity is null)
            return null;

        return new GetProviderInvoiceResponse(
            entity.Id,
            entity.Number!,
            entity.ProviderAccountId,
            entity.ProviderAccountNumber,
            entity.ContractingCompanyId,
            entity.ContractingCompanyName,
            entity.ProviderId,
            entity.ProviderName,
            entity.BillingCycleId,
            entity.BillingCycleName,
            entity.ProcessingMonthId,
            entity.ProcessingMonthName,
            entity.CostCenterId,
            entity.CostCenterName,
            entity.ParentInvoiceId,
            entity.IssueDate,
            entity.DueDate,
            entity.TotalAmount,
            entity.Status.GetDescription(),
            entity.SubtotalServices,
            entity.SubtotalUsage,
            entity.SubtotalTaxes,
            entity.SubtotalDiscounts,
            entity.SubtotalInstallments,
            [.. entity.PhoneLines.Select(l => new GetProviderPhoneLineResponse(
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
                l.CancellationDate))],
            [.. entity.ProviderInvoiceItems.Select(it => new GetProviderInvoiceItemResponse(
                it.Id,
                it.InvoiceId,
                it.ParentId,
                it.Description,
                it.Quantity,
                it.TotalPrice,
                it.ItemType.GetDescription(),
                it.QuotaAmount,
                it.ConsumedAmount,
                it.Unit?.GetDescription(),
                [.. it.Children.Select(c => new GetProviderInvoiceItemResponse(
                    c.Id,
                    c.InvoiceId,
                    c.ParentId,
                    c.Description,
                    c.Quantity,
                    c.TotalPrice,
                    c.ItemType.GetDescription(),
                    c.QuotaAmount,
                    c.ConsumedAmount,
                    c.Unit?.GetDescription(),
                    []))]))],
            [.. entity.ProviderInvoiceServices.Select(s => new GetProviderInvoiceServiceResponse(
                s.Id,
                s.InvoiceId,
                s.PlanId,
                s.PlanName,
                s.Description,
                s.Quantity,
                s.TotalPrice,
                s.QuotaAmount,
                s.ConsumedAmount,
                s.Unit?.GetDescription()))],
            [.. entity.ProviderInvoiceQuotaSharing.Select(q => new GetProviderInvoiceQuotaSharingResponse(
                q.Id,
                q.InvoiceId,
                q.PhoneLineId,
                q.Description,
                q.ConsumedAmount))]);
    }

    public async Task<IPagedList<ListProviderInvoiceResponse>> QueryAsync(
        string organizationId,
        PageSearch pageSearch,
        string? processingMonthId = null,
        CancellationToken cancellationToken = default)
    {
        IQueryable<ProviderInvoice> query = context
            .Set<ProviderInvoice>()
            .AsNoTracking()
            .Where(i => i.ContractingCompany.Provider.OrganizationId == organizationId);

        if (!string.IsNullOrWhiteSpace(processingMonthId))
        {
            query = query.Where(i => i.ProcessingMonthId == processingMonthId);
        }

        query = query.OrderByDescending(i => i.IssueDate);

        int totalCount = await query.CountAsync(cancellationToken);

        var items = await query
            .Select(i => new
            {
                i.Id,
                i.ProviderAccountId,
                ProviderAccountNumber = i.ProviderAccount.AccountNumber,
                ContractingCompanyId = i.ContractingCompany.ProviderId,
                ContractingCompanyName = i.ContractingCompany.LegalName,
                ProviderId = i.ContractingCompanyId,
                ProviderName = i.ContractingCompany.Provider.Name,
                i.BillingCycleId,
                BillingCycleName = i.BillingCycle.Name,
                i.ProcessingMonthId,
                i.CostCenterId,
                i.ParentInvoiceId,
                i.IssueDate,
                i.DueDate,
                i.TotalAmount,
                i.Status,
                i.SubtotalServices,
                i.SubtotalUsage,
                i.SubtotalTaxes,
                i.SubtotalDiscounts,
                i.SubtotalInstallments
            })
            .Paginate(pageSearch.PageIndex, pageSearch.PageSize)
            .ToListAsync(cancellationToken);

        return new PagedList<ListProviderInvoiceResponse>(
            items.Select(i => new ListProviderInvoiceResponse(
                i.Id,
                i.ProviderAccountId,
                i.ProviderAccountNumber,
                i.ContractingCompanyId,
                i.ContractingCompanyName,
                i.ProviderId,
                i.ProviderName,
                i.BillingCycleId,
                i.BillingCycleName,
                i.ProcessingMonthId,
                i.CostCenterId,
                i.ParentInvoiceId,
                i.IssueDate,
                i.DueDate,
                i.TotalAmount,
                i.Status.GetDescription(),
                i.SubtotalServices,
                i.SubtotalUsage,
                i.SubtotalTaxes,
                i.SubtotalDiscounts,
                i.SubtotalInstallments)),
            totalCount);
    }
}