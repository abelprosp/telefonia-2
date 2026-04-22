using Luxus.Connect.Domain.BillingCycles.Aggregates;
using Luxus.Connect.Domain.Controllership.Aggregates;
using Luxus.Connect.Domain.Customers.Aggregates;
using Luxus.Connect.Domain.PhoneLines.Aggregates;
using Luxus.Connect.Domain.ProcessingMonths.Aggregates;
using Luxus.Connect.Domain.Providers.Aggregates;
using Luxus.Connect.Domain.Providers.Enums;
using Luxus.Connect.Infra.Data.Repositories;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;

namespace Luxus.Connect.Infra.Data.Extensions.DependencyInjection;

public static class ServiceCollectionExtensions
{
    public static IServiceCollection AddAppData(this IServiceCollection services, Action<AppDataOptions> action)
    {
        var options = new AppDataOptions();

        action?.Invoke(options);

        services.AddAppDbContext(options.ConnectionString);

        services.AddRepositories();
        services.AddUnitOfWork();

        return services;
    }

    public static IServiceCollection AddUnitOfWork(this IServiceCollection services)
    {
        services.AddScoped<IAppUnitOfWork, AppUnitOfWork>();

        return services;
    }

    private static IServiceCollection AddRepositories(this IServiceCollection services)
    {
        services.AddScoped<IBillingCycleRepository, BillingCycleRepository>();
        services.AddScoped<IProcessingMonthRepository, ProcessingMonthRepository>();
        services.AddScoped<IContractingCompanyRepository, ContractingCompanyRepository>();
        services.AddScoped<ICostCenterRepository, CostCenterRepository>();
        services.AddScoped<ICustomerRepository, CustomerRepository>();
        services.AddScoped<ICustomerAttachmentRepository, CustomerAttachmentRepository>();
        services.AddScoped<ICustomerProcessingMonthManualReleaseRepository, CustomerProcessingMonthManualReleaseRepository>();
        services.AddScoped<IPhoneLineRepository, PhoneLineRepository>();
        services.AddScoped<IPhoneLineServiceRepository, PhoneLineServiceRepository>();
        services.AddScoped<IProviderAccountRepository, ProviderAccountRepository>();
        services.AddScoped<IProviderInvoiceImportRequestRepository, ProviderInvoiceImportRequestRepository>();
        services.AddScoped<IProviderInvoiceRepository, ProviderInvoiceRepository>();
        services.AddScoped<IProviderPlanRepository, ProviderPlanRepository>();
        services.AddScoped<IProviderPlanServiceRepository, ProviderPlanServiceRepository>();
        services.AddScoped<IProviderRepository, ProviderRepository>();

        return services;
    }

    private static IServiceCollection AddAppDbContext(this IServiceCollection services, string connectionString)
    {
        services.AddScoped<AppDbChangesInterceptor>();

        services.AddDbContext<AppDbContext>((provider, options) =>
        {
            options
                .UseNpgsql(connectionString, x =>
                {
                    x.MigrationsAssembly(typeof(AppDbContext).Assembly.GetName().Name);
                    x.MapEnum<CustomerType>();
                    x.MapEnum<CustomerDocumentType>();
                    x.MapEnum<ServiceType>();
                    x.MapEnum<ServiceApplicationType>();
                    x.MapEnum<ServiceAvailabilityRule>();
                    x.MapEnum<ExceedanceChargeType>();
                    x.MapEnum<BillingCycleStatus>();
                    x.MapEnum<ProcessingMonthStatus>();
                    x.MapEnum<LineClassification>();
                    x.MapEnum<PhoneLineStatus>();
                    x.MapEnum<TransitionSubStatus>();
                    x.MapEnum<ProviderInvoiceStatus>();
                    x.MapEnum<ProviderInvoiceItemType>();
                    x.MapEnum<InvoiceItemUnit>();
                })
                .AddInterceptors(provider.GetRequiredService<AppDbChangesInterceptor>())
                .EnableSensitiveDataLogging();
        });

        return services;
    }
}
