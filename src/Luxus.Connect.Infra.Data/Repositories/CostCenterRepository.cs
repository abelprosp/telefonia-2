using Goal.Infra.Data;
using Luxus.Connect.Domain.Controllership.Aggregates;

namespace Luxus.Connect.Infra.Data.Repositories;

internal sealed class CostCenterRepository(AppDbContext context)
    : Repository<CostCenter>(context), ICostCenterRepository
{
}
