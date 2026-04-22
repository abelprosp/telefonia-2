using Goal.Domain.Aggregates;

namespace Luxus.Connect.Domain.PhoneLines.Aggregates;

public interface IPhoneLineRepository : IRepository<PhoneLine>
{
    Task<PhoneLine?> GetByIdAsync(string organizationId, string id, CancellationToken cancellationToken = default);
    Task<PhoneLine?> GetByNumberAsync(string number, CancellationToken cancellationToken = default);
    Task<PhoneLine?> GetByAccountAndNumberAsync(string providerAccountId, string number, CancellationToken cancellationToken = default);
    Task<IEnumerable<PhoneLine>> ListByAccountIdAsync(string providerAccountId, CancellationToken cancellationToken = default);
}
