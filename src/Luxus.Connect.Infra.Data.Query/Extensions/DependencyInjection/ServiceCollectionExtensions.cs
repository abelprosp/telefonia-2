using Luxus.Connect.Infra.Data.Query.Repositories.BillingCycles;
using Luxus.Connect.Infra.Data.Query.Repositories.Controllership;
using Luxus.Connect.Infra.Data.Query.Repositories.Customers;
using Luxus.Connect.Infra.Data.Query.Repositories.PhoneLines;
using Luxus.Connect.Infra.Data.Query.Repositories.ProcessingMonths;
using Luxus.Connect.Infra.Data.Query.Repositories.Providers;
using Luxus.Connect.Infra.Data.Query.Repositories.Stats;
using Microsoft.Extensions.DependencyInjection;

namespace Luxus.Connect.Infra.Data.Query.Extensions.DependencyInjection;

public static class ServiceCollectionExtensions
{
    public static IServiceCollection AddQueryData(this IServiceCollection services)
    {
        services.AddRepositories();
        return services;
    }

    private static IServiceCollection AddRepositories(this IServiceCollection services)
    {
        services.AddScoped<IProviderQueryRepository, ProviderQueryRepository>();
        services.AddScoped<IPhoneLineQueryRepository, PhoneLineQueryRepository>();
        services.AddScoped<IBillingCycleQueryRepository, BillingCycleQueryRepository>();
        services.AddScoped<ICostCenterQueryRepository, CostCenterQueryRepository>();
        services.AddScoped<IProviderInvoiceQueryRepository, ProviderInvoiceQueryRepository>();
        services.AddScoped<IProcessingMonthQueryRepository, ProcessingMonthQueryRepository>();
        services.AddScoped<ICustomerQueryRepository, CustomerQueryRepository>();
        services.AddScoped<ICustomerProcessingMonthBillingReadinessQueryRepository, CustomerProcessingMonthBillingReadinessQueryRepository>();
        services.AddScoped<IStatsQueryRepository, StatsQueryRepository>();

        return services;
    }
}
