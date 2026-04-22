using Goal.Domain.Aggregates;
using Luxus.Connect.Domain.Providers.Enums;

namespace Luxus.Connect.Domain.Providers.Aggregates;

public class ProviderInvoiceService : Entity
{
    protected ProviderInvoiceService()
        : base()
    {
    }

    private ProviderInvoiceService(
        ProviderInvoice invoice,
        ProviderPlan plan,
        string description,
        decimal quantity,
        decimal totalPrice)
        : this()
    {
        Invoice = invoice;
        InvoiceId = invoice.Id;

        Plan = plan;
        PlanId = plan.Id;
        Description = description;
        Quantity = quantity;
        TotalPrice = totalPrice;
    }

    public string InvoiceId { get; private set; } = default!;
    public string PlanId { get; private set; } = default!;
    public string Description { get; private set; } = default!;
    public decimal Quantity { get; private set; }
    public decimal TotalPrice { get; private set; }
    public decimal? QuotaAmount { get; private set; }
    public decimal? ConsumedAmount { get; private set; }
    public InvoiceItemUnit? Unit { get; private set; }
    public ProviderInvoice Invoice { get; private set; } = default!;
    public ProviderPlan Plan { get; private set; } = default!;

    private void SetUnit(InvoiceItemUnit unit)
        => Unit = unit;

    private void SetConsumedAmount(decimal amount)
        => ConsumedAmount = amount;

    private void SetQuotaAmount(decimal amount)
        => QuotaAmount = amount;

    public static ProviderInvoiceService Create(
        ProviderInvoice invoice,
        ProviderPlan plan,
        string description,
        decimal quantity,
        decimal totalPrice,
        decimal? quotaAmount,
        decimal? consumedAmount,
        InvoiceItemUnit? unit)
    {
        ArgumentNullException.ThrowIfNull(invoice);
        ArgumentNullException.ThrowIfNull(plan);

        var service = new ProviderInvoiceService(
            invoice,
            plan,
            description,
            quantity,
            totalPrice);

        if (quotaAmount.HasValue)
        {
            service.SetQuotaAmount(quotaAmount.Value);
        }

        if (consumedAmount.HasValue)
        {
            service.SetConsumedAmount(consumedAmount.Value);
        }

        if (unit.HasValue)
        {
            service.SetUnit(unit.Value);
        }

        return service;
    }
}
