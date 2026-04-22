using Goal.Infra.Crosscutting.Collections;
using Goal.Infra.Crosscutting.Extensions;
using Luxus.Connect.Contracts.Controllership.Responses;
using Luxus.Connect.Domain.Controllership.Aggregates;
using Microsoft.EntityFrameworkCore;

namespace Luxus.Connect.Infra.Data.Query.Repositories.Controllership;

internal sealed class CostCenterQueryRepository(AppDbContext context)
    : AppQueryRepository(context), ICostCenterQueryRepository
{
    public async Task<ListCostCenterResponse?> LoadAsync(string organizationId, string id, CancellationToken cancellationToken = default)
    {
        var entity = await context
            .Set<CostCenter>()
            .AsNoTracking()
            .Select(c => new
            {
                c.Id,
                c.Name,
                c.Description,
                c.OrganizationId
            })
            .SingleOrDefaultAsync(c => c.OrganizationId == organizationId && c.Id == id, cancellationToken);

        return entity is null
            ? null
            : new ListCostCenterResponse(
                entity.Id,
                entity.Name,
                entity.Description);
    }

    public async Task<IPagedList<ListCostCenterResponse>> QueryAsync(string organizationId, PageSearch pageSearch, CancellationToken cancellationToken = default)
    {
        IQueryable<CostCenter> query = context
            .Set<CostCenter>()
            .AsNoTracking()
            .Where(c => c.OrganizationId == organizationId);

        int totalCount = await query.CountAsync(cancellationToken);

        var items = await query
            .Select(c => new
            {
                c.Id,
                c.Name,
                c.Description
            })
            .Paginate(pageSearch.PageIndex, pageSearch.PageSize)
            .ToListAsync(cancellationToken);

        return new PagedList<ListCostCenterResponse>(
            items.Select(c => new ListCostCenterResponse(
                c.Id,
                c.Name,
                c.Description)),
            totalCount);
    }
}