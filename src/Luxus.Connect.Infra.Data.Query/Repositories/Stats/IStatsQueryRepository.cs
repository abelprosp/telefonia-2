using Luxus.Connect.Contracts.Stats.Responses;

namespace Luxus.Connect.Infra.Data.Query.Repositories.Stats;

public interface IStatsQueryRepository
{
    Task<DashboardStatsResponse> GetDashboardStats(CancellationToken cancellationToken);
}
