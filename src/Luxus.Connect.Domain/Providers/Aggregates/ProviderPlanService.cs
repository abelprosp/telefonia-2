using Goal.Domain.Aggregates;

namespace Luxus.Connect.Domain.Providers.Aggregates;

public class ProviderPlanService : Entity
{
    protected ProviderPlanService()
        : base()
    {
    }

    private ProviderPlanService(ProviderPlan providerPlan, string name, bool recurring)
        : this()
    {
        ArgumentNullException.ThrowIfNull(providerPlan);
        ArgumentException.ThrowIfNullOrWhiteSpace(name);

        ProviderPlanId = providerPlan.Id;
        Name = name;
        Recurring = recurring;
        Active = true;
    }

    public string ProviderPlanId { get; private set; } = default!;
    public string Name { get; private set; } = default!;
    public bool Active { get; private set; }
    public bool Recurring { get; private set; }
    public decimal? Price { get; private set; }
    public ProviderPlan ProviderPlan { get; private set; } = default!;

    public void Update(string name, bool recurring, decimal? price)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(name);

        Name = name;
        Recurring = recurring;
        Price = price;
    }

    public void Inactivate()
        => Active = false;

    public void SetPrice(decimal price)
        => Price = price;

    public static ProviderPlanService Create(ProviderPlan providerPlan, string name, bool recurring)
        => new(providerPlan, name, recurring);
}
