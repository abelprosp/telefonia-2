using Goal.Domain.Aggregates;

namespace Luxus.Connect.Domain.ProcessingMonths.Aggregates;

public interface IProcessingMonthRepository : IRepository<ProcessingMonth>
{
    Task<ProcessingMonth?> GetByIdAsync(string organizationId, string id, CancellationToken cancellationToken = default);

    Task<ProcessingMonth?> GetByProviderAndCalendarAsync(
        string providerId,
        int year,
        int month,
        CancellationToken cancellationToken = default);

    /// <summary>
    /// Indica se existe algum mês de processamento <strong>fechado</strong> cuja competência civil
    /// (ano/mês) intersecta o intervalo <paramref name="rangeStart"/>–<paramref name="rangeEnd"/>.
    /// Usado na trava retroativa §11.3.
    /// </summary>
    Task<bool> ExistsClosedIntersectingDateRangeAsync(
        string organizationId,
        string providerId,
        DateOnly rangeStart,
        DateOnly rangeEnd,
        CancellationToken cancellationToken = default);
}
