using Luxus.Connect.Domain.Providers.Aggregates;

namespace Luxus.Connect.Contracts.Providers.Responses;

public sealed record GetProviderInvoiceQuotaSharingResponse(
    string Id,
    string InvoiceId,
    string PhoneLineId,
    string Description,
    decimal? ConsumedAmount)
{
    public static explicit operator GetProviderInvoiceQuotaSharingResponse(ProviderInvoiceQuotaSharing entity)
    {
        return new GetProviderInvoiceQuotaSharingResponse(
            entity.Id,
            entity.InvoiceId,
            entity.PhoneLineId,
            entity.Description,
            entity.ConsumedAmount);
    }
}