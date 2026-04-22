using Goal.Domain.Aggregates;
using Luxus.Connect.Domain.Customers.Aggregates;
using Luxus.Connect.Domain.Providers.Aggregates;

namespace Luxus.Connect.Domain.ProcessingMonths.Aggregates;

public class ProcessingMonth : Entity
{
    protected ProcessingMonth()
        : base()
    {
    }

    private ProcessingMonth(Provider provider, int year, int month, string displayName)
        : this()
    {
        Provider = provider;
        ProviderId = provider.Id;
        OrganizationId = provider.OrganizationId;

        Year = year;
        Month = month;
        DisplayName = displayName;
        Status = ProcessingMonthStatus.OPEN;
    }

    public string OrganizationId { get; private set; } = default!;
    public string ProviderId { get; private set; } = default!;
    public int Year { get; private set; }
    public int Month { get; private set; }
    public string DisplayName { get; private set; } = default!;
    public ProcessingMonthStatus Status { get; private set; } = ProcessingMonthStatus.OPEN;
    public DateTimeOffset? ClosedAt { get; private set; }
    public string? ClosedBy { get; private set; }
    public bool ClosedInContingency { get; private set; }
    public string? ContingencyJustification { get; private set; }
    public Provider Provider { get; private set; } = default!;
    public IEnumerable<ProviderInvoiceImportRequest> ProviderInvoiceImportRequests { get; private set; } = Enumerable.Empty<ProviderInvoiceImportRequest>().ToList();
    public IEnumerable<ProviderInvoice> ProviderInvoices { get; private set; } = Enumerable.Empty<ProviderInvoice>().ToList();
    public IEnumerable<CustomerProcessingMonthManualRelease> CustomerProcessingMonthManualReleases { get; private set; } = Enumerable.Empty<CustomerProcessingMonthManualRelease>().ToList();

    public static ProcessingMonth Create(Provider provider, int year, int month, string displayName)
    {
        ArgumentNullException.ThrowIfNull(provider);
        ValidateCalendar(year, month);
        ArgumentException.ThrowIfNullOrWhiteSpace(displayName);

        string trimmed = displayName.Trim();
        if (trimmed.Length > 128)
        {
            throw new ArgumentException("Nome de exibição não pode exceder 128 caracteres.", nameof(displayName));
        }

        return new ProcessingMonth(provider, year, month, trimmed);
    }

    public void Close(string closedBy)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(closedBy);

        if (Status == ProcessingMonthStatus.CLOSED)
        {
            throw new InvalidOperationException("O mês de processamento já está fechado.");
        }

        Status = ProcessingMonthStatus.CLOSED;
        ClosedAt = DateTimeOffset.UtcNow;
        ClosedBy = closedBy;
        ClosedInContingency = false;
        ContingencyJustification = null;
    }

    public void CloseInContingency(string closedBy, string justification)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(closedBy);
        ArgumentException.ThrowIfNullOrWhiteSpace(justification);

        if (Status == ProcessingMonthStatus.CLOSED)
        {
            throw new InvalidOperationException("O mês de processamento já está fechado.");
        }

        string trimmedJustification = justification.Trim();
        if (trimmedJustification.Length > 4000)
        {
            throw new ArgumentException("Justificativa não pode exceder 4000 caracteres.", nameof(justification));
        }

        Status = ProcessingMonthStatus.CLOSED;
        ClosedAt = DateTimeOffset.UtcNow;
        ClosedBy = closedBy;
        ClosedInContingency = true;
        ContingencyJustification = trimmedJustification;
    }

    private static void ValidateCalendar(int year, int month)
    {
        if (year is < 2000 or > 2100)
        {
            throw new ArgumentOutOfRangeException(nameof(year), "Ano deve estar entre 2000 e 2100.");
        }

        if (month is < 1 or > 12)
        {
            throw new ArgumentOutOfRangeException(nameof(month), "Mês deve estar entre 1 e 12.");
        }
    }
}
