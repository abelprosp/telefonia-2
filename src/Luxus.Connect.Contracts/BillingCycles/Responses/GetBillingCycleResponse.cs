using Goal.Infra.Crosscutting.Extensions;
using Luxus.Connect.Domain.BillingCycles.Aggregates;

namespace Luxus.Connect.Contracts.BillingCycles.Responses;

public sealed record GetBillingCycleResponse(
    string Id,
    string ProviderId,
    string Code,
    string Name,
    DateOnly StartDate,
    DateOnly EndDate,
    string Status,
    DateTimeOffset? ClosedAt,
    string? ClosedBy)
{
    public static explicit operator GetBillingCycleResponse(BillingCycle entity)
    {
        return new GetBillingCycleResponse(
            entity.Id,
            entity.ProviderId,
            entity.Code,
            entity.Name,
            entity.StartDate,
            entity.EndDate,
            entity.Status.GetDescription(),
            entity.ClosedAt,
            entity.ClosedBy);
    }
}
