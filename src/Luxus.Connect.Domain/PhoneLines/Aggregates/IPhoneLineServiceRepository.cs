using Goal.Domain.Aggregates;

namespace Luxus.Connect.Domain.PhoneLines.Aggregates;

public interface IPhoneLineServiceRepository : IRepository<PhoneLineService>
{
    Task<PhoneLineService?> GetByLineAndServiceAsync(
        string lineId,
        string serviceId,
        CancellationToken cancellationToken = default);
}
