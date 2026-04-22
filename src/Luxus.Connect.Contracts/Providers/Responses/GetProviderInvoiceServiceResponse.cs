using Goal.Infra.Crosscutting.Extensions;
using Luxus.Connect.Domain.Providers.Aggregates;

namespace Luxus.Connect.Contracts.Providers.Responses;

public sealed record GetProviderInvoiceServiceResponse(
    string Id,
    string InvoiceId,
    string PlanId,
    string PlanName,
    string Description,
    decimal Quantity,
    decimal TotalPrice,
    decimal? QuotaAmount,
    decimal? ConsumedAmount,
    string? Unit)
{
    public static explicit operator GetProviderInvoiceServiceResponse(ProviderInvoiceService entity)
    {
        return new GetProviderInvoiceServiceResponse(
            entity.Id,
            entity.InvoiceId,
            entity.PlanId,
            entity.Plan.Name,
            entity.Description,
            entity.Quantity,
            entity.TotalPrice,
            entity.QuotaAmount,
            entity.ConsumedAmount,
            entity.Unit?.GetDescription());
    }
}
