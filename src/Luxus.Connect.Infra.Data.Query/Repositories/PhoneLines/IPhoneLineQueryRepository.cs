using Goal.Infra.Crosscutting.Collections;
using Goal.Infra.Data.Query;
using Luxus.Connect.Contracts.PhoneLines.Responses;

namespace Luxus.Connect.Infra.Data.Query.Repositories.PhoneLines;

public interface IPhoneLineQueryRepository : IQueryRepository
{
    Task<GetPhoneLineResponse?> LoadAsync(string organizationId, string id, CancellationToken cancellationToken = default);
    Task<IEnumerable<PhoneLineCustomerLinkResponse>> ListCustomerLinksAsync(
        string organizationId,
        string phoneLineId,
        CancellationToken cancellationToken = default);
    Task<IPagedList<ListPhoneLineResponse>> QueryAsync(string organizationId, PageSearch pageSearch, CancellationToken cancellationToken = default);
    Task<IPagedList<ListPhoneLineResponse>> QueryByStatusAsync(string organizationId, string? status, PageSearch pageSearch, CancellationToken cancellationToken = default);
}
