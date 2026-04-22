using Goal.Domain.Aggregates;
using Luxus.Connect.Domain.BillingCycles.Aggregates;
using Luxus.Connect.Domain.Controllership.Aggregates;
using Luxus.Connect.Domain.PhoneLines.Aggregates;
using Luxus.Connect.Domain.ProcessingMonths.Aggregates;
using Luxus.Connect.Domain.Providers.Enums;

namespace Luxus.Connect.Domain.Providers.Aggregates;

public class ProviderInvoice : Entity
{
    protected ProviderInvoice()
        : base()
    {
    }

    private ProviderInvoice(
        ProviderAccount providerAccount,
        ContractingCompany contractingCompany,
        ProcessingMonth processingMonth,
        BillingCycle cycle,
        DateOnly issueDate,
        DateOnly dueDate,
        decimal totalAmount)
        : this()
    {
        ProcessingMonth = processingMonth;
        ProcessingMonthId = processingMonth.Id;

        BillingCycle = cycle;
        BillingCycleId = cycle.Id;

        ProviderAccount = providerAccount;
        ProviderAccountId = providerAccount.Id;

        ContractingCompany = contractingCompany;
        ContractingCompanyId = contractingCompany.Id;

        Number = GetInvoiceNumber();

        IssueDate = issueDate;
        DueDate = dueDate;
        TotalAmount = totalAmount;
    }

    public string ProviderAccountId { get; private set; } = default!;
    public string ContractingCompanyId { get; private set; } = default!;
    public string ProcessingMonthId { get; private set; } = default!;
    public string BillingCycleId { get; private set; } = default!;
    public string? CostCenterId { get; private set; }
    public string? ParentInvoiceId { get; private set; }
    public string Number { get; private set; } = default!;
    public DateOnly IssueDate { get; private set; }
    public DateOnly DueDate { get; private set; }
    public decimal TotalAmount { get; private set; }
    public ProviderInvoiceStatus Status { get; private set; } = ProviderInvoiceStatus.DRAFT;
    public decimal SubtotalServices { get; private set; } = 0;
    public decimal SubtotalUsage { get; private set; } = 0;
    public decimal SubtotalTaxes { get; private set; } = 0;
    public decimal SubtotalDiscounts { get; private set; } = 0;
    public decimal SubtotalInstallments { get; private set; } = 0;
    public ProviderAccount ProviderAccount { get; private set; } = default!;
    public ContractingCompany ContractingCompany { get; private set; } = default!;
    public ProcessingMonth ProcessingMonth { get; private set; } = default!;
    public BillingCycle BillingCycle { get; private set; } = default!;
    public ProviderInvoice? ParentInvoice { get; private set; } = default!;
    public CostCenter? CostCenter { get; private set; }
    public IEnumerable<PhoneLine> PhoneLines { get; private set; } = Enumerable.Empty<PhoneLine>().ToList();
    public IEnumerable<PhoneLine> LastPhoneLines { get; private set; } = Enumerable.Empty<PhoneLine>().ToList();
    public IEnumerable<ProviderInvoiceItem> ProviderInvoiceItems { get; private set; } = Enumerable.Empty<ProviderInvoiceItem>().ToList();
    public IEnumerable<ProviderInvoiceService> ProviderInvoiceServices { get; private set; } = Enumerable.Empty<ProviderInvoiceService>().ToList();
    public IEnumerable<ProviderInvoiceQuotaSharing> ProviderInvoiceQuotaSharing { get; private set; } = Enumerable.Empty<ProviderInvoiceQuotaSharing>().ToList();

    public void SetSubtotals(
        decimal subtotalServices,
        decimal subtotalUsage,
        decimal subtotalTaxes,
        decimal subtotalDiscounts,
        decimal subtotalInstallments)
    {
        SubtotalServices = subtotalServices;
        SubtotalUsage = subtotalUsage;
        SubtotalTaxes = subtotalTaxes;
        SubtotalDiscounts = subtotalDiscounts;
        SubtotalInstallments = subtotalInstallments;
    }

    public void SetStatus(ProviderInvoiceStatus status)
        => Status = status;

    public void LinkPlanLine(PhoneLine line)
    {
        ArgumentNullException.ThrowIfNull(line);

        if (PhoneLines.All(l => l.Id != line.Id))
        {
            PhoneLines = PhoneLines
                .Append(line)
                .ToList();
        }
    }

    public ProviderInvoiceItem AddItem(
        string description,
        decimal quantity,
        decimal totalPrice,
        ProviderInvoiceItemType itemType)
    {
        return AddItem(
            description,
            quantity,
            totalPrice,
            itemType,
            null,
            null,
            null,
            null);
    }

    public ProviderInvoiceItem AddItem(
        string description,
        decimal quantity,
        decimal totalPrice,
        ProviderInvoiceItemType itemType,
        ProviderInvoiceItem? parent)
    {
        return AddItem(
            description,
            quantity,
            totalPrice,
            itemType,
            parent,
            null,
            null,
            null);
    }

    public ProviderInvoiceItem AddItem(
        string description,
        decimal quantity,
        decimal totalPrice,
        ProviderInvoiceItemType itemType,
        ProviderInvoiceItem? parent,
        decimal? quotaAmount,
        decimal? consumedAmount,
        InvoiceItemUnit? unit)
    {
        var item = ProviderInvoiceItem.Create(
            this,
            description,
            quantity,
            totalPrice,
            itemType,
            parent,
            quotaAmount,
            consumedAmount,
            unit);

        ProviderInvoiceItems = ProviderInvoiceItems
            .Append(item)
            .ToList();

        return item;
    }

    public ProviderInvoiceService AddService(
            ProviderPlan plan,
            string description,
            decimal quantity,
            decimal totalPrice,
            decimal? quotaAmount,
            decimal? consumedAmount,
            InvoiceItemUnit? unit)
    {
        var service = ProviderInvoiceService.Create(
            this,
            plan,
            description,
            quantity,
            totalPrice,
            quotaAmount,
            consumedAmount,
            unit);

        ProviderInvoiceServices = ProviderInvoiceServices
            .Append(service)
            .ToList();

        return service;
    }

    public static ProviderInvoice Create(
        ProviderAccount providerAccount,
        ContractingCompany contractingCompany,
        ProcessingMonth processingMonth,
        BillingCycle cycle,
        DateOnly issueDate,
        DateOnly dueDate,
        decimal totalAmount)
    {
        ArgumentNullException.ThrowIfNull(providerAccount);
        ArgumentNullException.ThrowIfNull(contractingCompany);
        ArgumentNullException.ThrowIfNull(processingMonth);
        ArgumentNullException.ThrowIfNull(cycle);

        return new ProviderInvoice(
            providerAccount,
            contractingCompany,
            processingMonth,
            cycle,
            issueDate,
            dueDate,
            totalAmount);
    }

    private string GetInvoiceNumber()
        => $"{ProviderAccount?.AccountNumber}-{DueDate:yyyyMMdd}";
}
