using Goal.Infra.Crosscutting.Extensions;
using Luxus.Connect.Domain.ProcessingMonths.Aggregates;

namespace Luxus.Connect.Contracts.ProcessingMonths.Responses;

public sealed record GetProcessingMonthResponse(
    string Id,
    string ProviderId,
    int Year,
    int Month,
    string DisplayName,
    string Status,
    DateTimeOffset? ClosedAt,
    string? ClosedBy,
    bool ClosedInContingency,
    string? ContingencyJustification)
{
    public static explicit operator GetProcessingMonthResponse(ProcessingMonth entity)
    {
        return new GetProcessingMonthResponse(
            entity.Id,
            entity.ProviderId,
            entity.Year,
            entity.Month,
            entity.DisplayName,
            entity.Status.GetDescription(),
            entity.ClosedAt,
            entity.ClosedBy,
            entity.ClosedInContingency,
            entity.ContingencyJustification);
    }
}
