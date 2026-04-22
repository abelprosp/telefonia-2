using Goal.Infra.Crosscutting.Collections;
using Goal.Infra.Data.Query;
using Luxus.Connect.Contracts.ProcessingMonths.Responses;

namespace Luxus.Connect.Infra.Data.Query.Repositories.ProcessingMonths;

public interface IProcessingMonthQueryRepository : IQueryRepository
{
    Task<GetProcessingMonthResponse?> LoadAsync(string organizationId, string id, CancellationToken cancellationToken = default);

    Task<IPagedList<ListProcessingMonthResponse>> QueryAsync(
        string organizationId,
        PageSearch pageSearch,
        CancellationToken cancellationToken = default);
}
