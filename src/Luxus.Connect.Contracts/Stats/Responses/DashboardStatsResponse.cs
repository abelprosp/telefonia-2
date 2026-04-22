namespace Luxus.Connect.Contracts.Stats.Responses;

public sealed record DashboardStatsResponse(
        int BillingCyclesCount,
        int CustomersCount,
        int ProvidersCount,
        int ProviderInvoicesCount,
        int PhoneLinesCount)
{
}
