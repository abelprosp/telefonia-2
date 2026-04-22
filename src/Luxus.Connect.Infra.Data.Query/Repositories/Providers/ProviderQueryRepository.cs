using Goal.Infra.Crosscutting.Collections;
using Goal.Infra.Crosscutting.Extensions;
using Luxus.Connect.Contracts.Providers.Responses;
using Luxus.Connect.Domain.Providers.Aggregates;
using Microsoft.EntityFrameworkCore;

namespace Luxus.Connect.Infra.Data.Query.Repositories.Providers;

internal sealed class ProviderQueryRepository(AppDbContext context)
    : AppQueryRepository(context), IProviderQueryRepository
{
    public async Task<GetProviderResponse?> GetWithDetailsAsync(string organizationId, string id, CancellationToken cancellationToken = default)
    {
        var entity = await context
            .Set<Provider>()
            .AsNoTracking()
            .Select(p => new
            {
                p.Id,
                p.OrganizationId,
                p.Name,
                p.Slug,
                p.Active,
                Plans = p.ProviderPlans.Select(pp => new
                {
                    pp.Id,
                    pp.Name,
                    pp.Code,
                    Services = pp.ProviderPlanServices.Select(s => new
                    {
                        s.Id,
                        s.Name,
                        s.Active,
                        s.Recurring,
                        s.Price
                    })
                })
            })
            .SingleOrDefaultAsync(
                p => p.OrganizationId == organizationId && p.Id == id,
                cancellationToken);

        return entity is null
            ? null
            : new GetProviderResponse(
                entity.Id,
                entity.OrganizationId,
                entity.Name,
                entity.Slug,
                entity.Active,
                [.. entity.Plans.Select(p => new GetProviderPlanResponse(
                    p.Id,
                    p.Name,
                    p.Code,
                    [.. p.Services.Select(s => new GetProviderPlanServiceResponse(
                        s.Id,
                        s.Name,
                        s.Active,
                        s.Recurring,
                        s.Price
                    ))]
                ))]
            );
    }

    public async Task<IPagedList<ListProvidersResponse>> QueryAsync(string organizationId, PageSearch pageSearch, CancellationToken cancellationToken = default)
    {
        IQueryable<Provider> query = context
            .Set<Provider>()
            .AsNoTracking()
            .Where(p => p.OrganizationId == organizationId);

        int totalCount = await query.CountAsync(cancellationToken);

        var items = await query
            .Select(p => new
            {
                p.Id,
                p.Name,
                p.Slug,
                p.Active
            })
            .Paginate(pageSearch.PageIndex, pageSearch.PageSize)
            .ToListAsync(cancellationToken);

        return new PagedList<ListProvidersResponse>(
            items.Select(p => new ListProvidersResponse(p.Id, p.Name, p.Slug, p.Active)),
            totalCount);
    }
}