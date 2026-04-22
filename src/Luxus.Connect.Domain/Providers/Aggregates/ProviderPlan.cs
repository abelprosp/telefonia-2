using Goal.Domain.Aggregates;
using Luxus.Connect.Domain.PhoneLines.Aggregates;

namespace Luxus.Connect.Domain.Providers.Aggregates;

public class ProviderPlan : Entity
{
    protected ProviderPlan()
        : base()
    {
    }

    private ProviderPlan(Provider provider, string name, string code)
        : this()
    {
        ArgumentNullException.ThrowIfNull(provider);
        ArgumentException.ThrowIfNullOrWhiteSpace(name);
        ArgumentException.ThrowIfNullOrWhiteSpace(code);

        ProviderId = provider.Id;
        Name = name;
        Code = code;
    }

    public string ProviderId { get; private set; } = default!;
    public string Name { get; private set; } = default!;
    public string Code { get; private set; } = default!;
    public Provider Provider { get; private set; } = default!;
    public IEnumerable<PhoneLine> PhoneLines { get; private set; } = Enumerable.Empty<PhoneLine>().ToList();
    public IEnumerable<ProviderPlanService> ProviderPlanServices { get; private set; } = Enumerable.Empty<ProviderPlanService>().ToList();

    public static ProviderPlan Create(Provider provider, string name, string code)
        => new(provider, name, code);
}
