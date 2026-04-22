using Goal.Infra.Crosscutting.Collections;
using Goal.Infra.Crosscutting.Extensions;
using Luxus.Connect.Contracts.BillingCycles.Responses;
using Luxus.Connect.Domain.BillingCycles.Aggregates;
using Microsoft.EntityFrameworkCore;

namespace Luxus.Connect.Infra.Data.Query.Repositories.BillingCycles;

internal sealed class BillingCycleQueryRepository(AppDbContext context)
    : AppQueryRepository(context), IBillingCycleQueryRepository
{
    public async Task<GetBillingCycleResponse?> LoadAsync(string organizationId, string id, CancellationToken cancellationToken = default)
    {
        var entity = await context
            .Set<BillingCycle>()
            .AsNoTracking()
            .Select(b => new
            {
                b.Id,
                b.ProviderId,
                b.Code,
                b.Name,
                b.StartDate,
                b.EndDate,
                b.Status,
                b.ClosedAt,
                b.ClosedBy,
                b.OrganizationId
            })
            .SingleOrDefaultAsync(b => b.OrganizationId == organizationId && b.Id == id, cancellationToken);

        return entity is null
            ? null
            : new GetBillingCycleResponse(
                entity.Id,
                entity.ProviderId,
                entity.Code,
                entity.Name,
                entity.StartDate,
                entity.EndDate,
                entity.Status.GetDescription(),
                entity.ClosedAt,
                entity.ClosedBy);
    }

    public async Task<IPagedList<ListBillingCycleResponse>> QueryAsync(string organizationId, PageSearch pageSearch, CancellationToken cancellationToken = default)
    {
        IQueryable<BillingCycle> query = context
            .Set<BillingCycle>()
            .AsNoTracking()
            .Where(b => b.OrganizationId == organizationId);

        int totalCount = await query.CountAsync(cancellationToken);

        var items = await query
            .Select(b => new
            {
                b.Id,
                b.ProviderId,
                b.Code,
                b.Name,
                b.StartDate,
                b.EndDate,
                b.Status,
                b.ClosedAt,
                b.ClosedBy
            })
            .Paginate(pageSearch.PageIndex, pageSearch.PageSize)
            .ToListAsync(cancellationToken);

        return new PagedList<ListBillingCycleResponse>(
            items.Select(b => new ListBillingCycleResponse(
                b.Id,
                b.ProviderId,
                b.Code,
                b.Name,
                b.StartDate,
                b.EndDate,
                b.Status.GetDescription(),
                b.ClosedAt,
                b.ClosedBy)),
            totalCount);
    }
}