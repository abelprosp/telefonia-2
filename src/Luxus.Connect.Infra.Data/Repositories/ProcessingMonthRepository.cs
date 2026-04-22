using Goal.Infra.Data;
using Luxus.Connect.Domain.ProcessingMonths;
using Luxus.Connect.Domain.ProcessingMonths.Aggregates;
using Microsoft.EntityFrameworkCore;

namespace Luxus.Connect.Infra.Data.Repositories;

internal sealed class ProcessingMonthRepository(AppDbContext context)
    : Repository<ProcessingMonth>(context), IProcessingMonthRepository
{
    public Task<ProcessingMonth?> GetByIdAsync(string organizationId, string id, CancellationToken cancellationToken = default)
    {
        return Context
            .Set<ProcessingMonth>()
            .SingleOrDefaultAsync(
                p => p.OrganizationId == organizationId && p.Id == id,
                cancellationToken);
    }

    public Task<ProcessingMonth?> GetByProviderAndCalendarAsync(
        string providerId,
        int year,
        int month,
        CancellationToken cancellationToken = default)
    {
        return Context
            .Set<ProcessingMonth>()
            .SingleOrDefaultAsync(
                p => p.ProviderId == providerId && p.Year == year && p.Month == month,
                cancellationToken);
    }

    public async Task<bool> ExistsClosedIntersectingDateRangeAsync(
        string organizationId,
        string providerId,
        DateOnly rangeStart,
        DateOnly rangeEnd,
        CancellationToken cancellationToken = default)
    {
        if (rangeEnd < rangeStart)
            return false;

        var closedMonths = await Context
            .Set<ProcessingMonth>()
            .AsNoTracking()
            .Where(p =>
                p.OrganizationId == organizationId
                && p.ProviderId == providerId
                && p.Status == ProcessingMonthStatus.CLOSED)
            .Select(p => new { p.Year, p.Month })
            .ToListAsync(cancellationToken);

        foreach (var m in closedMonths)
        {
            if (ProcessingMonthDateRange.IntersectsCivilMonth(rangeStart, rangeEnd, m.Year, m.Month))
                return true;
        }

        return false;
    }
}
