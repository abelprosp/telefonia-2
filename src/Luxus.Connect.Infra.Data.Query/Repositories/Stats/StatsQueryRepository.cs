using Luxus.Connect.Contracts.Stats.Responses;
using Luxus.Connect.Domain.BillingCycles.Aggregates;
using Luxus.Connect.Domain.Customers.Aggregates;
using Luxus.Connect.Domain.PhoneLines.Aggregates;
using Luxus.Connect.Domain.Providers.Aggregates;
using Luxus.Connect.Infra.Crosscutting;
using Microsoft.EntityFrameworkCore;

namespace Luxus.Connect.Infra.Data.Query.Repositories.Stats;

internal sealed class StatsQueryRepository(AppDbContext context, AppState appState)
    : AppQueryRepository(context), IStatsQueryRepository
{
    public async Task<DashboardStatsResponse> GetDashboardStats(CancellationToken cancellationToken)
    {
        string organizationId = appState.Organization!.Id;

        int customersCount = await context
            .Set<Customer>()
            .AsNoTracking()
            .CountAsync(p => p.OrganizationId == organizationId, cancellationToken);

        int providersCount = await context
            .Set<Provider>()
            .AsNoTracking()
            .CountAsync(p => p.OrganizationId == organizationId, cancellationToken);

        int phoneLinesCount = await context
            .Set<PhoneLine>()
            .AsNoTracking()
            .CountAsync(p => p.ProviderPlan.Provider.OrganizationId == organizationId, cancellationToken);

        int invoicesCount = await context
            .Set<ProviderInvoice>()
            .AsNoTracking()
            .CountAsync(p => p.ContractingCompany.Provider.OrganizationId == organizationId, cancellationToken);

        int billingCyclesCount = await context
            .Set<BillingCycle>()
            .AsNoTracking()
            .CountAsync(p => p.OrganizationId == organizationId, cancellationToken);

        return new DashboardStatsResponse(
            billingCyclesCount,
            customersCount,
            providersCount,
            invoicesCount,
            phoneLinesCount);
    }
}