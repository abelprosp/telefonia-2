using Goal.Infra.Crosscutting.Collections;
using Goal.Infra.Crosscutting.Extensions;
using Luxus.Connect.Contracts.ProcessingMonths.Responses;
using Luxus.Connect.Domain.ProcessingMonths.Aggregates;
using Microsoft.EntityFrameworkCore;

namespace Luxus.Connect.Infra.Data.Query.Repositories.ProcessingMonths;

internal sealed class ProcessingMonthQueryRepository(AppDbContext context)
    : AppQueryRepository(context), IProcessingMonthQueryRepository
{
    public async Task<GetProcessingMonthResponse?> LoadAsync(string organizationId, string id, CancellationToken cancellationToken = default)
    {
        ProcessingMonth? entity = await context
            .Set<ProcessingMonth>()
            .AsNoTracking()
            .SingleOrDefaultAsync(m => m.OrganizationId == organizationId && m.Id == id, cancellationToken);

        return entity is null
            ? null
            : (GetProcessingMonthResponse)entity;
    }

    public async Task<IPagedList<ListProcessingMonthResponse>> QueryAsync(
        string organizationId,
        PageSearch pageSearch,
        CancellationToken cancellationToken = default)
    {
        IQueryable<ProcessingMonth> query = context
            .Set<ProcessingMonth>()
            .AsNoTracking()
            .Where(m => m.OrganizationId == organizationId)
            .OrderByDescending(m => m.Year)
            .ThenByDescending(m => m.Month);

        int totalCount = await query.CountAsync(cancellationToken);

        List<ProcessingMonth> items = await query
            .Paginate(pageSearch.PageIndex, pageSearch.PageSize)
            .ToListAsync(cancellationToken);

        return new PagedList<ListProcessingMonthResponse>(
            items.Select(m => (ListProcessingMonthResponse)m),
            totalCount);
    }
}
