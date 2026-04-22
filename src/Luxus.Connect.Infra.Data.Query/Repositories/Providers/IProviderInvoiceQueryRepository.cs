using Goal.Infra.Crosscutting.Collections;
using Goal.Infra.Data.Query;
using Luxus.Connect.Contracts.Providers.Responses;

namespace Luxus.Connect.Infra.Data.Query.Repositories.Providers;

public interface IProviderInvoiceQueryRepository : IQueryRepository
{
    Task<GetProviderInvoiceResponse?> LoadAsync(string organizationId, string id, CancellationToken cancellationToken = default);

    Task<IPagedList<ListProviderInvoiceResponse>> QueryAsync(
        string organizationId,
        PageSearch pageSearch,
        string? processingMonthId = null,
        CancellationToken cancellationToken = default);
}
