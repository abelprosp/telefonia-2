using Goal.Domain.Aggregates;
using Luxus.Connect.Domain.Providers.Aggregates;

namespace Luxus.Connect.Domain.PhoneLines.Aggregates;

public class PhoneLineService : Entity
{
    protected PhoneLineService()
        : base()
    {
    }

    private PhoneLineService(PhoneLine phoneLine, ProviderPlanService providerPlanService)
        : this()
    {
        PhoneLine = phoneLine;
        PhoneLineId = phoneLine.Id;

        ProviderPlanService = providerPlanService;
        ProviderPlanServiceId = providerPlanService.Id;

        Active = true;
        Recurring = false;
    }

    public string PhoneLineId { get; private set; } = default!;
    public string ProviderPlanServiceId { get; private set; } = default!;
    public string Name { get; private set; } = default!;
    public string Code { get; private set; } = default!;
    public bool Recurring { get; private set; }
    public decimal? Price { get; private set; }
    public bool Active { get; private set; } = true;
    public PhoneLine PhoneLine { get; private set; } = default!;
    public ProviderPlanService ProviderPlanService { get; private set; } = default!;

    public void ConfigureSubscription(
        bool recurring,
        decimal? price,
        bool active)
    {
        Recurring = recurring;
        Price = price;
        Active = active;
    }

    public void Inactivate()
        => Active = false;

    public static PhoneLineService Create(PhoneLine phoneLine, ProviderPlanService providerPlanService)
    {
        ArgumentNullException.ThrowIfNull(phoneLine);
        ArgumentNullException.ThrowIfNull(providerPlanService);

        return new(phoneLine, providerPlanService);
    }
}
