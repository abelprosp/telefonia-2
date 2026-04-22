using Luxus.Connect.Domain.Controllership.Aggregates;

namespace Luxus.Connect.Contracts.Controllership.Responses;

public sealed record ListCostCenterResponse(string Id, string Name, string Description)
{
    public static explicit operator ListCostCenterResponse(CostCenter entity)
    {
        return new ListCostCenterResponse(
            entity.Id,
            entity.Name,
            entity.Description);
    }
}
