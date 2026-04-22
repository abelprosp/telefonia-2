using Goal.Domain.Aggregates;
using Luxus.Connect.Domain.Providers.Enums;

namespace Luxus.Connect.Domain.Providers.Aggregates;

public class ProviderInvoiceItem : Entity
{
    protected ProviderInvoiceItem()
        : base()
    {
    }

    private ProviderInvoiceItem(
        ProviderInvoice invoice,
        string description,
        decimal quantity,
        decimal totalPrice,
        ProviderInvoiceItemType itemType)
        : this()
    {
        Invoice = invoice;
        InvoiceId = invoice.Id;

        Description = description;
        Quantity = quantity;
        TotalPrice = totalPrice;
        ItemType = itemType;
    }

    public string InvoiceId { get; private set; } = default!;
    public string? ParentId { get; private set; }
    public string Description { get; private set; } = default!;
    public decimal Quantity { get; private set; }
    public decimal TotalPrice { get; private set; }
    public ProviderInvoiceItemType ItemType { get; private set; }
    public decimal? QuotaAmount { get; private set; }
    public decimal? ConsumedAmount { get; private set; }
    public InvoiceItemUnit? Unit { get; private set; }
    public ProviderInvoice Invoice { get; private set; } = default!;
    public ProviderInvoiceItem? Parent { get; private set; } = default!;
    public IEnumerable<ProviderInvoiceItem> Children { get; private set; } = Enumerable.Empty<ProviderInvoiceItem>().ToList();

    private void SetUnit(InvoiceItemUnit unit)
        => Unit = unit;

    private void SetConsumedAmount(decimal amount)
        => ConsumedAmount = amount;

    private void SetQuotaAmount(decimal amount)
        => QuotaAmount = amount;

    private void SetParent(ProviderInvoiceItem parent)
    {
        Parent = parent;
        ParentId = parent.Id;
    }

    public static ProviderInvoiceItem Create(
        ProviderInvoice invoice,
        string description,
        decimal quantity,
        decimal totalPrice,
        ProviderInvoiceItemType itemType,
        ProviderInvoiceItem? parent,
        decimal? quotaAmount,
        decimal? consumedAmount,
        InvoiceItemUnit? unit)
    {
        ArgumentNullException.ThrowIfNull(invoice);
        ArgumentException.ThrowIfNullOrWhiteSpace(description);

        var item = new ProviderInvoiceItem(
            invoice,
            description,
            quantity,
            totalPrice,
            itemType);

        if (parent is not null)
        {
            item.SetParent(parent);
        }

        if (quotaAmount.HasValue)
        {
            item.SetQuotaAmount(quotaAmount.Value);
        }

        if (consumedAmount.HasValue)
        {
            item.SetConsumedAmount(consumedAmount.Value);
        }

        if (unit.HasValue)
        {
            item.SetUnit(unit.Value);
        }

        return item;
    }
}
