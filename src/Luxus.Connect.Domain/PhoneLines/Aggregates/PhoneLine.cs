using Goal.Domain.Aggregates;
using Luxus.Connect.Domain.Customers.Aggregates;
using Luxus.Connect.Domain.Controllership.Aggregates;
using Luxus.Connect.Domain.Providers.Aggregates;
using Luxus.Connect.Domain.Providers.Enums;

namespace Luxus.Connect.Domain.PhoneLines.Aggregates;

public class PhoneLine : Entity
{
    protected PhoneLine()
        : base()
    {
    }

    private PhoneLine(ProviderPlan providerPlan, ProviderAccount providerAccount, string number)
        : this()
    {
        ProviderPlan = providerPlan;
        ProviderPlanId = providerPlan.Id;

        ProviderAccount = providerAccount;
        ProviderAccountId = providerAccount.Id;

        Number = number;
    }

    public string ProviderPlanId { get; private set; } = default!;
    public string ProviderAccountId { get; private set; } = default!;
    public string? CostCenterId { get; private set; }
    public string? LastInvoiceId { get; private set; }
    public string? TitularLineId { get; private set; }
    public string Number { get; private set; } = default!;
    public LineClassification LineClassification { get; private set; } = LineClassification.Normal;
    public PhoneLineStatus Status { get; private set; } = PhoneLineStatus.INACTIVE;
    public TransitionSubStatus? TransitionSubStatus { get; private set; }
    public DateTimeOffset? TransitionStartedAt { get; private set; }
    public DateOnly? ActivationDate { get; private set; }
    public DateOnly? CancellationDate { get; private set; }
    public decimal? BaseCost { get; private set; }
    public decimal? CostWithConsumption { get; private set; }
    public ProviderPlan ProviderPlan { get; private set; } = default!;
    public ProviderAccount ProviderAccount { get; private set; } = default!;
    public PhoneLine? TitularLine { get; private set; }
    public ProviderInvoice? LastInvoice { get; private set; }
    public CostCenter? CostCenter { get; private set; }
    public IEnumerable<PhoneLine> ChildrenLines { get; private set; } = Enumerable.Empty<PhoneLine>().ToList();
    public IEnumerable<PhoneLineCustomerLink> CustomerLinks { get; private set; } = Enumerable.Empty<PhoneLineCustomerLink>().ToList();
    public IEnumerable<ProviderInvoice> ProviderInvoices { get; private set; } = Enumerable.Empty<ProviderInvoice>().ToList();
    public IEnumerable<ProviderInvoiceQuotaSharing> InvoiceQuotaSharings { get; private set; } = Enumerable.Empty<ProviderInvoiceQuotaSharing>().ToList();
    public IEnumerable<PhoneLineService> PhoneLineServices { get; private set; } = Enumerable.Empty<PhoneLineService>().ToList();

    public PhoneLineCustomerLink? ActiveCustomerLink =>
        CustomerLinks.FirstOrDefault(l => l.EndDate is null);

    public void AssignCustomer(Customer customer, DateOnly startDate)
    {
        ArgumentNullException.ThrowIfNull(customer);

        if (ActiveCustomerLink?.CustomerId == customer.Id)
        {
            return;
        }

        ActiveCustomerLink?.Close(startDate);

        CustomerLinks = CustomerLinks
            .Append(PhoneLineCustomerLink.Create(this, customer, startDate))
            .ToList();
    }

    public void UnassignCustomer(DateOnly endDate)
    {
        ActiveCustomerLink?.Close(endDate);
    }

    public void SetTitularLine(PhoneLine? titularLine)
    {
        TitularLine = titularLine;
        TitularLineId = titularLine?.Id;

        SyncClassificationWithHierarchy();
    }

    public void ConfigureOperationalDetails(string? costCenterId, LineClassification lineClassification)
    {
        CostCenterId = costCenterId;
        LineClassification = lineClassification;

        SyncClassificationWithHierarchy();
    }

    public void SetInitialLifecycleState(
        PhoneLineStatus status,
        TransitionSubStatus? transitionSubStatus,
        DateTimeOffset? transitionStartedAt,
        DateOnly? activationDate,
        DateOnly? cancellationDate)
    {
        Status = status;
        TransitionSubStatus = transitionSubStatus;
        TransitionStartedAt = transitionStartedAt;
        ActivationDate = activationDate;
        CancellationDate = cancellationDate;
    }

    public void MarkAsInStock()
    {
        Status = PhoneLineStatus.IN_STOCK;
        TransitionSubStatus = null;
    }

    public void MarkAsAwaitingInvoice()
    {
        Status = PhoneLineStatus.AWAITING_INVOICE;
        TransitionSubStatus = null;
    }

    public void MarkInactiveInStockWhenAbsentFromInvoice()
    {
        Status = PhoneLineStatus.INACTIVE;
        TransitionSubStatus = null;
    }

    public void RecordLastInvoice(ProviderInvoice invoice)
    {
        ArgumentNullException.ThrowIfNull(invoice);

        LastInvoice = invoice;
        LastInvoiceId = invoice.Id;
    }

    public void SetCostSnapshot(decimal? baseCost, decimal? costWithConsumption)
    {
        BaseCost = baseCost;
        CostWithConsumption = costWithConsumption;
    }

    public void ApplyImportedLinePresence(ProviderInvoice invoice, DateOnly? activationDateFromInvoice = null)
    {
        RecordLastInvoice(invoice);

        if (Status == PhoneLineStatus.IN_TRANSITION)
        {
            if (!activationDateFromInvoice.HasValue)
            {
                throw new ArgumentException(
                    "Data de ativação na fatura é obrigatória para linha em transição.",
                    nameof(activationDateFromInvoice));
            }

            ActivationDate = activationDateFromInvoice.Value;
            Status = PhoneLineStatus.ACTIVE;
            TransitionSubStatus = null;
            TransitionStartedAt = null;

            return;
        }

        if (Status == PhoneLineStatus.INACTIVE && string.IsNullOrWhiteSpace(ProviderAccountId))
        {
            MarkAsInStock();
        }

        if (Status is PhoneLineStatus.IN_STOCK or PhoneLineStatus.AWAITING_INVOICE)
        {
            Status = PhoneLineStatus.ACTIVE;
            TransitionSubStatus = null;
        }
    }

    public void SyncClassificationWithHierarchy()
    {
        if (LineClassification == LineClassification.Dependent && string.IsNullOrEmpty(TitularLineId))
            LineClassification = LineClassification.Normal;
    }

    public static PhoneLine Create(ProviderPlan providerPlan, ProviderAccount providerAccount, string number)
    {
        ArgumentNullException.ThrowIfNull(providerPlan);
        ArgumentNullException.ThrowIfNull(providerAccount);
        ArgumentException.ThrowIfNullOrWhiteSpace(number);

        return new PhoneLine(providerPlan, providerAccount, number);
    }
}
