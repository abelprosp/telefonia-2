using Goal.Domain.Aggregates;

namespace Luxus.Connect.Domain.Customers.Aggregates;

public interface ICustomerRepository : IRepository<Customer>
{
    Task<IEnumerable<Customer>> ListByDocumentAsync(string organizationId, string document, CancellationToken cancellationToken = default);

    Task<Customer?> GetByIdAsync(string organizationId, string id, CancellationToken cancellationToken = default);

    Task<bool> HasActivePhoneLinesAsync(
        string organizationId,
        string customerId,
        string? excludingPhoneLineId = null,
        CancellationToken cancellationToken = default);
}
