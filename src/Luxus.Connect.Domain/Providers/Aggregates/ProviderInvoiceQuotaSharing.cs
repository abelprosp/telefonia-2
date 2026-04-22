using Goal.Domain.Aggregates;
using Luxus.Connect.Domain.PhoneLines.Aggregates;

namespace Luxus.Connect.Domain.Providers.Aggregates;

public class ProviderInvoiceQuotaSharing : Entity
{
    protected ProviderInvoiceQuotaSharing()
        : base()
    {
    }

    private ProviderInvoiceQuotaSharing(
        ProviderInvoice invoice,
        PhoneLine line,
        string description)
        : this()
    {
        Invoice = invoice;
        InvoiceId = invoice.Id;

        PhoneLine = line;
        PhoneLineId = line.Id;

        Description = description;
    }

    public string InvoiceId { get; private set; } = default!;
    public string PhoneLineId { get; private set; } = default!;
    public string Description { get; private set; } = default!;
    public decimal? ConsumedAmount { get; private set; }
    public ProviderInvoice Invoice { get; private set; } = default!;
    public PhoneLine PhoneLine { get; private set; } = default!;

    private void SetConsumedAmount(decimal value)
        => ConsumedAmount = value;

    public static ProviderInvoiceQuotaSharing Create(
        ProviderInvoice invoice,
        PhoneLine line,
        string description,
        decimal? consumedAmount)
    {
        ArgumentNullException.ThrowIfNull(invoice);
        ArgumentNullException.ThrowIfNull(line);
        ArgumentException.ThrowIfNullOrWhiteSpace(description);

        var quotaSharing = new ProviderInvoiceQuotaSharing(invoice, line, description);

        if (consumedAmount.HasValue)
        {
            quotaSharing.SetConsumedAmount(consumedAmount.Value);
        }

        return quotaSharing;
    }
}