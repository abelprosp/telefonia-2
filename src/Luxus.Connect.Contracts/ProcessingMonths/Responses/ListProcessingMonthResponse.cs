using Goal.Infra.Crosscutting.Extensions;
using Luxus.Connect.Domain.ProcessingMonths.Aggregates;

namespace Luxus.Connect.Contracts.ProcessingMonths.Responses;

public sealed record ListProcessingMonthResponse(
    string Id,
    string ProviderId,
    int Year,
    int Month,
    string DisplayName,
    string Status,
    DateTimeOffset? ClosedAt,
    string? ClosedBy,
    bool ClosedInContingency)
{
    public static explicit operator ListProcessingMonthResponse(ProcessingMonth entity)
    {
        return new ListProcessingMonthResponse(
            entity.Id,
            entity.ProviderId,
            entity.Year,
            entity.Month,
            entity.DisplayName,
            entity.Status.GetDescription(),
            entity.ClosedAt,
            entity.ClosedBy,
            entity.ClosedInContingency);
    }
}
