using Goal.Infra.Crosscutting.Collections;
using Goal.Infra.Data.Query;
using Luxus.Connect.Contracts.Customers.Responses;

namespace Luxus.Connect.Infra.Data.Query.Repositories.Customers;

public interface ICustomerQueryRepository : IQueryRepository
{
    Task<ListCustomerResponse?> LoadAsync(string id, CancellationToken cancellationToken = default);
    Task<IReadOnlyList<CustomerProviderLinkResponse>> ListProviderLinksAsync(
        string organizationId,
        string customerId,
        CancellationToken cancellationToken = default);

    Task<IReadOnlyList<CustomerAttachmentResponse>?> ListAttachmentsAsync(
        string organizationId,
        string customerId,
        CancellationToken cancellationToken = default);
    Task<IPagedList<CustomerPhoneLineLinkResponse>> QueryPhoneLinesAsync(
        string organizationId,
        string customerId,
        PageSearch pageSearch,
        CancellationToken cancellationToken = default);
    Task<IPagedList<ListCustomerResponse>> QueryAsync(
        PageSearch pageSearch,
        string? providerId,
        CancellationToken cancellationToken = default);
}
