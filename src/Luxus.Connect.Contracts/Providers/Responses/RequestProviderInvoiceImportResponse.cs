using Goal.Infra.Crosscutting.Extensions;
using Luxus.Connect.Domain.Providers.Aggregates;

namespace Luxus.Connect.Contracts.Providers.Responses;

public sealed record RequestProviderInvoiceImportResponse(
    string Id,
    string ProcessingMonthId,
    string Status,
    string? Error,
    DateTimeOffset? CompletedAt)
{
    public static explicit operator RequestProviderInvoiceImportResponse(ProviderInvoiceImportRequest entity)
    {
        return new RequestProviderInvoiceImportResponse(
            entity.Id,
            entity.ProcessingMonthId,
            entity.Status.GetDescription(),
            entity.Error,
            entity.CompletedAt);
    }
}
