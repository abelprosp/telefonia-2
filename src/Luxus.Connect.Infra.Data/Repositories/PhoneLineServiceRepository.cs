using Goal.Infra.Data;
using Luxus.Connect.Domain.PhoneLines.Aggregates;
using Microsoft.EntityFrameworkCore;

namespace Luxus.Connect.Infra.Data.Repositories;

internal sealed class PhoneLineServiceRepository(AppDbContext context)
    : Repository<PhoneLineService>(context), IPhoneLineServiceRepository
{
    public Task<PhoneLineService?> GetByLineAndServiceAsync(
        string lineId,
        string serviceId,
        CancellationToken cancellationToken = default)
    {
        return Context
            .Set<PhoneLineService>()
            .FirstOrDefaultAsync(
                ls => ls.PhoneLineId == lineId && ls.ProviderPlanServiceId == serviceId,
                cancellationToken);
    }
}
