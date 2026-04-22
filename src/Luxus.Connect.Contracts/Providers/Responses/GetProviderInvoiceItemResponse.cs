using Goal.Infra.Crosscutting.Extensions;
using Luxus.Connect.Domain.Providers.Aggregates;

namespace Luxus.Connect.Contracts.Providers.Responses;

public sealed record GetProviderInvoiceItemResponse(
    string Id,
    string InvoiceId,
    string? ParentId,
    string Description,
    decimal Quantity,
    decimal TotalPrice,
    string ItemType,
    decimal? QuotaAmount,
    decimal? ConsumedAmount,
    string? Unit,
    IEnumerable<GetProviderInvoiceItemResponse> Children)
{
    public static explicit operator GetProviderInvoiceItemResponse(ProviderInvoiceItem entity)
    {
        return new GetProviderInvoiceItemResponse(
            entity.Id,
            entity.InvoiceId,
            entity.ParentId,
            entity.Description,
            entity.Quantity,
            entity.TotalPrice,
            entity.ItemType.GetDescription(),
            entity.QuotaAmount,
            entity.ConsumedAmount,
            entity.Unit?.GetDescription(),
            [.. entity.Children.Select(i => (GetProviderInvoiceItemResponse)i)]);
    }
}
