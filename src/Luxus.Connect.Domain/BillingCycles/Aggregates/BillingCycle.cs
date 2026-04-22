using Goal.Domain.Aggregates;
using Luxus.Connect.Domain.Providers.Aggregates;

namespace Luxus.Connect.Domain.BillingCycles.Aggregates;

public class BillingCycle : Entity
{
    protected BillingCycle()
        : base()
    {
    }

    private BillingCycle(Provider provider, string code, string name, DateOnly startDate, DateOnly endDate)
        : this()
    {
        Provider = provider;
        ProviderId = provider.Id;
        OrganizationId = provider.OrganizationId;

        Code = code;
        Name = name;
        StartDate = startDate;
        EndDate = endDate;
    }

    public string OrganizationId { get; private set; } = default!;
    public string ProviderId { get; private set; } = default!;
    public string Code { get; private set; } = default!;
    public string Name { get; private set; } = default!;
    public DateOnly StartDate { get; private set; }
    public DateOnly EndDate { get; private set; }
    public BillingCycleStatus Status { get; private set; } = BillingCycleStatus.OPEN;
    public DateTimeOffset? ClosedAt { get; private set; }
    public string? ClosedBy { get; private set; }
    public Provider Provider { get; private set; } = default!;
    public IEnumerable<ProviderAccount> ProviderAccounts { get; private set; } = Enumerable.Empty<ProviderAccount>().ToList();
    public IEnumerable<ProviderInvoice> ProviderInvoices { get; private set; } = Enumerable.Empty<ProviderInvoice>().ToList();

    public void Update(string code, string name, DateOnly startDate, DateOnly endDate)
    {
        EnsureNotConsolidated();
        ArgumentException.ThrowIfNullOrWhiteSpace(code);
        ArgumentException.ThrowIfNullOrWhiteSpace(name);

        if (endDate < startDate)
            throw new ArgumentException("Data final do ciclo deve ser maior ou igual à inicial.");

        Code = code;
        Name = name;
        StartDate = startDate;
        EndDate = endDate;
    }

    public void Open()
    {
        EnsureNotConsolidated();
        Status = BillingCycleStatus.OPEN;
    }

    public void Consolidate(string closedBy)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(closedBy);

        if (Status == BillingCycleStatus.CLOSED)
            throw new InvalidOperationException("O ciclo já está consolidado.");

        Status = BillingCycleStatus.CLOSED;
        ClosedAt = DateTimeOffset.UtcNow;
        ClosedBy = closedBy;
    }

    private void EnsureNotConsolidated()
    {
        if (Status == BillingCycleStatus.CLOSED)
            throw new InvalidOperationException("Não é possível alterar um ciclo consolidado.");
    }

    public static BillingCycle Create(Provider provider, string code, string name, DateOnly startDate, DateOnly endDate)
    {
        ArgumentNullException.ThrowIfNull(provider);
        ArgumentException.ThrowIfNullOrWhiteSpace(code);
        ArgumentException.ThrowIfNullOrWhiteSpace(name);

        if (endDate < startDate)
            throw new ArgumentException("Data final do ciclo deve ser maior ou igual à inicial.");

        return new(provider, code, name, startDate, endDate);
    }
}
