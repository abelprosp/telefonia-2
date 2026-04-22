using Goal.Domain.Aggregates;
using Luxus.Connect.Domain.ProcessingMonths.Aggregates;

namespace Luxus.Connect.Domain.Customers.Aggregates;

/// <summary>Liberação manual §11.2 — administrador com justificativa (auditoria).</summary>
public class CustomerProcessingMonthManualRelease : Entity
{
    protected CustomerProcessingMonthManualRelease()
        : base()
    {
    }

    private CustomerProcessingMonthManualRelease(
        Customer customer,
        ProcessingMonth processingMonth,
        string justification,
        string releasedByUserId)
        : this()
    {
        Customer = customer;
        CustomerId = customer.Id;

        ProcessingMonth = processingMonth;
        ProcessingMonthId = processingMonth.Id;

        OrganizationId = customer.OrganizationId;

        Justification = justification;
        ReleasedByUserId = releasedByUserId;
        ReleasedAt = DateTimeOffset.UtcNow;
    }

    public string OrganizationId { get; private set; } = default!;
    public string CustomerId { get; private set; } = default!;
    public string ProcessingMonthId { get; private set; } = default!;
    public string Justification { get; private set; } = default!;
    public string ReleasedByUserId { get; private set; } = default!;
    public DateTimeOffset ReleasedAt { get; private set; }

    public Customer Customer { get; private set; } = default!;
    public ProcessingMonth ProcessingMonth { get; private set; } = default!;

    public static CustomerProcessingMonthManualRelease Create(
        Customer customer,
        ProcessingMonth processingMonth,
        string justification,
        string releasedByUserId)
    {
        ArgumentNullException.ThrowIfNull(customer);
        ArgumentNullException.ThrowIfNull(processingMonth);
        ArgumentException.ThrowIfNullOrWhiteSpace(releasedByUserId);

        if (customer.OrganizationId != processingMonth.OrganizationId)
            throw new InvalidOperationException("Cliente e mês de processamento devem pertencer à mesma organização.");

        if (!customer.HasActiveProvider(processingMonth.ProviderId))
            throw new InvalidOperationException("Cliente e mês de processamento devem pertencer à mesma operadora.");

        string trimmed = justification.Trim();
        if (trimmed.Length < 10)
            throw new ArgumentException("Justificativa deve ter pelo menos 10 caracteres.", nameof(justification));

        if (trimmed.Length > 4000)
            throw new ArgumentException("Justificativa não pode exceder 4000 caracteres.", nameof(justification));

        return new CustomerProcessingMonthManualRelease(customer, processingMonth, trimmed, releasedByUserId);
    }
}
